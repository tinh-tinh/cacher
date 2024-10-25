package cacher

import (
	"context"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
)

type Cacher[M any] struct {
	Store *cache.Cache[M]
	ctx   context.Context
}

func New[M any](store store.StoreInterface) *Cacher[M] {
	return &Cacher[M]{
		Store: cache.New[M](store),
		ctx:   context.Background(),
	}
}

func (c *Cacher[M]) SetCtx(ctx context.Context) {
	c.ctx = ctx
}

func (c *Cacher[M]) GetCtx() context.Context {
	return c.ctx
}

func (c *Cacher[M]) Get(key string) (M, error) {
	return c.Store.Get(c.ctx, key)
}

func (c *Cacher[M]) Set(key string, value M, options ...store.Option) error {
	return c.Store.Set(c.ctx, key, value, options...)
}

func (c *Cacher[M]) Delete(key string) error {
	return c.Store.Delete(c.ctx, key)
}

func (c *Cacher[M]) Clear() error {
	return c.Store.Clear(c.ctx)
}
