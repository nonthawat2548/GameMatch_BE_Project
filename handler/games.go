package handler

import (
	"net/http"
	"strconv"
)

func (a *API) Games(w http.ResponseWriter, r *http.Request) {
	g, e := a.S.SearchGames(r.Context(), r.URL.Query().Get("q"))
	if e != nil {
		dbError(w, e)
		return
	}
	send(w, 200, g)
}
func (a *API) Detail(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if e != nil || id <= 0 {
		fail(w, 400, "id ไม่ถูกต้อง")
		return
	}
	g, e := a.S.PublishedGame(r.Context(), id)
	if e != nil {
		dbError(w, e)
		return
	}
	send(w, 200, g)
}
