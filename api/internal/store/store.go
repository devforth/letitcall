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
	ErrNotFound      = errors.New("not found")
	ErrExists        = errors.New("already exists")
	ErrCapacity      = errors.New("capacity reached")
	ErrLastRecipient = errors.New("user is the final event type recipient")
)

type Store struct {
	users       *leveldb.DB
	eventTypes  *leveldb.DB
	bookings    *leveldb.DB
	sessions    *leveldb.DB
	oauthStates *leveldb.DB
	mu          sync.Mutex
}

func Open(root string) (*Store, error) {
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, fmt.Errorf("create LevelDB root: %w", err)
	}

	opened := make([]*leveldb.DB, 0, 5)
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
	eventTypes, err := openTable("event_types")
	if err != nil {
		closeAll(opened)
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

	return &Store{users: users, eventTypes: eventTypes, bookings: bookings, sessions: sessions, oauthStates: oauthStates}, nil
}

func (s *Store) Close() error {
	return errors.Join(s.users.Close(), s.eventTypes.Close(), s.bookings.Close(), s.sessions.Close(), s.oauthStates.Close())
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

func (s *Store) CreateEventType(eventType model.EventType) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := []byte(eventType.EventSlug)
	exists, err := s.eventTypes.Has(key, nil)
	if err != nil {
		return err
	}
	if exists {
		return ErrExists
	}
	return putJSON(s.eventTypes, key, eventType)
}

func (s *Store) PutEventType(eventType model.EventType) error {
	return putJSON(s.eventTypes, []byte(eventType.EventSlug), eventType)
}

func (s *Store) GetEventType(slug string) (model.EventType, error) {
	var eventType model.EventType
	if err := getJSON(s.eventTypes, []byte(slug), &eventType); err != nil {
		return model.EventType{}, err
	}
	return eventType, nil
}

func (s *Store) ListEventTypes() ([]model.EventType, error) {
	iterator := s.eventTypes.NewIterator(nil, nil)
	defer iterator.Release()
	eventTypes := make([]model.EventType, 0)
	for iterator.Next() {
		var eventType model.EventType
		if err := json.Unmarshal(iterator.Value(), &eventType); err != nil {
			return nil, err
		}
		eventTypes = append(eventTypes, eventType)
	}
	if err := iterator.Error(); err != nil {
		return nil, err
	}
	sort.Slice(eventTypes, func(i, j int) bool { return eventTypes[i].Name < eventTypes[j].Name })
	return eventTypes, nil
}

func (s *Store) RemoveEventTypeRecipient(email string, updatedAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	normalized := normalizeEmail(email)
	iterator := s.eventTypes.NewIterator(nil, nil)
	defer iterator.Release()
	batch := new(leveldb.Batch)
	for iterator.Next() {
		var eventType model.EventType
		if err := json.Unmarshal(iterator.Value(), &eventType); err != nil {
			return err
		}
		index := -1
		for candidateIndex, candidate := range eventType.RecipientEmails {
			if normalizeEmail(candidate) == normalized {
				index = candidateIndex
				break
			}
		}
		if index < 0 {
			continue
		}
		if len(eventType.RecipientEmails) == 1 {
			return fmt.Errorf("%w: %s", ErrLastRecipient, eventType.EventSlug)
		}
		eventType.RecipientEmails = append(eventType.RecipientEmails[:index], eventType.RecipientEmails[index+1:]...)
		eventType.UpdatedAt = updatedAt
		encoded, err := json.Marshal(eventType)
		if err != nil {
			return err
		}
		batch.Put(append([]byte(nil), iterator.Key()...), encoded)
	}
	if err := iterator.Error(); err != nil {
		return err
	}
	return s.eventTypes.Write(batch, nil)
}

func (s *Store) CreateBooking(slotKey string, booking model.Booking, inviteeLimit *int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := []byte(slotKey)
	bookings := make([]model.Booking, 0)
	encoded, err := s.bookings.Get(key, nil)
	if err == nil {
		if err := json.Unmarshal(encoded, &bookings); err != nil {
			return err
		}
	} else if !errors.Is(err, leveldb.ErrNotFound) {
		return err
	}
	for _, existing := range bookings {
		if normalizeEmail(existing.AttendeeEmail) == normalizeEmail(booking.AttendeeEmail) {
			return ErrExists
		}
	}
	if inviteeLimit != nil && len(bookings) >= *inviteeLimit {
		return ErrCapacity
	}
	return putJSON(s.bookings, key, append(bookings, booking))
}

func (s *Store) GetBooking(id string) (model.Booking, error) {
	iterator := s.bookings.NewIterator(nil, nil)
	defer iterator.Release()
	for iterator.Next() {
		var bookings []model.Booking
		if err := json.Unmarshal(iterator.Value(), &bookings); err != nil {
			return model.Booking{}, err
		}
		for _, booking := range bookings {
			if booking.ID == id {
				return booking, nil
			}
		}
	}
	if err := iterator.Error(); err != nil {
		return model.Booking{}, err
	}
	return model.Booking{}, ErrNotFound
}

func (s *Store) ListBookings() ([]model.Booking, error) {
	iterator := s.bookings.NewIterator(nil, nil)
	defer iterator.Release()
	bookings := make([]model.Booking, 0)
	for iterator.Next() {
		var slotBookings []model.Booking
		if err := json.Unmarshal(iterator.Value(), &slotBookings); err != nil {
			return nil, err
		}
		bookings = append(bookings, slotBookings...)
	}
	if err := iterator.Error(); err != nil {
		return nil, err
	}
	sort.Slice(bookings, func(i, j int) bool { return bookings[i].Time.Before(bookings[j].Time) })
	return bookings, nil
}

func (s *Store) DeleteBooking(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	iterator := s.bookings.NewIterator(nil, nil)
	defer iterator.Release()
	for iterator.Next() {
		var bookings []model.Booking
		if err := json.Unmarshal(iterator.Value(), &bookings); err != nil {
			return err
		}
		for index, booking := range bookings {
			if booking.ID != id {
				continue
			}
			key := append([]byte(nil), iterator.Key()...)
			bookings = append(bookings[:index], bookings[index+1:]...)
			if len(bookings) == 0 {
				return s.bookings.Delete(key, nil)
			}
			return putJSON(s.bookings, key, bookings)
		}
	}
	if err := iterator.Error(); err != nil {
		return err
	}
	return ErrNotFound
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
