package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/HMasataka/gate/internal/domain"
)

const UserIDKey contextKey = "user_id"

func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(UserIDKey).(string); ok {
		return id
	}
	return ""
}

func JWTAuth(jwtManager domain.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeUnauthorized(w, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				writeUnauthorized(w, "invalid authorization header format")
				return
			}

			claims, err := jwtManager.ValidateToken(r.Context(), parts[1])
			if err != nil {
				writeUnauthorized(w, "invalid token")
				return
			}

			sub, ok := claims["sub"].(string)
			if !ok || sub == "" {
				writeUnauthorized(w, "invalid token claims")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, sub)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{
			"code":    "unauthorized",
			"message": message,
		},
	})
}
