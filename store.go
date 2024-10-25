package cacher

import (
	"context"
	"time"
)

type StoreOptions struct {
	Ttl         time.Duration
	CompressAlg string
}

type Store[M any] interface {
	Get(ctx context.Context, key string) (M, error)
	Set(ctx context.Context, key string, value M, opts ...StoreOptions) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
}
