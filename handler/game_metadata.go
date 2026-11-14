package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type screenshot struct {
	Thumbnail string `json:"path_thumbnail"`
	Full      string `json:"path_full"`
}
type requirements struct {
	Minimum     string `json:"minimum"`
	Recommended string `json:"recommended"`
}
type storePrice struct {
	Currency         string `json:"currency"`
	Initial          int64  `json:"initial"`
	Final            int64  `json:"final"`
	Discount         int    `json:"discount_percent"`
	InitialFormatted string `json:"initial_formatted"`
	FinalFormatted   string `json:"final_formatted"`
}
type gameMetadata struct {
	Developers  []string        `json:"developers"`
	Publishers  []string        `json:"publishers"`
	IsFree      *bool           `json:"is_free"`
	Price       *storePrice     `json:"price_overview"`
	Screenshots []screenshot    `json:"screenshots"`
	Header      string          `json:"header_image"`
	Backdrop    string          `json:"background_raw"`
	PC          json.RawMessage `json:"pc_requirements"`
	Mac         json.RawMessage `json:"mac_requirements"`
	Linux       json.RawMessage `json:"linux_requirements"`
	Release     struct {
		Date string `json:"date"`
		Soon bool   `json:"coming_soon"`
	} `json:"release_date"`
}

func steamImageURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.User != nil || u.Port() != "" || (u.Scheme != "https" && u.Scheme != "http") {
		return ""
	}
	host := u.Hostname()
	if !strings.HasSuffix(host, ".steamstatic.com") && !strings.HasSuffix(host, ".akamaihd.net") {
		return ""
	}
	u.Scheme = "https"
	return u.String()
}

var reqBreaks = regexp.MustCompile("(?i)<(?:br\\s*/?|/li|/p|/ul)>")
var reqTags = regexp.MustCompile("<[^>]*>")

func plainRequirements(s string) string {
	s = reqBreaks.ReplaceAllString(s, "\n")
	s = html.UnescapeString(reqTags.ReplaceAllString(s, ""))
	lines := []string{}
	for _, line := range strings.Split(s, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}
func requirementText(raw json.RawMessage) requirements {
	var r requirements
	if json.Unmarshal(raw, &r) != nil {
		return requirements{}
	}
	r.Minimum = plainRequirements(r.Minimum)
	r.Recommended = plainRequirements(r.Recommended)
	return r
}
func parseGameMetadata(data []byte, key string) (gameMetadata, error) {
	var response map[string]struct {
		Success bool          `json:"success"`
		Data    *gameMetadata `json:"data"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return gameMetadata{}, err
	}
	app, ok := response[key]
	if !ok || !app.Success || app.Data == nil {
		return gameMetadata{}, fmt.Errorf("missing Steam metadata")
	}
	return *app.Data, nil
}
func (a *API) GameMetadata(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		fail(w, 400, "รหัสเกมไม่ถูกต้อง")
		return
	}
	appID, err := a.S.PublishedGameSteamID(r.Context(), id)
	if err != nil {
		dbError(w, err)
		return
	}
	if appID == nil || *appID <= 0 {
		send(w, 200, nil)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()
	client := a.SteamClient
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}
	key := strconv.FormatInt(*appID, 10)
	data, err := steamFetch(ctx, client, "/api/appdetails?appids="+key+"&cc=th&l=english")
	if err != nil {
		fail(w, 502, "โหลดข้อมูล Steam ไม่สำเร็จ")
		return
	}
	m, err := parseGameMetadata(data, key)
	if err != nil {
		fail(w, 502, "ข้อมูล Steam ไม่พร้อม")
		return
	}
	shots := []screenshot{}
	for _, s := range m.Screenshots {
		s.Thumbnail = steamImageURL(s.Thumbnail)
		s.Full = steamImageURL(s.Full)
		if s.Full != "" {
			if s.Thumbnail == "" {
				s.Thumbnail = s.Full
			}
			shots = append(shots, s)
		}
		if len(shots) >= 24 {
			break
		}
	}
	specs := map[string]requirements{}
	for name, raw := range map[string]json.RawMessage{"Windows": m.PC, "macOS": m.Mac, "Linux": m.Linux} {
		if req := requirementText(raw); req.Minimum != "" || req.Recommended != "" {
			specs[name] = req
		}
	}
	if m.Price != nil && (m.Price.Currency != "THB" || m.Price.Final < 0 || m.Price.Initial < m.Price.Final || m.Price.Discount < 0 || m.Price.Discount > 100) {
		m.Price = nil
	}
	// Keep the response field, but do not invent a date absent from the source data.
	major := ""
	w.Header().Set("Cache-Control", "private, max-age=600")
	send(w, 200, map[string]any{"developers": m.Developers, "publishers": m.Publishers, "isFree": m.IsFree, "price": m.Price, "screenshots": shots, "headerImage": steamImageURL(m.Header), "backdrop": steamImageURL(m.Backdrop), "requirements": specs, "releaseDate": steamDate(m.Release.Date), "majorUpdateDate": major, "comingSoon": m.Release.Soon, "checkedAt": time.Now().UTC()})
}
