package cacher

import (
	"context"
)

type Cacher[M any] struct {
	Store       Store[M]
	ctx         context.Context
	CompressAlg CompressAlg
}

func New[M any](opt StoreOptions) *Cacher[M] {
	return &Cacher[M]{
		Store:       NewInMemory[M](opt),
		ctx:         context.Background(),
		CompressAlg: opt.CompressAlg,
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

func (c *Cacher[M]) MGet(keys ...string) ([]M, error) {
	return c.Store.MGet(c.ctx, keys...)
}

func (c *Cacher[M]) Set(key string, value M, opts ...StoreOptions) error {
	return c.Store.Set(c.ctx, key, value, opts...)
}

func (c *Cacher[M]) MSet(data ...Params[M]) error {
	return c.Store.MSet(c.ctx, data...)
}

func (c *Cacher[M]) Delete(key string) error {
	return c.Store.Delete(c.ctx, key)
}

func (c *Cacher[M]) Clear() error {
	return c.Store.Clear(c.ctx)
}
