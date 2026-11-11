package tests

import (
	"gamematch/mw"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOriginAllowlist(t *testing.T) {
	h := mw.CORS("http://localhost:3000", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	for _, tc := range []struct {
		origin string
		status int
	}{{"http://localhost:3000", 204}, {"https://untrusted.example", 403}} {
		req := httptest.NewRequest("OPTIONS", "/api/games", nil)
		req.Header.Set("Origin", tc.origin)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != tc.status {
			t.Fatalf("%s => %d", tc.origin, w.Code)
		}
	}
}
