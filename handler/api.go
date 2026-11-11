package handler

import (
	"context"
	"encoding/json"
	"errors"
	"gamematch/auth"
	"gamematch/mw"
	"gamematch/recommend"
	"gamematch/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type API struct {
	S               *store.Store
	Secret, RawgKey string
	SyncMu          sync.Mutex
	RawgClient      *http.Client
	SteamClient     *http.Client
}
type user struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"displayName"`
	Role  string `json:"role"`
}
type key int

const userKey key = 1

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
func (a *API) Guard(admin bool, h http.HandlerFunc) http.HandlerFunc {
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
		e = a.S.DB.QueryRow(r.Context(), `SELECT id,email,display_name,role,token_version FROM gm_members WHERE id=$1`, c.ID).Scan(&u.ID, &u.Email, &u.Name, &u.Role, &version)
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
		if admin && u.Role != "admin" {
			fail(w, 403, "สำหรับผู้ดูแลระบบเท่านั้น")
			return
		}
		h(w, r.WithContext(context.WithValue(r.Context(), userKey, u)))
	}
}
func current(r *http.Request) user { return r.Context().Value(userKey).(user) }
func (a *API) Routes() *http.ServeMux {
	m := http.NewServeMux()
	m.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if a.S.DB.Ping(ctx) != nil {
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
	m.HandleFunc("POST /api/recommendations", a.Recommend)
	m.HandleFunc("GET /api/me/favorites", a.Guard(false, a.Favorites))
	m.HandleFunc("PUT /api/me/favorites/{id}", a.Guard(false, a.Favorite))
	m.HandleFunc("DELETE /api/me/favorites/{id}", a.Guard(false, a.Favorite))
	m.HandleFunc("GET /api/me/preferences", a.Guard(false, a.Preferences))
	m.HandleFunc("PUT /api/me/preferences", a.Guard(false, a.Preferences))
	m.HandleFunc("GET /api/admin/games", a.Guard(true, a.AdminGames))
	m.HandleFunc("POST /api/admin/games", a.Guard(true, a.SaveGame))
	m.HandleFunc("PUT /api/admin/games/{id}", a.Guard(true, a.SaveGame))
	m.HandleFunc("DELETE /api/admin/games/{id}", a.Guard(true, a.DeleteGame))
	m.HandleFunc("POST /api/admin/sync", a.Guard(true, a.Sync))
	m.HandleFunc("POST /api/admin/sync/steam", a.Guard(true, a.SyncSteam))
	return m
}
func (a *API) Register(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"displayName"`
	}
	if !decode(w, r, &b) {
		return
	}
	b.Email = strings.ToLower(strings.TrimSpace(b.Email))
	b.Name = strings.TrimSpace(b.Name)
	address, e := mail.ParseAddress(b.Email)
	if e != nil || address.Address != b.Email || b.Name == "" || len(b.Name) > 100 || len(b.Password) < 8 || len(b.Password) > 72 {
		fail(w, 400, "ตรวจอีเมล ชื่อ และรหัสผ่าน 8–72 bytes")
		return
	}
	hash, e := bcrypt.GenerateFromPassword([]byte(b.Password), bcrypt.DefaultCost)
	if e != nil {
		fail(w, 500, "สร้างรหัสผ่านไม่สำเร็จ")
		return
	}
	var id int64
	e = a.S.DB.QueryRow(r.Context(), `INSERT INTO gm_members(email,password_hash,display_name) VALUES($1,$2,$3) RETURNING id`, b.Email, string(hash), b.Name).Scan(&id)
	if e != nil {
		dbError(w, e)
		return
	}
	send(w, 201, map[string]any{"id": id})
}
func (a *API) Login(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decode(w, r, &b) {
		return
	}
	var u user
	if len(b.Email) > 320 || len(b.Password) > 72 {
		fail(w, 401, "อีเมลหรือรหัสผ่านไม่ถูกต้อง")
		return
	}
	var hash string
	var version int
	e := a.S.DB.QueryRow(r.Context(), `SELECT id,email,display_name,role,password_hash,token_version FROM gm_members WHERE email=$1`, strings.ToLower(strings.TrimSpace(b.Email))).Scan(&u.ID, &u.Email, &u.Name, &u.Role, &hash, &version)
	if e != nil && !errors.Is(e, pgx.ErrNoRows) {
		dbError(w, e)
		return
	}
	if e != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(b.Password)) != nil {
		fail(w, 401, "อีเมลหรือรหัสผ่านไม่ถูกต้อง")
		return
	}
	token, e := auth.Issue(a.Secret, u.ID, version)
	if e != nil {
		fail(w, 500, "สร้าง session ไม่สำเร็จ")
		return
	}
	send(w, 200, map[string]any{"token": token, "user": u})
}
func (a *API) Logout(w http.ResponseWriter, r *http.Request) {
	_, e := a.S.DB.Exec(r.Context(), `UPDATE gm_members SET token_version=token_version+1 WHERE id=$1`, current(r).ID)
	if e != nil {
		dbError(w, e)
		return
	}
	send(w, 200, map[string]bool{"ok": true})
}
func (a *API) Profile(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Name string `json:"displayName"`
	}
	if !decode(w, r, &b) {
		return
	}
	b.Name = strings.TrimSpace(b.Name)
	if b.Name == "" || len(b.Name) > 100 {
		fail(w, 400, "กรุณาระบุชื่อไม่เกิน 100 bytes")
		return
	}
	_, e := a.S.DB.Exec(r.Context(), `UPDATE gm_members SET display_name=$1 WHERE id=$2`, b.Name, current(r).ID)
	if e != nil {
		dbError(w, e)
		return
	}
	send(w, 200, map[string]bool{"ok": true})
}
func (a *API) Password(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Current string `json:"currentPassword"`
		New     string `json:"newPassword"`
	}
	if !decode(w, r, &b) {
		return
	}
	if len(b.New) < 8 || len(b.New) > 72 {
		fail(w, 400, "รหัสผ่านใหม่ต้องมี 8–72 bytes")
		return
	}
	var old string
	e := a.S.DB.QueryRow(r.Context(), `SELECT password_hash FROM gm_members WHERE id=$1`, current(r).ID).Scan(&old)
	if e != nil {
		dbError(w, e)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(old), []byte(b.Current)) != nil {
		fail(w, 400, "รหัสผ่านปัจจุบันไม่ถูกต้อง")
		return
	}
	hash, e := bcrypt.GenerateFromPassword([]byte(b.New), bcrypt.DefaultCost)
	if e != nil {
		fail(w, 500, "สร้างรหัสผ่านไม่สำเร็จ")
		return
	}
	result, e := a.S.DB.Exec(r.Context(), `UPDATE gm_members SET password_hash=$1,token_version=token_version+1 WHERE id=$2 AND password_hash=$3`, string(hash), current(r).ID, old)
	if e != nil {
		dbError(w, e)
		return
	}
	if result.RowsAffected() != 1 {
		fail(w, 409, "ข้อมูลเปลี่ยนแล้ว กรุณาลองใหม่")
		return
	}
	send(w, 200, map[string]bool{"ok": true})
}
func (a *API) Games(w http.ResponseWriter, r *http.Request) {
	g, e := a.S.Games(r.Context(), false)
	if e != nil {
		dbError(w, e)
		return
	}
	q := strings.ToLower(r.URL.Query().Get("q"))
	out := []store.Game{}
	for _, v := range g {
		if strings.Contains(strings.ToLower(v.Title), q) {
			out = append(out, v)
		}
	}
	send(w, 200, out)
}
func (a *API) Detail(w http.ResponseWriter, r *http.Request) {
	games, e := a.S.Games(r.Context(), false)
	if e != nil {
		dbError(w, e)
		return
	}
	for _, g := range games {
		if strconv.FormatInt(g.ID, 10) == r.PathValue("id") {
			send(w, 200, g)
			return
		}
	}
	fail(w, 404, "ไม่พบเกม")
}
func (a *API) Recommend(w http.ResponseWriter, r *http.Request) {
	var p recommend.Preference
	if !decode(w, r, &p) {
		return
	}
	if !recommend.Valid(p) {
		fail(w, 400, "ตัวเลือกไม่ถูกต้อง")
		return
	}
	g, e := a.S.Games(r.Context(), false)
	if e != nil {
		dbError(w, e)
		return
	}
	send(w, 200, recommend.Rank(g, p))
}
func (a *API) Favorites(w http.ResponseWriter, r *http.Request) {
	rows, e := a.S.DB.Query(r.Context(), `SELECT game_id FROM gm_favorites WHERE member_id=$1`, current(r).ID)
	if e != nil {
		dbError(w, e)
		return
	}
	ids := map[int64]bool{}
	for rows.Next() {
		var id int64
		if e = rows.Scan(&id); e != nil {
			rows.Close()
			dbError(w, e)
			return
		}
		ids[id] = true
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		dbError(w, e)
		return
	}
	games, e := a.S.Games(r.Context(), false)
	if e != nil {
		dbError(w, e)
		return
	}
	out := []store.Game{}
	for _, g := range games {
		if ids[g.ID] {
			out = append(out, g)
		}
	}
	send(w, 200, out)
}
func (a *API) Favorite(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if e != nil || id <= 0 {
		fail(w, 400, "id ไม่ถูกต้อง")
		return
	}
	if r.Method == "DELETE" {
		_, e = a.S.DB.Exec(r.Context(), `DELETE FROM gm_favorites WHERE member_id=$1 AND game_id=$2`, current(r).ID, id)
	} else {
		var exists bool
		e = a.S.DB.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM gm_games WHERE id=$1 AND is_published)`, id).Scan(&exists)
		if e != nil {
			dbError(w, e)
			return
		}
		if !exists {
			fail(w, 404, "ไม่พบเกม")
			return
		}
		_, e = a.S.DB.Exec(r.Context(), `INSERT INTO gm_favorites VALUES($1,$2) ON CONFLICT DO NOTHING`, current(r).ID, id)
	}
	if e != nil {
		dbError(w, e)
		return
	}
	send(w, 200, map[string]bool{"ok": true})
}
func (a *API) Preferences(w http.ResponseWriter, r *http.Request) {
	p := recommend.Preference{Genre: "rpg", Mode: "solo", Style: "story"}
	if r.Method == "GET" {
		e := a.S.DB.QueryRow(r.Context(), `SELECT genre,mode,style FROM gm_preferences WHERE member_id=$1`, current(r).ID).Scan(&p.Genre, &p.Mode, &p.Style)
		if e != nil && !errors.Is(e, pgx.ErrNoRows) {
			dbError(w, e)
			return
		}
		send(w, 200, p)
		return
	}
	if !decode(w, r, &p) {
		return
	}
	if !recommend.Valid(p) {
		fail(w, 400, "ตัวเลือกไม่ถูกต้อง")
		return
	}
	_, e := a.S.DB.Exec(r.Context(), `INSERT INTO gm_preferences VALUES($1,$2,$3,$4) ON CONFLICT(member_id) DO UPDATE SET genre=$2,mode=$3,style=$4`, current(r).ID, p.Genre, p.Mode, p.Style)
	if e != nil {
		dbError(w, e)
		return
	}
	send(w, 200, p)
}
func (a *API) AdminGames(w http.ResponseWriter, r *http.Request) {
	g, e := a.S.Games(r.Context(), true)
	if e != nil {
		dbError(w, e)
		return
	}
	send(w, 200, g)
}
func (a *API) SaveGame(w http.ResponseWriter, r *http.Request) {
	var g store.Game
	if !decode(w, r, &g) {
		return
	}
	if r.Method == "PUT" {
		var e error
		g.ID, e = strconv.ParseInt(r.PathValue("id"), 10, 64)
		if e != nil || g.ID <= 0 {
			fail(w, 400, "id ไม่ถูกต้อง")
			return
		}
	} else {
		g.ID = 0
		g.RawgID = nil
	}
	g.Title = strings.TrimSpace(g.Title)
	if g.Title == "" || len(g.Title) > 250 {
		fail(w, 400, "กรุณาระบุชื่อเกม")
		return
	}
	if g.ImageURL != "" {
		u, e := url.Parse(g.ImageURL)
		if e != nil || u.Scheme != "https" || u.Host == "" {
			fail(w, 400, "รูปต้องเป็น HTTPS URL")
			return
		}
	}
	id, e := a.S.Save(r.Context(), g)
	if e != nil {
		dbError(w, e)
		return
	}
	send(w, 200, map[string]int64{"id": id})
}
func (a *API) DeleteGame(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if e != nil || id <= 0 {
		fail(w, 400, "id ไม่ถูกต้อง")
		return
	}
	result, e := a.S.DB.Exec(r.Context(), `DELETE FROM gm_games WHERE id=$1`, id)
	if e != nil {
		dbError(w, e)
		return
	}
	if result.RowsAffected() == 0 {
		fail(w, 404, "ไม่พบเกม")
		return
	}
	send(w, 200, map[string]bool{"ok": true})
}
func (a *API) Sync(w http.ResponseWriter, r *http.Request) {
	if a.RawgKey == "" {
		fail(w, 503, "ยังไม่ได้ตั้ง RAWG_API_KEY บน Backend")
		return
	}
	if !a.SyncMu.TryLock() {
		fail(w, 409, "กำลัง Sync อยู่")
		return
	}
	defer a.SyncMu.Unlock()
	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()
	u := "https://api.rawg.io/api/games?" + url.Values{"key": {a.RawgKey}, "page_size": {"40"}, "ordering": {"-added"}}.Encode()
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		fail(w, 500, "สร้างคำขอไม่สำเร็จ")
		return
	}
	client := a.RawgClient
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	res, e := client.Do(req)
	if e != nil {
		fail(w, 502, "ติดต่อ RAWG ไม่สำเร็จ")
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		fail(w, 502, "RAWG ปฏิเสธคำขอ ตรวจ Key หรือ Quota")
		return
	}
	var data struct {
		Results []struct {
			ID     int64  `json:"id"`
			Name   string `json:"name"`
			Image  string `json:"background_image"`
			Genres []struct {
				Slug string `json:"slug"`
			} `json:"genres"`
			Tags []struct {
				Slug string `json:"slug"`
			} `json:"tags"`
			Platforms []struct {
				Platform struct {
					Name string `json:"name"`
				} `json:"platform"`
			} `json:"platforms"`
		} `json:"results"`
	}
	body, readErr := io.ReadAll(io.LimitReader(res.Body, (8<<20)+1))
	if readErr != nil || len(body) > 8<<20 || json.Unmarshal(body, &data) != nil || data.Results == nil || len(data.Results) > 40 {
		fail(w, 502, "RAWG JSON ไม่ถูกต้อง")
		return
	}
	tx, e := a.S.DB.Begin(ctx)
	if e != nil {
		dbError(w, e)
		return
	}
	defer tx.Rollback(ctx)
	count := 0
	for _, v := range data.Results {
		if v.ID <= 0 || v.Name == "" {
			continue
		}
		genres, tags, platforms := []string{}, []string{}, []string{}
		for _, g := range v.Genres {
			genres = append(genres, g.Slug)
		}
		for _, t := range v.Tags {
			tags = append(tags, t.Slug)
		}
		for _, p := range v.Platforms {
			platforms = append(platforms, p.Platform.Name)
		}
		_, e = tx.Exec(ctx, `INSERT INTO gm_games(rawg_id,title,image_url,genres,tags,platforms) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(rawg_id) DO UPDATE SET title=EXCLUDED.title,image_url=EXCLUDED.image_url,genres=EXCLUDED.genres,tags=EXCLUDED.tags,platforms=EXCLUDED.platforms`, v.ID, v.Name, v.Image, genres, tags, platforms)
		if e != nil {
			dbError(w, e)
			return
		}
		count++
	}
	if e = tx.Commit(ctx); e != nil {
		dbError(w, e)
		return
	}
	send(w, 200, map[string]any{"imported": count, "note": "Imports first 40 popular results. Existing admin description/publication flags preserved."})
}
