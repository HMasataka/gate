package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const rateLimitLuaScript = `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local member = ARGV[4]

-- Remove entries outside the window
redis.call('ZREMRANGEBYSCORE', key, '-inf', now - window)

-- Count remaining entries
local count = redis.call('ZCARD', key)

if count >= limit then
    return 0
end

-- Add current timestamp as score and member
redis.call('ZADD', key, now, member)

-- Set expiry on the key
redis.call('PEXPIRE', key, math.ceil(window / 1000))

return 1
`

var rateLimitScript = redis.NewScript(rateLimitLuaScript)

type RateLimiterStore struct {
	client *redis.Client
}

func NewRateLimiterStore(client *redis.Client) *RateLimiterStore {
	return &RateLimiterStore{client: client}
}

func rateLimitKey(key string) string { return "ratelimit:" + key }

// Allow checks if the request is allowed under the sliding window rate limit.
// Returns true if allowed, false if denied.
func (s *RateLimiterStore) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	nowMicros := time.Now().UnixMicro()
	windowMicros := window.Microseconds()

	// member = timestamp:nanoseconds to avoid collisions
	member := fmt.Sprintf("%d:%d", nowMicros, time.Now().UnixNano())

	result, err := rateLimitScript.Run(
		ctx,
		s.client,
		[]string{rateLimitKey(key)},
		nowMicros,
		windowMicros,
		limit,
		member,
	).Int()
	if err != nil {
		return false, fmt.Errorf("rate limit script: %w", err)
	}

	return result == 1, nil
}

// Remaining returns the number of remaining allowed requests and the reset duration.
func (s *RateLimiterStore) Remaining(ctx context.Context, key string, limit int, window time.Duration) (int, time.Duration, error) {
	rKey := rateLimitKey(key)
	nowMicros := time.Now().UnixMicro()
	windowMicros := window.Microseconds()

	// Remove stale entries
	if err := s.client.ZRemRangeByScore(ctx, rKey, "-inf", fmt.Sprintf("%d", nowMicros-windowMicros)).Err(); err != nil {
		return 0, 0, fmt.Errorf("zremrangebyscore: %w", err)
	}

	count, err := s.client.ZCard(ctx, rKey).Result()
	if err != nil {
		return 0, 0, fmt.Errorf("zcard: %w", err)
	}

	remaining := max(limit-int(count), 0)

	return remaining, window, nil
}
