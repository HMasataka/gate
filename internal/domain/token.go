package domain

import "time"

type RefreshToken struct {
	ID        string
	TokenHash string
	UserID    string
	ClientID  string
	FamilyID  string
	Scopes    []string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type AuthorizationCode struct {
	ID                  string
	Code                string
	ClientID            string
	UserID              string
	RedirectURI         string
	Scopes              []string
	CodeChallenge       string
	CodeChallengeMethod string
	Nonce               string
	ExpiresAt           time.Time
	UsedAt              *time.Time
	CreatedAt           time.Time
}
