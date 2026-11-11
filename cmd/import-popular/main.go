package main

import (
	"context"
	"gamematch/handler"
	"gamematch/store"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" || strings.Contains(dsn, "ewfvqtflucuvtiajmmgb") {
		log.Fatal("Set the dedicated GameMatch DATABASE_URL; MovieHub is not allowed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal("Invalid database configuration")
	}
	defer db.Close()
	count, err := handler.ImportSteamPopular(ctx, &store.Store{DB: db}, nil, 100, func(imported, checked int, title string) {
		log.Printf("Verified %d/100 games (%d candidates checked): %s", imported, checked, title)
	})
	if err != nil {
		log.Printf("Import failed; transaction not committed: %v", err)
		os.Exit(1)
	}
	log.Printf("Committed %d popular Steam games; existing descriptions and publication preserved", count)
}
