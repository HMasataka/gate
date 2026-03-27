package usecase

import (
	"context"
	"time"

	"github.com/HMasataka/gate/internal/domain"
)

type RoleUsecase struct {
	roleRepo domain.RoleRepository
	permRepo domain.PermissionRepository
	resolver domain.PermissionResolver
	random   domain.RandomGenerator
}

func NewRoleUsecase(
	roleRepo domain.RoleRepository,
	permRepo domain.PermissionRepository,
	resolver domain.PermissionResolver,
	random domain.RandomGenerator,
) *RoleUsecase {
	return &RoleUsecase{
		roleRepo: roleRepo,
		permRepo: permRepo,
		resolver: resolver,
		random:   random,
	}
}

func (u *RoleUsecase) CreateRole(ctx context.Context, name, description, parentID string) (*domain.Role, error) {
	var parent *string

	if parentID != "" {
		if _, err := u.roleRepo.GetByID(ctx, parentID); err != nil {
			return nil, err
		}

		id := u.random.GenerateUUID()
		hasCycle, err := u.roleRepo.DetectCycle(ctx, id, parentID)
		if err != nil {
			return nil, err
		}
		if hasCycle {
			return nil, domain.ErrCyclicRole
		}

		parent = &parentID
	}

	now := time.Now()
	role := &domain.Role{
		ID:          u.random.GenerateUUID(),
		Name:        name,
		Description: description,
		ParentID:    parent,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := u.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}

	return role, nil
}

func (u *RoleUsecase) GetRole(ctx context.Context, id string) (*domain.Role, error) {
	return u.roleRepo.GetByID(ctx, id)
}

func (u *RoleUsecase) UpdateRole(ctx context.Context, id, name, description, parentID string) (*domain.Role, error) {
	role, err := u.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if parentID != "" {
		if _, err := u.roleRepo.GetByID(ctx, parentID); err != nil {
			return nil, err
		}

		hasCycle, err := u.roleRepo.DetectCycle(ctx, id, parentID)
		if err != nil {
			return nil, err
		}
		if hasCycle {
			return nil, domain.ErrCyclicRole
		}

		role.ParentID = &parentID
	} else {
		role.ParentID = nil
	}

	role.Name = name
	role.Description = description
	role.UpdatedAt = time.Now()

	if err := u.roleRepo.Update(ctx, role); err != nil {
		return nil, err
	}

	return role, nil
}

func (u *RoleUsecase) DeleteRole(ctx context.Context, id string) error {
	return u.roleRepo.Delete(ctx, id)
}

func (u *RoleUsecase) ListRoles(ctx context.Context, offset, limit int) ([]domain.Role, int, error) {
	return u.roleRepo.List(ctx, offset, limit)
}

func (u *RoleUsecase) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	if err := u.roleRepo.AssignToUser(ctx, userID, roleID); err != nil {
		return err
	}

	return u.resolver.Invalidate(ctx, userID)
}

func (u *RoleUsecase) RemoveRoleFromUser(ctx context.Context, userID, roleID string) error {
	if err := u.roleRepo.RemoveFromUser(ctx, userID, roleID); err != nil {
		return err
	}

	return u.resolver.Invalidate(ctx, userID)
}

func (u *RoleUsecase) GetUserRoles(ctx context.Context, userID string) ([]domain.Role, error) {
	return u.roleRepo.GetUserRoles(ctx, userID)
}
