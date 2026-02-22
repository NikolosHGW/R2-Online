package repository

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const sessionTTL = 5 * time.Minute

// redisKey returns the Redis key for a session token.
func redisKey(token int32) string {
	return fmt.Sprintf("session:%d", token)
}

// SessionRepository handles cross-server session tokens stored in Redis.
type SessionRepository struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewSessionRepository(db *pgxpool.Pool, rdb *redis.Client) *SessionRepository {
	return &SessionRepository{db: db, redis: rdb}
}

// CreateToken generates a session token in Redis, linking accountID to a game server.
// The token is a random int32 used as the "SessionId" in LoginUserReq (5100).
func (r *SessionRepository) CreateToken(ctx context.Context, accountID int32, serverID int16) (int32, error) {
	token := rand.Int32()

	// Redis: token → accountID (TTL = 5 minutes)
	err := r.redis.Set(ctx, redisKey(token), accountID, sessionTTL).Err()
	if err != nil {
		return 0, fmt.Errorf("CreateToken: redis: %w", err)
	}

	// Durable audit record (optional but useful for debugging)
	_, err = r.db.Exec(ctx,
		`INSERT INTO sessions (account_id, server_id, token) VALUES ($1, $2, $3)`,
		accountID, serverID, token,
	)
	if err != nil {
		return 0, fmt.Errorf("CreateToken: db: %w", err)
	}

	return token, nil
}

// ValidateToken checks that the token exists in Redis and returns the associated accountID.
// Consumes the token (deletes it) to prevent reuse.
func (r *SessionRepository) ValidateToken(ctx context.Context, token int32) (accountID int32, err error) {
	key := redisKey(token)

	val, err := r.redis.GetDel(ctx, key).Int()
	if err == redis.Nil {
		return 0, fmt.Errorf("ValidateToken: token not found or expired")
	}
	if err != nil {
		return 0, fmt.Errorf("ValidateToken: redis: %w", err)
	}

	return int32(val), nil
}

// IsAccountOnline returns true if there is an active (non-expired) session for the account.
func (r *SessionRepository) IsAccountOnline(ctx context.Context, accountID int32) (bool, error) {
	// Check for a token pattern in Redis (scan with account prefix)
	// Simpler approach: keep a separate Redis key "online:{accountID}"
	key := fmt.Sprintf("online:%d", accountID)
	exists, err := r.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// MarkAccountOnline sets the "online" flag for the account in Redis.
func (r *SessionRepository) MarkAccountOnline(ctx context.Context, accountID int32) error {
	key := fmt.Sprintf("online:%d", accountID)
	return r.redis.Set(ctx, key, 1, sessionTTL).Err()
}

// MarkAccountOffline clears the "online" flag when a game session ends.
func (r *SessionRepository) MarkAccountOffline(ctx context.Context, accountID int32) error {
	key := fmt.Sprintf("online:%d", accountID)
	return r.redis.Del(ctx, key).Err()
}
