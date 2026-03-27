package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"slices"
	"strings"
	"time"

	"github.com/HMasataka/gate/internal/config"
	"github.com/HMasataka/gate/internal/domain"
)

type OAuthUsecase struct {
	clientRepo   domain.OAuthClientRepository
	codeRepo     domain.AuthorizationCodeRepository
	tokenUsecase *TokenUsecase
	random       domain.RandomGenerator
	oauthCfg     config.OAuthConfig
	tokenCfg     config.TokenConfig
	txRunner     domain.TxRunner
}

func NewOAuthUsecase(
	clientRepo domain.OAuthClientRepository,
	codeRepo domain.AuthorizationCodeRepository,
	tokenUsecase *TokenUsecase,
	random domain.RandomGenerator,
	oauthCfg config.OAuthConfig,
	tokenCfg config.TokenConfig,
	txRunner domain.TxRunner,
) *OAuthUsecase {
	return &OAuthUsecase{
		clientRepo:   clientRepo,
		codeRepo:     codeRepo,
		tokenUsecase: tokenUsecase,
		random:       random,
		oauthCfg:     oauthCfg,
		tokenCfg:     tokenCfg,
		txRunner:     txRunner,
	}
}

func (u *OAuthUsecase) Authorize(
	ctx context.Context,
	clientID, redirectURI, responseType, scope, state, codeChallenge, codeChallengeMethod string,
) (string, error) {
	client, err := u.clientRepo.GetByID(ctx, clientID)
	if err != nil {
		return "", domain.ErrInvalidClient
	}

	if !slices.Contains(client.RedirectURIs, redirectURI) {
		return "", domain.ErrInvalidRedirectURI
	}

	if responseType != "code" {
		return "", domain.ErrInvalidGrantType
	}

	if codeChallengeMethod != "" && codeChallengeMethod != "S256" {
		return "", domain.ErrInvalidGrantType
	}

	rawCode, err := u.random.GenerateToken(32)
	if err != nil {
		return "", err
	}

	codeHash := sha256hex(rawCode)

	scopes := splitScope(scope)

	authCode := &domain.AuthorizationCode{
		ID:                  u.random.GenerateUUID(),
		Code:                codeHash,
		ClientID:            clientID,
		UserID:              "",
		RedirectURI:         redirectURI,
		Scopes:              scopes,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		Nonce:               state,
		ExpiresAt:           time.Now().Add(u.tokenCfg.AuthCodeExpiry),
	}

	if err := u.codeRepo.Create(ctx, authCode); err != nil {
		return "", err
	}

	return rawCode, nil
}

func (u *OAuthUsecase) ExchangeCode(
	ctx context.Context,
	clientID, clientSecret, code, codeVerifier, redirectURI string,
) (string, string, error) {
	client, err := u.clientRepo.GetByID(ctx, clientID)
	if err != nil {
		return "", "", domain.ErrInvalidClient
	}

	if client.Secret != sha256hex(clientSecret) {
		return "", "", domain.ErrInvalidClient
	}

	codeHash := sha256hex(code)

	authCode, err := u.codeRepo.GetByCode(ctx, codeHash)
	if err != nil {
		return "", "", domain.ErrInvalidToken
	}

	if time.Now().After(authCode.ExpiresAt) {
		return "", "", domain.ErrTokenExpired
	}

	if authCode.UsedAt != nil {
		return "", "", domain.ErrCodeReuse
	}

	if authCode.CodeChallenge != "" {
		if codeVerifier == "" {
			return "", "", domain.ErrInvalidToken
		}

		sum := sha256.Sum256([]byte(codeVerifier))
		computed := base64.RawURLEncoding.EncodeToString(sum[:])

		if computed != authCode.CodeChallenge {
			return "", "", domain.ErrInvalidToken
		}
	}

	var accessToken, refreshToken string
	if err := u.txRunner.RunInTx(ctx, func(ctx context.Context) error {
		if err := u.codeRepo.MarkUsed(ctx, authCode.ID); err != nil {
			return err
		}
		var err error
		accessToken, refreshToken, err = u.tokenUsecase.IssueTokenPair(ctx, authCode.UserID, clientID, authCode.Scopes)
		return err
	}); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (u *OAuthUsecase) ClientCredentials(
	ctx context.Context,
	clientID, clientSecret, scope string,
) (string, error) {
	client, err := u.clientRepo.GetByID(ctx, clientID)
	if err != nil {
		return "", domain.ErrInvalidClient
	}

	if client.Type != domain.ClientTypeConfidential {
		return "", domain.ErrInvalidGrantType
	}

	if client.Secret != sha256hex(clientSecret) {
		return "", domain.ErrInvalidClient
	}

	scopes := splitScope(scope)

	claims := map[string]any{
		"sub":   clientID,
		"scope": strings.Join(scopes, " "),
	}

	accessToken, err := u.tokenUsecase.jwt.GenerateAccessToken(ctx, claims)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func (u *OAuthUsecase) Introspect(ctx context.Context, token string) (map[string]any, error) {
	claims, err := u.tokenUsecase.jwt.ValidateToken(ctx, token)
	if err != nil {
		return map[string]any{"active": false}, nil
	}

	claims["active"] = true

	return claims, nil
}

func sha256hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func splitScope(scope string) []string {
	if scope == "" {
		return nil
	}

	parts := strings.Fields(scope)

	return parts
}
