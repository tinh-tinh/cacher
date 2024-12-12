package cacher

import (
	"context"
	"time"
)

type StoreOptions struct {
	Ttl         time.Duration
	CompressAlg CompressAlg
	Hooks       []Hook
}

type Params[M any] struct {
	Key     string
	Val     M
	Options StoreOptions
}

type Store[M any] interface {
	SetOptions(option StoreOptions)
	Get(ctx context.Context, key string) (M, error)
	MGet(ctx context.Context, keys ...string) ([]M, error)
	Set(ctx context.Context, key string, value M, opts ...StoreOptions) error
	MSet(ctx context.Context, data ...Params[M]) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	GetHooks() []Hook
}
