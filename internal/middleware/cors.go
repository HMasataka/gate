package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
)

func (m *Middleware) CORS(next http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins: m.cfg.CORS.AllowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		MaxAge:         m.cfg.CORS.MaxAge,
	})
	return c.Handler(next)
}
