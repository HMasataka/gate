package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/HMasataka/gate/internal/domain"
)

type UserUsecase struct {
	userRepo domain.UserRepository
	sessions domain.SessionStore
	token    *TokenUsecase
	random   domain.RandomGenerator
}

func NewUserUsecase(
	userRepo domain.UserRepository,
	sessions domain.SessionStore,
	token *TokenUsecase,
	random domain.RandomGenerator,
) *UserUsecase {
	return &UserUsecase{
		userRepo: userRepo,
		sessions: sessions,
		token:    token,
		random:   random,
	}
}

func (u *UserUsecase) List(ctx context.Context, offset, limit int) ([]domain.User, int, error) {
	return u.userRepo.List(ctx, offset, limit)
}

func (u *UserUsecase) Get(ctx context.Context, id string) (*domain.User, error) {
	return u.userRepo.GetByID(ctx, id)
}

func (u *UserUsecase) Update(ctx context.Context, id, email string) (*domain.User, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.Email = email
	user.UpdatedAt = time.Now()

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *UserUsecase) SoftDelete(ctx context.Context, id string) error {
	return u.userRepo.SoftDelete(ctx, id)
}

func (u *UserUsecase) Lock(ctx context.Context, id string) error {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	user.Status = domain.UserStatusLocked
	user.UpdatedAt = time.Now()

	return u.userRepo.Update(ctx, user)
}

func (u *UserUsecase) Unlock(ctx context.Context, id string) error {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	user.Status = domain.UserStatusActive
	user.UpdatedAt = time.Now()

	return u.userRepo.Update(ctx, user)
}

func (u *UserUsecase) ResetMFA(ctx context.Context, id string) error {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	user.TOTPSecret = ""
	user.TOTPEnabled = false
	user.RecoveryCodes = nil
	user.UpdatedAt = time.Now()

	return u.userRepo.Update(ctx, user)
}

func (u *UserUsecase) RevokeAllTokens(ctx context.Context, id string) error {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	user.TokensRevokedAt = &now
	user.UpdatedAt = now

	if err := u.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("update user tokens_revoked_at: %w", err)
	}

	if err := u.sessions.DeleteByUserID(ctx, id); err != nil {
		return fmt.Errorf("delete user sessions: %w", err)
	}

	if u.token != nil {
		if err := u.token.RevokeUserTokens(ctx, id); err != nil {
			return fmt.Errorf("revoke user refresh tokens: %w", err)
		}
	}

	return nil
}
