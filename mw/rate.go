package mw

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type rateWindow struct {
	count int
	end   time.Time
}

// RateLimiter limits attempts per TCP peer, shared by login and registration.
// Forwarded headers are deliberately ignored: callers can forge them.
// A Render proxy/NAT can share one bucket; this is a single-process demo limit.
type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]rateWindow
	limit   int
	window  time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	if limit < 1 || window <= 0 {
		panic("rate limit and window must be positive")
	}
	return &RateLimiter{entries: make(map[string]rateWindow), limit: limit, window: window}
}

// Allow accepts an explicit time so expiry can be tested without sleeps.
func (l *RateLimiter) Allow(peer string, now time.Time) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	bucket, exists := l.entries[peer]
	if !exists || !now.Before(bucket.end) {
		if !exists && len(l.entries) >= 10000 {
			for key, value := range l.entries {
				if !now.Before(value.end) {
					delete(l.entries, key)
				}
			}
			if len(l.entries) >= 10000 {
				return false, l.window
			}
		}
		bucket = rateWindow{end: now.Add(l.window)}
	}
	if bucket.count >= l.limit {
		return false, bucket.end.Sub(now)
	}
	bucket.count++
	l.entries[peer] = bucket
	return true, 0
}

func (l *RateLimiter) Wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		peer, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			peer = r.RemoteAddr
		}
		if ip := net.ParseIP(peer); ip != nil {
			peer = ip.String()
		}
		if ok, wait := l.Allow(peer, time.Now()); !ok {
			seconds := int((wait + time.Second - 1) / time.Second)
			w.Header().Set("Retry-After", strconv.Itoa(seconds))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{"error": "ลองเข้าสู่ระบบหรือสมัครสมาชิกบ่อยเกินไป กรุณารอสักครู่แล้วลองใหม่"})
			return
		}
		next(w, r)
	}
}
