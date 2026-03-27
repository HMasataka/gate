package usecase

import (
	"context"
	"time"

	"github.com/HMasataka/gate/internal/domain"
)

type PermissionUsecase struct {
	permRepo domain.PermissionRepository
	resolver domain.PermissionResolver
	random   domain.RandomGenerator
}

func NewPermissionUsecase(
	permRepo domain.PermissionRepository,
	resolver domain.PermissionResolver,
	random domain.RandomGenerator,
) *PermissionUsecase {
	return &PermissionUsecase{
		permRepo: permRepo,
		resolver: resolver,
		random:   random,
	}
}

func (u *PermissionUsecase) CreatePermission(ctx context.Context, name, description string) (*domain.Permission, error) {
	now := time.Now()
	perm := &domain.Permission{
		ID:          u.random.GenerateUUID(),
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := u.permRepo.Create(ctx, perm); err != nil {
		return nil, err
	}

	return perm, nil
}

func (u *PermissionUsecase) GetPermission(ctx context.Context, id string) (*domain.Permission, error) {
	return u.permRepo.GetByID(ctx, id)
}

func (u *PermissionUsecase) UpdatePermission(ctx context.Context, id, name, description string) (*domain.Permission, error) {
	perm, err := u.permRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	perm.Name = name
	perm.Description = description
	perm.UpdatedAt = time.Now()

	if err := u.permRepo.Update(ctx, perm); err != nil {
		return nil, err
	}

	return perm, nil
}

func (u *PermissionUsecase) DeletePermission(ctx context.Context, id string) error {
	return u.permRepo.Delete(ctx, id)
}

func (u *PermissionUsecase) ListPermissions(ctx context.Context, offset, limit int) ([]domain.Permission, int, error) {
	return u.permRepo.List(ctx, offset, limit)
}

func (u *PermissionUsecase) AssignPermissionToRole(ctx context.Context, roleID, permID string) error {
	return u.permRepo.AssignToRole(ctx, roleID, permID)
}

func (u *PermissionUsecase) RemovePermissionFromRole(ctx context.Context, roleID, permID string) error {
	return u.permRepo.RemoveFromRole(ctx, roleID, permID)
}

func (u *PermissionUsecase) AssignPermissionToUser(ctx context.Context, userID, permID string) error {
	if err := u.permRepo.AssignToUser(ctx, userID, permID); err != nil {
		return err
	}

	return u.resolver.Invalidate(ctx, userID)
}

func (u *PermissionUsecase) RemovePermissionFromUser(ctx context.Context, userID, permID string) error {
	if err := u.permRepo.RemoveFromUser(ctx, userID, permID); err != nil {
		return err
	}

	return u.resolver.Invalidate(ctx, userID)
}

func (u *PermissionUsecase) ResolveForUser(ctx context.Context, userID string) ([]string, error) {
	return u.resolver.Resolve(ctx, userID)
}
