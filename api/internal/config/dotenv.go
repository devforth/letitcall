package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

func LoadDotEnv() error {
	if err := godotenv.Load(".env.local"); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := godotenv.Overload(".env"); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
