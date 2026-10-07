package cache

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestOpenWithoutRedisUsesMemoryFallback(t *testing.T) {
	client, err := Open(context.Background(), "")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if client.RedisBacked() || client.Backend() != BackendMemory {
		t.Fatalf("backend = %s, want memory", client.Backend())
	}
	if err := client.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestOpenWithUnreachableRedisFallsBackWithoutFailing(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	client, err := Open(ctx, "redis://127.0.0.1:1/0")
	if err == nil {
		t.Fatal("Open() did not report the unreachable Redis server")
	}
	if client == nil {
		t.Fatal("Open() returned a nil client for an unreachable Redis server")
	}
	if client.RedisBacked() || client.Backend() != BackendMemory {
		t.Fatalf("backend = %s, want memory fallback", client.Backend())
	}
	if pingErr := client.Ping(ctx); pingErr != nil {
		t.Fatalf("memory fallback Ping() error = %v", pingErr)
	}
}

func TestOpenRejectsMalformedRedisURL(t *testing.T) {
	client, err := Open(context.Background(), "://not-a-url")
	if err == nil {
		t.Fatal("Open() accepted a malformed Redis URL")
	}
	if client == nil || client.RedisBacked() {
		t.Fatal("Open() did not fall back to memory for a malformed URL")
	}
}

func TestClientSatisfiesCacheInterfaces(t *testing.T) {
	client, err := Open(context.Background(), "")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	var _ Store = client
	var _ RateLimiterStore = client
	var _ Locker = client

	ctx := context.Background()
	if err := client.Set(ctx, "k", []byte("v"), time.Minute); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	value, err := client.Get(ctx, "k")
	if err != nil || string(value) != "v" {
		t.Fatalf("Get() = %q, %v", value, err)
	}
	count, _, err := client.Increment(ctx, "counter", time.Minute)
	if err != nil || count != 1 {
		t.Fatalf("Increment() = %d, %v", count, err)
	}
	if err := client.Delete(ctx, "k"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := client.Get(ctx, "k"); !errors.Is(err, ErrCacheMiss) {
		t.Fatalf("Get() after Delete error = %v, want ErrCacheMiss", err)
	}
}

func TestWithLockRunsOnceAndReleases(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	ran, err := WithLock(ctx, store, "matchmaking:join", time.Minute, func(context.Context) error {
		taken, lockErr := store.TryLock(ctx, LockKey("matchmaking:join"), "other", time.Minute)
		if lockErr != nil {
			return lockErr
		}
		if taken {
			t.Fatal("lock was re-acquired while held by WithLock")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WithLock() error = %v", err)
	}
	if !ran {
		t.Fatal("WithLock() did not run the locked function")
	}
	taken, err := store.TryLock(ctx, LockKey("matchmaking:join"), "after", time.Minute)
	if err != nil || !taken {
		t.Fatalf("lock not released after WithLock: taken=%v err=%v", taken, err)
	}
}

func TestWithLockSkipsWhenAlreadyHeld(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	if err := store.Set(ctx, LockKey("job:1"), nil, time.Minute); err != nil {
		t.Fatalf("seed lock error = %v", err)
	}
	called := false
	ran, err := WithLock(ctx, store, "job:1", time.Minute, func(context.Context) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("WithLock() error = %v", err)
	}
	if ran || called {
		t.Fatal("WithLock() ran while the lock was already held")
	}
}

func TestWithLockRejectsStoreWithoutLocker(t *testing.T) {
	ran, err := WithLock(context.Background(), onlyStore{}, "x", time.Minute, func(context.Context) error {
		return nil
	})
	if !errors.Is(err, ErrNoLocker) {
		t.Fatalf("WithLock() error = %v, want ErrNoLocker", err)
	}
	if ran {
		t.Fatal("WithLock() ran against a store without lock support")
	}
}

// onlyStore implements Store but deliberately not Locker.
type onlyStore struct{}

func (onlyStore) Get(context.Context, string) ([]byte, error) { return nil, ErrCacheMiss }
func (onlyStore) Set(context.Context, string, []byte, time.Duration) error {
	return nil
}
func (onlyStore) Delete(context.Context, string) error { return nil }
func (onlyStore) Ping(context.Context) error           { return nil }
