package bootstrap

import (
	"errors"
	"fmt"
	"time"

	"github.com/letitcall/letitcall/api/internal/config"
	"github.com/letitcall/letitcall/api/internal/security"
	"github.com/letitcall/letitcall/api/internal/store"
)

func EnsureFirstUser(database *store.Store, credentials config.FirstUser, now time.Time) error {
	count, err := database.UserCount()
	if err != nil {
		return fmt.Errorf("count users: %w", err)
	}
	if count > 0 {
		return nil
	}
	if credentials.Email == "" || credentials.Password == "" {
		return fmt.Errorf(
			"users table is empty; set %s and %s to create the first user",
			config.EnvFirstUserEmail,
			config.EnvFirstUserPassword,
		)
	}
	user, err := security.NewUser(credentials.Email, credentials.Password, "UTC", now)
	if err != nil {
		return fmt.Errorf("validate first user: %w", err)
	}
	if err := database.CreateUser(user); err != nil && !errors.Is(err, store.ErrExists) {
		return fmt.Errorf("create first user: %w", err)
	}
	return nil
}
