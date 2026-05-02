package relay

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCooldownStore struct {
	rdb *redis.Client
}

func NewRedisCooldownStore(rdb *redis.Client) *RedisCooldownStore {
	return &RedisCooldownStore{rdb: rdb}
}

func (s *RedisCooldownStore) SetCooldown(ctx context.Context, channelID int64, duration time.Duration) error {
	key := fmt.Sprintf("channel:cooldown:%d", channelID)
	return s.rdb.Set(ctx, key, "1", duration).Err()
}

func (s *RedisCooldownStore) IsCoolingDown(ctx context.Context, channelID int64) bool {
	key := fmt.Sprintf("channel:cooldown:%d", channelID)
	val, err := s.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false
	}
	return val > 0
}
