package tests

import (
	"gamematch/handler"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUnauthenticatedCannotReachAdmin(t *testing.T) {
	a := &handler.API{Secret: "test-secret-at-least-32-characters"}
	for _, header := range []string{"", "Bearer broken"} {
		req := httptest.NewRequest("POST", "/api/admin/sync", nil)
		req.Header.Set("Authorization", header)
		w := httptest.NewRecorder()
		a.Guard(true, func(w http.ResponseWriter, r *http.Request) { t.Fatal("unauthenticated request reached handler") })(w, req)
		if w.Code != 401 {
			t.Fatalf("want 401 got %d", w.Code)
		}
	}
}
