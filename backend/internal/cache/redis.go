package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore is the Redis-backed Store implementation.
type RedisStore struct {
	client *redis.Client
}

type redisSubscription struct {
	pubsub   *redis.PubSub
	messages chan []byte
	stop     context.CancelFunc
}

func (s *redisSubscription) Messages() <-chan []byte { return s.messages }

func (s *redisSubscription) Close() error {
	s.stop()
	return s.pubsub.Close()
}

// OpenStore selects a backend for rawURL. An empty rawURL yields the in-memory
// store with a nil error. A configured but unusable Redis also yields the
// in-memory store; the returned bool is false and the error explains why, so
// callers can fall back without failing startup.
func OpenStore(ctx context.Context, rawURL string) (Store, bool, error) {
	if rawURL == "" {
		return NewMemoryStore(), false, nil
	}
	options, err := redis.ParseURL(rawURL)
	if err != nil {
		return NewMemoryStore(), false, fmt.Errorf("parse Redis URL: %w", err)
	}
	client := redis.NewClient(options)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return NewMemoryStore(), false, fmt.Errorf("ping Redis: %w", err)
	}
	return &RedisStore{client: client}, true, nil
}

func (s *RedisStore) Close() error {
	return s.client.Close()
}

func (s *RedisStore) Get(ctx context.Context, key string) ([]byte, error) {
	value, err := s.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrCacheMiss
	}
	return value, err
}

func (s *RedisStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return s.client.Set(ctx, key, value, ttl).Err()
}

func (s *RedisStore) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}

func (s *RedisStore) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

func (s *RedisStore) Increment(ctx context.Context, key string, ttl time.Duration) (int64, time.Duration, error) {
	if ttl <= 0 {
		ttl = time.Minute
	}
	count, err := s.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, 0, err
	}
	if count == 1 {
		if err := s.client.Expire(ctx, key, ttl).Err(); err != nil {
			return 0, 0, err
		}
		return count, ttl, nil
	}
	remaining, err := s.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, 0, err
	}
	if remaining < 0 {
		remaining = ttl
		_ = s.client.Expire(ctx, key, ttl).Err()
	}
	return count, remaining, nil
}

func (s *RedisStore) TryLock(ctx context.Context, key, token string, ttl time.Duration) (bool, error) {
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	return s.client.SetNX(ctx, key, token, ttl).Result()
}

func (s *RedisStore) Unlock(ctx context.Context, key, token string) error {
	const script = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0`
	return s.client.Eval(ctx, script, []string{key}, token).Err()
}

func (s *RedisStore) Publish(ctx context.Context, topic string, message []byte) error {
	return s.client.Publish(ctx, topic, message).Err()
}

func (s *RedisStore) Subscribe(ctx context.Context, topic string) (PubSubSubscription, error) {
	pubsub := s.client.Subscribe(ctx, topic)
	if _, err := pubsub.Receive(ctx); err != nil {
		_ = pubsub.Close()
		return nil, err
	}
	subscriptionCtx, stop := context.WithCancel(ctx)
	subscription := &redisSubscription{
		pubsub: pubsub, messages: make(chan []byte, 128), stop: stop,
	}
	go func() {
		defer close(subscription.messages)
		for message := range pubsub.Channel() {
			select {
			case subscription.messages <- []byte(message.Payload):
			case <-subscriptionCtx.Done():
				return
			}
		}
	}()
	return subscription, nil
}
