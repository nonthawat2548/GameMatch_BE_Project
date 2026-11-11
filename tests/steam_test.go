package tests

import (
	"context"
	"encoding/json"
	"gamematch/handler"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func steamResponse(id int64, kind, date string, soon bool) string {
	data := map[string]any{strconv.FormatInt(id, 10): map[string]any{"success": true, "data": map[string]any{"type": kind, "steam_appid": id, "name": "Demo: Steam fixture " + strconv.FormatInt(id, 10), "short_description": "<b>Fixture &amp; plain text</b>", "header_image": "https://cdn.steamstatic.com/fixture.jpg", "release_date": map[string]any{"coming_soon": soon, "date": date}, "genres": []map[string]string{{"description": "RPG"}}, "categories": []map[string]string{{"description": "Single-player"}, {"description": "Multi-player"}}, "platforms": map[string]bool{"windows": true}}}}
	encoded, _ := json.Marshal(data)
	return string(encoded)
}

func steamClient(feed *string, detail func(int64) string, page *string) *http.Client {
	return &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		body := *page
		if r.URL.Path == "/api/featuredcategories/" {
			body = *feed
		}
		if r.URL.Path == "/api/appdetails" {
			id, _ := strconv.ParseInt(r.URL.Query().Get("appids"), 10, 64)
			body = detail(id)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
}

func TestSteamRejectsUnreleasedNonGameAndMissingTags(t *testing.T) {
	feed := `{"new_releases":{"items":[{"id":501}]}}`
	for _, tc := range []struct {
		name, kind, date, page string
		soon                   bool
	}{
		{"DLC", "dlc", "Jan 1, 2020", `<a class="app_tag">RPG</a>`, false},
		{"upcoming", "game", "Jan 1, 2020", `<a class="app_tag">RPG</a>`, true},
		{"future date", "game", "Jan 1, 2999", `<a class="app_tag">RPG</a>`, false},
		{"ambiguous date", "game", "Coming soon", `<a class="app_tag">RPG</a>`, false},
		{"changed page", "game", "Jan 1, 2020", `<html>Age check or missing tags</html>`, false},
		{"explicit catalog tag", "game", "Jan 1, 2020", `<a class="app_tag">Sexual Content</a>`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := &handler.API{SteamClient: steamClient(&feed, func(id int64) string { return steamResponse(id, tc.kind, tc.date, tc.soon) }, &tc.page)}
			w := httptest.NewRecorder()
			a.SyncSteam(w, httptest.NewRequest("POST", "/api/admin/sync/steam", nil))
			if w.Code != 502 {
				t.Fatalf("want 502, got %d", w.Code)
			}
		})
	}
}

func TestSteamRejectsBadFeedAndConcurrentImport(t *testing.T) {
	for _, body := range []string{`{}`, `{"new_releases":{"items":[]}}`, `not JSON`} {
		page := ""
		a := &handler.API{SteamClient: steamClient(&body, func(int64) string { return "" }, &page)}
		w := httptest.NewRecorder()
		a.SyncSteam(w, httptest.NewRequest("POST", "/api/admin/sync/steam", nil))
		if w.Code != 502 {
			t.Fatal(w.Code)
		}
	}
	a := &handler.API{}
	a.SyncMu.Lock()
	defer a.SyncMu.Unlock()
	w := httptest.NewRecorder()
	a.SyncSteam(w, httptest.NewRequest("POST", "/api/admin/sync/steam", nil))
	if w.Code != 409 {
		t.Fatal(w.Code)
	}
}

func TestDatabaseSteamSyncMetadataPreservationAndRollback(t *testing.T) {
	a, server := databaseAPI(t)
	ctx := context.Background()
	id := registerMember(t, server, "steam-admin@example.test")
	token := loginMember(t, server, "steam-admin@example.test", "Password123!")
	requestJSON(t, server, "POST", "/api/admin/sync/steam", "", nil, 401, nil)
	requestJSON(t, server, "POST", "/api/admin/sync/steam", token, nil, 403, nil)
	if _, err := a.S.DB.Exec(ctx, `UPDATE gm_members SET role='admin' WHERE id=$1`, id); err != nil {
		t.Fatal(err)
	}
	feed := `{"new_releases":{"items":[{"id":501,"large_capsule_image":"https://cdn.steamstatic.com/capsule_616x353.jpg"},{"id":502}]}}`
	page := `<a class="app_tag">Story Rich</a>`
	a.SteamClient = steamClient(&feed, func(id int64) string { return steamResponse(id, "game", "Jan 1, 2020", false) }, &page)
	if _, err := a.S.DB.Exec(ctx, `INSERT INTO gm_games(steam_id,title,description,is_published) VALUES(501,'Demo: Original','Admin edit',false)`); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		requestJSON(t, server, "POST", "/api/admin/sync/steam", token, nil, 200, nil)
	}
	games, err := a.S.Games(ctx, true)
	if err != nil || len(games) != 2 {
		t.Fatal("dedupe failed: ", err)
	}
	for _, g := range games {
		if g.SteamID == nil || g.ReleaseDate != "2020-01-01" || g.SyncedAt == nil || g.ImageURL == "" {
			t.Fatal("missing source/date/image")
		}
		if *g.SteamID == 501 && (g.Description != "Admin edit" || g.Published) {
			t.Fatal("admin edits overwritten")
		}
		if *g.SteamID == 501 && !strings.Contains(g.ImageURL, "capsule_616x353") {
			t.Fatal("did not prefer the Store capsule")
		}
		if *g.SteamID == 502 && g.Description != "Fixture & plain text" {
			t.Fatal("Steam HTML was not converted to text")
		}
		joined := strings.Join(g.Tags, ",")
		if !strings.Contains(joined, "story-rich") || !strings.Contains(joined, "singleplayer") || strings.Contains(joined, "co-op") {
			t.Fatal("incorrect tag/category mapping")
		}
	}
	if _, err := a.S.DB.Exec(ctx, `UPDATE gm_games SET description='' WHERE steam_id=501`); err != nil {
		t.Fatal(err)
	}
	requestJSON(t, server, "POST", "/api/admin/sync/steam", token, nil, 200, nil)
	var cleared string
	if err := a.S.DB.QueryRow(ctx, `SELECT description FROM gm_games WHERE steam_id=501`).Scan(&cleared); err != nil || cleared != "" {
		t.Fatal("sync overwrote an intentionally cleared description")
	}
	if _, err := a.S.DB.Exec(ctx, `CREATE FUNCTION reject_steam_fixture() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.steam_id=503 THEN RAISE EXCEPTION 'intentional Steam QA rollback'; END IF; RETURN NEW; END $$; CREATE TRIGGER reject_steam BEFORE INSERT ON gm_games FOR EACH ROW EXECUTE FUNCTION reject_steam_fixture();`); err != nil {
		t.Fatal(err)
	}
	feed = `{"new_releases":{"items":[{"id":503},{"id":504}]}}`
	requestJSON(t, server, "POST", "/api/admin/sync/steam", token, nil, 500, nil)
	var count int
	if err := a.S.DB.QueryRow(ctx, `SELECT count(*) FROM gm_games`).Scan(&count); err != nil || count != 2 {
		t.Fatal("partial Steam import survived rollback")
	}
}
