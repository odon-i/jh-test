package cache

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"forum/internal/config"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
}

func Open(ctx context.Context, cfg config.RedisConfig) (*Cache, error) {
	if !cfg.Enabled {
		return &Cache{}, nil
	}
	client := redis.NewClient(&redis.Options{Addr: cfg.Address, Password: cfg.Password, DB: cfg.DB})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("connect redis: %w", err)
	}
	return &Cache{client: client}, nil
}

func (c *Cache) Enabled() bool { return c != nil && c.client != nil }

func (c *Cache) Close() error {
	if !c.Enabled() {
		return nil
	}
	return c.client.Close()
}

func (c *Cache) SetLikeState(ctx context.Context, userID, postID, version uint64, liked bool) error {
	if !c.Enabled() {
		return nil
	}
	key := fmt.Sprintf("post:%d:liked:%d", postID, userID)
	likedValue := "0"
	if liked {
		likedValue = "1"
	}
	value := fmt.Sprintf("%d:%s", version, likedValue)
	return c.client.Set(ctx, key, value, 10*time.Minute).Err()
}

func (c *Cache) GetLikeState(ctx context.Context, userID, postID uint64) (uint64, bool, bool, error) {
	if !c.Enabled() {
		return 0, false, false, nil
	}
	key := fmt.Sprintf("post:%d:liked:%d", postID, userID)
	value, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, false, false, nil
	}
	if err != nil {
		return 0, false, false, err
	}
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		return 0, false, false, nil
	}
	version, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil || (parts[1] != "0" && parts[1] != "1") {
		return 0, false, false, nil
	}
	return version, parts[1] == "1", true, nil
}
