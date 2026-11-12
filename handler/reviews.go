package handler

import (
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"
)

func (a *API) Reviews(w http.ResponseWriter, r *http.Request) {
	gameID, e := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if e != nil || gameID <= 0 {
		fail(w, 400, "id ไม่ถูกต้อง")
		return
	}
	var published bool
	e = a.S.DB.QueryRow(r.Context(), `SELECT is_published FROM gm_games WHERE id=$1`, gameID).Scan(&published)
	if e != nil {
		dbError(w, e)
		return
	}
	if !published {
		fail(w, 404, "ไม่พบเกม")
		return
	}
	if r.Method == "GET" {
		rows, e := a.S.DB.Query(r.Context(), `SELECT r.id,r.member_id,m.display_name,r.rating,r.comment,r.created_at FROM gm_reviews r JOIN gm_members m ON m.id=r.member_id WHERE r.game_id=$1 ORDER BY r.created_at DESC,r.id DESC`, gameID)
		if e != nil {
			dbError(w, e)
			return
		}
		defer rows.Close()
		out := []review{}
		for rows.Next() {
			var item review
			if e = rows.Scan(&item.ID, &item.MemberID, &item.DisplayName, &item.Rating, &item.Comment, &item.CreatedAt); e != nil {
				dbError(w, e)
				return
			}
			out = append(out, item)
		}
		if e = rows.Err(); e != nil {
			dbError(w, e)
			return
		}
		send(w, 200, out)
		return
	}
	var reviewID int64
	if r.Method == "PUT" || r.Method == "DELETE" {
		reviewID, e = strconv.ParseInt(r.PathValue("reviewId"), 10, 64)
		if e != nil || reviewID <= 0 {
			fail(w, 400, "id รีวิวไม่ถูกต้อง")
			return
		}
	}
	if r.Method == "DELETE" {
		who := current(r)
		result, err := a.S.DB.Exec(r.Context(), `DELETE FROM gm_reviews WHERE id=$1 AND game_id=$2 AND (member_id=$3 OR $4)`, reviewID, gameID, who.ID, who.Role == "admin")
		if err != nil {
			dbError(w, err)
			return
		}
		if result.RowsAffected() == 0 {
			fail(w, 404, "ไม่พบรีวิวหรือไม่มีสิทธิ์ลบ")
			return
		}
		send(w, 200, map[string]bool{"ok": true})
		return
	}
	var b struct {
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
	}
	if !decode(w, r, &b) {
		return
	}
	b.Comment = strings.TrimSpace(b.Comment)
	if b.Rating < 1 || b.Rating > 5 || utf8.RuneCountInString(b.Comment) < 10 || len(b.Comment) > 2000 {
		fail(w, 400, "ให้คะแนน 1–5 ดาว และเขียนรีวิวอย่างน้อย 10 ตัวอักษร ไม่เกิน 2000 ไบต์")
		return
	}
	who := current(r)
	item := review{MemberID: who.ID, DisplayName: who.Name, Rating: b.Rating, Comment: b.Comment}
	status := 201
	if r.Method == "PUT" {
		status = 200
		e = a.S.DB.QueryRow(r.Context(), `UPDATE gm_reviews SET rating=$1,comment=$2 WHERE id=$3 AND game_id=$4 AND member_id=$5 RETURNING id,created_at`, b.Rating, b.Comment, reviewID, gameID, who.ID).Scan(&item.ID, &item.CreatedAt)
	} else {
		e = a.S.DB.QueryRow(r.Context(), `INSERT INTO gm_reviews(member_id,game_id,rating,comment) VALUES($1,$2,$3,$4) RETURNING id,created_at`, who.ID, gameID, b.Rating, b.Comment).Scan(&item.ID, &item.CreatedAt)
	}
	if e != nil {
		var pgErr *pgconn.PgError
		if errors.As(e, &pgErr) && pgErr.Code == "23505" {
			fail(w, 409, "คุณรีวิวเกมนี้ไปแล้ว")
			return
		}
		dbError(w, e)
		return
	}
	send(w, status, item)
}
