package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func DecodeJSON(r *http.Request, v any) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}
	return nil
}

func PathParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

func QueryParam(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

func Paginate(r *http.Request) (offset, limit int) {
	offset = 0
	limit = 20

	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	return offset, limit
}
