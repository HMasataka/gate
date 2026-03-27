package domain

import "time"

type ClientType string

const (
	ClientTypeConfidential ClientType = "confidential"
	ClientTypePublic       ClientType = "public"
)

type OAuthClient struct {
	ID              string
	Secret          string
	Name            string
	Type            ClientType
	RedirectURIs    []string
	AllowedScopes   []string
	TokensRevokedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
