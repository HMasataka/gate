package handler

import (
	"net/http"

	"github.com/HMasataka/gate/internal/usecase"
)

// SocialHandler handles social OAuth2 login flows.
type SocialHandler struct {
	social *usecase.SocialUsecase
}

// NewSocialHandler creates a new SocialHandler.
func NewSocialHandler(social *usecase.SocialUsecase) *SocialHandler {
	return &SocialHandler{social: social}
}

// Authorize handles GET /api/v1/auth/social/{provider}/authorize
// It redirects the user to the provider's authorization page.
func (h *SocialHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	provider := PathParam(r, "provider")
	if provider == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "provider is required")
		return
	}

	// Use a random UUID as the state parameter to prevent CSRF.
	// In production this should be stored in the session for verification on callback.
	state := QueryParam(r, "state")
	if state == "" {
		state = "default"
	}

	authURL, err := h.social.GetAuthURL(provider, state)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid_provider", err.Error())
		return
	}

	http.Redirect(w, r, authURL, http.StatusFound)
}

// Callback handles GET /api/v1/auth/social/{provider}/callback
// It exchanges the authorization code for user info and returns session tokens.
func (h *SocialHandler) Callback(w http.ResponseWriter, r *http.Request) {
	provider := PathParam(r, "provider")
	if provider == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "provider is required")
		return
	}

	code := QueryParam(r, "code")
	if code == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "code is required")
		return
	}

	result, err := h.social.HandleCallback(r.Context(), provider, code, r.RemoteAddr, r.UserAgent())
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]any{
		"session_id":    result.Session.ID,
		"expires_at":    result.Session.ExpiresAt,
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
	})
}
