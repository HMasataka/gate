package handler

import (
	"net/http"

	"github.com/HMasataka/gate/internal/domain"
	"github.com/HMasataka/gate/internal/middleware"
	"github.com/HMasataka/gate/internal/usecase"
)

type MFAHandler struct {
	mfa    *usecase.MFAUsecase
	hasher domain.PasswordHasher
}

func NewMFAHandler(mfa *usecase.MFAUsecase, hasher domain.PasswordHasher) *MFAHandler {
	return &MFAHandler{mfa: mfa, hasher: hasher}
}

func (h *MFAHandler) SetupTOTP(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	provisioningURI, secret, err := h.mfa.SetupTOTP(r.Context(), userID)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]any{
		"provisioning_uri": provisioningURI,
		"secret":           secret,
	})
}

type confirmTOTPRequest struct {
	Code string `json:"code"`
}

func (h *MFAHandler) ConfirmTOTP(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	var req confirmTOTPRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.Code == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "code is required")
		return
	}

	recoveryCodes, err := h.mfa.ConfirmTOTP(r.Context(), userID, req.Code)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]any{
		"recovery_codes": recoveryCodes,
	})
}

type disableTOTPRequest struct {
	Password string `json:"password"`
}

func (h *MFAHandler) DisableTOTP(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	var req disableTOTPRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.Password == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "password is required")
		return
	}

	if err := h.mfa.DisableTOTP(r.Context(), userID, req.Password, h.hasher); err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "totp disabled"})
}

func (h *MFAHandler) RegenerateRecoveryCodes(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		Error(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	recoveryCodes, err := h.mfa.RegenerateRecoveryCodes(r.Context(), userID)
	if err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]any{
		"recovery_codes": recoveryCodes,
	})
}
