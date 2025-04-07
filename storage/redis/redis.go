package redis

import (
	"context"
	"time"

	"github.com/tinh-tinh/cacher/v2"

	redis_store "github.com/redis/go-redis/v9"
)

type Options struct {
	Connect *redis_store.Options
	Ttl     time.Duration
}

func New(opt Options) cacher.Store {
	client := redis_store.NewClient(opt.Connect)
	return &Redis{
		client: client,
		ttl:    opt.Ttl,
	}
}

type Redis struct {
	client *redis_store.Client
	ttl    time.Duration
}

func (r *Redis) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis_store.Nil {
			return nil, nil
		}
		return nil, err
	}

	return val, nil
}

func (r *Redis) Set(ctx context.Context, key string, val []byte, opts ...cacher.StoreOptions) error {
	var ttl time.Duration
	if len(opts) > 0 && opts[0].Ttl > 0 {
		ttl = opts[0].Ttl
	} else {
		ttl = r.ttl
	}
	err := r.client.Set(ctx, key, val, ttl).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *Redis) Delete(ctx context.Context, key string) error {
	// Handler
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *Redis) Clear(ctx context.Context) error {
	// Handler
	r.client.FlushDB(ctx).Err()
	return nil
}
