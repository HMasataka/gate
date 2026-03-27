package middleware

import (
	"encoding/json"
	"net/http"
	"slices"

	"github.com/HMasataka/gate/internal/domain"
)

// RequirePermission returns a middleware that checks if the authenticated user
// has the given permission. Returns 401 if not authenticated, 403 if not authorized.
func RequirePermission(resolver domain.PermissionResolver, permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			userID := GetUserID(ctx)
			if userID == "" {
				writeUnauthorized(w, "authentication required")
				return
			}

			permissions, err := resolver.Resolve(ctx, userID)
			if err != nil {
				writeForbidden(w, "insufficient permissions")
				return
			}

			if !slices.Contains(permissions, permission) {
				writeForbidden(w, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func writeForbidden(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{
			"code":    "forbidden",
			"message": message,
		},
	})
}
