package tests

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"gamematch/handler"
	"gamematch/mw"
	"gamematch/store"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// Each test owns a new schema. Never truncate public tables or use MovieHub.
func databaseAPI(t *testing.T) (*handler.API, *httptest.Server) {
	t.Helper()
	dsn := os.Getenv("GM_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set GM_TEST_DATABASE_URL to run real PostgreSQL integration checks")
	}
	if strings.Contains(dsn, "ewfvqtflucuvtiajmmgb") {
		t.Fatal("refusing MovieHub database")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	owner, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal("invalid test database configuration")
	}
	var random [8]byte
	if _, err = rand.Read(random[:]); err != nil {
		t.Fatal(err)
	}
	schema := "gm_test_" + hex.EncodeToString(random[:])
	if !regexp.MustCompile(`^gm_test_[0-9a-f]{16}$`).MatchString(schema) {
		t.Fatal("invalid test schema")
	}
	if _, err = owner.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		owner.Close()
		t.Fatal("cannot create isolated test schema: ", err)
	}
	t.Cleanup(func() {
		cleanup, done := context.WithTimeout(context.Background(), 20*time.Second)
		defer done()
		if _, err := owner.Exec(cleanup, "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Error("test schema cleanup failed: ", err)
		}
		owner.Close()
	})
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal("invalid test pool configuration")
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	config.MaxConns = 3
	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	sql, err := os.ReadFile("../database/schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, string(sql)); err != nil {
		t.Fatal("apply test schema: ", err)
	}
	a := &handler.API{S: &store.Store{DB: db}, Secret: "integration-test-secret-at-least-32-characters"}
	server := httptest.NewServer(mw.CORS("http://localhost:3000", a.Routes()))
	t.Cleanup(server.Close)
	return a, server
}

