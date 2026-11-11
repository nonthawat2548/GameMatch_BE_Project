package main

import (
	"context"
	"errors"
	"gamematch/handler"
	"gamematch/mw"
	"gamematch/store"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		log.Fatal("JWT_SECRET must contain at least 32 characters")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if os.Getenv("DATABASE_URL") == "" {
		log.Fatal("DATABASE_URL must be set on the backend")
	}
	config, e := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if e != nil {
		log.Fatal("Invalid DATABASE_URL")
	}
	config.MaxConns = 5
	config.MinConns = 0
	config.MaxConnIdleTime = 5 * time.Minute
	db, e := pgxpool.NewWithConfig(ctx, config)
	if e != nil {
		log.Fatal("Cannot initialize PostgreSQL pool")
	}
	defer db.Close()
	if e = db.Ping(ctx); e != nil {
		log.Fatal("Cannot connect to PostgreSQL: check DATABASE_URL and SSL")
	}
	a := &handler.API{S: &store.Store{DB: db}, Secret: secret, RawgKey: os.Getenv("RAWG_API_KEY")}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	origins := os.Getenv("ALLOWED_ORIGINS")
	if origins == "" {
		origins = "http://localhost:3000"
	}
	server := &http.Server{Addr: ":" + port, Handler: mw.CORS(origins, a.Routes()), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 35 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		log.Println("GameMatch API listening on :" + port)
		if e = server.ListenAndServe(); e != nil && !errors.Is(e, http.ErrServerClosed) {
			log.Fatal(e)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
	<-stop
	shutdown, done := context.WithTimeout(context.Background(), 10*time.Second)
	defer done()
	server.Shutdown(shutdown)
}
