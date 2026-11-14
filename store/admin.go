package store

import "context"

type AdminStats struct {
	Games     int `json:"games"`
	Published int `json:"published"`
	Members   int `json:"members"`
	Admins    int `json:"admins"`
	Reviews   int `json:"reviews"`
}

func (s *Store) Stats(ctx context.Context) (AdminStats, error) {
	var stats AdminStats
	err := s.DB.QueryRow(ctx, `SELECT
		(SELECT count(*) FROM gm_games),
		(SELECT count(*) FROM gm_games WHERE is_published),
		(SELECT count(*) FROM gm_members),
		(SELECT count(*) FROM gm_members WHERE role='admin'),
		(SELECT count(*) FROM gm_reviews)`).Scan(&stats.Games, &stats.Published, &stats.Members, &stats.Admins, &stats.Reviews)
	return stats, err
}
