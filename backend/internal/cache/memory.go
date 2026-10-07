package cache

import (
	"bytes"
	"context"
	"sync"
	"time"
)

type memoryItem struct {
	value     []byte
	expiresAt time.Time
}

type MemoryStore struct {
	mu    sync.Mutex
	items map[string]memoryItem
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{items: make(map[string]memoryItem)}
}

func (s *MemoryStore) Get(ctx context.Context, key string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[key]
	if !ok || expired(item, time.Now()) {
		delete(s.items, key)
		return nil, ErrCacheMiss
	}
	return bytes.Clone(item.value), nil
}

func (s *MemoryStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key] = memoryItem{value: bytes.Clone(value), expiresAt: expiresAt}
	return nil
}

func (s *MemoryStore) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key)
	return nil
}

func (s *MemoryStore) Ping(ctx context.Context) error {
	return ctx.Err()
}

func (s *MemoryStore) Increment(ctx context.Context, key string, ttl time.Duration) (int64, time.Duration, error) {
	if err := ctx.Err(); err != nil {
		return 0, 0, err
	}
	if ttl <= 0 {
		ttl = time.Minute
	}
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[key]
	if !ok || expired(item, now) {
		item = memoryItem{value: []byte("0"), expiresAt: now.Add(ttl)}
	}
	count := parsePositiveInt(item.value) + 1
	item.value = []byte(formatInt(count))
	s.items[key] = item
	return count, time.Until(item.expiresAt), nil
}

func (s *MemoryStore) TryLock(ctx context.Context, key, token string, ttl time.Duration) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[key]
	if ok && !expired(item, now) {
		return false, nil
	}
	s.items[key] = memoryItem{value: []byte(token), expiresAt: now.Add(ttl)}
	return true, nil
}

func (s *MemoryStore) Unlock(ctx context.Context, key, token string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[key]
	if ok && string(item.value) == token {
		delete(s.items, key)
	}
	return nil
}

func expired(item memoryItem, now time.Time) bool {
	return !item.expiresAt.IsZero() && !item.expiresAt.After(now)
}

func parsePositiveInt(value []byte) int64 {
	var count int64
	for _, digit := range value {
		if digit < '0' || digit > '9' {
			return 0
		}
		count = count*10 + int64(digit-'0')
	}
	return count
}

func formatInt(value int64) string {
	if value == 0 {
		return "0"
	}
	var buf [20]byte
	index := len(buf)
	for value > 0 {
		index--
		buf[index] = byte('0' + value%10)
		value /= 10
	}
	return string(buf[index:])
}
