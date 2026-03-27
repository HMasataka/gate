package social

import "context"

// UserInfo holds the normalized user information from a social provider.
type UserInfo struct {
	ProviderUserID string
	Email          string
	Name           string
	AvatarURL      string
	EmailVerified  bool
}

// SocialProvider is the interface that all OAuth2/OIDC providers must implement.
type SocialProvider interface {
	// AuthURL generates the OAuth2 authorization URL with the given state parameter.
	AuthURL(state string) string
	// ExchangeCode exchanges the authorization code for user info.
	ExchangeCode(ctx context.Context, code string) (*UserInfo, error)
	// Name returns the provider name (e.g. "google", "github").
	Name() string
}
