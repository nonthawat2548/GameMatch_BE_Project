package handler

import (
	"github.com/jackc/pgx/v5"
	"context"
	"encoding/json"
	"errors"
	"gamematch/mw"
	"gamematch/store"
	"github.com/jackc/pgx/v5/pgconn"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type API struct {
	S               *store.Store
	Secret, RawgKey string
	SteamKey        string
	SyncMu          sync.Mutex
	RawgClient      *http.Client
	SteamClient     *http.Client
}
type user = store.Member

func send(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
func fail(w http.ResponseWriter, status int, msg string) {
	send(w, status, map[string]string{"error": msg})
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if d.Decode(v) != nil {
		fail(w, 400, "ข้อมูล JSON ไม่ถูกต้อง")
		return false
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		fail(w, 400, "JSON ต้องมีเพียง object เดียว")
		return false
	}
	return true
}
func dbError(w http.ResponseWriter, e error) {
	if errors.Is(e, pgx.ErrNoRows) {
		fail(w, 404, "ไม่พบข้อมูล")
		return
	}
	var p *pgconn.PgError
	if errors.As(e, &p) && p.Code == "23505" {
		fail(w, 409, "ข้อมูลนี้มีอยู่แล้ว")
		return
	}
	log.Printf("database operation failed: %T", e)
	fail(w, 500, "ระบบฐานข้อมูลขัดข้อง")
}
func (a *API) Routes() *http.ServeMux {
	m := http.NewServeMux()
	m.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if a.S.Ping(ctx) != nil {
			fail(w, 503, "Database unavailable")
			return
		}
		send(w, 200, map[string]string{"status": "ok"})
	})
	authLimit := mw.NewRateLimiter(30, time.Minute)
	m.HandleFunc("POST /api/auth/register", authLimit.Wrap(a.Register))
	m.HandleFunc("POST /api/auth/login", authLimit.Wrap(a.Login))
	m.HandleFunc("POST /api/auth/logout", a.Guard(false, a.Logout))
	m.HandleFunc("GET /api/me", a.Guard(false, func(w http.ResponseWriter, r *http.Request) { send(w, 200, current(r)) }))
	m.HandleFunc("PUT /api/me", a.Guard(false, a.Profile))
	m.HandleFunc("PUT /api/me/password", a.Guard(false, a.Password))
	m.HandleFunc("GET /api/games", a.Games)
	m.HandleFunc("GET /api/games/{id}", a.Detail)
	m.HandleFunc("GET /api/games/{id}/trailers", a.GameTrailers)
	m.HandleFunc("GET /api/games/{id}/metadata", a.GameMetadata)
	m.HandleFunc("GET /api/games/{id}/steam-rating", a.SteamRating)
	m.HandleFunc("GET /api/games/{id}/reviews", a.Reviews)
	m.HandleFunc("POST /api/games/{id}/reviews", a.Guard(false, a.Reviews))
	m.HandleFunc("PUT /api/games/{id}/reviews/{reviewId}", a.Guard(false, a.Reviews))
	m.HandleFunc("DELETE /api/games/{id}/reviews/{reviewId}", a.Guard(false, a.Reviews))
	m.HandleFunc("GET /api/me/review-votes/{id}", a.Guard(false,a.ReviewVoteIDs))
	m.HandleFunc("PUT /api/games/{id}/reviews/{reviewId}/helpful", a.Guard(false,a.ReviewVote))
	m.HandleFunc("DELETE /api/games/{id}/reviews/{reviewId}/helpful", a.Guard(false,a.ReviewVote))
	m.HandleFunc("POST /api/recommendations", a.Recommend)
	m.HandleFunc("GET /api/games/{id}/similar", a.SimilarGames)
	m.HandleFunc("GET /api/me/recommendations", a.Guard(false, a.Favorites))
	m.HandleFunc("GET /api/me/favorites", a.Guard(false, a.Favorites))
	m.HandleFunc("PUT /api/me/favorites/{id}", a.Guard(false, a.Favorite))
	m.HandleFunc("DELETE /api/me/favorites/{id}", a.Guard(false, a.Favorite))
	m.HandleFunc("GET /api/me/preferences", a.Guard(false, a.Preferences))
	m.HandleFunc("PUT /api/me/preferences", a.Guard(false, a.Preferences))
	m.HandleFunc("GET /api/admin/stats", a.Guard(true, a.AdminStats))
	m.HandleFunc("GET /api/admin/members", a.Guard(true, a.AdminMembers))
	m.HandleFunc("PUT /api/admin/members/{id}/role", a.Guard(true, a.AdminSetRole))
	m.HandleFunc("GET /api/admin/games", a.Guard(true, a.AdminGames))
	m.HandleFunc("POST /api/admin/games", a.Guard(true, a.SaveGame))
	m.HandleFunc("PUT /api/admin/games/{id}", a.Guard(true, a.SaveGame))
	m.HandleFunc("DELETE /api/admin/games/{id}", a.Guard(true, a.DeleteGame))
	m.HandleFunc("GET /api/admin/sync/status", a.Guard(true, a.SyncStatus))
	m.HandleFunc("POST /api/admin/sync", a.Guard(true, a.Sync))
	m.HandleFunc("POST /api/admin/sync/steam", a.Guard(true, a.SyncSteam))
	return m
}
