package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"gamematch/store"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Catalog IDs come from the keyed Steam Web API; details and tags use public Store endpoints.
// A failed fetch/parse leaves the previous catalog intact.
const steamStore = "https://store.steampowered.com"
const steamLimit = 12

var steamTagPattern = regexp.MustCompile(`(?s)<a[^>]*class="app_tag"[^>]*>(.*?)</a>`)
var htmlTagPattern = regexp.MustCompile(`<[^>]*>`)

type steamDetail struct {
	Type        string `json:"type"`
	ID          int64  `json:"steam_appid"`
	Name        string `json:"name"`
	Description string `json:"short_description"`
	Image       string `json:"header_image"`
	Release     struct {
		Soon bool   `json:"coming_soon"`
		Date string `json:"date"`
	} `json:"release_date"`
	Genres []struct {
		Description string `json:"description"`
	} `json:"genres"`
	Categories []struct {
		Description string `json:"description"`
	} `json:"categories"`
	Platforms struct {
		Windows bool `json:"windows"`
		Mac     bool `json:"mac"`
		Linux   bool `json:"linux"`
	} `json:"platforms"`
}

func steamFetch(ctx context.Context, client *http.Client, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", steamStore+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "GameMatch/1.0 (student game catalog)")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Steam status %d", res.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, (4<<20)+1))
	if err != nil || len(data) > 4<<20 {
		return nil, fmt.Errorf("invalid Steam response size")
	}
	return data, nil
}

func steamDate(value string) string {
	for _, format := range []string{"Jan 2, 2006", "2 Jan, 2006", "2 Jan 2006", "2006-01-02"} {
		if date, err := time.Parse(format, strings.TrimSpace(value)); err == nil {
			return date.Format("2006-01-02")
		}
	}
	return ""
}

func steamSlug(value string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(value)), " ", "-")
}

func steamGame(detail steamDetail, page []byte) (store.Game, bool) {
	date := steamDate(detail.Release.Date)
	image, err := url.Parse(detail.Image)
	if detail.Type != "game" || detail.ID <= 0 || detail.Release.Soon || date == "" || date > time.Now().UTC().Format("2006-01-02") || strings.TrimSpace(detail.Name) == "" || len(detail.Name) > 250 || err != nil || image.Scheme != "https" || !strings.HasSuffix(image.Hostname(), ".steamstatic.com") {
		return store.Game{}, false
	}
	tags := []string{}
	seen := map[string]bool{}
	add := func(value string) {
		value = steamSlug(value)
		if value != "" && !seen[value] {
			seen[value] = true
			tags = append(tags, value)
		}
	}
	for _, match := range steamTagPattern.FindAllSubmatch(page, 40) {
		value := html.UnescapeString(strings.TrimSpace(htmlTagPattern.ReplaceAllString(string(match[1]), "")))
		if value == "Sexual Content" || value == "Nudity" || value == "Hentai" || value == "Adult Only Sexual Content" {
			return store.Game{}, false
		}
		add(value)
	}
	// A missing/changed tag page must not silently become invented preference metadata.
	if len(tags) == 0 {
		return store.Game{}, false
	}
	genres, platforms := []string{}, []string{}
	for _, g := range detail.Genres {
		switch g.Description {
		case "Utilities", "Animation & Modeling", "Audio Production", "Design & Illustration", "Photo Editing", "Software Training", "Video Production", "Web Publishing", "Game Development":
			return store.Game{}, false
		}
		genres = append(genres, steamSlug(g.Description))
	}
	for _, category := range detail.Categories {
		value := strings.ToLower(category.Description)
		if value == "single-player" {
			add("singleplayer")
		}
		if strings.Contains(value, "co-op") {
			add("co-op")
		}
		if strings.Contains(value, "pvp") {
			add("pvp")
		}
	}
	if detail.Platforms.Windows {
		platforms = append(platforms, "Windows")
	}
	if detail.Platforms.Mac {
		platforms = append(platforms, "macOS")
	}
	if detail.Platforms.Linux {
		platforms = append(platforms, "Linux")
	}
	return store.Game{SteamID: &detail.ID, Title: detail.Name, Description: html.UnescapeString(htmlTagPattern.ReplaceAllString(detail.Description, "")), ImageURL: detail.Image, Genres: genres, Tags: tags, Platforms: platforms, Published: true, ReleaseDate: date}, true
}

type steamApp struct {
	ID int64 `json:"appid"`
}

