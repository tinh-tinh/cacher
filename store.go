package cacher

import (
	"context"
	"time"
)

type StoreOptions struct {
	Ttl time.Duration
}

type Store interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, opts ...StoreOptions) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
}
