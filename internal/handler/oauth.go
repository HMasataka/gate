package handler

import (
	"errors"
	"net/http"

	"github.com/HMasataka/gate/internal/domain"
	"github.com/HMasataka/gate/internal/usecase"
)

type OAuthHandler struct {
	token *usecase.TokenUsecase
}

func NewOAuthHandler(token *usecase.TokenUsecase) *OAuthHandler {
	return &OAuthHandler{token: token}
}

type tokenRequest struct {
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
}

// Token handles POST /oauth/token (refresh_token grant only)
func (h *OAuthHandler) Token(w http.ResponseWriter, r *http.Request) {
	var req tokenRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.GrantType != "refresh_token" {
		Error(w, http.StatusBadRequest, "unsupported_grant_type", "only refresh_token grant is supported")
		return
	}

	if req.RefreshToken == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "refresh_token is required")
		return
	}

	accessToken, newRefreshToken, err := h.token.RefreshTokens(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidToken) || errors.Is(err, domain.ErrTokenRevoked) {
			Error(w, http.StatusUnauthorized, "invalid_grant", "refresh token is invalid or revoked")
			return
		}
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]any{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    int(h.token.AccessTokenExpiry().Seconds()),
		"refresh_token": newRefreshToken,
	})
}

type revokeRequest struct {
	Token string `json:"token"`
}

// Revoke handles POST /oauth/revoke
func (h *OAuthHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	var req revokeRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.Token == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "token is required")
		return
	}

	// RFC 7009: always return 200 even if token not found
	if err := h.token.RevokeToken(r.Context(), req.Token); err != nil && !errors.Is(err, domain.ErrNotFound) {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "token revoked"})
}
