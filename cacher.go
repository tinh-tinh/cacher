package cacher

import (
	"context"
	"errors"
	"time"

	"github.com/tinh-tinh/tinhtinh/common/memory"
)

type Cacher[M any] struct {
	Store memory.Store
	ctx   context.Context
}

func New[M any](opt memory.Options) *Cacher[M] {
	return &Cacher[M]{
		Store: *memory.New(opt),
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
	data, ok := c.Store.Get(key).(M)
	if !ok {
		return *new(M), errors.New("key not found")
	}
	return data, nil
}

func (c *Cacher[M]) Set(key string, value M, ttl ...time.Duration) {
	c.Store.Set(key, value, ttl...)
}

func (c *Cacher[M]) Delete(key string) {
	c.Store.Delete(key)
}

func (c *Cacher[M]) Clear() {
	c.Store.Clear()
}
