package tests

import (
	"context"
	"gamematch/handler"
	"gamematch/store"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func TestDetailRejectsInvalidIDBeforeDatabase(t *testing.T) {
	api := &handler.API{}
	routes := api.Routes()
	for _, id := range []string{"abc", "0", "-1", "9223372036854775808"} {
		t.Run(id, func(t *testing.T) {
			w := httptest.NewRecorder()
			routes.ServeHTTP(w, httptest.NewRequest("GET", "/api/games/"+id, nil))
			if w.Code != http.StatusBadRequest {
				t.Fatalf("invalid ID should be rejected before accessing the store: %d", w.Code)
			}
		})
	}
}

func TestDatabaseCatalogFilteringAndDetailIDs(t *testing.T) {
	api, server := databaseAPI(t)
	ctx := context.Background()
	save := func(title string, published bool, genre string) int64 {
		t.Helper()
		id, err := api.S.Save(ctx, store.Game{Title: title, Published: published, Genres: []string{genre}})
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	literalID := save("Demo: 100%_Action", true, "action")
	otherID := save("Demo: ACTION adventure", true, "action")
	hiddenID := save("Demo: Hidden 100%_action", false, "action")
	save("Demo: Puzzle", true, "puzzle")

	var matches []store.Game
	requestJSON(t, server, "GET", "/api/games?q=aCtIoN", "", nil, 200, &matches)
	if len(matches) != 2 || matches[0].ID != otherID || matches[1].ID != literalID {
		t.Fatal("search must remain case-insensitive, ordered and limited to published games")
	}
	requestJSON(t, server, "GET", "/api/games?q="+url.QueryEscape("%_"), "", nil, 200, &matches)
	if len(matches) != 1 || matches[0].ID != literalID {
		t.Fatal("search must treat SQL wildcard characters as literal text")
	}
	requestJSON(t, server, "GET", "/api/games?q="+url.QueryEscape("' OR true --"), "", nil, 200, &matches)
	if matches == nil || len(matches) != 0 {
		t.Fatal("a search without matches must return an empty JSON array")
	}

	var detail store.Game
	requestJSON(t, server, "GET", "/api/games/00"+strconv.FormatInt(literalID, 10), "", nil, 200, &detail)
	if detail.ID != literalID || detail.Genres == nil || detail.Tags == nil || detail.Platforms == nil {
		t.Fatal("numeric ID parsing or game JSON arrays changed")
	}
	requestJSON(t, server, "GET", "/api/games/"+strconv.FormatInt(hiddenID, 10), "", nil, 404, nil)
	requestJSON(t, server, "GET", "/api/games/9223372036854775807", "", nil, 404, nil)

	memberA := registerMember(t, server, "catalog-a@example.test")
	memberB := registerMember(t, server, "catalog-b@example.test")
	tokenA := loginMember(t, server, "catalog-a@example.test", "Password123!")
	tokenB := loginMember(t, server, "catalog-b@example.test", "Password123!")
	for _, pair := range [][2]int64{{memberA, literalID}, {memberA, hiddenID}, {memberB, otherID}} {
		if err := api.S.AddFavorite(ctx, pair[0], pair[1]); err != nil {
			t.Fatal(err)
		}
	}
	requestJSON(t, server, "GET", "/api/me/favorites", tokenA, nil, 200, &matches)
	if len(matches) != 1 || matches[0].ID != literalID {
		t.Fatal("favorites must exclude hidden games and other members' selections")
	}
	requestJSON(t, server, "GET", "/api/me/favorites", tokenB, nil, 200, &matches)
	if len(matches) != 1 || matches[0].ID != otherID {
		t.Fatal("favorites must be filtered to the authenticated member")
	}
	requestJSON(t, server, "GET", "/api/me/recommendations", tokenA, nil, 200, &matches)
	if len(matches) != 1 || matches[0].ID != otherID {
		t.Fatal("favorite recommendations must still use the remaining public catalog")
	}
}

func TestDatabaseMetadataKeepsUnknownMajorUpdateEmpty(t *testing.T) {
	api, server := databaseAPI(t)
	var gameID int64
	if err := api.S.DB.QueryRow(context.Background(), `INSERT INTO gm_games(title,steam_id) VALUES('Demo: Metadata fixture',730) RETURNING id`).Scan(&gameID); err != nil {
		t.Fatal(err)
	}
	api.SteamClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != "store.steampowered.com" || r.URL.Query().Get("appids") != "730" {
			t.Fatal("unexpected metadata request")
		}
		body := `{"730":{"success":true,"data":{"release_date":{"date":"Aug 21, 2012","coming_soon":false}}}}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
	var metadata map[string]any
	requestJSON(t, server, "GET", "/api/games/"+strconv.FormatInt(gameID, 10)+"/metadata", "", nil, 200, &metadata)
	if metadata["majorUpdateDate"] != "" || metadata["releaseDate"] != "2012-08-21" {
		t.Fatal("unknown major update must stay empty without changing the sourced release date")
	}
}
