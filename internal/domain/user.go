package domain

import "time"

type UserStatus string

const (
	UserStatusUnverified UserStatus = "unverified"
	UserStatusActive     UserStatus = "active"
	UserStatusLocked     UserStatus = "locked"
	UserStatusDeleted    UserStatus = "deleted"
)

type User struct {
	ID                  string
	Email               string
	PasswordHash        string
	Status              UserStatus
	EmailVerified       bool
	EmailVerifyToken    string
	EmailVerifyExpiry   *time.Time
	PasswordResetToken  string
	PasswordResetExpiry *time.Time
	TOTPSecret          string
	TOTPEnabled         bool
	RecoveryCodes       []string
	TokensRevokedAt     *time.Time
	Version             int
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           *time.Time
}
