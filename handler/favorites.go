package handler

import (
	"github.com/jackc/pgx/v5"
	"errors"
	"gamematch/recommend"
	"gamematch/store"
	"net/http"
	"strconv"
)

func (a *API) Favorites(w http.ResponseWriter, r *http.Request) {
	rows, e := a.S.DB.Query(r.Context(), `SELECT game_id FROM gm_favorites WHERE member_id=$1`, current(r).ID)
	if e != nil {
		dbError(w, e)
		return
	}
	ids := map[int64]bool{}
	for rows.Next() {
		var id int64
		if e = rows.Scan(&id); e != nil {
			rows.Close()
			dbError(w, e)
			return
		}
		ids[id] = true
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		dbError(w, e)
		return
	}
	games, e := a.S.Games(r.Context(), false)
	if e != nil {
		dbError(w, e)
		return
	}
	out := []store.Game{}
	for _, g := range games {
		if ids[g.ID] {
			out = append(out, g)
		}
	}
	if r.URL.Path == "/api/me/recommendations" {
		send(w, 200, recommend.Similar(games, out))
		return
	}
	send(w, 200, out)
}
func (a *API) Favorite(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if e != nil || id <= 0 {
		fail(w, 400, "id ไม่ถูกต้อง")
		return
	}
	if r.Method == "DELETE" {
		_, e = a.S.DB.Exec(r.Context(), `DELETE FROM gm_favorites WHERE member_id=$1 AND game_id=$2`, current(r).ID, id)
	} else {
		var exists bool
		e = a.S.DB.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM gm_games WHERE id=$1 AND is_published)`, id).Scan(&exists)
		if e != nil {
			dbError(w, e)
			return
		}
		if !exists {
			fail(w, 404, "ไม่พบเกม")
			return
		}
		_, e = a.S.DB.Exec(r.Context(), `INSERT INTO gm_favorites VALUES($1,$2) ON CONFLICT DO NOTHING`, current(r).ID, id)
	}
	if e != nil {
		dbError(w, e)
		return
	}
	send(w, 200, map[string]bool{"ok": true})
}
func (a *API) Preferences(w http.ResponseWriter, r *http.Request) {
	p := recommend.Preference{Genre: "rpg", Mode: "solo", Style: "story"}
	if r.Method == "GET" {
		e := a.S.DB.QueryRow(r.Context(), `SELECT genre,mode,style FROM gm_preferences WHERE member_id=$1`, current(r).ID).Scan(&p.Genre, &p.Mode, &p.Style)
		if e != nil && !errors.Is(e, pgx.ErrNoRows) {
			dbError(w, e)
			return
		}
		send(w, 200, p)
		return
	}
	if !decode(w, r, &p) {
		return
	}
	if !recommend.Valid(p) {
		fail(w, 400, "ตัวเลือกไม่ถูกต้อง")
		return
	}
	_, e := a.S.DB.Exec(r.Context(), `INSERT INTO gm_preferences VALUES($1,$2,$3,$4) ON CONFLICT(member_id) DO UPDATE SET genre=$2,mode=$3,style=$4`, current(r).ID, p.Genre, p.Mode, p.Style)
	if e != nil {
		dbError(w, e)
		return
	}
	send(w, 200, p)
}
