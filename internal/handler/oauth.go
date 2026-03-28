package handler

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/HMasataka/gate/internal/domain"
	"github.com/HMasataka/gate/internal/usecase"
)

type OAuthHandler struct {
	oauth *usecase.OAuthUsecase
	token *usecase.TokenUsecase
}

func NewOAuthHandler(oauth *usecase.OAuthUsecase, token *usecase.TokenUsecase) *OAuthHandler {
	return &OAuthHandler{oauth: oauth, token: token}
}

// Authorize handles GET /oauth/authorize
func (h *OAuthHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	clientID := QueryParam(r, "client_id")
	redirectURI := QueryParam(r, "redirect_uri")
	responseType := QueryParam(r, "response_type")
	scope := QueryParam(r, "scope")
	state := QueryParam(r, "state")
	codeChallenge := QueryParam(r, "code_challenge")
	codeChallengeMethod := QueryParam(r, "code_challenge_method")

	if redirectURI == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "redirect_uri is required")
		return
	}

	parsedURI, err := url.Parse(redirectURI)
	if err != nil || parsedURI.Scheme == "" || parsedURI.Host == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "redirect_uri is not a valid URI")
		return
	}

	code, err := h.oauth.Authorize(r.Context(), clientID, redirectURI, responseType, scope, state, codeChallenge, codeChallengeMethod)
	if err != nil {
		errCode := "server_error"
		errDesc := "internal server error"
		switch {
		case errors.Is(err, domain.ErrInvalidClient):
			errCode = "invalid_client"
			errDesc = "client not found"
		case errors.Is(err, domain.ErrInvalidRedirectURI):
			// Cannot redirect - redirect_uri is invalid
			Error(w, http.StatusBadRequest, "invalid_request", "redirect_uri does not match registered URIs")
			return
		case errors.Is(err, domain.ErrInvalidGrantType):
			errCode = "unsupported_response_type"
			errDesc = "unsupported response_type"
		}

		q := parsedURI.Query()
		q.Set("error", errCode)
		q.Set("error_description", errDesc)
		if state != "" {
			q.Set("state", state)
		}
		parsedURI.RawQuery = q.Encode()
		http.Redirect(w, r, parsedURI.String(), http.StatusFound)
		return
	}

	q := parsedURI.Query()
	q.Set("code", code)
	if state != "" {
		q.Set("state", state)
	}
	parsedURI.RawQuery = q.Encode()
	http.Redirect(w, r, parsedURI.String(), http.StatusFound)
}

type oauthTokenRequest struct {
	GrantType string `json:"grant_type"`
	// authorization_code
	Code         string `json:"code"`
	RedirectURI  string `json:"redirect_uri"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	CodeVerifier string `json:"code_verifier"`
	// refresh_token
	RefreshToken string `json:"refresh_token"`
	// client_credentials
	Scope string `json:"scope"`
}

// Token handles POST /oauth/token
func (h *OAuthHandler) Token(w http.ResponseWriter, r *http.Request) {
	var req oauthTokenRequest

	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "application/x-www-form-urlencoded") {
		if err := r.ParseForm(); err != nil {
			Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
			return
		}
		req.GrantType = r.PostFormValue("grant_type")
		req.Code = r.PostFormValue("code")
		req.RedirectURI = r.PostFormValue("redirect_uri")
		req.ClientID = r.PostFormValue("client_id")
		req.ClientSecret = r.PostFormValue("client_secret")
		req.CodeVerifier = r.PostFormValue("code_verifier")
		req.RefreshToken = r.PostFormValue("refresh_token")
		req.Scope = r.PostFormValue("scope")
	} else {
		if err := DecodeJSON(r, &req); err != nil {
			Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
			return
		}
	}

	switch req.GrantType {
	case "authorization_code":
		if req.Code == "" {
			Error(w, http.StatusBadRequest, "invalid_request", "code is required")
			return
		}

		accessToken, refreshToken, err := h.oauth.ExchangeCode(
			r.Context(),
			req.ClientID, req.ClientSecret, req.Code, req.CodeVerifier, req.RedirectURI,
		)
		if err != nil {
			if errors.Is(err, domain.ErrInvalidClient) {
				Error(w, http.StatusUnauthorized, "invalid_client", "invalid client credentials")
				return
			}
			if errors.Is(err, domain.ErrInvalidToken) || errors.Is(err, domain.ErrTokenExpired) || errors.Is(err, domain.ErrCodeReuse) {
				Error(w, http.StatusBadRequest, "invalid_grant", "authorization code is invalid or expired")
				return
			}
			HandleError(w, err)
			return
		}

		JSON(w, http.StatusOK, map[string]any{
			"access_token":  accessToken,
			"token_type":    "Bearer",
			"expires_in":    int(h.token.AccessTokenExpiry().Seconds()),
			"refresh_token": refreshToken,
		})

	case "refresh_token":
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

	case "client_credentials":
		if req.ClientID == "" || req.ClientSecret == "" {
			Error(w, http.StatusBadRequest, "invalid_request", "client_id and client_secret are required")
			return
		}

		accessToken, err := h.oauth.ClientCredentials(r.Context(), req.ClientID, req.ClientSecret, req.Scope)
		if err != nil {
			if errors.Is(err, domain.ErrInvalidClient) {
				Error(w, http.StatusUnauthorized, "invalid_client", "invalid client credentials")
				return
			}
			if errors.Is(err, domain.ErrInvalidGrantType) {
				Error(w, http.StatusBadRequest, "unauthorized_client", "client is not authorized for client_credentials grant")
				return
			}
			HandleError(w, err)
			return
		}

		JSON(w, http.StatusOK, map[string]any{
			"access_token": accessToken,
			"token_type":   "Bearer",
			"expires_in":   int(h.token.AccessTokenExpiry().Seconds()),
		})

	default:
		Error(w, http.StatusBadRequest, "unsupported_grant_type", "grant type is not supported")
	}
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

type introspectRequest struct {
	Token string `json:"token"`
}

// Introspect handles POST /oauth/introspect (RFC 7662)
func (h *OAuthHandler) Introspect(w http.ResponseWriter, r *http.Request) {
	var req introspectRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.Token == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "token is required")
		return
	}

	claims, err := h.oauth.Introspect(r.Context(), req.Token)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, claims)
}
