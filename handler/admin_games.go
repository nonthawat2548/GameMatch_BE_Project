package handler

import (
	"gamematch/store"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (a *API) AdminGames(w http.ResponseWriter, r *http.Request) {
	g, e := a.S.Games(r.Context(), true)
	if e != nil {
		dbError(w, e)
		return
	}
	send(w, 200, g)
}
func (a *API) SaveGame(w http.ResponseWriter, r *http.Request) {
	var g store.Game
	if !decode(w, r, &g) {
		return
	}
	if r.Method == "PUT" {
		var e error
		g.ID, e = strconv.ParseInt(r.PathValue("id"), 10, 64)
		if e != nil || g.ID <= 0 {
			fail(w, 400, "id ไม่ถูกต้อง")
			return
		}
	} else {
		g.ID = 0
		g.RawgID = nil
	}
	g.Title = strings.TrimSpace(g.Title)
	if g.Title == "" || len(g.Title) > 250 {
		fail(w, 400, "กรุณาระบุชื่อเกม")
		return
	}
	if g.ImageURL != "" {
		u, e := url.Parse(g.ImageURL)
		if e != nil || u.Scheme != "https" || u.Host == "" {
			fail(w, 400, "รูปต้องเป็น HTTPS URL")
			return
		}
	}
	id, e := a.S.Save(r.Context(), g)
	if e != nil {
		dbError(w, e)
		return
	}
	send(w, 200, map[string]int64{"id": id})
}
func (a *API) DeleteGame(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if e != nil || id <= 0 {
		fail(w, 400, "id ไม่ถูกต้อง")
		return
	}
	result, e := a.S.DB.Exec(r.Context(), `DELETE FROM gm_games WHERE id=$1`, id)
	if e != nil {
		dbError(w, e)
		return
	}
	if result.RowsAffected() == 0 {
		fail(w, 404, "ไม่พบเกม")
		return
	}
	send(w, 200, map[string]bool{"ok": true})
}

// SyncStatus reports configuration presence only; it never returns the key.
