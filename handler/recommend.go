package handler

import (
	"gamematch/recommend"
	"gamematch/store"
	"net/http"
	"strconv"
)

func (a *API) Recommend(w http.ResponseWriter, r *http.Request) {
	var p recommend.Preference
	if !decode(w, r, &p) {
		return
	}
	p = recommend.Normalize(p)
	if !recommend.Valid(p) {
		fail(w, 400, "ตัวเลือกไม่ถูกต้อง")
		return
	}
	g, e := a.S.Games(r.Context(), false)
	if e != nil {
		dbError(w, e)
		return
	}
	send(w, 200, recommend.Rank(g, p))
}
func (a *API) SimilarGames(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		fail(w, 400, "id ไม่ถูกต้อง")
		return
	}
	games, err := a.S.Games(r.Context(), false)
	if err != nil {
		dbError(w, err)
		return
	}
	for _, g := range games {
		if g.ID == id {
			send(w, 200, recommend.Similar(games, []store.Game{g}))
			return
		}
	}
	fail(w, 404, "ไม่พบเกม")
}
