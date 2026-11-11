package store

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Game struct {
	ID                  int64      `json:"id"`
	RawgID              *int64     `json:"rawgId,omitempty"`
	SteamID             *int64     `json:"steamId,omitempty"`
	ReleaseDate         string     `json:"releaseDate,omitempty"`
	SyncedAt            *time.Time `json:"syncedAt,omitempty"`
	PopularityOrder     *int       `json:"popularityOrder,omitempty"`
	PopularitySource    string     `json:"popularitySource,omitempty"`
	PopularityUpdatedAt *time.Time `json:"popularityUpdatedAt,omitempty"`
	Title               string     `json:"title"`
	Description         string     `json:"description"`
	ImageURL            string     `json:"imageUrl"`
	Genres              []string   `json:"genres"`
	Tags                []string   `json:"tags"`
	Platforms           []string   `json:"platforms"`
	Published           bool       `json:"isPublished"`
	Score               int        `json:"matchScore"`
	Reasons             []string   `json:"reasons,omitempty"`
}
type Store struct{ DB *pgxpool.Pool }

func (s *Store) Games(ctx context.Context, admin bool) ([]Game, error) {
	rows, e := s.DB.Query(ctx, `SELECT id,rawg_id,title,description,image_url,genres,tags,platforms,is_published,steam_id,COALESCE(release_date::text,''),synced_at,popularity_order,COALESCE(popularity_source,''),popularity_updated_at FROM gm_games WHERE ($1 OR is_published) ORDER BY popularity_order ASC NULLS LAST,release_date DESC NULLS LAST,id DESC`, admin)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Game{}
	for rows.Next() {
		var g Game
		if e = rows.Scan(&g.ID, &g.RawgID, &g.Title, &g.Description, &g.ImageURL, &g.Genres, &g.Tags, &g.Platforms, &g.Published, &g.SteamID, &g.ReleaseDate, &g.SyncedAt, &g.PopularityOrder, &g.PopularitySource, &g.PopularityUpdatedAt); e != nil {
			return nil, e
		}
		out = append(out, g)
	}
	return out, rows.Err()
}
func (s *Store) Save(ctx context.Context, g Game) (int64, error) {
	if g.Genres == nil {
		g.Genres = []string{}
	}
	if g.Tags == nil {
		g.Tags = []string{}
	}
	if g.Platforms == nil {
		g.Platforms = []string{}
	}
	if g.ID > 0 {
		var id int64
		e := s.DB.QueryRow(ctx, `UPDATE gm_games SET title=$1,description=$2,image_url=$3,genres=$4,tags=$5,platforms=$6,is_published=$7 WHERE id=$8 RETURNING id`, g.Title, g.Description, g.ImageURL, g.Genres, g.Tags, g.Platforms, g.Published, g.ID).Scan(&id)
		return id, e
	}
	var id int64
	e := s.DB.QueryRow(ctx, `INSERT INTO gm_games(rawg_id,title,description,image_url,genres,tags,platforms,is_published) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`, g.RawgID, g.Title, g.Description, g.ImageURL, g.Genres, g.Tags, g.Platforms, g.Published).Scan(&id)
	return id, e
}
