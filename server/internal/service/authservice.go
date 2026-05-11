package service

import (
	"context"
	"file-uploader/internal/models"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthDbI interface {
	CreatePassword(password string) error
	GetPassword(ctx context.Context) (string, error)
	CreateSession(ctx context.Context, id string, expiresAt time.Time) error
	GetSession(ctx context.Context, id string) (*models.Session, error)
}

type AuthService struct {
	db AuthDbI
}

func NewAuthService(db AuthDbI) *AuthService {
	return &AuthService{db: db}
}

func (as *AuthService) CreatePassword(password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return as.db.CreatePassword(string(bytes))
}

func (as *AuthService) VerifyPassword(ctx context.Context, password string) error {
	storedPassword, err := as.db.GetPassword(ctx)
	if err != nil {
		return err
	}
	return bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
}

func (as *AuthService) CreateSession(ctx context.Context) (string, error) {
	sessionID := uuid.NewString()
	return sessionID, as.db.CreateSession(ctx, sessionID, time.Now().Add(7*24*time.Hour))
}

func (as *AuthService) ValidateSession(ctx context.Context, sessionID string) error {
	session, err := as.db.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return nil
	}
	if time.Now().After(session.ExpiresAt) {
		return nil
	}
	return nil
}
