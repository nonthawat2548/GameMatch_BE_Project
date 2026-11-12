package handler

import (
	"gamematch/mw"
	"net/http"
)

func current(r *http.Request) user {
	u, _ := mw.CurrentMember(r.Context())
	return u
}

// Compose the two middlewares just like MovieHub. Keep Guard for existing routes/tests.
func (a *API) Guard(admin bool, next http.HandlerFunc) http.HandlerFunc {
	if admin {
		return mw.RequireLogin(a.Secret, a.S, mw.RequireRole("admin", next))
	}
	return mw.RequireLogin(a.Secret, a.S, next)
}
