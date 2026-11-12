package store

import "context"

func (s *Store) Games(ctx context.Context, admin bool) ([]Game, error) {
	rows, e := s.DB.Query(ctx, `SELECT id,rawg_id,title,description,description_th,image_url,genres,tags,platforms,is_published,steam_id,COALESCE(release_date::text,''),synced_at,popularity_order,COALESCE(popularity_source,''),popularity_updated_at FROM gm_games WHERE ($1 OR is_published) ORDER BY popularity_order ASC NULLS LAST,release_date DESC NULLS LAST,id DESC`, admin)
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
		e := s.DB.QueryRow(ctx, `UPDATE gm_games SET title=$1,description=$2,description_th=$3,image_url=$4,genres=$5,tags=$6,platforms=$7,is_published=$8 WHERE id=$9 RETURNING id`, g.Title, g.Description, g.DescriptionTh, g.ImageURL, g.Genres, g.Tags, g.Platforms, g.Published, g.ID).Scan(&id)
		return id, e
	}
	var id int64
	e := s.DB.QueryRow(ctx, `INSERT INTO gm_games(rawg_id,title,description,description_th,image_url,genres,tags,platforms,is_published) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`, g.RawgID, g.Title, g.Description, g.DescriptionTh, g.ImageURL, g.Genres, g.Tags, g.Platforms, g.Published).Scan(&id)
	return id, e
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
