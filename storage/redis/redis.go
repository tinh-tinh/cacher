package redis

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/tinh-tinh/cacher"

	redis_store "github.com/redis/go-redis/v9"
)

type Options struct {
	Connect     *redis_store.Options
	Ttl         time.Duration
	CompressAlg cacher.CompressAlg
	Hooks       []cacher.Hook
}

func New[M any](opt Options) cacher.Store[M] {
	if opt.CompressAlg != "" && !cacher.IsValidAlg(opt.CompressAlg) {
		return nil
	}

	client := redis_store.NewClient(opt.Connect)
	return &Redis[M]{
		client:      client,
		ttl:         opt.Ttl,
		hooks:       opt.Hooks,
		CompressAlg: opt.CompressAlg,
	}
}

type Redis[M any] struct {
	client      *redis_store.Client
	ttl         time.Duration
	hooks       []cacher.Hook
	CompressAlg cacher.CompressAlg
}

func (r *Redis[M]) Get(ctx context.Context, key string) (M, error) {
	cacher.HandlerBeforeGet(r, key)

	// Handler
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis_store.Nil {
			return *new(M), nil
		}
		return *new(M), err
	}

	var schema M
	err = json.Unmarshal([]byte(val), &schema)
	if err != nil {
		if r.CompressAlg != "" {
			schema, err = cacher.Decompress[M]([]byte(val), r.CompressAlg)
			if err != nil {
				return *new(M), err
			}
		} else {
			return *new(M), err
		}
	}

	cacher.HandlerAfterGet(r, key, schema)
	return schema, nil
}

func (r *Redis[M]) MGet(ctx context.Context, keys ...string) ([]M, error) {
	var output []M
	for _, key := range keys {
		schema, err := r.Get(ctx, key)
		if err != nil {
			return nil, err
		}
		if !reflect.ValueOf(schema).IsZero() {
			output = append(output, schema)
		}
	}
	return output, nil
}

func (r *Redis[M]) Set(ctx context.Context, key string, val M, opts ...cacher.StoreOptions) error {
	cacher.HandlerBeforeSet(r, key, val)

	var value interface{}
	valStr, err := json.Marshal(&val)
	if err != nil {
		return err
	}
	value = string(valStr)
	// Handler
	if r.CompressAlg != "" {
		b, err := cacher.Compress(val, r.CompressAlg)
		if err != nil {
			return err
		}
		value = b
	}

	var ttl time.Duration
	if len(opts) > 0 {
		ttl = opts[0].Ttl
	} else {
		ttl = r.ttl
	}
	err = r.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return err
	}

	cacher.HandlerAfterSet(r, key, val)
	return nil
}

func (r *Redis[M]) MSet(ctx context.Context, data ...cacher.Params[M]) error {
	for _, d := range data {
		err := r.Set(ctx, d.Key, d.Val, d.Options)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Redis[M]) Delete(ctx context.Context, key string) error {
	cacher.HandlerBeforeDelete(r, key)
	// Handler
	r.client.Del(ctx, key).Err()

	cacher.HandlerAfterDelete(r, key)
	return nil
}

func (r *Redis[M]) Clear(ctx context.Context) error {
	// Handler
	r.client.FlushDB(ctx).Err()
	return nil
}

func (r *Redis[M]) SetOptions(option cacher.StoreOptions) {
	if option.CompressAlg != "" && cacher.IsValidAlg(option.CompressAlg) {
		r.CompressAlg = option.CompressAlg
	}

	if option.Ttl > 0 {
		r.ttl = option.Ttl
	}

	if option.Hooks != nil {
		r.hooks = option.Hooks
	}
}

func (r *Redis[M]) GetHooks() []cacher.Hook {
	return r.hooks
}

func (r *Redis[M]) GetConnect() interface{} {
	return r.client
}
