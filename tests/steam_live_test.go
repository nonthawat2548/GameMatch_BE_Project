package tests

import (
	"context"
	"gamematch/handler"
	"gamematch/store"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Explicit opt-in import into the real catalog, never run by the default test command.
func TestLiveSteamImport(t *testing.T) {
	if os.Getenv("GM_STEAM_IMPORT") != "1" {
		t.Skip("set GM_STEAM_IMPORT=1 and DATABASE_URL to import real Steam data")
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" || strings.Contains(dsn, "ewfvqtflucuvtiajmmgb") {
		t.Fatal("dedicated GameMatch database required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal("invalid database configuration")
	}
	defer db.Close()
	a := &handler.API{S: &store.Store{DB: db}}
	w := httptest.NewRecorder()
	a.SyncSteam(w, httptest.NewRequest("POST", "/api/admin/sync/steam", nil))
	t.Log(w.Body.String())
	if w.Code != 200 {
		t.Fatal("live Steam import failed with status ", w.Code)
	}
	games, err := a.S.Games(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, g := range games {
		if g.SteamID != nil {
			count++
			if g.ImageURL == "" || g.ReleaseDate == "" {
				t.Fatal("missing live image/date")
			}
			t.Logf("Steam %d | %s | %s", *g.SteamID, g.Title, g.ReleaseDate)
		}
	}
	if count == 0 {
		t.Fatal("no real Steam games were persisted")
	}
}
