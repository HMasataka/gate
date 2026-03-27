package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/HMasataka/gate/internal/domain"
)

type OIDCUsecase struct {
	userRepo  domain.UserRepository
	jwt       domain.JWTManager
	serverURL string
}

func NewOIDCUsecase(userRepo domain.UserRepository, jwt domain.JWTManager, serverURL string) *OIDCUsecase {
	return &OIDCUsecase{
		userRepo:  userRepo,
		jwt:       jwt,
		serverURL: serverURL,
	}
}

// UserInfo returns standard OIDC claims for the given user ID.
func (u *OIDCUsecase) UserInfo(ctx context.Context, userID string) (map[string]any, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	return map[string]any{
		"sub":            user.ID,
		"email":          user.Email,
		"email_verified": user.EmailVerified,
		"updated_at":     user.UpdatedAt.Unix(),
	}, nil
}

// Discovery returns the OIDC Discovery document as per RFC 8414.
func (u *OIDCUsecase) Discovery() map[string]any {
	return map[string]any{
		"issuer":                                u.serverURL,
		"authorization_endpoint":               u.serverURL + "/api/v1/oauth/authorize",
		"token_endpoint":                        u.serverURL + "/oauth/token",
		"userinfo_endpoint":                     u.serverURL + "/oauth/userinfo",
		"jwks_uri":                              u.serverURL + "/.well-known/jwks.json",
		"response_types_supported":              []string{"code"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"ES256", "RS256"},
		"scopes_supported":                      []string{"openid", "email", "profile"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_post", "client_secret_basic"},
		"claims_supported":                      []string{"sub", "email", "email_verified", "updated_at"},
	}
}

// IssueIDToken generates a signed ID token for the given user, client, and nonce.
func (u *OIDCUsecase) IssueIDToken(ctx context.Context, userID, clientID, nonce string) (string, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("get user: %w", err)
	}

	now := time.Now()
	claims := map[string]any{
		"sub":            user.ID,
		"email":          user.Email,
		"email_verified": user.EmailVerified,
		"aud":            clientID,
		"iat":            now.Unix(),
		"exp":            now.Add(time.Hour).Unix(),
	}

	if nonce != "" {
		claims["nonce"] = nonce
	}

	token, err := u.jwt.GenerateIDToken(ctx, claims)
	if err != nil {
		return "", fmt.Errorf("generate id token: %w", err)
	}

	return token, nil
}
