package handler

import (
	"net/http"

	"github.com/HMasataka/gate/internal/usecase"
)

type AdminClientHandler struct {
	client *usecase.ClientUsecase
}

func NewAdminClientHandler(client *usecase.ClientUsecase) *AdminClientHandler {
	return &AdminClientHandler{client: client}
}

type createClientRequest struct {
	Name          string   `json:"name"`
	ClientType    string   `json:"client_type"`
	RedirectURIs  []string `json:"redirect_uris"`
	AllowedScopes []string `json:"allowed_scopes"`
	GrantTypes    []string `json:"grant_types"`
}

// Create handles POST /api/v1/admin/clients
func (h *AdminClientHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createClientRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.Name == "" || req.ClientType == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "name and client_type are required")
		return
	}

	client, err := h.client.Register(r.Context(), req.Name, req.ClientType, req.RedirectURIs, req.AllowedScopes, req.GrantTypes)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusCreated, client)
}

// Get handles GET /api/v1/admin/clients/{id}
func (h *AdminClientHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := PathParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}

	client, err := h.client.Get(r.Context(), id)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, client)
}

// List handles GET /api/v1/admin/clients
func (h *AdminClientHandler) List(w http.ResponseWriter, r *http.Request) {
	offset, limit := Paginate(r)

	clients, total, err := h.client.List(r.Context(), offset, limit)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]any{
		"clients": clients,
		"total":   total,
		"offset":  offset,
		"limit":   limit,
	})
}

type updateClientRequest struct {
	Name          string   `json:"name"`
	RedirectURIs  []string `json:"redirect_uris"`
	AllowedScopes []string `json:"allowed_scopes"`
}

// Update handles PUT /api/v1/admin/clients/{id}
func (h *AdminClientHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := PathParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}

	var req updateClientRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.Name == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "name is required")
		return
	}

	client, err := h.client.Update(r.Context(), id, req.Name, req.RedirectURIs, req.AllowedScopes)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, client)
}

// Delete handles DELETE /api/v1/admin/clients/{id}
func (h *AdminClientHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := PathParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}

	if err := h.client.Delete(r.Context(), id); err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "client deleted"})
}

// RotateSecret handles POST /api/v1/admin/clients/{id}/rotate-secret
func (h *AdminClientHandler) RotateSecret(w http.ResponseWriter, r *http.Request) {
	id := PathParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}

	client, err := h.client.RotateSecret(r.Context(), id)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, client)
}
