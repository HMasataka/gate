package domain

import "time"

type AuditAction string

const (
	AuditActionLogin            AuditAction = "login"
	AuditActionLogout           AuditAction = "logout"
	AuditActionLoginFailed      AuditAction = "login_failed"
	AuditActionRegister         AuditAction = "register"
	AuditActionPasswordChange   AuditAction = "password_change"
	AuditActionPasswordReset    AuditAction = "password_reset"
	AuditActionTokenIssue       AuditAction = "token_issue"
	AuditActionTokenRevoke      AuditAction = "token_revoke"
	AuditActionRoleChange       AuditAction = "role_change"
	AuditActionPermissionChange AuditAction = "permission_change"
	AuditActionMFASetup         AuditAction = "mfa_setup"
	AuditActionMFADisable       AuditAction = "mfa_disable"
	AuditActionAdminAction      AuditAction = "admin_action"
	AuditActionSocialLogin      AuditAction = "social_login"
)

type AuditLog struct {
	ID        string
	UserID    *string
	Action    AuditAction
	IPAddress string
	UserAgent string
	Metadata  map[string]any
	CreatedAt time.Time
}
