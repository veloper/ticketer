package internal

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userKey contextKey = "user"

func PATMiddleware(store *Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only API routes need auth
			if !strings.HasPrefix(r.URL.Path, "/api/") {
				next.ServeHTTP(w, r)
				return
			}
			// WebSocket endpoint authenticates via ?pat= query param instead.
			if r.URL.Path == "/api/ws" || r.URL.Path == "/mcp" {
				next.ServeHTTP(w, r)
				return
			}

			pat := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if pat == "" {
				http.Error(w, `{"error":"missing authorization"}`, http.StatusUnauthorized)
				return
			}

			user, err := store.GetUserByPAT(pat)
			if err != nil {
				http.Error(w, `{"error":"invalid pat"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserFromCtx(ctx context.Context) *User {
	return ctx.Value(userKey).(*User)
}
