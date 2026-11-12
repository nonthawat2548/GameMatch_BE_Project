package store

import "context"

func (s *Store) Games(ctx context.Context, admin bool) ([]Game, error) {
	rows, e := s.DB.Query(ctx, `SELECT id,rawg_id,title,description,description_th,image_url,COALESCE((SELECT array_agg(d.slug ORDER BY x.position) FROM gm_game_genres x JOIN gm_genres d ON d.id=x.genre_id WHERE x.game_id=gm_games.id),genres),tags,platforms,is_published,steam_id,COALESCE(release_date::text,''),synced_at,popularity_order,COALESCE(popularity_source,''),popularity_updated_at FROM gm_games WHERE ($1 OR is_published) ORDER BY popularity_order ASC NULLS LAST,release_date DESC NULLS LAST,id DESC`, admin)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Game{}
	for rows.Next() {
		var g Game
		if e = rows.Scan(&g.ID, &g.RawgID, &g.Title, &g.Description, &g.DescriptionTh, &g.ImageURL, &g.Genres, &g.Tags, &g.Platforms, &g.Published, &g.SteamID, &g.ReleaseDate, &g.SyncedAt, &g.PopularityOrder, &g.PopularitySource, &g.PopularityUpdatedAt); e != nil {
			return nil, e
		}
		out = append(out, g)
	}
	return out, rows.Err()
}
func (s *Store) Save(ctx context.Context, g Game) (int64, error) {
	g.Genres = cleanGenres(g.Genres)
	if g.Tags == nil { g.Tags = []string{} }
	if g.Platforms == nil { g.Platforms = []string{} }
	tx, err := s.DB.Begin(ctx)
	if err != nil { return 0, err }
	defer tx.Rollback(ctx)
	var id int64
	if g.ID > 0 {
		err = tx.QueryRow(ctx, `UPDATE gm_games SET title=$1,description=$2,description_th=$3,image_url=$4,genres=$5,tags=$6,platforms=$7,is_published=$8 WHERE id=$9 RETURNING id`, g.Title, g.Description, g.DescriptionTh, g.ImageURL, g.Genres, g.Tags, g.Platforms, g.Published, g.ID).Scan(&id)
	} else {
		err = tx.QueryRow(ctx, `INSERT INTO gm_games(rawg_id,title,description,description_th,image_url,genres,tags,platforms,is_published) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`, g.RawgID, g.Title, g.Description, g.DescriptionTh, g.ImageURL, g.Genres, g.Tags, g.Platforms, g.Published).Scan(&id)
	}
	if err != nil { return 0, err }
	if err = replaceGameGenres(ctx, tx, id, g.Genres); err != nil { return 0, err }
	if err = tx.Commit(ctx); err != nil { return 0, err }
	return id, nil
}

func (s *Store) GamePublished(ctx context.Context, id int64) (bool, error) {
	var published bool
	err := s.DB.QueryRow(ctx, `SELECT is_published FROM gm_games WHERE id=$1`, id).Scan(&published)
	return published, err
}
func (s *Store) PublishedGameExists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := s.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM gm_games WHERE id=$1 AND is_published)`, id).Scan(&exists)
	return exists, err
}
func (s *Store) DeleteGame(ctx context.Context, id int64) (bool, error) {
	result, err := s.DB.Exec(ctx, `DELETE FROM gm_games WHERE id=$1`, id)
	return result.RowsAffected() > 0, err
}

func (s *Store) PublishedGameSteamID(ctx context.Context, id int64) (*int64, error) {
	var steamID *int64
	err := s.DB.QueryRow(ctx, `SELECT steam_id FROM gm_games WHERE id=$1 AND is_published`, id).Scan(&steamID)
	return steamID, err
}
