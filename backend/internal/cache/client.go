package cache

import (
	"context"
	"fmt"
	"time"
)

// Backend identifies which store is backing a Client.
type Backend string

const (
	// BackendRedis is used when a configured Redis server is reachable.
	BackendRedis Backend = "redis"
	// BackendMemory is the fallback used when Redis is absent or unreachable.
	BackendMemory Backend = "memory"
)

// Client is the process-wide cache handle shared by the API and worker
// processes. It always wraps a usable Store: Redis when one is configured and
// reachable, an in-memory store otherwise. Requests never observe a nil Client
// after Open, so Redis stays strictly optional.
type Client struct {
	store   Store
	backend Backend
	closer  Closer
}

// Open connects to Redis when rawURL is set and falls back to the in-memory
// store when Redis is not configured, cannot be parsed, or cannot be reached.
// The returned Client is always usable; a non-nil error only explains why a
// configured Redis was skipped so callers can log it without failing startup.
func Open(ctx context.Context, rawURL string) (*Client, error) {
	store, redisBacked, err := OpenStore(ctx, rawURL)
	if !redisBacked {
		return &Client{store: store, backend: BackendMemory}, err
	}
	client := &Client{store: store, backend: BackendRedis}
	if closer, ok := store.(Closer); ok {
		client.closer = closer
	}
	return client, nil
}

// Backend reports which store is active.
func (c *Client) Backend() Backend { return c.backend }

// RedisBacked reports whether the Client is backed by Redis.
func (c *Client) RedisBacked() bool { return c.backend == BackendRedis }

// Store exposes the underlying store for consumers that depend directly on the
// cache contract.
func (c *Client) Store() Store { return c.store }

// Ping reports cache health. The in-memory backend is always healthy.
func (c *Client) Ping(ctx context.Context) error { return c.store.Ping(ctx) }

// Close releases the Redis connection pool when Redis is in use. It is a no-op
// for the in-memory fallback.
func (c *Client) Close() error {
	if c.closer == nil {
		return nil
	}
	return c.closer.Close()
}

// Describe returns a short, log-safe summary of the active backend.
func (c *Client) Describe() string {
	return fmt.Sprintf("cache backend=%s", c.backend)
}

// Get implements Store.
func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	return c.store.Get(ctx, key)
}

// Set implements Store.
func (c *Client) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.store.Set(ctx, key, value, ttl)
}

// Delete implements Store.
func (c *Client) Delete(ctx context.Context, key string) error {
	return c.store.Delete(ctx, key)
}

// Increment implements RateLimiterStore by forwarding to the active store.
func (c *Client) Increment(ctx context.Context, key string, ttl time.Duration) (int64, time.Duration, error) {
	limiter, ok := c.store.(RateLimiterStore)
	if !ok {
		return 0, 0, ErrUnsupported
	}
	return limiter.Increment(ctx, key, ttl)
}

// TryLock implements Locker by forwarding to the active store.
func (c *Client) TryLock(ctx context.Context, key, token string, ttl time.Duration) (bool, error) {
	locker, ok := c.store.(Locker)
	if !ok {
		return false, ErrUnsupported
	}
	return locker.TryLock(ctx, key, token, ttl)
}

// Unlock implements Locker by forwarding to the active store.
func (c *Client) Unlock(ctx context.Context, key, token string) error {
	locker, ok := c.store.(Locker)
	if !ok {
		return ErrUnsupported
	}
	return locker.Unlock(ctx, key, token)
}

func (c *Client) Publish(ctx context.Context, topic string, message []byte) error {
	broker, ok := c.store.(PubSubStore)
	if !ok {
		return ErrUnsupported
	}
	return broker.Publish(ctx, topic, message)
}

func (c *Client) Subscribe(ctx context.Context, topic string) (PubSubSubscription, error) {
	broker, ok := c.store.(PubSubStore)
	if !ok {
		return nil, ErrUnsupported
	}
	return broker.Subscribe(ctx, topic)
}
