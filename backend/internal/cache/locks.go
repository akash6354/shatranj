package cache

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// ErrNoLocker is returned when a lock is created without a backend that
// implements the Locker contract.
var ErrNoLocker = errors.New("cache: store does not support locks")

// DefaultLockTTL bounds how long a lock survives if its owner crashes without
// releasing it.
const DefaultLockTTL = 30 * time.Second

// Lock is a named, token-owned mutual exclusion handle. Locks are best-effort:
// they reduce duplicate work across processes but never replace the durable
// guarantees of the database.
type Lock struct {
	locker Locker
	key    string
	token  string
	ttl    time.Duration
	held   bool
}

// NewLock prepares a lock for name. It does not acquire the lock; call
// Acquire. The store must implement Locker or NewLock reports ErrNoLocker.
func NewLock(store Store, name string, ttl time.Duration) (*Lock, error) {
	locker, ok := store.(Locker)
	if !ok {
		return nil, ErrNoLocker
	}
	if name == "" {
		return nil, errors.New("cache: lock name is required")
	}
	if ttl <= 0 {
		ttl = DefaultLockTTL
	}
	token, err := randomToken()
	if err != nil {
		return nil, err
	}
	return &Lock{locker: locker, key: LockKey(name), token: token, ttl: ttl}, nil
}

// Acquire attempts to take the lock. It returns false (without error) when the
// lock is already held by another owner.
func (l *Lock) Acquire(ctx context.Context) (bool, error) {
	acquired, err := l.locker.TryLock(ctx, l.key, l.token, l.ttl)
	if err != nil {
		return false, err
	}
	l.held = acquired
	return acquired, nil
}

// Release frees the lock if this handle still owns it. Releasing a lock that
// was never acquired, or that has already expired and been re-taken by another
// owner, is a safe no-op.
func (l *Lock) Release(ctx context.Context) error {
	if !l.held {
		return nil
	}
	l.held = false
	return l.locker.Unlock(ctx, l.key, l.token)
}

// Held reports whether this handle currently believes it owns the lock.
func (l *Lock) Held() bool { return l.held }

// Key returns the fully-qualified cache key for the lock.
func (l *Lock) Key() string { return l.key }

// WithLock runs fn while holding the lock for name, releasing it afterwards.
// When the lock cannot be taken, fn is not called and ran is false. A backend
// that does not implement Locker yields ErrNoLocker.
func WithLock(ctx context.Context, store Store, name string, ttl time.Duration, fn func(context.Context) error) (ran bool, err error) {
	lock, err := NewLock(store, name, ttl)
	if err != nil {
		return false, err
	}
	acquired, err := lock.Acquire(ctx)
	if err != nil || !acquired {
		return false, err
	}
	defer func() {
		if releaseErr := lock.Release(ctx); releaseErr != nil && err == nil {
			err = fmt.Errorf("release lock %q: %w", name, releaseErr)
		}
	}()
	return true, fn(ctx)
}

func randomToken() (string, error) {
	var buffer [16]byte
	if _, err := rand.Read(buffer[:]); err != nil {
		return "", fmt.Errorf("generate lock token: %w", err)
	}
	return hex.EncodeToString(buffer[:]), nil
}
