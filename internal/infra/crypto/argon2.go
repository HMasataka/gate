package crypto

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"

	"github.com/HMasataka/gate/internal/config"
)

type Argon2Hasher struct {
	cfg config.Argon2Config
	sem chan struct{}
}

func NewArgon2Hasher(cfg config.Argon2Config) *Argon2Hasher {
	return &Argon2Hasher{
		cfg: cfg,
		sem: make(chan struct{}, cfg.Semaphore),
	}
}

func (h *Argon2Hasher) Hash(ctx context.Context, password string) (string, error) {
	select {
	case h.sem <- struct{}{}:
		defer func() { <-h.sem }()
	case <-ctx.Done():
		return "", ctx.Err()
	}

	salt := make([]byte, h.cfg.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, h.cfg.Time, h.cfg.Memory, h.cfg.Parallelism, h.cfg.KeyLength)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.cfg.Memory, h.cfg.Time, h.cfg.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func (h *Argon2Hasher) Compare(ctx context.Context, password, encodedHash string) (bool, error) {
	select {
	case h.sem <- struct{}{}:
		defer func() { <-h.sem }()
	case <-ctx.Done():
		return false, ctx.Err()
	}

	// Parse PHC format: $argon2id$v=19$m=65536,t=1,p=4$salt$hash
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, nil
	}
	if parts[1] != "argon2id" {
		return false, nil
	}

	var memory uint32
	var time uint32
	var parallelism uint8
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &parallelism)
	if err != nil {
		return false, nil
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, nil
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, nil
	}

	hash := argon2.IDKey([]byte(password), salt, time, memory, parallelism, uint32(len(expectedHash)))

	return subtle.ConstantTimeCompare(hash, expectedHash) == 1, nil
}
