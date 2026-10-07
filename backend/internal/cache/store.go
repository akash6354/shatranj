package cache

import (
	"context"
	"errors"
	"time"
)

// ErrCacheMiss is returned by Get when a key is absent or expired.
var ErrCacheMiss = errors.New("cache miss")

// ErrUnsupported is returned when a backend does not implement an optional
// capability a caller requested.
var ErrUnsupported = errors.New("cache: operation not supported by store")

// Store is the minimal key/value contract every cache backend implements.
type Store interface {
	Get(context.Context, string) ([]byte, error)
	Set(context.Context, string, []byte, time.Duration) error
	Delete(context.Context, string) error
	Ping(context.Context) error
}

// RateLimiterStore is implemented by stores that can atomically count requests
// within a fixed window.
type RateLimiterStore interface {
	Increment(context.Context, string, time.Duration) (int64, time.Duration, error)
}

// Locker is implemented by stores that can hand out short-lived mutual
// exclusion locks keyed by name.
type Locker interface {
	TryLock(context.Context, string, string, time.Duration) (bool, error)
	Unlock(context.Context, string, string) error
}

// PubSubSubscription provides messages from a shared broker topic.
type PubSubSubscription interface {
	Messages() <-chan []byte
	Close() error
}

// PubSubStore is an optional shared-event capability implemented by Redis.
type PubSubStore interface {
	Publish(context.Context, string, []byte) error
	Subscribe(context.Context, string) (PubSubSubscription, error)
}

// Closer is implemented by stores that own an external resource such as a
// Redis connection pool.
type Closer interface {
	Close() error
}

func prefixed(key string) string {
	return keyPrefix + key
}
