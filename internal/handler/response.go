package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/HMasataka/gate/internal/domain"
)

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func Error(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, map[string]any{
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	})
}

func HandleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		Error(w, http.StatusNotFound, "not_found", "resource not found")
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrDuplicateEmail):
		Error(w, http.StatusConflict, "conflict", err.Error())
	case errors.Is(err, domain.ErrOptimisticLock):
		Error(w, http.StatusConflict, "conflict", "resource was modified by another request")
	case errors.Is(err, domain.ErrInvalidCredentials):
		Error(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
	case errors.Is(err, domain.ErrAccountLocked):
		Error(w, http.StatusForbidden, "account_locked", "account is locked")
	case errors.Is(err, domain.ErrAccountUnverified):
		Error(w, http.StatusForbidden, "account_unverified", "email not verified")
	case errors.Is(err, domain.ErrAccountDeleted):
		Error(w, http.StatusForbidden, "account_deleted", "account has been deleted")
	case errors.Is(err, domain.ErrTokenExpired):
		Error(w, http.StatusUnauthorized, "token_expired", "token has expired")
	case errors.Is(err, domain.ErrTokenRevoked):
		Error(w, http.StatusUnauthorized, "token_revoked", "token has been revoked")
	case errors.Is(err, domain.ErrTokenReused):
		Error(w, http.StatusUnauthorized, "token_reused", "token reuse detected")
	case errors.Is(err, domain.ErrInvalidToken):
		Error(w, http.StatusUnauthorized, "invalid_token", "invalid token")
	case errors.Is(err, domain.ErrMFARequired):
		Error(w, http.StatusForbidden, "mfa_required", "multi-factor authentication required")
	case errors.Is(err, domain.ErrInvalidMFACode):
		Error(w, http.StatusUnauthorized, "invalid_mfa_code", "invalid MFA code")
	case errors.Is(err, domain.ErrRateLimited):
		Error(w, http.StatusTooManyRequests, "rate_limited", "too many requests")
	case errors.Is(err, domain.ErrForbidden):
		Error(w, http.StatusForbidden, "forbidden", "access denied")
	case errors.Is(err, domain.ErrInvalidRedirectURI):
		Error(w, http.StatusBadRequest, "invalid_redirect_uri", "invalid redirect URI")
	case errors.Is(err, domain.ErrInvalidScope):
		Error(w, http.StatusBadRequest, "invalid_scope", "invalid scope")
	case errors.Is(err, domain.ErrInvalidGrantType):
		Error(w, http.StatusBadRequest, "invalid_grant", "unsupported grant type")
	case errors.Is(err, domain.ErrInvalidClient):
		Error(w, http.StatusUnauthorized, "invalid_client", "invalid client credentials")
	case errors.Is(err, domain.ErrCodeReuse):
		Error(w, http.StatusBadRequest, "invalid_grant", "authorization code has been used")
	default:
		Error(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
