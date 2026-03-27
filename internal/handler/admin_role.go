package handler

import (
	"net/http"

	"github.com/HMasataka/gate/internal/usecase"
)

type AdminRoleHandler struct {
	role       *usecase.RoleUsecase
	permission *usecase.PermissionUsecase
}

func NewAdminRoleHandler(role *usecase.RoleUsecase, permission *usecase.PermissionUsecase) *AdminRoleHandler {
	return &AdminRoleHandler{role: role, permission: permission}
}

// --- Role request types ---

type createRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ParentID    string `json:"parent_id"`
}

type updateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ParentID    string `json:"parent_id"`
}

// --- Permission request types ---

type createPermissionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type updatePermissionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateRole handles POST /api/v1/admin/roles
func (h *AdminRoleHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req createRoleRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.Name == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "name is required")
		return
	}

	role, err := h.role.CreateRole(r.Context(), req.Name, req.Description, req.ParentID)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusCreated, role)
}

// GetRole handles GET /api/v1/admin/roles/{id}
func (h *AdminRoleHandler) GetRole(w http.ResponseWriter, r *http.Request) {
	id := PathParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}

	role, err := h.role.GetRole(r.Context(), id)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, role)
}

// UpdateRole handles PUT /api/v1/admin/roles/{id}
func (h *AdminRoleHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	id := PathParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}

	var req updateRoleRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.Name == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "name is required")
		return
	}

	role, err := h.role.UpdateRole(r.Context(), id, req.Name, req.Description, req.ParentID)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, role)
}

// DeleteRole handles DELETE /api/v1/admin/roles/{id}
func (h *AdminRoleHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	id := PathParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}

	if err := h.role.DeleteRole(r.Context(), id); err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "role deleted"})
}

// ListRoles handles GET /api/v1/admin/roles
func (h *AdminRoleHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	offset, limit := Paginate(r)

	roles, total, err := h.role.ListRoles(r.Context(), offset, limit)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]any{
		"roles":  roles,
		"total":  total,
		"offset": offset,
		"limit":  limit,
	})
}

// AssignRoleToUser handles POST /api/v1/admin/roles/{id}/users/{userID}
func (h *AdminRoleHandler) AssignRoleToUser(w http.ResponseWriter, r *http.Request) {
	roleID := PathParam(r, "id")
	userID := PathParam(r, "userID")

	if roleID == "" || userID == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "id and userID are required")
		return
	}

	if err := h.role.AssignRoleToUser(r.Context(), userID, roleID); err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "role assigned"})
}

// RemoveRoleFromUser handles DELETE /api/v1/admin/roles/{id}/users/{userID}
func (h *AdminRoleHandler) RemoveRoleFromUser(w http.ResponseWriter, r *http.Request) {
	roleID := PathParam(r, "id")
	userID := PathParam(r, "userID")

	if roleID == "" || userID == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "id and userID are required")
		return
	}

	if err := h.role.RemoveRoleFromUser(r.Context(), userID, roleID); err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "role removed"})
}

// CreatePermission handles POST /api/v1/admin/permissions
func (h *AdminRoleHandler) CreatePermission(w http.ResponseWriter, r *http.Request) {
	var req createPermissionRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.Name == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "name is required")
		return
	}

	perm, err := h.permission.CreatePermission(r.Context(), req.Name, req.Description)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusCreated, perm)
}

// GetPermission handles GET /api/v1/admin/permissions/{id}
func (h *AdminRoleHandler) GetPermission(w http.ResponseWriter, r *http.Request) {
	id := PathParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}

	perm, err := h.permission.GetPermission(r.Context(), id)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, perm)
}

// UpdatePermission handles PUT /api/v1/admin/permissions/{id}
func (h *AdminRoleHandler) UpdatePermission(w http.ResponseWriter, r *http.Request) {
	id := PathParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}

	var req updatePermissionRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.Name == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "name is required")
		return
	}

	perm, err := h.permission.UpdatePermission(r.Context(), id, req.Name, req.Description)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, perm)
}

// DeletePermission handles DELETE /api/v1/admin/permissions/{id}
func (h *AdminRoleHandler) DeletePermission(w http.ResponseWriter, r *http.Request) {
	id := PathParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}

	if err := h.permission.DeletePermission(r.Context(), id); err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "permission deleted"})
}

// ListPermissions handles GET /api/v1/admin/permissions
func (h *AdminRoleHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	offset, limit := Paginate(r)

	permissions, total, err := h.permission.ListPermissions(r.Context(), offset, limit)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]any{
		"permissions": permissions,
		"total":       total,
		"offset":      offset,
		"limit":       limit,
	})
}

// AssignPermissionToRole handles POST /api/v1/admin/roles/{id}/permissions/{permID}
func (h *AdminRoleHandler) AssignPermissionToRole(w http.ResponseWriter, r *http.Request) {
	roleID := PathParam(r, "id")
	permID := PathParam(r, "permID")

	if roleID == "" || permID == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "id and permID are required")
		return
	}

	if err := h.permission.AssignPermissionToRole(r.Context(), roleID, permID); err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "permission assigned to role"})
}

// RemovePermissionFromRole handles DELETE /api/v1/admin/roles/{id}/permissions/{permID}
func (h *AdminRoleHandler) RemovePermissionFromRole(w http.ResponseWriter, r *http.Request) {
	roleID := PathParam(r, "id")
	permID := PathParam(r, "permID")

	if roleID == "" || permID == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "id and permID are required")
		return
	}

	if err := h.permission.RemovePermissionFromRole(r.Context(), roleID, permID); err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "permission removed from role"})
}
