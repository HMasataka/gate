package handler

import (
	"errors"
	"net/http"

	"github.com/HMasataka/gate/internal/domain"
	"github.com/HMasataka/gate/internal/usecase"
)

type AuthHandler struct {
	auth *usecase.AuthUsecase
}

func NewAuthHandler(auth *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "email and password are required")
		return
	}

	user, err := h.auth.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicateEmail) {
			// ユーザー列挙防止: 成功と同じレスポンスを返す
			JSON(w, http.StatusCreated, map[string]any{
				"message": "registration successful, please check your email",
			})
			return
		}
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusCreated, map[string]any{
		"id":    user.ID,
		"email": user.Email,
		"message": "registration successful, please check your email",
	})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "email and password are required")
		return
	}

	result, err := h.auth.Login(r.Context(), req.Email, req.Password, r.RemoteAddr, r.UserAgent())
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

type logoutRequest struct {
	SessionID    string `json:"session_id"`
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.SessionID == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "session_id is required")
		return
	}

	if err := h.auth.Logout(r.Context(), req.SessionID, req.RefreshToken); err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

type verifyEmailRequest struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req verifyEmailRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.Email == "" || req.Token == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "email and token are required")
		return
	}

	if err := h.auth.VerifyEmail(r.Context(), req.Email, req.Token); err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "email verified"})
}

type resendVerificationRequest struct {
	Email string `json:"email"`
}

func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req resendVerificationRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.Email == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "email is required")
		return
	}

	if err := h.auth.ResendVerification(r.Context(), req.Email); err != nil {
		HandleError(w, err)
		return
	}

	// ユーザー列挙防止: 常に同じレスポンス
	JSON(w, http.StatusOK, map[string]string{"message": "if the email exists, a verification email has been sent"})
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.Email == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "email is required")
		return
	}

	if err := h.auth.ForgotPassword(r.Context(), req.Email); err != nil {
		HandleError(w, err)
		return
	}

	// ユーザー列挙防止: 常に同じレスポンス
	JSON(w, http.StatusOK, map[string]string{"message": "if the email exists, a password reset email has been sent"})
}

type resetPasswordRequest struct {
	Email       string `json:"email"`
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.Email == "" || req.Token == "" || req.NewPassword == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "email, token, and new_password are required")
		return
	}

	if err := h.auth.ResetPassword(r.Context(), req.Email, req.Token, req.NewPassword); err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "password reset successful"})
}

type changePasswordRequest struct {
	UserID          string `json:"user_id"`
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req changePasswordRequest
	if err := DecodeJSON(r, &req); err != nil {
		Error(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	if req.UserID == "" || req.CurrentPassword == "" || req.NewPassword == "" {
		Error(w, http.StatusBadRequest, "invalid_request", "user_id, current_password, and new_password are required")
		return
	}

	if err := h.auth.ChangePassword(r.Context(), req.UserID, req.CurrentPassword, req.NewPassword); err != nil {
		HandleError(w, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{"message": "password changed"})
}
