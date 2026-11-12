package handler

import (
	"github.com/jackc/pgx/v5"
	"context"
	"errors"
	"gamematch/auth"
	"net/http"
	"strings"
)

func (a *API) RequireLogin(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			fail(w, 401, "กรุณาเข้าสู่ระบบ")
			return
		}
		c, e := auth.Parse(a.Secret, strings.TrimPrefix(header, "Bearer "))
		if e != nil {
			fail(w, 401, "Session หมดอายุ")
			return
		}
		var u user
		var version int
		e = a.S.DB.QueryRow(r.Context(), `SELECT id,email,username,display_name,role,token_version FROM gm_members WHERE id=$1`, c.ID).Scan(&u.ID, &u.Email, &u.Username, &u.Name, &u.Role, &version)
		if errors.Is(e, pgx.ErrNoRows) {
			fail(w, 401, "ไม่พบบัญชี")
			return
		}
		if e != nil {
			dbError(w, e)
			return
		}
		if version != c.Version {
			fail(w, 401, "กรุณาเข้าสู่ระบบใหม่")
			return
		}
		h(w, r.WithContext(context.WithValue(r.Context(), userKey, u)))
	}
}
func current(r *http.Request) user { return r.Context().Value(userKey).(user) }

func RequireRole(role string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if current(r).Role != role {
			fail(w, 403, "สำหรับผู้ดูแลระบบเท่านั้น")
			return
		}
		next(w, r)
	}
}

// Preserve Guard for existing callers; compose the two checks like MovieHub.
func (a *API) Guard(admin bool, h http.HandlerFunc) http.HandlerFunc {
	if admin {
		return a.RequireLogin(RequireRole("admin", h))
	}
	return a.RequireLogin(h)
}
