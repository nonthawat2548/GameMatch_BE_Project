package tests

import (
	"context"
	"gamematch/handler"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func popularClient() *http.Client {
	return &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		body := `<a class="app_tag">Story Rich</a>`
		if r.URL.Host == "api.steampowered.com" {
			body = `{"response":{"ranks":[{"rank":2,"appid":502},{"rank":1,"appid":501}]}}`
		}
		if r.URL.Path == "/search/results/" {
			body = `{"success":1,"results_html":"<a data-ds-appid=\"501\"></a><a data-ds-appid=\"503\"></a>"}`
		}
		if r.URL.Path == "/api/appdetails" {
			id, _ := strconv.ParseInt(r.URL.Query().Get("appids"), 10, 64)
			body = steamResponse(id, "game", "Jan 1, 2020", false)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
}

func TestSteamRejectsSoftwareEvenWhenStoreTypeIsGame(t *testing.T) {
	feed := `{"new_releases":{"items":[{"id":501}]}}`
	page := `<a class="app_tag">Casual</a>`
	a := &handler.API{SteamClient: steamClient(&feed, func(id int64) string {
		return strings.ReplaceAll(steamResponse(id, "game", "Jan 1, 2020", false), "RPG", "Utilities")
	}, &page)}
	w := httptest.NewRecorder()
	a.SyncSteam(w, httptest.NewRequest("POST", "/api/admin/sync/steam", nil))
	if w.Code != 502 {
		t.Fatal("software should not reach the database")
	}
}

func TestDatabasePopularImportDeduplicatesAndRollsBackRanking(t *testing.T) {
	a, _ := databaseAPI(t)
	ctx := context.Background()
	if _, err := a.S.DB.Exec(ctx, `INSERT INTO gm_games(steam_id,title,description,is_published) VALUES(501,'Demo: Original','',false); INSERT INTO gm_games(steam_id,title,popularity_order,popularity_source) VALUES(999,'Demo: Old ranking',1,'old source');`); err != nil {
		t.Fatal(err)
	}
	count, err := handler.ImportSteamPopular(ctx, a.S, popularClient(), 2, nil)
	if err != nil || count != 2 {
		t.Fatal("popular import failed: ", err)
	}
	games, err := a.S.Games(ctx, true)
	if err != nil || len(games) != 3 {
		t.Fatal("duplicate or missing catalog rows: ", err)
	}
	if *games[0].SteamID != 501 || games[0].PopularityOrder == nil || *games[0].PopularityOrder != 1 || games[0].Published || games[0].Description != "" || games[0].PopularitySource != "Steam Most Played" || games[0].PopularityUpdatedAt == nil {
		t.Fatal("source order or admin fields not preserved")
	}
	if *games[2].SteamID != 999 || games[2].PopularityOrder != nil {
		t.Fatal("old ranking metadata not cleared")
	}
	// An incomplete fetch must retain the existing complete ranking.
	if _, err = handler.ImportSteamPopular(ctx, a.S, popularClient(), 100, nil); err == nil {
		t.Fatal("incomplete popular set was accepted")
	}
	var order int
	if err = a.S.DB.QueryRow(ctx, `SELECT popularity_order FROM gm_games WHERE steam_id=501`).Scan(&order); err != nil || order != 1 {
		t.Fatal("failed fetch changed stored ranking")
	}
	if _, err = a.S.DB.Exec(ctx, `CREATE FUNCTION reject_popular() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.steam_id=502 THEN RAISE EXCEPTION 'intentional popular import rollback'; END IF; RETURN NEW; END $$; CREATE TRIGGER reject_popular BEFORE INSERT ON gm_games FOR EACH ROW EXECUTE FUNCTION reject_popular();`); err != nil {
		t.Fatal(err)
	}
	if _, err = handler.ImportSteamPopular(ctx, a.S, popularClient(), 2, nil); err == nil {
		t.Fatal("forced transaction failure was ignored")
	}
	if err = a.S.DB.QueryRow(ctx, `SELECT popularity_order FROM gm_games WHERE steam_id=502`).Scan(&order); err != nil || order != 2 {
		t.Fatal("ranking reset survived a rollback")
	}
}
