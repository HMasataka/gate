package middleware

import (
	"github.com/HMasataka/gate/internal/config"
)

type Middleware struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Middleware {
	return &Middleware{cfg: cfg}
}
