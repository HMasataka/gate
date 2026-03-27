package domain

import (
	"context"
	"time"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	SoftDelete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]User, int, error)
	ListPendingPurge(ctx context.Context, before time.Time) ([]User, error)
	HardDelete(ctx context.Context, id string) error
}

type OAuthClientRepository interface {
	Create(ctx context.Context, client *OAuthClient) error
	GetByID(ctx context.Context, id string) (*OAuthClient, error)
	Update(ctx context.Context, client *OAuthClient) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]OAuthClient, int, error)
}

type RoleRepository interface {
	Create(ctx context.Context, role *Role) error
	GetByID(ctx context.Context, id string) (*Role, error)
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]Role, int, error)
	AssignToUser(ctx context.Context, userID, roleID string) error
	RemoveFromUser(ctx context.Context, userID, roleID string) error
	GetUserRoles(ctx context.Context, userID string) ([]Role, error)
	DetectCycle(ctx context.Context, roleID, parentID string) (bool, error)
}

type PermissionRepository interface {
	Create(ctx context.Context, perm *Permission) error
	GetByID(ctx context.Context, id string) (*Permission, error)
	Update(ctx context.Context, perm *Permission) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]Permission, int, error)
	AssignToRole(ctx context.Context, roleID, permID string) error
	RemoveFromRole(ctx context.Context, roleID, permID string) error
	AssignToUser(ctx context.Context, userID, permID string) error
	RemoveFromUser(ctx context.Context, userID, permID string) error
	ResolveForUser(ctx context.Context, userID string) ([]string, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	GetByTokenHash(ctx context.Context, hash string) (*RefreshToken, error)
	RevokeByID(ctx context.Context, id string) error
	RevokeByFamilyID(ctx context.Context, familyID string) error
	RevokeByUserID(ctx context.Context, userID string) error
	RevokeByClientID(ctx context.Context, clientID string) error
	DeleteExpired(ctx context.Context) (int64, error)
}

type AuthorizationCodeRepository interface {
	Create(ctx context.Context, code *AuthorizationCode) error
	GetByCode(ctx context.Context, codeHash string) (*AuthorizationCode, error)
	MarkUsed(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context) (int64, error)
}

type SocialConnectionRepository interface {
	Create(ctx context.Context, conn *SocialConnection) error
	GetByProviderAndProviderUserID(ctx context.Context, provider, providerUserID string) (*SocialConnection, error)
	GetByUserID(ctx context.Context, userID string) ([]SocialConnection, error)
	Update(ctx context.Context, conn *SocialConnection) error
	Delete(ctx context.Context, id string) error
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	List(ctx context.Context, filter AuditLogFilter) ([]AuditLog, int, error)
	DeleteBefore(ctx context.Context, before time.Time) (int64, error)
}

type AuditLogFilter struct {
	UserID *string
	Action *AuditAction
	From   *time.Time
	To     *time.Time
	Offset int
	Limit  int
}
