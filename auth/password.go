package auth

import (
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

// Same password rules as MovieHub: minimum characters, maximum UTF-8 bytes.
func ValidPassword(password string) bool {
	return utf8.RuneCountInString(password) >= 8 && len(password) <= 72
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// Perform a bcrypt check even when the account does not exist, like MovieHub.
var DummyHash = func() string {
	hash, err := HashPassword("ไม่มีใครใช้รหัสนี้")
	if err != nil {
		panic(err)
	}
	return hash
}()
