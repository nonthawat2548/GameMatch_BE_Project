package mw

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"gamematch/auth"
	"gamematch/store"
	"github.com/jackc/pgx/v5"
)

type authKey int
const memberKey authKey = 1

func CurrentMember(ctx context.Context) (store.Member, bool) {
	member, ok := ctx.Value(memberKey).(store.Member)
	return member, ok
}

func authError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// RequireLogin validates the JWT, then checks the current DB account and version.
func RequireLogin(secret string, s *store.Store, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			authError(w, 401, "กรุณาเข้าสู่ระบบ")
			return
		}
		claims, err := auth.Parse(secret, strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			authError(w, 401, "Session หมดอายุ")
			return
		}
		member, version, err := s.MemberByID(r.Context(), claims.ID)
		if errors.Is(err, pgx.ErrNoRows) {
			authError(w, 401, "ไม่พบบัญชี")
			return
		}
		if err != nil {
			log.Printf("member lookup failed: %T", err)
			authError(w, 500, "ระบบฐานข้อมูลขัดข้อง")
			return
		}
		if version != claims.Version {
			authError(w, 401, "กรุณาเข้าสู่ระบบใหม่")
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), memberKey, member)))
	}
}

func RequireRole(role string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		member, ok := CurrentMember(r.Context())
		if !ok {
			authError(w, 401, "กรุณาเข้าสู่ระบบ")
			return
		}
		if member.Role != role {
			authError(w, 403, "สำหรับผู้ดูแลระบบเท่านั้น")
			return
		}
		next(w, r)
	}
}
