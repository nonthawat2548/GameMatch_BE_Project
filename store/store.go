package store

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type MatchPart struct {
	Label  string `json:"label"`
	Points int    `json:"points"`
	Max    int    `json:"max"`
}

type Game struct {
	ID                  int64       `json:"id"`
	RawgID              *int64      `json:"rawgId,omitempty"`
	SteamID             *int64      `json:"steamId,omitempty"`
	ReleaseDate         string      `json:"releaseDate,omitempty"`
	SyncedAt            *time.Time  `json:"syncedAt,omitempty"`
	PopularityOrder     *int        `json:"popularityOrder,omitempty"`
	PopularitySource    string      `json:"popularitySource,omitempty"`
	PopularityUpdatedAt *time.Time  `json:"popularityUpdatedAt,omitempty"`
	Title               string      `json:"title"`
	Description         string      `json:"description"`
	DescriptionTh       string      `json:"descriptionTh"`
	ImageURL            string      `json:"imageUrl"`
	Genres              []string    `json:"genres"`
	Tags                []string    `json:"tags"`
	Platforms           []string    `json:"platforms"`
	Published           bool        `json:"isPublished"`
	Score               int         `json:"matchScore"`
	Reasons             []string    `json:"reasons,omitempty"`
	MatchBreakdown      []MatchPart `json:"matchBreakdown,omitempty"`
	SimilarityScore     int         `json:"similarityScore,omitempty"`
}
type Store struct{ DB *pgxpool.Pool }

func (s *Store) Ping(ctx context.Context) error { return s.DB.Ping(ctx) }
