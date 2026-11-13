package store

import (
	"context"
	"time"
)

type Review struct {
	MemberID    int64     `json:"memberId"`
	ID          int64     `json:"id"`
	DisplayName string    `json:"displayName"`
	Rating      int       `json:"rating"`
	Comment     string    `json:"comment"`
	CreatedAt   time.Time `json:"createdAt"`
	HelpfulCount int64 `json:"helpfulCount"`
}

func (s *Store) ListReviews(ctx context.Context, gameID int64) ([]Review, error) {
	rows, err := s.DB.Query(ctx, `SELECT r.id,r.member_id,m.display_name,r.rating,r.comment,r.created_at,(SELECT count(*) FROM gm_review_votes v WHERE v.review_id=r.id) FROM gm_reviews r JOIN gm_members m ON m.id=r.member_id WHERE r.game_id=$1 ORDER BY r.created_at DESC,r.id DESC`, gameID)
	if err != nil { return nil, err }
	defer rows.Close()
	out := []Review{}
	for rows.Next() {
		var item Review
		if err = rows.Scan(&item.ID, &item.MemberID, &item.DisplayName, &item.Rating, &item.Comment, &item.CreatedAt, &item.HelpfulCount); err != nil { return nil, err }
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) AddReview(ctx context.Context, gameID int64, item Review) (Review, error) {
	err := s.DB.QueryRow(ctx, `INSERT INTO gm_reviews(member_id,game_id,rating,comment) VALUES($1,$2,$3,$4) RETURNING id,created_at`, item.MemberID, gameID, item.Rating, item.Comment).Scan(&item.ID, &item.CreatedAt)
	return item, err
}
func (s *Store) UpdateReview(ctx context.Context, reviewID, gameID int64, item Review) (Review, error) {
	err := s.DB.QueryRow(ctx, `UPDATE gm_reviews SET rating=$1,comment=$2 WHERE id=$3 AND game_id=$4 AND member_id=$5 RETURNING id,created_at`, item.Rating, item.Comment, reviewID, gameID, item.MemberID).Scan(&item.ID, &item.CreatedAt)
	return item, err
}
func (s *Store) DeleteReview(ctx context.Context, reviewID, gameID, memberID int64, admin bool) (bool, error) {
	result, err := s.DB.Exec(ctx, `DELETE FROM gm_reviews WHERE id=$1 AND game_id=$2 AND (member_id=$3 OR $4)`, reviewID, gameID, memberID, admin)
	return result.RowsAffected() > 0, err
}
