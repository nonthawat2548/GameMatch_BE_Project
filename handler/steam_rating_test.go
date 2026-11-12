package handler

import "testing"

func TestSteamReviewSummary(t *testing.T) {
	s,err := parseSteamReviewSummary([]byte(`{"success":1,"query_summary":{"total_positive":87,"total_negative":14,"total_reviews":101,"review_score":8,"review_score_desc":"Very Positive"}}`))
	if err!=nil || s.PositivePercent==nil || *s.PositivePercent!=86 || s.Total!=101 || s.Score!=8 { t.Fatalf("summary: %#v, %v",s,err) }
	empty,err := parseSteamReviewSummary([]byte(`{"success":1,"query_summary":{"total_positive":0,"total_negative":0,"total_reviews":0,"review_score":0,"review_score_desc":"No user reviews"}}`))
	if err!=nil || empty.PositivePercent!=nil || empty.Total!=0 { t.Fatalf("empty: %#v %v",empty,err) }
	for _,body:=range []string{`{"success":0}`,`{"success":1,"query_summary":{}}`,`{"success":1,"query_summary":{"total_positive":5,"total_negative":1,"total_reviews":2}}`} {
		if _,err:=parseSteamReviewSummary([]byte(body));err==nil { t.Fatal("accepted invalid response") }
	}
}
