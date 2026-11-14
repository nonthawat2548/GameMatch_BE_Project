package config

import (
	"errors"
	"os"
	"strings"
)

// LoadJWTSecret prefers an explicitly configured environment value over the file.
func LoadJWTSecret(path string) (string, error) {
	secret, configured := os.LookupEnv("JWT_SECRET")
	if !configured {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", errors.New("cannot read .jwt_secret: set JWT_SECRET or create the secret file")
		}
		secret = strings.TrimSpace(string(data))
	}
	if len(secret) < 32 {
		return "", errors.New("JWT_SECRET or .jwt_secret must contain at least 32 characters")
	}
	return secret, nil
}
