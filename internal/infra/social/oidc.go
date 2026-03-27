package social

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// OIDCProvider is a generic OIDC/OAuth2 provider implementation.
type OIDCProvider struct {
	name         string
	clientID     string
	clientSecret string
	redirectURI  string
	authURL      string
	tokenURL     string
	userInfoURL  string
	scopes       []string
}

// NewOIDCProvider creates a new OIDCProvider with the given configuration.
func NewOIDCProvider(
	name, clientID, clientSecret, redirectURI,
	authURL, tokenURL, userInfoURL string,
	scopes []string,
) *OIDCProvider {
	return &OIDCProvider{
		name:         name,
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		authURL:      authURL,
		tokenURL:     tokenURL,
		userInfoURL:  userInfoURL,
		scopes:       scopes,
	}
}

// Name returns the provider name.
func (p *OIDCProvider) Name() string {
	return p.name
}

// AuthURL builds the OAuth2 authorization URL with the given state.
func (p *OIDCProvider) AuthURL(state string) string {
	params := url.Values{}
	params.Set("client_id", p.clientID)
	params.Set("redirect_uri", p.redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", strings.Join(p.scopes, " "))
	params.Set("state", state)
	return p.authURL + "?" + params.Encode()
}

// ExchangeCode exchanges the authorization code for user info.
func (p *OIDCProvider) ExchangeCode(ctx context.Context, code string) (*UserInfo, error) {
	accessToken, err := p.fetchAccessToken(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("fetch access token: %w", err)
	}

	info, err := p.fetchUserInfo(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("fetch user info: %w", err)
	}

	return info, nil
}

func (p *OIDCProvider) fetchAccessToken(ctx context.Context, code string) (string, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", p.redirectURI)
	form.Set("client_id", p.clientID)
	form.Set("client_secret", p.clientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("post token endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("parse token response: %w", err)
	}
	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("empty access token in response")
	}

	return tokenResp.AccessToken, nil
}

func (p *OIDCProvider) fetchUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get userinfo endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read userinfo response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var raw struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse userinfo response: %w", err)
	}

	return &UserInfo{
		ProviderUserID: raw.Sub,
		Email:          raw.Email,
		Name:           raw.Name,
		AvatarURL:      raw.Picture,
		EmailVerified:  raw.EmailVerified,
	}, nil
}
