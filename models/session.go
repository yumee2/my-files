package models

import "time"

type Session struct {
	ID        string
	ExpiresAt time.Time
	CreatedAt time.Time
}
