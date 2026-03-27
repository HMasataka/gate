package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/HMasataka/gate/internal/config"
	"github.com/HMasataka/gate/internal/domain"
)

type SessionStore struct {
	client      *redis.Client
	maxSessions int
}

func NewSessionStore(client *redis.Client, cfg config.SessionConfig) *SessionStore {
	return &SessionStore{
		client:      client,
		maxSessions: cfg.MaxSessions,
	}
}

func sessionKey(id string) string          { return "session:" + id }
func userSessionsKey(userID string) string { return "user:sessions:" + userID }

func (s *SessionStore) Create(ctx context.Context, session *domain.Session) error {
	key := sessionKey(session.ID)
	userKey := userSessionsKey(session.UserID)

	fields := map[string]any{
		"user_id":    session.UserID,
		"ip_address": session.IPAddress,
		"user_agent": session.UserAgent,
		"expires_at": session.ExpiresAt.UTC().Format(time.RFC3339),
		"created_at": session.CreatedAt.UTC().Format(time.RFC3339),
	}

	// セッション数チェックと古いセッションの削除
	count, err := s.client.SCard(ctx, userKey).Result()
	if err != nil {
		return fmt.Errorf("scard user sessions: %w", err)
	}

	if int(count) >= s.maxSessions {
		// 最も古いセッションを取得して削除
		sessionIDs, err := s.client.SMembers(ctx, userKey).Result()
		if err != nil {
			return fmt.Errorf("smembers user sessions: %w", err)
		}

		oldestID, oldestTime := findOldestSession(ctx, s.client, sessionIDs)
		if oldestID != "" && !oldestTime.IsZero() {
			if err := s.deleteSession(ctx, oldestID); err != nil {
				return fmt.Errorf("delete oldest session: %w", err)
			}
		}
	}

	pipe := s.client.Pipeline()
	pipe.HSet(ctx, key, fields)
	pipe.ExpireAt(ctx, key, session.ExpiresAt)
	pipe.SAdd(ctx, userKey, session.ID)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("create session: %w", err)
	}

	return nil
}

// findOldestSession は sessionIDs の中から created_at が最も古いセッション ID と時刻を返す。
func findOldestSession(ctx context.Context, client *redis.Client, sessionIDs []string) (string, time.Time) {
	var oldestID string
	var oldestTime time.Time

	for _, id := range sessionIDs {
		createdAtStr, err := client.HGet(ctx, sessionKey(id), "created_at").Result()
		if err != nil {
			continue
		}
		t, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			continue
		}
		if oldestTime.IsZero() || t.Before(oldestTime) {
			oldestTime = t
			oldestID = id
		}
	}

	return oldestID, oldestTime
}

func (s *SessionStore) Get(ctx context.Context, id string) (*domain.Session, error) {
	key := sessionKey(id)

	fields, err := s.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("hgetall session: %w", err)
	}
	if len(fields) == 0 {
		return nil, domain.ErrNotFound
	}

	session, err := mapToSession(id, fields)
	if err != nil {
		return nil, fmt.Errorf("map session fields: %w", err)
	}

	return session, nil
}

func (s *SessionStore) Delete(ctx context.Context, id string) error {
	return s.deleteSession(ctx, id)
}

func (s *SessionStore) deleteSession(ctx context.Context, id string) error {
	key := sessionKey(id)

	userID, err := s.client.HGet(ctx, key, "user_id").Result()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return fmt.Errorf("hget user_id: %w", err)
	}

	pipe := s.client.Pipeline()
	pipe.Del(ctx, key)
	pipe.SRem(ctx, userSessionsKey(userID), id)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	return nil
}

func (s *SessionStore) DeleteByUserID(ctx context.Context, userID string) error {
	userKey := userSessionsKey(userID)

	sessionIDs, err := s.client.SMembers(ctx, userKey).Result()
	if err != nil {
		return fmt.Errorf("smembers user sessions: %w", err)
	}

	if len(sessionIDs) > 0 {
		keys := make([]string, len(sessionIDs))
		for i, id := range sessionIDs {
			keys[i] = sessionKey(id)
		}

		pipe := s.client.Pipeline()
		pipe.Del(ctx, keys...)
		pipe.Del(ctx, userKey)

		if _, err := pipe.Exec(ctx); err != nil {
			return fmt.Errorf("delete user sessions: %w", err)
		}
	} else {
		if err := s.client.Del(ctx, userKey).Err(); err != nil {
			return fmt.Errorf("delete user sessions key: %w", err)
		}
	}

	return nil
}

func (s *SessionStore) ListByUserID(ctx context.Context, userID string) ([]domain.Session, error) {
	userKey := userSessionsKey(userID)

	sessionIDs, err := s.client.SMembers(ctx, userKey).Result()
	if err != nil {
		return nil, fmt.Errorf("smembers user sessions: %w", err)
	}

	sessions := make([]domain.Session, 0, len(sessionIDs))
	expiredIDs := make([]string, 0)

	for _, id := range sessionIDs {
		key := sessionKey(id)

		fields, err := s.client.HGetAll(ctx, key).Result()
		if err != nil {
			continue
		}
		if len(fields) == 0 {
			// 期限切れまたは削除済み — クリーンアップ対象
			expiredIDs = append(expiredIDs, id)
			continue
		}

		session, err := mapToSession(id, fields)
		if err != nil {
			continue
		}

		if session.ExpiresAt.Before(time.Now()) {
			// 期限切れセッション — クリーンアップ対象
			expiredIDs = append(expiredIDs, id)
			s.client.Del(ctx, key) //nolint:errcheck
			continue
		}

		sessions = append(sessions, *session)
	}

	// 期限切れセッションを user:sessions:{userID} から除去
	if len(expiredIDs) > 0 {
		args := make([]any, len(expiredIDs))
		for i, id := range expiredIDs {
			args[i] = id
		}
		s.client.SRem(ctx, userKey, args...) //nolint:errcheck
	}

	return sessions, nil
}

// mapToSession は Redis Hash のフィールドマップを domain.Session に変換する。
func mapToSession(id string, fields map[string]string) (*domain.Session, error) {
	expiresAt, err := time.Parse(time.RFC3339, fields["expires_at"])
	if err != nil {
		return nil, fmt.Errorf("parse expires_at: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, fields["created_at"])
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}

	return &domain.Session{
		ID:        id,
		UserID:    fields["user_id"],
		IPAddress: fields["ip_address"],
		UserAgent: fields["user_agent"],
		ExpiresAt: expiresAt,
		CreatedAt: createdAt,
	}, nil
}