func requestJSON(t *testing.T, server *httptest.Server, method, path, token string, body any, want int, result any) {
	t.Helper()
	var input io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		input = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, server.URL+path, input)
	if err != nil {
		t.Fatal(err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != want {
		t.Fatalf("%s %s: want %d, got %d: %s", method, path, want, res.StatusCode, data)
	}
	if result != nil && json.Unmarshal(data, result) != nil {
		t.Fatalf("invalid response JSON for %s", path)
	}
}

func registerMember(t *testing.T, server *httptest.Server, email string) int64 {
	t.Helper()
	var created struct {
		ID int64 `json:"id"`
	}
	requestJSON(t, server, "POST", "/api/auth/register", "", map[string]string{"email": email, "password": "Password123!", "displayName": "Demo: Integration member"}, 201, &created)
	if created.ID <= 0 {
		t.Fatal("missing registered member id")
	}
	return created.ID
}

func loginMember(t *testing.T, server *httptest.Server, email, password string) string {
	t.Helper()
	var session struct {
		Token string `json:"token"`
	}
	requestJSON(t, server, "POST", "/api/auth/login", "", map[string]string{"email": email, "password": password}, 200, &session)
	if session.Token == "" {
		t.Fatal("missing login token")
	}
	return session.Token
}

func TestDatabaseMemberOwnershipAndSessionLifecycle(t *testing.T) {
	a, server := databaseAPI(t)
	ctx := context.Background()
	registerMember(t, server, "a@example.test")
	registerMember(t, server, "b@example.test")
	requestJSON(t, server, "POST", "/api/auth/register", "", map[string]string{"email": "A@EXAMPLE.TEST", "password": "Password123!", "displayName": "Duplicate"}, 409, nil)
	tokenA := loginMember(t, server, "a@example.test", "Password123!")
	tokenB := loginMember(t, server, "b@example.test", "Password123!")
	requestJSON(t, server, "GET", "/api/me", "", nil, 401, nil)
	requestJSON(t, server, "GET", "/api/admin/games", tokenA, nil, 403, nil)
	id, err := a.S.Save(ctx, store.Game{Title: "Demo: Ownership fixture", Genres: []string{"rpg"}, Tags: []string{"singleplayer", "story-rich"}, Published: true})
	if err != nil {
		t.Fatal(err)
	}
	favoritePath := "/api/me/favorites/" + strconv.FormatInt(id, 10)
	requestJSON(t, server, "PUT", favoritePath, tokenA, nil, 200, nil)
	requestJSON(t, server, "PUT", favoritePath, tokenA, nil, 200, nil)
	requestJSON(t, server, "DELETE", favoritePath, tokenB, nil, 200, nil)
	var games []store.Game
	requestJSON(t, server, "GET", "/api/me/favorites", tokenB, nil, 200, &games)
	if len(games) != 0 {
		t.Fatal("B can see A's favorite")
	}
	requestJSON(t, server, "GET", "/api/me/favorites", tokenA, nil, 200, &games)
	if len(games) != 1 || games[0].ID != id {
		t.Fatal("B deleted A's favorite or duplicate save was not idempotent")
	}
	pref := map[string]string{"genre": "strategy", "mode": "coop", "style": "challenge"}
	requestJSON(t, server, "PUT", "/api/me/preferences", tokenA, pref, 200, nil)
	// A new server/router reads persisted rows, without any handler-local cache.
	restarted := httptest.NewServer(a.Routes())
	defer restarted.Close()
	var actual map[string]string
	requestJSON(t, restarted, "GET", "/api/me/preferences", tokenA, nil, 200, &actual)
	if actual["genre"] != "strategy" {
		t.Fatal("preference did not persist")
	}
	requestJSON(t, restarted, "GET", "/api/me/preferences", tokenB, nil, 200, &actual)
	if actual["genre"] != "rpg" {
		t.Fatal("preferences leaked to B")
	}
	requestJSON(t, server, "PUT", "/api/me", tokenA, map[string]string{"displayName": "Demo: Updated member"}, 200, nil)
	var member struct{ DisplayName, Role string }
	requestJSON(t, server, "GET", "/api/me", tokenA, nil, 200, &member)
	if member.DisplayName != "Demo: Updated member" || member.Role != "member" {
		t.Fatal("profile/role incorrect")
	}
	requestJSON(t, server, "PUT", "/api/me/password", tokenA, map[string]string{"currentPassword": "wrong", "newPassword": "ChangedPassword123!"}, 400, nil)
	requestJSON(t, server, "GET", "/api/me", tokenA, nil, 200, nil)
	requestJSON(t, server, "PUT", "/api/me/password", tokenA, map[string]string{"currentPassword": "Password123!", "newPassword": "ChangedPassword123!"}, 200, nil)
	requestJSON(t, server, "GET", "/api/me", tokenA, nil, 401, nil)
	requestJSON(t, server, "POST", "/api/auth/login", "", map[string]string{"email": "a@example.test", "password": "Password123!"}, 401, nil)
	newToken := loginMember(t, server, "a@example.test", "ChangedPassword123!")
	secondToken := loginMember(t, server, "a@example.test", "ChangedPassword123!")
	requestJSON(t, server, "POST", "/api/auth/logout", newToken, nil, 200, nil)
	requestJSON(t, server, "GET", "/api/me", secondToken, nil, 401, nil)
	requestJSON(t, server, "GET", "/api/me", tokenB, nil, 200, nil)
	var hash string
	if err = a.S.DB.QueryRow(ctx, `SELECT password_hash FROM gm_members WHERE email=$1`, "a@example.test").Scan(&hash); err != nil {
		t.Fatal(err)
	}
	if err = bcrypt.CompareHashAndPassword([]byte(hash), []byte("ChangedPassword123!")); err != nil {
		t.Fatal("password was not persisted with bcrypt")
	}
}

func TestDatabaseAdminVisibilityAndConstraints(t *testing.T) {
	a, server := databaseAPI(t)
	ctx := context.Background()
	id := registerMember(t, server, "admin@example.test")
	token := loginMember(t, server, "admin@example.test", "Password123!")
	requestJSON(t, server, "POST", "/api/admin/sync", token, nil, 403, nil)
	if _, err := a.S.DB.Exec(ctx, `UPDATE gm_members SET role='admin' WHERE id=$1`, id); err != nil {
		t.Fatal(err)
	}
	// The same JWT gains/loses admin access from the current DB role.
	requestJSON(t, server, "GET", "/api/admin/games", token, nil, 200, nil)
	var created struct {
		ID int64 `json:"id"`
	}
	game := store.Game{Title: "Demo: Admin fixture", Genres: []string{"rpg"}, Tags: []string{"singleplayer", "story-rich"}, Platforms: []string{"PC"}, Published: true}
	requestJSON(t, server, "POST", "/api/admin/games", token, game, 200, &created)
	path := "/api/admin/games/" + strconv.FormatInt(created.ID, 10)
	fav := "/api/me/favorites/" + strconv.FormatInt(created.ID, 10)
	requestJSON(t, server, "PUT", fav, token, nil, 200, nil)
	game.Published = false
	requestJSON(t, server, "PUT", path, token, game, 200, nil)
	requestJSON(t, server, "GET", "/api/games/"+strconv.FormatInt(created.ID, 10), "", nil, 404, nil)
	var games []store.Game
	for _, publicPath := range []string{"/api/games", "/api/me/favorites"} {
		requestJSON(t, server, "GET", publicPath, token, nil, 200, &games)
		if len(games) != 0 {
			t.Fatal("hidden game is visible")
		}
	}
	requestJSON(t, server, "PUT", fav, token, nil, 404, nil)
	preference := map[string]string{"genre": "rpg", "mode": "solo", "style": "story"}
	requestJSON(t, server, "POST", "/api/recommendations", "", preference, 200, &games)
	if len(games) != 0 {
		t.Fatal("hidden game recommended")
	}
	game.Published = true
	requestJSON(t, server, "PUT", path, token, game, 200, nil)
	requestJSON(t, server, "POST", "/api/recommendations", "", preference, 200, &games)
	if len(games) != 1 || games[0].Score != 100 {
		t.Fatal("database-backed recommendation incorrect")
	}
	if _, err := a.S.DB.Exec(ctx, `INSERT INTO gm_favorites(member_id,game_id) VALUES($1,$2)`, id, created.ID); err == nil {
		t.Fatal("duplicate favorite accepted")
	}
	if _, err := a.S.DB.Exec(ctx, `INSERT INTO gm_favorites(member_id,game_id) VALUES($1,$2)`, id, created.ID+10000); err == nil {
		t.Fatal("missing game FK accepted")
	}
	if _, err := a.S.DB.Exec(ctx, `UPDATE gm_members SET role='owner' WHERE id=$1`, id); err == nil {
		t.Fatal("invalid DB role accepted")
	}
	requestJSON(t, server, "DELETE", path, token, nil, 200, nil)
	var count int
	if err := a.S.DB.QueryRow(ctx, `SELECT count(*) FROM gm_favorites`).Scan(&count); err != nil || count != 0 {
		t.Fatal("favorite cascade failed: ", err)
	}
	requestJSON(t, server, "DELETE", path, token, nil, 404, nil)
	if _, err := a.S.DB.Exec(ctx, `UPDATE gm_members SET role='member' WHERE id=$1`, id); err != nil {
		t.Fatal(err)
	}
	requestJSON(t, server, "GET", "/api/admin/games", token, nil, 403, nil)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestDatabaseRawgSyncPreservesEditsAndRollsBack(t *testing.T) {
	a, _ := databaseAPI(t)
	ctx := context.Background()
	a.RawgKey = "test-only-not-a-live-key"
	var payload string
	a.RawgClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host != "api.rawg.io" || req.URL.Query().Get("page_size") != "40" {
			t.Fatal("unexpected RAWG request")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(payload)), Header: make(http.Header)}, nil
	})}
	if _, err := a.S.DB.Exec(ctx, `INSERT INTO gm_games(rawg_id,title,description,is_published) VALUES(101,'Demo: Old','Admin description',false)`); err != nil {
		t.Fatal(err)
	}
	payload = `{"results":[{"id":101,"name":"Demo: Updated metadata"},{"id":102,"name":"Demo: Imported fixture"}]}`
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		a.Sync(w, httptest.NewRequest("POST", "/api/admin/sync", nil))
		if w.Code != 200 {
			t.Fatalf("sync: %d %s", w.Code, w.Body.String())
		}
	}
	var count int
	var description string
	var published bool
	if err := a.S.DB.QueryRow(ctx, `SELECT count(*) FROM gm_games`).Scan(&count); err != nil || count != 2 {
		t.Fatal("RAWG deduplication failed: ", err)
	}
	if err := a.S.DB.QueryRow(ctx, `SELECT description,is_published FROM gm_games WHERE rawg_id=101`).Scan(&description, &published); err != nil || description != "Admin description" || published {
		t.Fatal("admin fields overwritten: ", err)
	}
	// Force a real SQL failure on row 2, after row 1 was inserted into the transaction.
	if _, err := a.S.DB.Exec(ctx, `CREATE FUNCTION reject_fixture() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.rawg_id=104 THEN RAISE EXCEPTION 'intentional QA rollback'; END IF; RETURN NEW; END $$; CREATE TRIGGER qa_reject BEFORE INSERT ON gm_games FOR EACH ROW EXECUTE FUNCTION reject_fixture();`); err != nil {
		t.Fatal(err)
	}
	payload = `{"results":[{"id":103,"name":"Demo: Must roll back"},{"id":104,"name":"Demo: Forced failure"}]}`
	w := httptest.NewRecorder()
	a.Sync(w, httptest.NewRequest("POST", "/api/admin/sync", nil))
	if w.Code != 500 {
		t.Fatalf("wanted failed transaction, got %d", w.Code)
	}
	if err := a.S.DB.QueryRow(ctx, `SELECT count(*) FROM gm_games`).Scan(&count); err != nil || count != 2 {
		t.Fatal("partial import persisted: ", err)
	}
	payload = `{}`
	w = httptest.NewRecorder()
	a.Sync(w, httptest.NewRequest("POST", "/api/admin/sync", nil))
	if w.Code != 502 {
		t.Fatal("missing results was reported as a successful import")
	}
}

