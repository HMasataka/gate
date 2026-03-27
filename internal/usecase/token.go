package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/HMasataka/gate/internal/config"
	"github.com/HMasataka/gate/internal/domain"
)

type TokenUsecase struct {
	tokenRepo domain.RefreshTokenRepository
	userRepo  domain.UserRepository
	jwt       domain.JWTManager
	random    domain.RandomGenerator
	tokenCfg  config.TokenConfig
	txRunner  domain.TxRunner
}

func NewTokenUsecase(
	tokenRepo domain.RefreshTokenRepository,
	userRepo domain.UserRepository,
	jwt domain.JWTManager,
	random domain.RandomGenerator,
	tokenCfg config.TokenConfig,
	txRunner domain.TxRunner,
) *TokenUsecase {
	return &TokenUsecase{
		tokenRepo: tokenRepo,
		userRepo:  userRepo,
		jwt:       jwt,
		random:    random,
		tokenCfg:  tokenCfg,
		txRunner:  txRunner,
	}
}

func (u *TokenUsecase) IssueTokenPair(ctx context.Context, userID, clientID string, scopes []string) (string, string, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", "", err
	}

	claims := map[string]any{
		"sub":   userID,
		"email": user.Email,
		"scope": strings.Join(scopes, " "),
	}

	accessToken, err := u.jwt.GenerateAccessToken(ctx, claims)
	if err != nil {
		return "", "", err
	}

	rawToken, err := u.random.GenerateToken(32)
	if err != nil {
		return "", "", err
	}

	tokenHash := hashToken(rawToken)
	familyID := u.random.GenerateUUID()

	refreshToken := &domain.RefreshToken{
		ID:        u.random.GenerateUUID(),
		TokenHash: tokenHash,
		UserID:    userID,
		ClientID:  clientID,
		FamilyID:  familyID,
		Scopes:    scopes,
		ExpiresAt: time.Now().Add(u.tokenCfg.RefreshTokenExpiry),
	}

	if err := u.tokenRepo.Create(ctx, refreshToken); err != nil {
		return "", "", err
	}

	return accessToken, rawToken, nil
}

func (u *TokenUsecase) RefreshTokens(ctx context.Context, rawToken string) (string, string, error) {
	tokenHash := hashToken(rawToken)

	oldToken, err := u.tokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return "", "", domain.ErrInvalidToken
	}

	user, err := u.userRepo.GetByID(ctx, oldToken.UserID)
	if err != nil {
		return "", "", err
	}

	if user.TokensRevokedAt != nil && user.TokensRevokedAt.After(oldToken.CreatedAt) {
		return "", "", domain.ErrTokenRevoked
	}

	newRawToken, err := u.random.GenerateToken(32)
	if err != nil {
		return "", "", err
	}

	newTokenHash := hashToken(newRawToken)

	newRefreshToken := &domain.RefreshToken{
		ID:        u.random.GenerateUUID(),
		TokenHash: newTokenHash,
		UserID:    oldToken.UserID,
		ClientID:  oldToken.ClientID,
		FamilyID:  oldToken.FamilyID,
		Scopes:    oldToken.Scopes,
		ExpiresAt: time.Now().Add(u.tokenCfg.RefreshTokenExpiry),
	}

	if err := u.txRunner.RunInTx(ctx, func(ctx context.Context) error {
		if err := u.tokenRepo.RevokeByID(ctx, oldToken.ID); err != nil {
			return err
		}
		return u.tokenRepo.Create(ctx, newRefreshToken)
	}); err != nil {
		return "", "", err
	}

	claims := map[string]any{
		"sub":   user.ID,
		"email": user.Email,
		"scope": strings.Join(oldToken.Scopes, " "),
	}

	accessToken, err := u.jwt.GenerateAccessToken(ctx, claims)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRawToken, nil
}

func (u *TokenUsecase) RevokeToken(ctx context.Context, rawToken string) error {
	tokenHash := hashToken(rawToken)

	token, err := u.tokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return err
	}

	return u.tokenRepo.RevokeByID(ctx, token.ID)
}

func (u *TokenUsecase) RevokeUserTokens(ctx context.Context, userID string) error {
	return u.tokenRepo.RevokeByUserID(ctx, userID)
}

func (u *TokenUsecase) AccessTokenExpiry() time.Duration {
	return u.tokenCfg.AccessTokenExpiry
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
