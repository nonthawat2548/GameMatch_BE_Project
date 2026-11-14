package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Steam's own review summary, not a GameMatch rating or a critic score.
// https://partner.steamgames.com/doc/store/getreviews
type SteamReviewSummary struct {
	Positive        int64  `json:"positive"`
	Negative        int64  `json:"negative"`
	Total           int64  `json:"total"`
	Score           int    `json:"score"`
	Description     string `json:"description"`
	PositivePercent *int   `json:"positivePercent"`
}

type steamReviewCounts struct {
	Positive    *int64 `json:"total_positive"`
	Negative    *int64 `json:"total_negative"`
	Total       *int64 `json:"total_reviews"`
	Score       int    `json:"review_score"`
	Description string `json:"review_score_desc"`
}

func parseSteamReviewSummary(data []byte) (SteamReviewSummary, error) {
	var envelope struct {
		Response *struct {
			Summary *steamReviewCounts `json:"query_summary"`
		} `json:"response"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil || envelope.Response == nil || envelope.Response.Summary == nil {
		return SteamReviewSummary{}, fmt.Errorf("invalid Steam review response")
	}
	s := envelope.Response.Summary
	if s.Positive == nil || s.Negative == nil || s.Total == nil ||
		*s.Positive < 0 || *s.Negative < 0 || *s.Total < 0 ||
		*s.Total > (1<<63-1)/100 || *s.Positive > *s.Total || *s.Negative != *s.Total-*s.Positive ||
		s.Score < 0 || s.Score > 9 {
		return SteamReviewSummary{}, fmt.Errorf("invalid Steam review counts")
	}
	out := SteamReviewSummary{Positive: *s.Positive, Negative: *s.Negative, Total: *s.Total, Score: s.Score, Description: s.Description}
	if out.Total > 0 {
		// Whole percent rounded down, as commonly displayed by Steam.
		percent := int(out.Positive * 100 / out.Total)
		out.PositivePercent = &percent
	}
	return out, nil
}

func (a *API) SteamRating(w http.ResponseWriter, r *http.Request) {
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
	// Match Steam's default purchase filter; all languages and all review types.
	// Only request one review, and return summary counts (never user identities/text).
	data, err := fetchSteamReviewSummary(ctx, client, *appID)
	if err != nil {
		fail(w, 502, "โหลดคะแนนจาก Steam ไม่สำเร็จ")
		return
	}
	summary, err := parseSteamReviewSummary(data)
	if err != nil {
		fail(w, 502, "Steam ยังไม่ส่งข้อมูลคะแนนที่สมบูรณ์")
		return
	}
	w.Header().Set("Cache-Control", "private, max-age=600")
	send(w, 200, summary)
}

// GetAppReviews replaces the deprecated store /appreviews endpoint.
// No publisher key is needed for public review summaries.
func fetchSteamReviewSummary(ctx context.Context, client *http.Client, appID int64) ([]byte, error) {
	input, err := json.Marshal(map[string]any{
		"appid": appID, "filter": 1, "languages": []string{"all"},
		"review_type": 0, "purchase_type": 0, "num_per_page": 1,
		"filter_offtopic_activity": true, "display_language": "english",
	})
	if err != nil {
		return nil, err
	}
	q := url.Values{"input_json": []string{string(input)}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.steampowered.com/IUserReviewsService/GetAppReviews/v1/?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "GameMatch/1.0 (student game catalog)")
	req.Header.Set("Accept", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Steam review status %d", res.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, (4<<20)+1))
	if err != nil || len(data) > 4<<20 {
		return nil, fmt.Errorf("invalid Steam review response size")
	}
	return data, nil
}
