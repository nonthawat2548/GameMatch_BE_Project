package mw

import (
	"net/http"
	"strings"
)

func CORS(origins string, next http.Handler) http.Handler {
	allow := map[string]bool{}
	for _, s := range strings.Split(origins, ",") {
		allow[strings.TrimSpace(s)] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		w.Header().Add("Vary", "Origin")
		if origin != "" {
			if !allow[origin] {
				http.Error(w, "Origin not allowed", 403)
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		}
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}
