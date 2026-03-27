package domain

import "time"

type Session struct {
	ID        string
	UserID    string
	IPAddress string
	UserAgent string
	ExpiresAt time.Time
	CreatedAt time.Time
}
