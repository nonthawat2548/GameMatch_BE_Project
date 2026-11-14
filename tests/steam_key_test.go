package tests

import (
	"encoding/json"
	"gamematch/handler"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSteamMissingKeyAndInvalidCursor(t *testing.T) {
	for _, tc := range []struct {
		key, query string
		status     int
	}{
		{"", "", 503}, {"  ", "", 503}, {"test-key", "?after=-1", 400},
		{"test-key", "?after=abc", 400}, {"test-key", "?after=4294967296", 400},
	} {
		a := &handler.API{SteamKey: tc.key, SteamClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			t.Fatal("invalid config reached upstream")
			return nil, nil
		})}}
		w := httptest.NewRecorder()
		a.SyncSteam(w, httptest.NewRequest("POST", "/api/admin/sync/steam"+tc.query, nil))
		if w.Code != tc.status {
			t.Fatalf("want %d, got %d", tc.status, w.Code)
		}
	}
}

func TestSteamKeyIsolationAndCursorOnSkippedBatch(t *testing.T) {
	apiCalls, storeCalls := 0, 0
	a := &handler.API{SteamKey: " test-private-key "}
	a.SteamClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		body := ""
		if r.URL.Host == "api.steampowered.com" {
			apiCalls++
			if r.Header.Get("x-webapi-key") != "test-private-key" || strings.Contains(r.URL.String(), "test-private-key") {
				t.Fatal("wrong credential transport")
			}
			if r.URL.Path != "/IStoreService/GetAppList/v1/" {
				t.Fatal("wrong Steam API method")
			}
			var args struct {
				After int  `json:"last_appid"`
				Limit int  `json:"max_results"`
				Games bool `json:"include_games"`
				DLC   bool `json:"include_dlc"`
			}
			if json.Unmarshal([]byte(r.URL.Query().Get("input_json")), &args) != nil || args.After != 500 || args.Limit != 12 || !args.Games || args.DLC {
				t.Fatal("incorrect catalog parameters")
			}
			body = `{"response":{"apps":[{"appid":501},{"appid":502}],"last_appid":502,"have_more_results":true}}`
		} else {
			storeCalls++
			if r.URL.Host != "store.steampowered.com" || r.Header.Get("x-webapi-key") != "" {
				t.Fatal("credential leaked to Store")
			}
			id := r.URL.Query().Get("appids")
			body = `{"` + id + `":{"success":false}}`
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
	w := httptest.NewRecorder()
	a.SyncSteam(w, httptest.NewRequest("POST", "/api/admin/sync/steam?after=500", nil))
	var data struct {
		Imported int  `json:"imported"`
		Next     int  `json:"nextAppId"`
		More     bool `json:"hasMore"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &data); err != nil {
		t.Fatal(err)
	}
	if w.Code != 200 || data.Imported != 0 || data.Next != 502 || !data.More || apiCalls != 1 || storeCalls != 2 {
		t.Fatalf("incorrect skipped batch response: %d %s", w.Code, w.Body.String())
	}
}

func TestSteamRejectsBadKeyedResponsesBeforeStoreOrDatabase(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
	}{
		{"invalid key", `{"error":"secret must never be relayed"}`, 403},
		{"quota", `{}`, 429}, {"unavailable", `{}`, 503}, {"redirect", ``, 302},
		{"missing apps", `{"response":{}}`, 200},
		{"duplicate ids", `{"response":{"apps":[{"appid":501},{"appid":501}]}}`, 200},
		{"out of order", `{"response":{"apps":[{"appid":502},{"appid":501}]}}`, 200},
		{"stale cursor", `{"response":{"apps":[{"appid":499}]}}`, 200},
		{"no progress", `{"response":{"apps":[],"have_more_results":true}}`, 200},
		{"oversized", strings.Repeat(" ", (1<<20)+1), 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			a := &handler.API{SteamKey: "private-test-key", SteamClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if calls != 1 || r.URL.Host != "api.steampowered.com" {
					t.Fatal("bad API response reached another endpoint")
				}
				headers := make(http.Header)
				headers.Set("Location", "https://example.com/")
				return &http.Response{StatusCode: tc.status, Body: io.NopCloser(strings.NewReader(tc.body)), Header: headers}, nil
			})}}
			w := httptest.NewRecorder()
			a.SyncSteam(w, httptest.NewRequest("POST", "/api/admin/sync/steam?after=500", nil))
			if w.Code != 502 || strings.Contains(w.Body.String(), "private-test-key") || strings.Contains(w.Body.String(), "secret must never be relayed") {
				t.Fatalf("unsafe failure response: %d %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestSteamEmptyCatalogEndsPagination(t *testing.T) {
	feed := `{"response":{"apps":[],"have_more_results":false}}`
	page := ""
	a := &handler.API{SteamKey: "test-key", SteamClient: steamClient(&feed, func(int64) string { t.Fatal("empty batch fetched details"); return "" }, &page)}
	w := httptest.NewRecorder()
	a.SyncSteam(w, httptest.NewRequest("POST", "/api/admin/sync/steam?after=500", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"hasMore":false`) || !strings.Contains(w.Body.String(), `"nextAppId":500`) {
		t.Fatal("incorrect empty catalog response")
	}
}
