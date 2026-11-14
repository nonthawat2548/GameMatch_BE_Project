package store

import "context"

func (s *Store) FavoriteIDs(ctx context.Context, memberID int64) (map[int64]bool, error) {
	rows, err := s.DB.Query(ctx, `SELECT game_id FROM gm_favorites WHERE member_id=$1`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := map[int64]bool{}
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids[id] = true
	}
	return ids, rows.Err()
}
func (s *Store) AddFavorite(ctx context.Context, memberID, gameID int64) error {
	_, err := s.DB.Exec(ctx, `INSERT INTO gm_favorites(member_id,game_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, memberID, gameID)
	return err
}
func (s *Store) RemoveFavorite(ctx context.Context, memberID, gameID int64) error {
	_, err := s.DB.Exec(ctx, `DELETE FROM gm_favorites WHERE member_id=$1 AND game_id=$2`, memberID, gameID)
	return err
}
func (s *Store) MemberPreferences(ctx context.Context, memberID int64) (genre, mode, style string, err error) {
	err = s.DB.QueryRow(ctx, `SELECT genre,mode,style FROM gm_preferences WHERE member_id=$1`, memberID).Scan(&genre, &mode, &style)
	return
}
func (s *Store) SavePreferences(ctx context.Context, memberID int64, genre, mode, style string) error {
	_, err := s.DB.Exec(ctx, `INSERT INTO gm_preferences(member_id,genre,mode,style) VALUES($1,$2,$3,$4) ON CONFLICT(member_id) DO UPDATE SET genre=$2,mode=$3,style=$4`, memberID, genre, mode, style)
	return err
}
