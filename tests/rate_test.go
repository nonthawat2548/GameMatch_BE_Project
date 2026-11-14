package tests

import (
	"gamematch/handler"
	"gamematch/mw"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRateLimitWindowAndSeparatePeers(t *testing.T) {
	l := mw.NewRateLimiter(2, time.Minute)
	now := time.Unix(100, 0)
	for i := 0; i < 2; i++ {
		if ok, _ := l.Allow("one", now); !ok {
			t.Fatal("allowed attempts were rejected")
		}
	}
	if ok, wait := l.Allow("one", now.Add(10*time.Second)); ok || wait != 50*time.Second {
		t.Fatalf("blocked=%v retry=%v", !ok, wait)
	}
	if ok, _ := l.Allow("two", now); !ok {
		t.Fatal("unrelated peer was blocked")
	}
	if ok, _ := l.Allow("one", now.Add(time.Minute)); !ok {
		t.Fatal("window did not expire")
	}
}

func TestRateLimitConcurrentRequests(t *testing.T) {
	l := mw.NewRateLimiter(10, time.Minute)
	var accepted atomic.Int32
	var group sync.WaitGroup
	for i := 0; i < 100; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if ok, _ := l.Allow("same", time.Unix(100, 0)); ok {
				accepted.Add(1)
			}
		}()
	}
	group.Wait()
	if accepted.Load() != 10 {
		t.Fatalf("accepted %d concurrent requests", accepted.Load())
	}
}

func TestRateLimitIgnoresSpoofedForwardedHeadersAndPorts(t *testing.T) {
	l := mw.NewRateLimiter(1, time.Minute)
	wrapped := l.Wrap(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
	for i, peer := range []string{"192.0.2.1:1000", "192.0.2.1:2000"} {
		r := httptest.NewRequest("POST", "/login", nil)
		r.RemoteAddr = peer
		r.Header.Set("X-Forwarded-For", "198.51.100."+string(rune('1'+i)))
		w := httptest.NewRecorder()
		wrapped(w, r)
		if i == 0 && w.Code != 204 {
			t.Fatal(w.Code)
		}
		if i == 1 && (w.Code != 429 || w.Header().Get("Retry-After") == "" || !strings.Contains(w.Body.String(), "error")) {
			t.Fatalf("missing JSON 429/retry: %s", w.Body.String())
		}
	}
}

func TestLoginAndRegisterShareRouteLimit(t *testing.T) {
	a := &handler.API{}
	routes := a.Routes()
	for i := 0; i < 31; i++ {
		path := "/api/auth/login"
		if i%2 == 0 {
			path = "/api/auth/register"
		}
		r := httptest.NewRequest("POST", path, strings.NewReader("not JSON"))
		r.RemoteAddr = "192.0.2.10:1000"
		w := httptest.NewRecorder()
		routes.ServeHTTP(w, r)
		want := 400
		if i == 30 {
			want = 429
		}
		if w.Code != want {
			t.Fatalf("attempt %d want %d got %d", i+1, want, w.Code)
		}
	}
}

func TestRateLimitBoundsNewPeersAndReclaimsExpiredEntries(t *testing.T) {
	l := mw.NewRateLimiter(1, time.Minute)
	now := time.Unix(100, 0)
	for i := 0; i < 10000; i++ {
		if ok, _ := l.Allow(string(rune(i)), now); !ok {
			t.Fatal("capacity reached early")
		}
	}
	if ok, _ := l.Allow("overflow", now); ok {
		t.Fatal("unbounded peer map")
	}
	if ok, _ := l.Allow("reclaimed", now.Add(time.Minute)); !ok {
		t.Fatal("expired peers were not reclaimed")
	}
}
