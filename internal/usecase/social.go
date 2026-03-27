package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/HMasataka/gate/internal/config"
	"github.com/HMasataka/gate/internal/domain"
)

// SocialProvider is the usecase-layer interface for social OAuth2 providers.
// This intentionally does NOT import the infra/social package (Clean Architecture).
type SocialProvider interface {
	Name() string
	AuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*SocialUserInfo, error)
}

// SocialUserInfo holds the normalized user information returned by a social provider.
type SocialUserInfo struct {
	ProviderUserID string
	Email          string
	Name           string
	AvatarURL      string
	EmailVerified  bool
}

// SocialUsecase handles social login flows.
type SocialUsecase struct {
	socialRepo domain.SocialConnectionRepository
	userRepo   domain.UserRepository
	sessions   domain.SessionStore
	token      *TokenUsecase
	random     domain.RandomGenerator
	providers  map[string]SocialProvider
	sessCfg    config.SessionConfig
}

// NewSocialUsecase creates a new SocialUsecase.
func NewSocialUsecase(
	socialRepo domain.SocialConnectionRepository,
	userRepo domain.UserRepository,
	sessions domain.SessionStore,
	token *TokenUsecase,
	random domain.RandomGenerator,
	providers map[string]SocialProvider,
	sessCfg config.SessionConfig,
) *SocialUsecase {
	return &SocialUsecase{
		socialRepo: socialRepo,
		userRepo:   userRepo,
		sessions:   sessions,
		token:      token,
		random:     random,
		providers:  providers,
		sessCfg:    sessCfg,
	}
}

// GetAuthURL returns the OAuth2 authorization URL for the given provider.
func (u *SocialUsecase) GetAuthURL(provider, state string) (string, error) {
	p, ok := u.providers[provider]
	if !ok {
		return "", fmt.Errorf("unknown social provider: %s", provider)
	}
	return p.AuthURL(state), nil
}

// HandleCallback handles the OAuth2 callback, creating or linking a user account,
// and returns a LoginResult with session and JWT tokens.
func (u *SocialUsecase) HandleCallback(ctx context.Context, provider, code, ipAddr, userAgent string) (*LoginResult, error) {
	p, ok := u.providers[provider]
	if !ok {
		return nil, fmt.Errorf("unknown social provider: %s", provider)
	}

	userInfo, err := p.ExchangeCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchange social code: %w", err)
	}

	// Try to find existing social connection.
	conn, err := u.socialRepo.GetByProviderAndProviderUserID(ctx, provider, userInfo.ProviderUserID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("lookup social connection: %w", err)
	}

	var user *domain.User

	if conn != nil {
		// Existing connection: load the user.
		user, err = u.userRepo.GetByID(ctx, conn.UserID)
		if err != nil {
			return nil, fmt.Errorf("get user by id: %w", err)
		}
	} else {
		// No connection found. Try email-based auto-linking if email is available.
		if userInfo.Email != "" {
			user, err = u.userRepo.GetByEmail(ctx, userInfo.Email)
			if err != nil && !errors.Is(err, domain.ErrNotFound) {
				return nil, fmt.Errorf("lookup user by email: %w", err)
			}
		}

		if user != nil {
			// Auto-link: create a social connection for the existing user.
			newConn := &domain.SocialConnection{
				ID:             u.random.GenerateUUID(),
				UserID:         user.ID,
				Provider:       provider,
				ProviderUserID: userInfo.ProviderUserID,
				Email:          userInfo.Email,
				Name:           userInfo.Name,
				AvatarURL:      userInfo.AvatarURL,
			}
			if err := u.socialRepo.Create(ctx, newConn); err != nil {
				return nil, fmt.Errorf("create social connection for existing user: %w", err)
			}
		} else {
			// New user: create account and social connection.
			now := time.Now()
			status := domain.UserStatusUnverified
			if userInfo.EmailVerified {
				status = domain.UserStatusActive
			}

			user = &domain.User{
				ID:            u.random.GenerateUUID(),
				Email:         userInfo.Email,
				PasswordHash:  "",
				Status:        status,
				EmailVerified: userInfo.EmailVerified,
				Version:       1,
				CreatedAt:     now,
				UpdatedAt:     now,
			}
			if err := u.userRepo.Create(ctx, user); err != nil {
				return nil, fmt.Errorf("create user for social login: %w", err)
			}

			newConn := &domain.SocialConnection{
				ID:             u.random.GenerateUUID(),
				UserID:         user.ID,
				Provider:       provider,
				ProviderUserID: userInfo.ProviderUserID,
				Email:          userInfo.Email,
				Name:           userInfo.Name,
				AvatarURL:      userInfo.AvatarURL,
			}
			if err := u.socialRepo.Create(ctx, newConn); err != nil {
				return nil, fmt.Errorf("create social connection for new user: %w", err)
			}
		}
	}

	// Create a session.
	session := &domain.Session{
		ID:        u.random.GenerateUUID(),
		UserID:    user.ID,
		IPAddress: ipAddr,
		UserAgent: userAgent,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	if err := u.sessions.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	result := &LoginResult{
		Session: session,
	}

	if u.token != nil {
		accessToken, refreshToken, err := u.token.IssueTokenPair(ctx, user.ID, "", nil)
		if err != nil {
			return nil, fmt.Errorf("issue token pair: %w", err)
		}
		result.AccessToken = accessToken
		result.RefreshToken = refreshToken
	}

	return result, nil
}
