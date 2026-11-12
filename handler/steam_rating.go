package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// Steam's own review summary, not a GameMatch rating or a critic score.
// https://partner.steamgames.com/doc/store/getreviews
type SteamReviewSummary struct {
	Positive int64 `json:"positive"`
	Negative int64 `json:"negative"`
	Total int64 `json:"total"`
	Score int `json:"score"`
	Description string `json:"description"`
	PositivePercent *int `json:"positivePercent"`
}

func parseSteamReviewSummary(data []byte) (SteamReviewSummary, error) {
	var response struct {
		Success int `json:"success"`
		Summary *struct {
			Positive *int64 `json:"total_positive"`
			Negative *int64 `json:"total_negative"`
			Total *int64 `json:"total_reviews"`
			Score int `json:"review_score"`
			Description string `json:"review_score_desc"`
		} `json:"query_summary"`
	}
	if err := json.Unmarshal(data, &response); err != nil || response.Success != 1 || response.Summary == nil {
		return SteamReviewSummary{}, fmt.Errorf("invalid Steam review response")
	}
	s := response.Summary
	if s.Positive == nil || s.Negative == nil || s.Total == nil ||
		*s.Positive < 0 || *s.Negative < 0 || *s.Total < 0 ||
		*s.Total > (1<<63-1)/100 || *s.Positive > *s.Total || *s.Negative != *s.Total - *s.Positive ||
		s.Score < 0 || s.Score > 9 {
		return SteamReviewSummary{}, fmt.Errorf("invalid Steam review counts")
	}
	out := SteamReviewSummary{Positive:*s.Positive, Negative:*s.Negative, Total:*s.Total, Score:s.Score, Description:s.Description}
	if out.Total > 0 {
		// Whole percent rounded down, as commonly displayed by Steam.
		percent := int(out.Positive * 100 / out.Total)
		out.PositivePercent = &percent
	}
	return out, nil
}

func (a *API) SteamRating(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"),10,64)
	if err != nil || id <= 0 { fail(w,400,"รหัสเกมไม่ถูกต้อง"); return }
	appID, err := a.S.PublishedGameSteamID(r.Context(), id)
	if err != nil { dbError(w,err); return }
	if appID == nil || *appID <= 0 { send(w,200,nil); return }
	ctx,cancel := context.WithTimeout(r.Context(),8*time.Second)
	defer cancel()
	client := a.SteamClient
	if client == nil { client = &http.Client{Timeout:8*time.Second} }
	// Match Steam's default purchase filter; all languages and all review types.
	// Only request one review, and return summary counts (never user identities/text).
	data,err := steamFetch(ctx,client,"/appreviews/"+strconv.FormatInt(*appID,10)+"?json=1&filter=all&language=all&review_type=all&purchase_type=steam&num_per_page=1&cursor=%2A")
	if err != nil { fail(w,502,"โหลดคะแนนจาก Steam ไม่สำเร็จ"); return }
	summary,err := parseSteamReviewSummary(data)
	if err != nil { fail(w,502,"Steam ยังไม่ส่งข้อมูลคะแนนที่สมบูรณ์"); return }
	w.Header().Set("Cache-Control","private, max-age=600")
	send(w,200,summary)
}
