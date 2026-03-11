package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
}

type SessionData struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

func NewRedis(url, password string) (*Cache, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis url: %w", err)
	}

	if password != "" {
		opts.Password = password
	}

	client := redis.NewClient(opts)

	// Ping to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Cache{client: client}, nil
}

func (c *Cache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *Cache) Close() error {
	return c.client.Close()
}

func (c *Cache) SetSession(ctx context.Context, userID string, data SessionData, ttl time.Duration) error {
	key := fmt.Sprintf("session:%s", userID)

	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	return c.client.Set(ctx, key, string(b), ttl).Err()
}

func (c *Cache) GetSession(ctx context.Context, userID string) (*SessionData, error) {
	key := fmt.Sprintf("session:%s", userID)

	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Not found
	} else if err != nil {
		return nil, fmt.Errorf("redis get session error: %w", err)
	}

	var data SessionData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return &data, nil
}

func (c *Cache) DeleteSession(ctx context.Context, userID string) error {
	key := fmt.Sprintf("session:%s", userID)
	return c.client.Del(ctx, key).Err()
}

func (c *Cache) IncrRateLimit(ctx context.Context, ip, endpoint string) (int64, error) {
	key := fmt.Sprintf("ratelimit:%s:%s", ip, endpoint)

	count, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("redis incr error: %w", err)
	}

	// Set TTL only on the first increment
	if count == 1 {
		c.client.Expire(ctx, key, 60*time.Second)
	}

	return count, nil
}

func (c *Cache) GetRateLimit(ctx context.Context, ip, endpoint string) (int64, error) {
	key := fmt.Sprintf("ratelimit:%s:%s", ip, endpoint)

	count, err := c.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	} else if err != nil {
		return 0, fmt.Errorf("redis get rate limit error: %w", err)
	}

	return count, nil
}

func (c *Cache) SetVideoMeta(ctx context.Context, videoID string, video interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("video:meta:%s", videoID)

	b, err := json.Marshal(video)
	if err != nil {
		return fmt.Errorf("failed to marshal video meta: %w", err)
	}

	return c.client.Set(ctx, key, string(b), ttl).Err()
}

func (c *Cache) GetVideoMeta(ctx context.Context, videoID string) (string, error) {
	key := fmt.Sprintf("video:meta:%s", videoID)

	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // Not found
	} else if err != nil {
		return "", fmt.Errorf("redis get video meta error: %w", err)
	}

	return val, nil
}

func (c *Cache) DeleteVideoMeta(ctx context.Context, videoID string) error {
	key := fmt.Sprintf("video:meta:%s", videoID)
	return c.client.Del(ctx, key).Err()
}

func (c *Cache) InvalidateUserVideoList(ctx context.Context, userID string) error {
	pattern := fmt.Sprintf("videos:list:%s:*", userID)

	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("redis scan error: %w", err)
		}

		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("redis del error: %w", err)
			}
		}

		if cursor == 0 {
			break
		}
	}

	return nil
}
