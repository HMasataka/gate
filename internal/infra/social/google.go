package social

// NewGoogleProvider creates an OIDCProvider configured for Google.
func NewGoogleProvider(clientID, clientSecret, redirectURI string) *OIDCProvider {
	return NewOIDCProvider(
		"google",
		clientID, clientSecret, redirectURI,
		"https://accounts.google.com/o/oauth2/v2/auth",
		"https://oauth2.googleapis.com/token",
		"https://openidconnect.googleapis.com/v1/userinfo",
		[]string{"openid", "email", "profile"},
	)
}
