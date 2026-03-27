package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/HMasataka/gate/internal/domain"
)

type PermissionCache struct {
	client *redis.Client
	repo   domain.PermissionRepository
	ttl    time.Duration
}

func NewPermissionCache(client *redis.Client, repo domain.PermissionRepository, ttl time.Duration) *PermissionCache {
	return &PermissionCache{
		client: client,
		repo:   repo,
		ttl:    ttl,
	}
}

func permissionKey(userID string) string { return "permissions:" + userID }

func (c *PermissionCache) Resolve(ctx context.Context, userID string) ([]string, error) {
	key := permissionKey(userID)

	exists, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("check permissions cache existence: %w", err)
	}

	if exists > 0 {
		perms, err := c.client.SMembers(ctx, key).Result()
		if err != nil {
			return nil, fmt.Errorf("smembers permissions cache: %w", err)
		}
		return perms, nil
	}

	// Cache miss: fetch from repository and populate cache.
	perms, err := c.repo.ResolveForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("resolve permissions for user: %w", err)
	}

	if len(perms) > 0 {
		members := make([]any, len(perms))
		for i, p := range perms {
			members[i] = p
		}

		pipe := c.client.Pipeline()
		pipe.SAdd(ctx, key, members...)
		pipe.Expire(ctx, key, c.ttl)
		if _, err := pipe.Exec(ctx); err != nil {
			return nil, fmt.Errorf("populate permissions cache: %w", err)
		}
	}

	return perms, nil
}

func (c *PermissionCache) Invalidate(ctx context.Context, userID string) error {
	key := permissionKey(userID)

	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("invalidate permissions cache: %w", err)
	}

	return nil
}
