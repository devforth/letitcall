package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/letitcall/letitcall/api/internal/model"
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	ErrNotFound = errors.New("not found")
	ErrExists   = errors.New("already exists")
)

type Store struct {
	users       *leveldb.DB
	bookings    *leveldb.DB
	sessions    *leveldb.DB
	oauthStates *leveldb.DB
	mu          sync.Mutex
}

func Open(root string) (*Store, error) {
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, fmt.Errorf("create LevelDB root: %w", err)
	}

	opened := make([]*leveldb.DB, 0, 4)
	openTable := func(name string) (*leveldb.DB, error) {
		db, err := leveldb.OpenFile(filepath.Join(root, name+".leveldb"), nil)
		if err != nil {
			return nil, fmt.Errorf("open %s table: %w", name, err)
		}
		opened = append(opened, db)
		return db, nil
	}

	users, err := openTable("users")
	if err != nil {
		return nil, err
	}
	bookings, err := openTable("bookings")
	if err != nil {
		closeAll(opened)
		return nil, err
	}
	sessions, err := openTable("sessions")
	if err != nil {
		closeAll(opened)
		return nil, err
	}
	oauthStates, err := openTable("oauth_states")
	if err != nil {
		closeAll(opened)
		return nil, err
	}

	return &Store{users: users, bookings: bookings, sessions: sessions, oauthStates: oauthStates}, nil
}

func (s *Store) Close() error {
	return errors.Join(s.users.Close(), s.bookings.Close(), s.sessions.Close(), s.oauthStates.Close())
}

func closeAll(databases []*leveldb.DB) {
	for _, database := range databases {
		_ = database.Close()
	}
}

func (s *Store) UserCount() (int, error) {
	iterator := s.users.NewIterator(nil, nil)
	defer iterator.Release()
	count := 0
	for iterator.Next() {
		count++
	}
	return count, iterator.Error()
}

func (s *Store) CreateUser(user model.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := []byte(normalizeEmail(user.Email))
	exists, err := s.users.Has(key, nil)
	if err != nil {
		return err
	}
	if exists {
		return ErrExists
	}
	return putJSON(s.users, key, user)
}

func (s *Store) PutUser(user model.User) error {
	return putJSON(s.users, []byte(normalizeEmail(user.Email)), user)
}

func (s *Store) GetUser(email string) (model.User, error) {
	var user model.User
	if err := getJSON(s.users, []byte(normalizeEmail(email)), &user); err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (s *Store) ListUsers() ([]model.User, error) {
	iterator := s.users.NewIterator(nil, nil)
	defer iterator.Release()
	users := make([]model.User, 0)
	for iterator.Next() {
		var user model.User
		if err := json.Unmarshal(iterator.Value(), &user); err != nil {
			return nil, fmt.Errorf("decode user: %w", err)
		}
		users = append(users, user)
	}
	if err := iterator.Error(); err != nil {
		return nil, err
	}
	sort.Slice(users, func(i, j int) bool { return users[i].Email < users[j].Email })
	return users, nil
}

func (s *Store) DeleteUser(email string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := []byte(normalizeEmail(email))
	exists, err := s.users.Has(key, nil)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}
	return s.users.Delete(key, nil)
}

func (s *Store) DeleteUserIfMoreThanOne(email string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := []byte(normalizeEmail(email))
	exists, err := s.users.Has(key, nil)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}
	iterator := s.users.NewIterator(nil, nil)
	count := 0
	for iterator.Next() {
		count++
	}
	err = iterator.Error()
	iterator.Release()
	if err != nil {
		return err
	}
	if count <= 1 {
		return errors.New("cannot delete the last user")
	}
	return s.users.Delete(key, nil)
}

func (s *Store) PutSession(token string, session model.Session) error {
	return putJSON(s.sessions, []byte(token), session)
}

func (s *Store) GetSession(token string, now time.Time) (model.Session, error) {
	var session model.Session
	if err := getJSON(s.sessions, []byte(token), &session); err != nil {
		return model.Session{}, err
	}
	if !session.ExpiresAt.After(now) {
		_ = s.sessions.Delete([]byte(token), nil)
		return model.Session{}, ErrNotFound
	}
	return session, nil
}

func (s *Store) DeleteSession(token string) error {
	return s.sessions.Delete([]byte(token), nil)
}

func (s *Store) DeleteSessionsForUser(email string) error {
	normalized := normalizeEmail(email)
	iterator := s.sessions.NewIterator(nil, nil)
	defer iterator.Release()
	batch := new(leveldb.Batch)
	for iterator.Next() {
		var session model.Session
		if err := json.Unmarshal(iterator.Value(), &session); err != nil {
			return err
		}
		if normalizeEmail(session.Email) == normalized {
			batch.Delete(append([]byte(nil), iterator.Key()...))
		}
	}
	if err := iterator.Error(); err != nil {
		return err
	}
	return s.sessions.Write(batch, nil)
}

func (s *Store) PutOAuthState(state string, value model.OAuthState) error {
	return putJSON(s.oauthStates, []byte(state), value)
}

func (s *Store) ConsumeOAuthState(state string, now time.Time) (model.OAuthState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := []byte(state)
	var value model.OAuthState
	if err := getJSON(s.oauthStates, key, &value); err != nil {
		return model.OAuthState{}, err
	}
	if err := s.oauthStates.Delete(key, nil); err != nil {
		return model.OAuthState{}, err
	}
	if !value.ExpiresAt.After(now) {
		return model.OAuthState{}, ErrNotFound
	}
	return value, nil
}

func (s *Store) CreateBooking(key string, booking model.Booking) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	encodedKey := []byte(key)
	exists, err := s.bookings.Has(encodedKey, nil)
	if err != nil {
		return err
	}
	if exists {
		return ErrExists
	}
	return putJSON(s.bookings, encodedKey, booking)
}

func (s *Store) GetBooking(key string) (model.Booking, error) {
	var booking model.Booking
	if err := getJSON(s.bookings, []byte(key), &booking); err != nil {
		return model.Booking{}, err
	}
	return booking, nil
}

func (s *Store) ListBookings() ([]model.Booking, error) {
	iterator := s.bookings.NewIterator(nil, nil)
	defer iterator.Release()
	bookings := make([]model.Booking, 0)
	for iterator.Next() {
		var booking model.Booking
		if err := json.Unmarshal(iterator.Value(), &booking); err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}
	return bookings, iterator.Error()
}

func (s *Store) DeleteBooking(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	encodedKey := []byte(key)
	exists, err := s.bookings.Has(encodedKey, nil)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}
	return s.bookings.Delete(encodedKey, nil)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func putJSON(database *leveldb.DB, key []byte, value any) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return database.Put(key, encoded, nil)
}

func getJSON(database *leveldb.DB, key []byte, destination any) error {
	encoded, err := database.Get(key, nil)
	if errors.Is(err, leveldb.ErrNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(encoded, destination)
}
