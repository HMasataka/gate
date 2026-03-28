package handler

import (
	"net/http"

	"github.com/HMasataka/gate/internal/middleware"
	"github.com/HMasataka/gate/internal/usecase"
)

type ClientHandler struct {
	client *usecase.ClientUsecase
}

func NewClientHandler(client *usecase.ClientUsecase) *ClientHandler {
	return &ClientHandler{client: client}
}

type userCreateClientRequest struct {
	Name          string   `json:"name"`
	ClientType    string   `json:"client_type"`
	RedirectURIs  []string `json:"redirect_uris"`
	AllowedScopes []string `json:"allowed_scopes"`
	GrantTypes    []string `json:"grant_types"`
}

// Create handles POST /api/v1/clients
func (h *ClientHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req userCreateClientRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.Name == "" || req.ClientType == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "name and client_type are required")
		return
	}

	client, err := h.client.RegisterForUser(r.Context(), userID, req.Name, req.ClientType, req.RedirectURIs, req.AllowedScopes, req.GrantTypes)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusCreated, client)
}

// List handles GET /api/v1/clients
func (h *ClientHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	offset, limit := Paginate(r)

	clients, total, err := h.client.ListByOwner(r.Context(), userID, offset, limit)
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

// Get handles GET /api/v1/clients/{id}
func (h *ClientHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	id := PathParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}

	client, err := h.client.GetOwned(r.Context(), userID, id)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, client)
}

type userUpdateClientRequest struct {
	Name          string   `json:"name"`
	RedirectURIs  []string `json:"redirect_uris"`
	AllowedScopes []string `json:"allowed_scopes"`
}

// Update handles PUT /api/v1/clients/{id}
func (h *ClientHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	id := PathParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}

	var req userUpdateClientRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.Name == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "name is required")
		return
	}

	client, err := h.client.UpdateOwned(r.Context(), userID, id, req.Name, req.RedirectURIs, req.AllowedScopes)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, client)
}

// Delete handles DELETE /api/v1/clients/{id}
func (h *ClientHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	id := PathParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}

	if err := h.client.DeleteOwned(r.Context(), userID, id); err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "client deleted"})
}

// RotateSecret handles POST /api/v1/clients/{id}/rotate-secret
func (h *ClientHandler) RotateSecret(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	id := PathParam(r, "id")
	if id == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "id is required")
		return
	}

	client, err := h.client.RotateSecretOwned(r.Context(), userID, id)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, client)
}
