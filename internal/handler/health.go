package handler

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	db  *sqlx.DB
	rdb *redis.Client
}

func NewHealthHandler(db *sqlx.DB, rdb *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, rdb: rdb}
}

func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	checks := map[string]string{}
	healthy := true

	if err := h.db.PingContext(ctx); err != nil {
		checks["database"] = "unavailable"
		healthy = false
	} else {
		checks["database"] = "ok"
	}

	if err := h.rdb.Ping(ctx).Err(); err != nil {
		checks["redis"] = "unavailable"
		healthy = false
	} else {
		checks["redis"] = "ok"
	}

	status := "ok"
	httpStatus := http.StatusOK
	if !healthy {
		status = "unavailable"
		httpStatus = http.StatusServiceUnavailable
	}

	JSON(w, httpStatus, map[string]any{
		"status": status,
		"checks": checks,
	})
}