// Steam Web API credentials are sent only to this fixed HTTPS host, never to the Store.
func steamAppList(ctx context.Context, client *http.Client, key string, after uint32) ([]steamApp, bool, error) {
	input, _ := json.Marshal(map[string]any{
		"last_appid": after, "max_results": steamLimit,
		"include_games": true, "include_dlc": false, "include_software": false,
		"include_videos": false, "include_hardware": false,
	})
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.steampowered.com/IStoreService/GetAppList/v1/?"+url.Values{"input_json": {string(input)}}.Encode(), nil)
	if err != nil {
		return nil, false, fmt.Errorf("cannot create Steam API request")
	}
	req.Header.Set("x-webapi-key", strings.TrimSpace(key))
	webClient := *client
	webClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	res, err := webClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("Steam Web API unavailable")
	}
	defer res.Body.Close()
	if res.StatusCode == 401 || res.StatusCode == 403 {
		return nil, false, fmt.Errorf("Steam ปฏิเสธ API key กรุณาตรวจสอบ STEAM_API_KEY")
	}
	if res.StatusCode != 200 {
		return nil, false, fmt.Errorf("Steam Web API ตอบกลับสถานะ %d ลองใหม่ภายหลัง", res.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, (1<<20)+1))
	var result struct {
		Response struct {
			Apps []steamApp `json:"apps"`
			More bool       `json:"have_more_results"`
		} `json:"response"`
	}
	if err != nil || len(data) > 1<<20 || json.Unmarshal(data, &result) != nil || result.Response.Apps == nil || len(result.Response.Apps) > steamLimit {
		return nil, false, fmt.Errorf("รูปแบบรายชื่อเกมจาก Steam Web API ไม่ถูกต้อง")
	}
	previous := int64(after)
	for _, app := range result.Response.Apps {
		if app.ID <= previous || app.ID > int64(^uint32(0)) {
			return nil, false, fmt.Errorf("ลำดับ AppID จาก Steam Web API ไม่ถูกต้อง")
		}
		previous = app.ID
	}
	if result.Response.More && len(result.Response.Apps) == 0 {
		return nil, false, fmt.Errorf("Steam Web API ไม่ส่งข้อมูลสำหรับหน้าถัดไป")
	}
	return result.Response.Apps, result.Response.More, nil
}

func (a *API) SyncSteam(w http.ResponseWriter, r *http.Request) {
	if strings.TrimSpace(a.SteamKey) == "" {
		fail(w, 503, "ยังไม่ได้ตั้ง STEAM_API_KEY บน Backend")
		return
	}
	var after uint32
	if value := r.URL.Query().Get("after"); value != "" {
		parsed, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			fail(w, 400, "after ต้องเป็น AppID จำนวนเต็มบวกหรือศูนย์")
			return
		}
		after = uint32(parsed)
	}
	if !a.SyncMu.TryLock() {
		fail(w, 409, "กำลังอัปเดตข้อมูลอยู่")
		return
	}
	defer a.SyncMu.Unlock()
	ctx, cancel := context.WithTimeout(r.Context(), 28*time.Second)
	defer cancel()
	client := a.SteamClient
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}
	items, more, err := steamAppList(ctx, client, a.SteamKey, after)
	if err != nil {
		fail(w, 502, err.Error())
		return
	}
	games := []store.Game{}
	next := after
	for _, item := range items {
		next = uint32(item.ID)
		id := strconv.FormatInt(item.ID, 10)
		data, err := steamFetch(ctx, client, "/api/appdetails?appids="+id+"&cc=th&l=english")
		var response map[string]struct {
			Success bool        `json:"success"`
			Data    steamDetail `json:"data"`
		}
		if err != nil || json.Unmarshal(data, &response) != nil {
			fail(w, 502, "โหลดรายละเอียดเกมจาก Steam ไม่สำเร็จ ข้อมูลเดิมยังอยู่")
			return
		}
		entry, ok := response[id]
		if !ok {
			fail(w, 502, "Steam ยังไม่ส่งข้อมูลเกมครบ ข้อมูลเดิมยังอยู่")
			return
		}
		if !entry.Success {
			continue
		}
		if entry.Data.ID != item.ID || entry.Data.Type != "game" || entry.Data.Release.Soon {
			continue
		}
		page, err := steamFetch(ctx, client, "/app/"+id+"/?l=english&cc=th")
		if err != nil {
			fail(w, 502, "โหลดแท็ก Steam ไม่สำเร็จ ข้อมูลเดิมยังอยู่")
			return
		}
		if game, valid := steamGame(entry.Data, page); valid {
			games = append(games, game)
		}
	}
	if len(games) == 0 {
		send(w, 200, map[string]any{"imported": 0, "nextAppId": next, "hasMore": more, "note": "ไม่มีเกมที่ผ่านเงื่อนไขในชุดนี้"})
		return
	}
	tx, err := a.S.DB.Begin(ctx)
	if err != nil {
		dbError(w, err)
		return
	}
	defer tx.Rollback(ctx)
	for i := len(games) - 1; i >= 0; i-- {
		g := games[i]
		_, err = tx.Exec(ctx, `INSERT INTO gm_games(steam_id,title,description,image_url,genres,tags,platforms,release_date,synced_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8::date,now()) ON CONFLICT(steam_id) DO UPDATE SET title=EXCLUDED.title,image_url=EXCLUDED.image_url,genres=EXCLUDED.genres,tags=EXCLUDED.tags,platforms=EXCLUDED.platforms,release_date=EXCLUDED.release_date,synced_at=EXCLUDED.synced_at`, g.SteamID, g.Title, g.Description, g.ImageURL, g.Genres, g.Tags, g.Platforms, g.ReleaseDate)
		if err != nil {
			dbError(w, err)
			return
		}
	}
	if err = tx.Commit(ctx); err != nil {
		dbError(w, err)
		return
	}
	send(w, 200, map[string]any{"imported": len(games), "nextAppId": next, "hasMore": more, "note": "นำเข้าจาก Steam Web API และ Steam Store แล้ว คงคำอธิบายและสถานะเผยแพร่ที่ผู้ดูแลแก้ไว้"})
}