func TestReviewOwnershipAndLifecycle(t *testing.T) {
	a, server := databaseAPI(t)
	ctx := context.Background()
	owner := registerMember(t, server, "review-owner@example.test")
	other := registerMember(t, server, "review-other@example.test")
	token := loginMember(t, server, "review-owner@example.test", "Password123!")
	otherToken := loginMember(t, server, "review-other@example.test", "Password123!")
	var gameID int64
	if err := a.S.DB.QueryRow(ctx, "INSERT INTO gm_games(title) VALUES('Review fixture') RETURNING id").Scan(&gameID); err != nil {
		t.Fatal(err)
	}
	path := "/api/games/" + strconv.FormatInt(gameID, 10) + "/reviews"
	body := map[string]any{"rating": 5, "comment": "เกมสนุกมากเล่นกับเพื่อนได้"}
	var created struct {
		ID int64 `json:"id"`
		MemberID int64 `json:"memberId"`
	}
	requestJSON(t, server, "POST", path, "", body, 401, nil)
	requestJSON(t, server, "POST", path, token, map[string]any{"rating": 5, "comment": "สั้น"}, 400, nil)
	requestJSON(t, server, "POST", path, token, body, 201, &created)
	if created.MemberID != owner || created.ID <= 0 {
		t.Fatal("missing review owner")
	}
	requestJSON(t, server, "POST", path, token, body, 409, nil)
	itemPath := path + "/" + strconv.FormatInt(created.ID, 10)
	changed := map[string]any{"rating": 3, "comment": "Updated review"}
	requestJSON(t, server, "PUT", itemPath, "", changed, 401, nil)
	requestJSON(t, server, "PUT", itemPath, otherToken, changed, 404, nil)
	requestJSON(t, server, "DELETE", itemPath, otherToken, nil, 404, nil)
	requestJSON(t, server, "PUT", itemPath, token, changed, 200, nil)
	var items []struct {
		ID int64
		MemberID int64
		Rating int
		Comment string
	}
	requestJSON(t, server, "GET", path, "", nil, 200, &items)
	if len(items) != 1 || items[0].Rating != 3 || items[0].Comment != "Updated review" || items[0].MemberID != owner {
		t.Fatal("updated review not persisted")
	}
	requestJSON(t, server, "DELETE", itemPath, token, nil, 200, nil)
	requestJSON(t, server, "GET", path, "", nil, 200, &items)
	if len(items) != 0 { t.Fatal("review still exists") }
	requestJSON(t, server, "POST", path, token, body, 201, &created)
	itemPath = path + "/" + strconv.FormatInt(created.ID, 10)
	if _, err := a.S.DB.Exec(ctx, "UPDATE gm_members SET role='admin' WHERE id=$1", other); err != nil {
		t.Fatal(err)
	}
	requestJSON(t, server, "PUT", itemPath, otherToken, changed, 404, nil)
	requestJSON(t, server, "DELETE", itemPath, otherToken, nil, 200, nil)
}

