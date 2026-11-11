package tests

import (
	"gamematch/auth"
	"github.com/golang-jwt/jwt/v5"
	"testing"
	"time"
)

const secret = "test-secret-at-least-32-characters-long"

func TestTokenRoundTrip(t *testing.T) {
	token, e := auth.Issue(secret, 42, 3)
	if e != nil {
		t.Fatal(e)
	}
	claims, e := auth.Parse(secret, token)
	if e != nil || claims.ID != 42 || claims.Version != 3 {
		t.Fatalf("claims=%#v error=%v", claims, e)
	}
}
func TestWrongKeyAndExpired(t *testing.T) {
	token, _ := auth.Issue(secret, 1, 0)
	if _, e := auth.Parse("different-secret", token); e == nil {
		t.Fatal("accepted wrong signature")
	}
	expired, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, auth.Claims{ID: 1, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)), Issuer: "gamematch"}}).SignedString([]byte(secret))
	if _, e := auth.Parse(secret, expired); e == nil {
		t.Fatal("accepted expired token")
	}
}
func TestRejectUnexpectedSigningMethod(t *testing.T) {
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS384, jwt.MapClaims{"uid": 1, "exp": time.Now().Add(time.Hour).Unix(), "iss": "gamematch"}).SignedString([]byte(secret))
	if _, e := auth.Parse(secret, token); e == nil {
		t.Fatal("accepted HS384")
	}
}
