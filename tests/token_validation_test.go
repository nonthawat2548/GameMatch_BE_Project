package tests

import (
	"gamematch/auth"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestTokenRejectsInvalidMemberClaims(t *testing.T) {
	for _, pair := range [][2]int{{0, 0}, {-1, 0}, {1, -1}} {
		claims := auth.Claims{ID: int64(pair[0]), Version: pair[1], RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "gamematch", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		}}
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := auth.Parse(secret, token); err == nil {
			t.Fatal("accepted a signed token with invalid member claims")
		}
	}
}