func TestAdminMemberRoles(t *testing.T) {
	a, server := databaseAPI(t)
	ctx := context.Background()
	admin := registerMember(t, server, "role-admin@example.test")
	member := registerMember(t, server, "role-member@example.test")
	if _, err := a.S.DB.Exec(ctx, "UPDATE gm_members SET role='admin' WHERE id=$1", admin); err != nil {
		t.Fatal(err)
	}
	adminToken := loginMember(t, server, "role-admin@example.test", "Password123!")
	memberToken := loginMember(t, server, "role-member@example.test", "Password123!")
	listPath := "/api/admin/members"
	targetPath := listPath + "/" + strconv.FormatInt(member, 10) + "/role"
	promote := map[string]string{"role": "admin"}
	demote := map[string]string{"role": "member"}
	requestJSON(t, server, "GET", listPath, "", nil, 401, nil)
	requestJSON(t, server, "GET", listPath, memberToken, nil, 403, nil)
	requestJSON(t, server, "PUT", targetPath, "", promote, 401, nil)
	requestJSON(t, server, "PUT", targetPath, memberToken, promote, 403, nil)
	var items []map[string]any
	requestJSON(t, server, "GET", listPath, adminToken, nil, 200, &items)
	if len(items) != 2 {
		t.Fatal("expected two members")
	}
	for _, item := range items {
		if _, ok := item["passwordHash"]; ok {
			t.Fatal("password hash leaked")
		}
		if _, ok := item["tokenVersion"]; ok {
			t.Fatal("token version leaked")
		}
	}
	requestJSON(t, server, "PUT", listPath+"/"+strconv.FormatInt(admin, 10)+"/role", adminToken, demote, 400, nil)
	requestJSON(t, server, "PUT", targetPath, adminToken, map[string]string{"role": "owner"}, 400, nil)
	requestJSON(t, server, "PUT", listPath+"/999999/role", adminToken, promote, 404, nil)
	var updated struct{ Role string }
	requestJSON(t, server, "PUT", targetPath, adminToken, promote, 200, &updated)
	if updated.Role != "admin" {
		t.Fatal("promotion failed")
	}
	requestJSON(t, server, "GET", "/api/me", memberToken, nil, 401, nil)
	promotedToken := loginMember(t, server, "role-member@example.test", "Password123!")
	requestJSON(t, server, "GET", listPath, promotedToken, nil, 200, nil)
	requestJSON(t, server, "PUT", targetPath, adminToken, demote, 200, &updated)
	if updated.Role != "member" {
		t.Fatal("demotion failed")
	}
	requestJSON(t, server, "GET", listPath, promotedToken, nil, 401, nil)
	newToken := loginMember(t, server, "role-member@example.test", "Password123!")
	requestJSON(t, server, "GET", listPath, newToken, nil, 403, nil)
	requestJSON(t, server, "GET", "/api/me", newToken, nil, 200, &updated)
	if updated.Role != "member" {
		t.Fatal("role not persisted")
	}
}

func TestAdminStatsUsesRealDataAndRequiresRole(t *testing.T) {
	a, server := databaseAPI(t)
	id := registerMember(t, server, "stats-admin@example.test")
	registerMember(t, server, "stats-member@example.test")
	if _, err := a.S.DB.Exec(context.Background(), "UPDATE gm_members SET role='admin' WHERE id=$1", id); err != nil { t.Fatal(err) }
	admin := loginMember(t, server, "stats-admin@example.test", "Password123!")
	member := loginMember(t, server, "stats-member@example.test", "Password123!")
	requestJSON(t, server, "GET", "/api/admin/stats", "", nil, 401, nil)
	requestJSON(t, server, "GET", "/api/admin/stats", member, nil, 403, nil)
	var stats store.AdminStats
	requestJSON(t, server, "GET", "/api/admin/stats", admin, nil, 200, &stats)
	if stats.Members != 2 || stats.Admins != 1 || stats.Games != 0 || stats.Reviews != 0 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
}
