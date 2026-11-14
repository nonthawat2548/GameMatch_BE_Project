package store

import (
	"context"
	"errors"
)

var ErrOwnReview = errors.New("cannot vote on own review")

func (s *Store) ReviewVoteIDs(ctx context.Context, gameID, memberID int64) ([]int64, error) {
	rows, err := s.DB.Query(ctx, `SELECT v.review_id FROM gm_review_votes v JOIN gm_reviews r ON r.id=v.review_id JOIN gm_games g ON g.id=r.game_id WHERE r.game_id=$1 AND g.is_published AND v.member_id=$2`, gameID, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []int64{}
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
func (s *Store) SetReviewVote(ctx context.Context, gameID, reviewID, memberID int64, add bool) (int64, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	var owner int64
	err = tx.QueryRow(ctx, `SELECT r.member_id FROM gm_reviews r JOIN gm_games g ON g.id=r.game_id WHERE r.id=$1 AND r.game_id=$2 AND g.is_published FOR UPDATE OF r`, reviewID, gameID).Scan(&owner)
	if err != nil {
		return 0, err
	}
	if owner == memberID {
		return 0, ErrOwnReview
	}
	if add {
		_, err = tx.Exec(ctx, `INSERT INTO gm_review_votes(review_id,member_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, reviewID, memberID)
	} else {
		_, err = tx.Exec(ctx, `DELETE FROM gm_review_votes WHERE review_id=$1 AND member_id=$2`, reviewID, memberID)
	}
	if err != nil {
		return 0, err
	}
	var count int64
	if err = tx.QueryRow(ctx, `SELECT count(*) FROM gm_review_votes WHERE review_id=$1`, reviewID).Scan(&count); err != nil {
		return 0, err
	}
	return count, tx.Commit(ctx)
}
