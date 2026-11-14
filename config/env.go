package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads optional local configuration without overriding the process environment.
func LoadEnv(path string) error {
	err := godotenv.Load(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
