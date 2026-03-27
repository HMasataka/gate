package domain

import "time"

type SocialConnection struct {
	ID             string
	UserID         string
	Provider       string
	ProviderUserID string
	Email          string
	Name           string
	AvatarURL      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
