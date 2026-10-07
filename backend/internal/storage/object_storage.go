package storage

import (
	"bytes"
	"context"
	"io"
	"path"
	"strings"
	"sync"
)

// MemoryStore is a process-local ObjectStore for development and tests.
// It intentionally avoids cloud provider behavior until a provider is wired.
type MemoryStore struct {
	mu      sync.RWMutex
	objects map[string]memoryObject
}

type memoryObject struct {
	contentType string
	body        []byte
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{objects: make(map[string]memoryObject)}
}

func (s *MemoryStore) Put(ctx context.Context, object PutObject) (ObjectInfo, error) {
	key, err := cleanKey(object.Key)
	if err != nil {
		return ObjectInfo{}, err
	}
	body, err := io.ReadAll(object.Body)
	if err != nil {
		return ObjectInfo{}, err
	}
	if err := ctx.Err(); err != nil {
		return ObjectInfo{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.objects[key] = memoryObject{contentType: strings.TrimSpace(object.ContentType), body: append([]byte(nil), body...)}
	return ObjectInfo{Key: key, ContentType: object.ContentType, Size: int64(len(body))}, nil
}

func (s *MemoryStore) Get(ctx context.Context, key string) (Object, error) {
	cleaned, err := cleanKey(key)
	if err != nil {
		return Object{}, err
	}
	if err := ctx.Err(); err != nil {
		return Object{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	object, ok := s.objects[cleaned]
	if !ok {
		return Object{}, ErrNotFound
	}
	body := append([]byte(nil), object.body...)
	return Object{
		Key: cleaned, ContentType: object.contentType, Size: int64(len(body)),
		Body: io.NopCloser(bytes.NewReader(body)),
	}, nil
}

func (s *MemoryStore) Delete(ctx context.Context, key string) error {
	cleaned, err := cleanKey(key)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.objects, cleaned)
	return nil
}

func cleanKey(key string) (string, error) {
	key = strings.TrimSpace(strings.ReplaceAll(key, "\\", "/"))
	for _, segment := range strings.Split(key, "/") {
		if segment == ".." {
			return "", ErrInvalidKey
		}
	}
	key = path.Clean("/" + key)
	key = strings.TrimPrefix(key, "/")
	if key == "" || key == "." {
		return "", ErrInvalidKey
	}
	return key, nil
}
