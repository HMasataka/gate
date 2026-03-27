package handler

import (
	"net/http"
	"time"

	"github.com/HMasataka/gate/internal/domain"
	"github.com/HMasataka/gate/internal/usecase"
)

type AdminUserHandler struct {
	user  *usecase.UserUsecase
	audit *usecase.AuditUsecase
}

func NewAdminUserHandler(user *usecase.UserUsecase, audit *usecase.AuditUsecase) *AdminUserHandler {
	return &AdminUserHandler{user: user, audit: audit}
}

// List handles GET /api/v1/admin/users
func (h *AdminUserHandler) List(w http.ResponseWriter, r *http.Request) {
	offset, limit := Paginate(r)

	users, total, err := h.user.List(r.Context(), offset, limit)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]any{
		"users":  users,
		"total":  total,
		"offset": offset,
		"limit":  limit,
	})
}

// Get handles GET /api/v1/admin/users/{id}
func (h *AdminUserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := PathParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}

	user, err := h.user.Get(r.Context(), id)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, user)
}

type updateUserRequest struct {
	Email string `json:"email"`
}

// Update handles PUT /api/v1/admin/users/{id}
func (h *AdminUserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := PathParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}

	var req updateUserRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.Email == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "email is required")
		return
	}

	user, err := h.user.Update(r.Context(), id, req.Email)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, user)
}

// Delete handles DELETE /api/v1/admin/users/{id}
func (h *AdminUserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := PathParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}

	if err := h.user.SoftDelete(r.Context(), id); err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "user deleted"})
}

// Lock handles POST /api/v1/admin/users/{id}/lock
func (h *AdminUserHandler) Lock(w http.ResponseWriter, r *http.Request) {
	id := PathParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}

	if err := h.user.Lock(r.Context(), id); err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "user locked"})
}

// Unlock handles POST /api/v1/admin/users/{id}/unlock
func (h *AdminUserHandler) Unlock(w http.ResponseWriter, r *http.Request) {
	id := PathParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}

	if err := h.user.Unlock(r.Context(), id); err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "user unlocked"})
}

// ResetMFA handles POST /api/v1/admin/users/{id}/reset-mfa
func (h *AdminUserHandler) ResetMFA(w http.ResponseWriter, r *http.Request) {
	id := PathParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}

	if err := h.user.ResetMFA(r.Context(), id); err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "mfa reset"})
}

// RevokeAllTokens handles POST /api/v1/admin/users/{id}/revoke-tokens
func (h *AdminUserHandler) RevokeAllTokens(w http.ResponseWriter, r *http.Request) {
	id := PathParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}

	if err := h.user.RevokeAllTokens(r.Context(), id); err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "tokens revoked"})
}

// ListAuditLogs handles GET /api/v1/admin/audit-logs
func (h *AdminUserHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	offset, limit := Paginate(r)

	filter := domain.AuditLogFilter{
		Offset: offset,
		Limit:  limit,
	}

	if v := QueryParam(r, "user_id"); v != "" {
		filter.UserID = &v
	}
	if v := QueryParam(r, "action"); v != "" {
		action := domain.AuditAction(v)
		filter.Action = &action
	}
	if v := QueryParam(r, "from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.From = &t
		}
	}
	if v := QueryParam(r, "to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.To = &t
		}
	}

	logs, total, err := h.audit.List(r.Context(), filter)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]any{
		"audit_logs": logs,
		"total":      total,
		"offset":     offset,
		"limit":      limit,
	})
}
