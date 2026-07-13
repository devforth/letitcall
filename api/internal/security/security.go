package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/letitcall/letitcall/api/internal/model"
	"golang.org/x/crypto/bcrypt"
)

const MinPasswordLength = 12

func NormalizeEmail(value string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(value))
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return "", errors.New("email must be a valid address")
	}
	return email, nil
}

func ValidateTimezone(value string) (string, error) {
	timezone := strings.TrimSpace(value)
	if timezone == "" {
		return "", errors.New("timezone is required")
	}
	if _, err := time.LoadLocation(timezone); err != nil {
		return "", errors.New("timezone must be a valid IANA timezone")
	}
	return timezone, nil
}

func HashPassword(password string) (string, error) {
	if len(password) < MinPasswordLength {
		return "", fmt.Errorf("password must be at least %d characters", MinPasswordLength)
	}
	if len([]byte(password)) > 72 {
		return "", errors.New("password must be at most 72 bytes")
	}
	return hashPassword(password)
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func NewUser(email, password, timezone string, now time.Time) (model.User, error) {
	normalizedEmail, err := NormalizeEmail(email)
	if err != nil {
		return model.User{}, err
	}
	normalizedTimezone := "UTC"
	if timezone != "" {
		normalizedTimezone, err = ValidateTimezone(timezone)
		if err != nil {
			return model.User{}, err
		}
	}
	passwordHash := ""
	if password != "" {
		passwordHash, err = HashPassword(password)
		if err != nil {
			return model.User{}, err
		}
	}
	return newUser(normalizedEmail, passwordHash, normalizedTimezone, now), nil
}

func NewFirstUser(identifier, password string, now time.Time) (model.User, error) {
	passwordHash, err := hashPassword(password)
	if err != nil {
		return model.User{}, err
	}
	return newUser(strings.ToLower(strings.TrimSpace(identifier)), passwordHash, "UTC", now), nil
}

func newUser(email, passwordHash, timezone string, now time.Time) model.User {
	now = now.UTC().Truncate(time.Second)
	return model.User{
		Email:        email,
		PasswordHash: passwordHash,
		Timezone:     timezone,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func RandomToken(bytes int) (string, error) {
	buffer := make([]byte, bytes)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("generate secure token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func TokenDigest(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}

type TokenCipher struct {
	aead cipher.AEAD
}

func NewTokenCipher(key []byte) (*TokenCipher, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create token cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create token cipher GCM: %w", err)
	}
	return &TokenCipher{aead: aead}, nil
}

func (c *TokenCipher) Encrypt(plaintext []byte) (string, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("generate token nonce: %w", err)
	}
	ciphertext := c.aead.Seal(nonce, nonce, plaintext, nil)
	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

func (c *TokenCipher) Decrypt(value string) ([]byte, error) {
	ciphertext, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("decode encrypted token: %w", err)
	}
	if len(ciphertext) < c.aead.NonceSize() {
		return nil, errors.New("encrypted token is too short")
	}
	nonce, payload := ciphertext[:c.aead.NonceSize()], ciphertext[c.aead.NonceSize():]
	plaintext, err := c.aead.Open(nil, nonce, payload, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt token: %w", err)
	}
	return plaintext, nil
}
