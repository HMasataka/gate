package handler

import (
	"net/http"

	"github.com/HMasataka/gate/internal/middleware"
	"github.com/HMasataka/gate/internal/usecase"
)

type OIDCHandler struct {
	oidc *usecase.OIDCUsecase
}

func NewOIDCHandler(oidc *usecase.OIDCUsecase) *OIDCHandler {
	return &OIDCHandler{oidc: oidc}
}

// Discovery handles GET /.well-known/openid-configuration
func (h *OIDCHandler) Discovery(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, h.oidc.Discovery())
}

// UserInfo handles GET /oauth/userinfo (JWT auth required)
func (h *OIDCHandler) UserInfo(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	claims, err := h.oidc.UserInfo(r.Context(), userID)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, claims)
}
