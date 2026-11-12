package handler

import (
	"gamematch/store"
	"net/http"
	"strconv"
	"strings"
)

func (a *API) Games(w http.ResponseWriter, r *http.Request) {
	g, e := a.S.Games(r.Context(), false)
	if e != nil {
		dbError(w, e)
		return
	}
	q := strings.ToLower(r.URL.Query().Get("q"))
	out := []store.Game{}
	for _, v := range g {
		if strings.Contains(strings.ToLower(v.Title), q) {
			out = append(out, v)
		}
	}
	send(w, 200, out)
}
func (a *API) Detail(w http.ResponseWriter, r *http.Request) {
	games, e := a.S.Games(r.Context(), false)
	if e != nil {
		dbError(w, e)
		return
	}
	for _, g := range games {
		if strconv.FormatInt(g.ID, 10) == r.PathValue("id") {
			send(w, 200, g)
			return
		}
	}
	fail(w, 404, "ไม่พบเกม")
}
