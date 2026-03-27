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

const (
	githubAuthorizeURL   = "https://github.com/login/oauth/authorize"
	githubTokenURL       = "https://github.com/login/oauth/access_token"
	githubUserURL        = "https://api.github.com/user"
	githubUserEmailsURL  = "https://api.github.com/user/emails"
)

// GitHubProvider implements the SocialProvider interface for GitHub OAuth2.
type GitHubProvider struct {
	clientID     string
	clientSecret string
	redirectURI  string
}

// NewGitHubProvider creates a new GitHubProvider.
func NewGitHubProvider(clientID, clientSecret, redirectURI string) *GitHubProvider {
	return &GitHubProvider{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
	}
}

// Name returns "github".
func (p *GitHubProvider) Name() string {
	return "github"
}

// AuthURL builds the GitHub OAuth2 authorization URL with the given state.
func (p *GitHubProvider) AuthURL(state string) string {
	params := url.Values{}
	params.Set("client_id", p.clientID)
	params.Set("redirect_uri", p.redirectURI)
	params.Set("scope", "read:user,user:email")
	params.Set("state", state)
	return githubAuthorizeURL + "?" + params.Encode()
}

// ExchangeCode exchanges the GitHub authorization code for user info.
func (p *GitHubProvider) ExchangeCode(ctx context.Context, code string) (*UserInfo, error) {
	accessToken, err := p.fetchAccessToken(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("fetch github access token: %w", err)
	}

	info, err := p.fetchUserInfo(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("fetch github user info: %w", err)
	}

	return info, nil
}

func (p *GitHubProvider) fetchAccessToken(ctx context.Context, code string) (string, error) {
	form := url.Values{}
	form.Set("client_id", p.clientID)
	form.Set("client_secret", p.clientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", p.redirectURI)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, githubTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("post github token endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read github token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github token endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("parse github token response: %w", err)
	}
	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("empty access token in github response")
	}

	return tokenResp.AccessToken, nil
}

func (p *GitHubProvider) fetchUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubUserURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create github user request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get github user endpoint: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read github user response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github user endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var raw struct {
		ID        int64  `json:"id"`
		Email     string `json:"email"`
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
		Name      string `json:"name"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse github user response: %w", err)
	}

	email := raw.Email
	emailVerified := false

	// If email is empty, fetch primary verified email from /user/emails
	if email == "" {
		email, emailVerified = p.fetchPrimaryEmail(ctx, accessToken)
	}

	name := raw.Name
	if name == "" {
		name = raw.Login
	}

	return &UserInfo{
		ProviderUserID: fmt.Sprintf("%d", raw.ID),
		Email:          email,
		Name:           name,
		AvatarURL:      raw.AvatarURL,
		EmailVerified:  emailVerified,
	}, nil
}

func (p *GitHubProvider) fetchPrimaryEmail(ctx context.Context, accessToken string) (string, bool) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, githubUserEmailsURL, nil)
	if err != nil {
		return "", false
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", false
	}

	if resp.StatusCode != http.StatusOK {
		return "", false
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", false
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, true
		}
	}

	// Fallback: return first verified email
	for _, e := range emails {
		if e.Verified {
			return e.Email, true
		}
	}

	return "", false
}
