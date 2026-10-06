package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// OpenRedis creates and verifies a Redis client. An empty URL disables Redis
// and returns a nil client without error.
func OpenRedis(ctx context.Context, rawURL string) (*redis.Client, error) {
	if rawURL == "" {
		return nil, nil
	}
	options, err := redis.ParseURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse Redis URL: %w", err)
	}
	client := redis.NewClient(options)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping Redis: %w", err)
	}
	return client, nil
}
