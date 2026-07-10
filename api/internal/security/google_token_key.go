package security

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const GoogleTokenKeyFile = "google-token.key"

func LoadGoogleTokenKey(dataPath string) ([]byte, error) {
	path := filepath.Join(dataPath, GoogleTokenKeyFile)
	key, err := os.ReadFile(path)
	if err == nil {
		if len(key) != 32 {
			return nil, fmt.Errorf("%s must contain exactly 32 bytes", path)
		}
		return key, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read Google token key: %w", err)
	}

	key = make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generate Google token key: %w", err)
	}
	if err := os.WriteFile(path, key, 0o600); err != nil {
		return nil, fmt.Errorf("write Google token key: %w", err)
	}
	return key, nil
}
