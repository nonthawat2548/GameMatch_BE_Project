package handler

import (
	"gamematch/store"
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
	published, e := a.S.GamePublished(r.Context(), gameID)
	if e != nil {
		dbError(w, e)
		return
	}
	if !published {
		fail(w, 404, "ไม่พบเกม")
		return
	}
	if r.Method == "GET" {
		out, e := a.S.ListReviews(r.Context(), gameID)
		if e != nil {
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
		deleted, err := a.S.DeleteReview(r.Context(), reviewID, gameID, who.ID, who.Role == "admin")
		if err != nil {
			dbError(w, err)
			return
		}
		if !deleted {
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
	item := store.Review{MemberID: who.ID, DisplayName: who.Name, Rating: b.Rating, Comment: b.Comment}
	status := 201
	if r.Method == "PUT" {
		status = 200
		item, e = a.S.UpdateReview(r.Context(), reviewID, gameID, item)
	} else {
		item, e = a.S.AddReview(r.Context(), gameID, item)
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
