package tests

import (
	"gamematch/handler"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSyncMissingKeyAndConcurrentImport(t *testing.T) {
	a := &handler.API{}
	w := httptest.NewRecorder()
	a.Sync(w, httptest.NewRequest("POST", "/api/admin/sync", nil))
	if w.Code != 503 {
		t.Fatal("missing key should return 503")
	}
	a.RawgKey = "test-only-key"
	a.SyncMu.Lock()
	defer a.SyncMu.Unlock()
	w = httptest.NewRecorder()
	a.Sync(w, httptest.NewRequest("POST", "/api/admin/sync", nil))
	if w.Code != 409 {
		t.Fatal("concurrent import should return 409")
	}
}

func TestSyncRejectsBadUpstreamBeforeDatabase(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
	}{
		{"upstream quota", `{"error":"quota"}`, 429},
		{"missing results", `{}`, 200},
		{"null results", `{"results":null}`, 200},
		{"trailing JSON", `{"results":[]} {}`, 200},
		{"oversized body", strings.Repeat(" ", (8<<20)+1), 200},
		{"more than first page", `{"results":[` + strings.Repeat(`{"id":1,"name":"Demo: Fixture"},`, 40) + `{"id":41,"name":"Demo: Extra"}]}`, 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := &handler.API{RawgKey: "test-only-key", RawgClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.status, Body: io.NopCloser(strings.NewReader(tc.body)), Header: make(http.Header)}, nil
			})}}
			w := httptest.NewRecorder()
			a.Sync(w, httptest.NewRequest("POST", "/api/admin/sync", nil))
			if w.Code != 502 {
				t.Fatalf("wanted 502, got %d", w.Code)
			}
		})
	}
}

func TestSyncStatusDoesNotExposeKey(t *testing.T) {
	for _, key := range []string{"", "   ", "private-test-key"} {
		a := &handler.API{SteamKey: key}
		w := httptest.NewRecorder()
		a.SyncStatus(w, httptest.NewRequest("GET", "/api/admin/sync/status", nil))
		expected := "{\"steamConfigured\":false}\n"
		if strings.TrimSpace(key) != "" {
			expected = "{\"steamConfigured\":true}\n"
		}
		if w.Code != 200 || w.Body.String() != expected || w.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("unexpected status response: %d %s", w.Code, w.Body.String())
		}
	}
}

func TestSyncStatusRequiresAuthentication(t *testing.T) {
	a := &handler.API{Secret: "test-secret-at-least-32-characters", SteamKey: "private-test-key"}
	w := httptest.NewRecorder()
	a.Routes().ServeHTTP(w, httptest.NewRequest("GET", "/api/admin/sync/status", nil))
	if w.Code != 401 {
		t.Fatalf("want 401 got %d", w.Code)
	}
}
