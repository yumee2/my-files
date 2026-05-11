package http

import (
	"context"
	"net/http"
)

type SessionValidator interface {
	ValidateSession(ctx context.Context, sessionID string) error
}

func RequireAuth(authService SessionValidator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if err := authService.ValidateSession(r.Context(), cookie.Value); err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
