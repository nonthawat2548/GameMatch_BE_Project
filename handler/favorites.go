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
	ids, e := a.S.FavoriteIDs(r.Context(), current(r).ID)
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
		e = a.S.RemoveFavorite(r.Context(), current(r).ID, id)
	} else {
		exists, err := a.S.PublishedGameExists(r.Context(), id)
		e = err
		if e != nil {
			dbError(w, e)
			return
		}
		if !exists {
			fail(w, 404, "ไม่พบเกม")
			return
		}
		e = a.S.AddFavorite(r.Context(), current(r).ID, id)
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
		genre, mode, style, e := a.S.MemberPreferences(r.Context(), current(r).ID)
		if e == nil { p.Genre, p.Mode, p.Style = genre, mode, style }
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
	e := a.S.SavePreferences(r.Context(), current(r).ID, p.Genre, p.Mode, p.Style)
	if e != nil {
		dbError(w, e)
		return
	}
	send(w, 200, p)
}
