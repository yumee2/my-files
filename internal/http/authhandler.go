package http

import (
	"context"
	"net/http"
)

type AuthServiceI interface {
	VerifyPassword(ctx context.Context, password string) error
	CreateSession(ctx context.Context) (string, error)
}

type AuthHandler struct {
	service AuthServiceI
}

func NewAuthHandler(service AuthServiceI) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("password")
	if err := h.service.VerifyPassword(r.Context(), password); err != nil {
		http.Error(w, "invalid password", http.StatusUnauthorized)
		return
	}

	sessionID, err := h.service.CreateSession(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60 * 60 * 24 * 7,
	})

	w.WriteHeader(http.StatusNoContent)

}
