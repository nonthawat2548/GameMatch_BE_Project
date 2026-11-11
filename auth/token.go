package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type Claims struct {
	ID      int64 `json:"uid"`
	Version int   `json:"ver"`
	jwt.RegisteredClaims
}

func Issue(secret string, id int64, version int) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{ID: id, Version: version, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)), IssuedAt: jwt.NewNumericDate(time.Now()), Issuer: "gamematch"}}).SignedString([]byte(secret))
}
func Parse(secret, token string) (*Claims, error) {
	c := &Claims{}
	t, e := jwt.ParseWithClaims(token, c, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("invalid method")
		}
		return []byte(secret), nil
	}, jwt.WithExpirationRequired(), jwt.WithIssuer("gamematch"))
	if e != nil || !t.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return c, nil
}
