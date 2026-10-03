package server

import (
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

// BearerMiddleware checks the Authorization header. HTTP mode had no
// authentication before, so this is added for the tereo instance. The compare is
// timing-safe (crypto/subtle.ConstantTimeCompare).
//
// When token is empty, no authentication is performed. That is the loopback-only
// case: the server relies on binding to 127.0.0.1 only. Startup logs a WARN once.
func BearerMiddleware(token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		got := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
			slog.Warn("auth failed", "path", r.URL.Path, "remote", r.RemoteAddr)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("WWW-Authenticate", "Bearer")
			w.WriteHeader(http.StatusUnauthorized)
			resp := wireResponse{
				JSONRPC: "2.0",
				Error:   &wireError{Code: errInvalidRequest, Message: "unauthorized"},
			}
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				slog.Error("auth error encode failed", "error", err)
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}
