package storage

import (
	"context"
	"errors"
	"io"
)

var (
	ErrInvalidKey = errors.New("invalid storage key")
	ErrNotFound   = errors.New("object not found")
)

// Object describes a stored blob and its metadata.
type Object struct {
	Key         string
	ContentType string
	Size        int64
	Body        io.ReadCloser
}

// PutObject contains data and metadata for a stored object.
type PutObject struct {
	Key         string
	ContentType string
	Body        io.Reader
}

// ObjectStore is the storage boundary used by upload-capable modules.
type ObjectStore interface {
	Put(context.Context, PutObject) (ObjectInfo, error)
	Get(context.Context, string) (Object, error)
	Delete(context.Context, string) error
}

// ObjectInfo is safe to return from write operations without holding a body.
type ObjectInfo struct {
	Key         string `json:"key"`
	ContentType string `json:"content_type,omitempty"`
	Size        int64  `json:"size"`
	URL         string `json:"url,omitempty"`
}
