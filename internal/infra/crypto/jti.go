package crypto

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/HMasataka/gate/internal/domain"
)

// JTIStore prevents JWT replay attacks by tracking used JTI (JWT ID) values in Redis.
type JTIStore struct {
	client *redis.Client
}

// NewJTIStore creates a new JTIStore backed by the given Redis client.
func NewJTIStore(client *redis.Client) *JTIStore {
	return &JTIStore{client: client}
}

func jtiKey(jti string) string { return "jti:" + jti }

// MarkUsed marks a JTI as used with the given TTL.
// Returns domain.ErrTokenRevoked if the JTI has already been used.
func (s *JTIStore) MarkUsed(ctx context.Context, jti string, ttl time.Duration) error {
	key := jtiKey(jti)

	// Use Set with NX option (Redis >= 2.6.12): returns redis.Nil if key already exists.
	err := s.client.SetArgs(ctx, key, "1", redis.SetArgs{TTL: ttl, Mode: "NX"}).Err()
	if err == redis.Nil {
		return domain.ErrTokenRevoked
	}
	if err != nil {
		return fmt.Errorf("jti mark used: %w", err)
	}

	return nil
}
