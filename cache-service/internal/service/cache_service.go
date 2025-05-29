package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheService struct {
	cli *redis.Client
	ttl time.Duration
}

func New(addr, pass string, db int, ttlSec int) *CacheService {
	return &CacheService{
		cli: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: pass,
			DB:       db,
		}),
		ttl: time.Duration(ttlSec) * time.Second,
	}
}

func conversationKey(u1, u2 string) string {
	u1 = strings.ToLower(u1)
	u2 = strings.ToLower(u2)
	if u1 < u2 {
		return "conversation:" + u1 + ":" + u2
	}
	return "conversation:" + u2 + ":" + u1
}

func (c *CacheService) GetConversation(ctx context.Context, u1, u2 string, dst any) (bool, error) {
	key := conversationKey(u1, u2)

	val, err := c.cli.Get(ctx, key).Bytes()
	switch {
	case errors.Is(err, redis.Nil):
		return false, nil
	case err != nil:
		return false, err
	}

	_ = c.cli.Expire(ctx, key, c.ttl).Err()

	return true, json.Unmarshal(val, dst)
}

func (c *CacheService) SetConversation(ctx context.Context, u1, u2 string, v any) error {
	b, _ := json.Marshal(v)
	return c.cli.Set(ctx, conversationKey(u1, u2), b, c.ttl).Err()
}

func (c *CacheService) DeleteConversation(ctx context.Context, u1, u2 string) error {
	return c.cli.Del(ctx, conversationKey(u1, u2)).Err()
}
