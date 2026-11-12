package store

import "context"

// Import methods keep each batch atomic and preserve manually edited descriptions/publication.
func (s *Store) ImportRAWG(ctx context.Context, games []Game) (int, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil { return 0, err }
	defer tx.Rollback(ctx)
	for _, g := range games {
		g.Genres = cleanGenres(g.Genres)
		var id int64
		err = tx.QueryRow(ctx, `INSERT INTO gm_games(rawg_id,title,image_url,genres,tags,platforms) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(rawg_id) DO UPDATE SET title=EXCLUDED.title,image_url=EXCLUDED.image_url,genres=EXCLUDED.genres,tags=EXCLUDED.tags,platforms=EXCLUDED.platforms RETURNING id`, g.RawgID, g.Title, g.ImageURL, g.Genres, g.Tags, g.Platforms).Scan(&id)
		if err != nil { return 0, err }
		if err = replaceGameGenres(ctx, tx, id, g.Genres); err != nil { return 0, err }
	}
	if err = tx.Commit(ctx); err != nil { return 0, err }
	return len(games), nil
}
func (s *Store) ImportSteam(ctx context.Context, games []Game) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	for i := len(games)-1; i >= 0; i-- {
		g := games[i]
		g.Genres = cleanGenres(g.Genres)
		var id int64
		err = tx.QueryRow(ctx, `INSERT INTO gm_games(steam_id,title,description,image_url,genres,tags,platforms,release_date,synced_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8::date,now()) ON CONFLICT(steam_id) DO UPDATE SET title=EXCLUDED.title,image_url=EXCLUDED.image_url,genres=EXCLUDED.genres,tags=EXCLUDED.tags,platforms=EXCLUDED.platforms,release_date=EXCLUDED.release_date,synced_at=EXCLUDED.synced_at RETURNING id`, g.SteamID, g.Title, g.Description, g.ImageURL, g.Genres, g.Tags, g.Platforms, g.ReleaseDate).Scan(&id)
		if err != nil { return err }
		if err = replaceGameGenres(ctx, tx, id, g.Genres); err != nil { return err }
	}
	return tx.Commit(ctx)
}
func (s *Store) ImportPopular(ctx context.Context, games []Game) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `UPDATE gm_games SET popularity_order=NULL,popularity_source=NULL,popularity_updated_at=NULL WHERE popularity_order IS NOT NULL`); err != nil { return err }
	for _, g := range games {
		g.Genres = cleanGenres(g.Genres)
		var id int64
		err = tx.QueryRow(ctx, `INSERT INTO gm_games(steam_id,title,description,image_url,genres,tags,platforms,release_date,synced_at,popularity_order,popularity_source,popularity_updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8::date,now(),$9,$10,now()) ON CONFLICT(steam_id) DO UPDATE SET title=EXCLUDED.title,image_url=EXCLUDED.image_url,genres=EXCLUDED.genres,tags=EXCLUDED.tags,platforms=EXCLUDED.platforms,release_date=EXCLUDED.release_date,synced_at=EXCLUDED.synced_at,popularity_order=EXCLUDED.popularity_order,popularity_source=EXCLUDED.popularity_source,popularity_updated_at=EXCLUDED.popularity_updated_at RETURNING id`, g.SteamID, g.Title, g.Description, g.ImageURL, g.Genres, g.Tags, g.Platforms, g.ReleaseDate, g.PopularityOrder, g.PopularitySource).Scan(&id)
		if err != nil { return err }
		if err = replaceGameGenres(ctx, tx, id, g.Genres); err != nil { return err }
	}
	return tx.Commit(ctx)
}
