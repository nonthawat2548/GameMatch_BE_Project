package handler

import (
	"context"
	"encoding/json"
	"gamematch/store"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (a *API) SyncStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	send(w, http.StatusOK, map[string]bool{"steamConfigured": strings.TrimSpace(a.SteamKey) != ""})
}

func (a *API) Sync(w http.ResponseWriter, r *http.Request) {
	if strings.TrimSpace(a.RawgKey) == "" {
		fail(w, 503, "ยังไม่ได้ตั้ง RAWG_API_KEY บน Backend")
		return
	}
	if !a.SyncMu.TryLock() {
		fail(w, 409, "กำลัง Sync อยู่")
		return
	}
	defer a.SyncMu.Unlock()
	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()
	u := "https://api.rawg.io/api/games?" + url.Values{"key": {a.RawgKey}, "page_size": {"40"}, "ordering": {"-added"}}.Encode()
	req, e := http.NewRequestWithContext(ctx, "GET", u, nil)
	if e != nil {
		fail(w, 500, "สร้างคำขอไม่สำเร็จ")
		return
	}
	client := a.RawgClient
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	res, e := client.Do(req)
	if e != nil {
		fail(w, 502, "ติดต่อ RAWG ไม่สำเร็จ")
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		fail(w, 502, "RAWG ปฏิเสธคำขอ ตรวจ Key หรือ Quota")
		return
	}
	var data struct {
		Results []struct {
			ID     int64  `json:"id"`
			Name   string `json:"name"`
			Image  string `json:"background_image"`
			Genres []struct {
				Slug string `json:"slug"`
			} `json:"genres"`
			Tags []struct {
				Slug string `json:"slug"`
			} `json:"tags"`
			Platforms []struct {
				Platform struct {
					Name string `json:"name"`
				} `json:"platform"`
			} `json:"platforms"`
		} `json:"results"`
	}
	body, readErr := io.ReadAll(io.LimitReader(res.Body, (8<<20)+1))
	if readErr != nil || len(body) > 8<<20 || json.Unmarshal(body, &data) != nil || data.Results == nil || len(data.Results) > 40 {
		fail(w, 502, "RAWG JSON ไม่ถูกต้อง")
		return
	}
	games := []store.Game{}
	for _, v := range data.Results {
		if v.ID <= 0 || v.Name == "" {
			continue
		}
		genres, tags, platforms := []string{}, []string{}, []string{}
		for _, g := range v.Genres {
			genres = append(genres, g.Slug)
		}
		for _, t := range v.Tags {
			tags = append(tags, t.Slug)
		}
		for _, p := range v.Platforms {
			platforms = append(platforms, p.Platform.Name)
		}
		id := int64(v.ID)
		games = append(games, store.Game{RawgID: &id, Title: v.Name, ImageURL: v.Image, Genres: genres, Tags: tags, Platforms: platforms})
	}
	count, e := a.S.ImportRAWG(ctx, games)
	if e != nil {
		dbError(w, e)
		return
	}
	send(w, 200, map[string]any{"imported": count, "note": "Imports first 40 popular results. Existing admin description/publication flags preserved."})
}
