package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/HMasataka/gate/internal/config"
	"github.com/HMasataka/gate/internal/domain"
)

type AuthUsecase struct {
	userRepo domain.UserRepository
	hasher   domain.PasswordHasher
	mailer   domain.Mailer
	sessions domain.SessionStore
	random   domain.RandomGenerator
	authCfg  config.AuthConfig
	sessCfg  config.SessionConfig
}

func NewAuthUsecase(
	userRepo domain.UserRepository,
	hasher domain.PasswordHasher,
	mailer domain.Mailer,
	sessions domain.SessionStore,
	random domain.RandomGenerator,
	authCfg config.AuthConfig,
	sessCfg config.SessionConfig,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo: userRepo,
		hasher:   hasher,
		mailer:   mailer,
		sessions: sessions,
		random:   random,
		authCfg:  authCfg,
		sessCfg:  sessCfg,
	}
}

func (u *AuthUsecase) validatePassword(password string) error {
	if len(password) < u.authCfg.PasswordMinLength {
		return domain.ErrPasswordTooShort
	}

	if len(password) > u.authCfg.PasswordMaxLength {
		return domain.ErrPasswordTooLong
	}

	return nil
}

func (u *AuthUsecase) Register(ctx context.Context, email, password string) (*domain.User, error) {
	if err := u.validatePassword(password); err != nil {
		return nil, err
	}

	hash, err := u.hasher.Hash(ctx, password)
	if err != nil {
		return nil, err
	}

	token, err := u.random.GenerateToken(32)
	if err != nil {
		return nil, err
	}

	expiry := time.Now().Add(u.authCfg.EmailVerificationExpiry)
	now := time.Now()

	user := &domain.User{
		ID:               u.random.GenerateUUID(),
		Email:            email,
		PasswordHash:     hash,
		Status:           domain.UserStatusUnverified,
		EmailVerified:    false,
		EmailVerifyToken: token,
		EmailVerifyExpiry: &expiry,
		Version:          1,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	if err := u.mailer.SendEmailVerification(ctx, email, token); err != nil {
		slog.ErrorContext(ctx, "failed to send email verification", slog.String("email", email), slog.Any("error", err))
	}

	return user, nil
}

func (u *AuthUsecase) Login(ctx context.Context, email, password, ipAddress, userAgent string) (*domain.Session, error) {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	switch user.Status {
	case domain.UserStatusLocked:
		return nil, domain.ErrAccountLocked
	case domain.UserStatusDeleted:
		return nil, domain.ErrAccountDeleted
	}

	match, err := u.hasher.Compare(ctx, password, user.PasswordHash)
	if err != nil {
		return nil, err
	}

	if !match {
		return nil, domain.ErrInvalidCredentials
	}

	session := &domain.Session{
		ID:        u.random.GenerateUUID(),
		UserID:    user.ID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	if err := u.sessions.Create(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

func (u *AuthUsecase) Logout(ctx context.Context, sessionID string) error {
	return u.sessions.Delete(ctx, sessionID)
}

func (u *AuthUsecase) VerifyEmail(ctx context.Context, email, token string) error {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}

	if user.EmailVerifyToken != token {
		return domain.ErrInvalidToken
	}

	if user.EmailVerifyExpiry != nil && time.Now().After(*user.EmailVerifyExpiry) {
		return domain.ErrTokenExpired
	}

	user.EmailVerified = true
	user.Status = domain.UserStatusActive
	user.EmailVerifyToken = ""
	user.EmailVerifyExpiry = nil
	user.UpdatedAt = time.Now()

	return u.userRepo.Update(ctx, user)
}

func (u *AuthUsecase) ResendVerification(ctx context.Context, email string) error {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil
	}

	if user.EmailVerified {
		return nil
	}

	token, err := u.random.GenerateToken(32)
	if err != nil {
		return err
	}

	expiry := time.Now().Add(u.authCfg.EmailVerificationExpiry)

	user.EmailVerifyToken = token
	user.EmailVerifyExpiry = &expiry
	user.UpdatedAt = time.Now()

	if err := u.userRepo.Update(ctx, user); err != nil {
		return err
	}

	if err := u.mailer.SendEmailVerification(ctx, email, token); err != nil {
		slog.ErrorContext(ctx, "failed to send email verification", slog.String("email", email), slog.Any("error", err))
	}

	return nil
}

func (u *AuthUsecase) ForgotPassword(ctx context.Context, email string) error {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil
	}

	token, err := u.random.GenerateToken(32)
	if err != nil {
		return err
	}

	expiry := time.Now().Add(u.authCfg.PasswordResetExpiry)

	user.PasswordResetToken = token
	user.PasswordResetExpiry = &expiry
	user.UpdatedAt = time.Now()

	if err := u.userRepo.Update(ctx, user); err != nil {
		return err
	}

	if err := u.mailer.SendPasswordReset(ctx, email, token); err != nil {
		slog.ErrorContext(ctx, "failed to send password reset", slog.String("email", email), slog.Any("error", err))
	}

	return nil
}

func (u *AuthUsecase) ResetPassword(ctx context.Context, email, token, newPassword string) error {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}

	if user.PasswordResetToken != token {
		return domain.ErrInvalidToken
	}

	if user.PasswordResetExpiry != nil && time.Now().After(*user.PasswordResetExpiry) {
		return domain.ErrTokenExpired
	}

	if err := u.validatePassword(newPassword); err != nil {
		return err
	}

	hash, err := u.hasher.Hash(ctx, newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hash
	user.PasswordResetToken = ""
	user.PasswordResetExpiry = nil
	user.UpdatedAt = time.Now()

	return u.userRepo.Update(ctx, user)
}

func (u *AuthUsecase) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	match, err := u.hasher.Compare(ctx, currentPassword, user.PasswordHash)
	if err != nil {
		return err
	}

	if !match {
		return domain.ErrInvalidCredentials
	}

	if err := u.validatePassword(newPassword); err != nil {
		return err
	}

	hash, err := u.hasher.Hash(ctx, newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hash
	user.UpdatedAt = time.Now()

	return u.userRepo.Update(ctx, user)
}
