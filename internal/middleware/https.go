package middleware

import (
	"net/http"
)

// HTTPSRedirect redirects HTTP requests to HTTPS when enabled.
// Only active when X-Forwarded-Proto is not "https".
// Returns 301 Moved Permanently to the HTTPS equivalent URL.
func HTTPSRedirect(enabled bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if enabled && r.Header.Get("X-Forwarded-Proto") != "https" {
				target := "https://" + r.Host + r.URL.RequestURI()
				http.Redirect(w, r, target, http.StatusMovedPermanently)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
