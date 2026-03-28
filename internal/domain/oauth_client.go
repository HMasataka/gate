package domain

import "time"

type ClientType string

const (
	ClientTypeConfidential ClientType = "confidential"
	ClientTypePublic       ClientType = "public"
)

type OAuthClient struct {
	ID              string     `json:"client_id"`
	Secret          string     `json:"client_secret,omitempty"`
	Name            string     `json:"name"`
	Type            ClientType `json:"client_type"`
	OwnerID         string     `json:"owner_id,omitempty"`
	RedirectURIs    []string   `json:"redirect_uris"`
	AllowedScopes   []string   `json:"allowed_scopes"`
	TokensRevokedAt *time.Time `json:"tokens_revoked_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
