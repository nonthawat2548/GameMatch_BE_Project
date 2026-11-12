package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type steamMovie struct {
	ID int64 `json:"id"`
	Name string `json:"name"`
	MP4 map[string]string `json:"mp4"`
	WebM map[string]string `json:"webm"`
	HLS string `json:"hls_h264"`
	DASH string `json:"dash_h264"`
	DASHAV1 string `json:"dash_av1"`
	Highlight bool `json:"highlight"`
}

type Trailer struct {
	ID int64 `json:"id"`
	Title string `json:"title"`
	Sources []string `json:"sources"`
	Preview bool `json:"preview"`
}

// Only allow Steam-controlled video hosts. Never proxy arbitrary URLs.
func steamVideoURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.User != nil || (u.Scheme != "http" && u.Scheme != "https") ||
		u.Port() != "" || !strings.HasSuffix(u.Hostname(), ".steamstatic.com") ||
		!strings.HasPrefix(u.Path, "/store_trailers/") {
		return ""
	}
	u.Scheme = "https"
	return u.String()
}

var adaptiveTrailerPath = regexp.MustCompile(`^/store_trailers/([0-9]+)/[0-9]+/[a-fA-F0-9]{40}/[0-9]+/(hls_264_master\.m3u8|dash_h264\.mpd|dash_av1\.mpd)$`)

func movieSources(movie steamMovie, appID int64) Trailer {
	t := Trailer{ID: movie.ID, Title: movie.Name, Sources: []string{}}
	if strings.TrimSpace(t.Title) == "" { t.Title = "ตัวอย่างเกม" }
	seen := map[string]bool{}
	add := func(value string) {
		if safe := steamVideoURL(value); safe != "" && !seen[safe] {
			t.Sources = append(t.Sources, safe)
			seen[safe] = true
		}
	}
	// Progressive files play directly in HTML video, including full trailers where available.
	for _, value := range []string{movie.MP4["max"],movie.WebM["max"],movie.MP4["480"],movie.WebM["480"]} {
		if u, err := url.Parse(value); err == nil && (strings.HasSuffix(u.Path,".mp4") || strings.HasSuffix(u.Path,".webm")) { add(value) }
	}
	if len(t.Sources)>0 { return t }
	// Modern Steam trailers are adaptive streams. Their official short preview
	// uses the same asset directory (no HLS/DASH player dependency).
	for _, value := range []string{movie.HLS,movie.DASH,movie.DASHAV1} {
		safe := steamVideoURL(value)
		if safe == "" { continue }
		u, _ := url.Parse(safe)
		match := adaptiveTrailerPath.FindStringSubmatch(u.Path)
		if len(match)!=3 || match[1]!=strconv.FormatInt(appID,10) { continue }
		u.Path = strings.TrimSuffix(u.Path,match[2])+"microtrailer.mp4"
		u.RawPath = ""
		u.RawQuery = ""
		add(u.String())
		u.Path = strings.TrimSuffix(u.Path,".mp4")+".webm"
		add(u.String())
		t.Preview = true
		break
	}
	return t
}

func (a *API) GameTrailers(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"),10,64)
	if err != nil || id <= 0 { fail(w,400,"รหัสเกมไม่ถูกต้อง"); return }
	appID, err := a.S.PublishedGameSteamID(r.Context(),id)
	if err != nil { dbError(w,err); return }
	out := []Trailer{}
	if appID == nil || *appID <= 0 { send(w,200,out); return }
	ctx,cancel := context.WithTimeout(r.Context(),8*time.Second)
	defer cancel()
	client := a.SteamClient
	if client == nil { client = &http.Client{Timeout:8*time.Second} }
	key := strconv.FormatInt(*appID,10)
	data,err := steamFetch(ctx,client,"/api/appdetails?appids="+key+"&cc=th&l=english&filters=movies")
	var result map[string]struct {
		Success bool `json:"success"`
		Data struct { Movies []steamMovie `json:"movies"` } `json:"data"`
	}
	if err != nil || json.Unmarshal(data,&result)!=nil { fail(w,502,"โหลดคลิปจาก Steam ไม่สำเร็จ ลองใหม่ได้"); return }
	entry,ok := result[key]
	if !ok || !entry.Success { fail(w,502,"Steam ยังไม่ส่งข้อมูลคลิปเกมนี้"); return }
	// Prefer gameplay explicitly labeled by the publisher, then highlighted videos.
	sort.SliceStable(entry.Data.Movies,func(i,j int)bool{
		rank:=func(m steamMovie)int{
			if strings.Contains(strings.ToLower(m.Name),"gameplay"){return 0}
			if m.Highlight{return 1}
			return 2
		}
		return rank(entry.Data.Movies[i])<rank(entry.Data.Movies[j])
	})
	for _, movie := range entry.Data.Movies {
		t := movieSources(movie,*appID)
		if len(t.Sources)>0 { out=append(out,t) }
		if len(out)>=8 { break }
	}
	w.Header().Set("Cache-Control","public, max-age=600")
	send(w,200,out)
}
