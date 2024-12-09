package redis

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
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
	findHook := slices.IndexFunc(r.hooks, func(h cacher.Hook) bool {
		return h.Key == cacher.BeforeGet
	})
	if findHook != -1 {
		r.hooks[findHook].Fnc(key, nil)
	}
	// Handler
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return *new(M), err
	}

	var schema M
	err = json.Unmarshal([]byte(val), &schema)
	if err != nil {
		if r.CompressAlg != "" {
			schema, err = r.decompress([]byte(val))
			if err != nil {
				return *new(M), err
			}
		} else {
			return *new(M), err
		}
	}

	findHook = slices.IndexFunc(r.hooks, func(h cacher.Hook) bool {
		return h.Key == cacher.AfterGet
	})
	if findHook != -1 {
		r.hooks[findHook].Fnc(key, schema)
	}

	return schema, nil
}

func (r *Redis[M]) MGet(ctx context.Context, keys ...string) ([]M, error) {
	var output []M
	for _, key := range keys {
		schema, err := r.Get(ctx, key)
		if err != nil {
			return nil, err
		}
		output = append(output, schema)
	}
	return output, nil
}

func (r *Redis[M]) Set(ctx context.Context, key string, val M, opts ...cacher.StoreOptions) error {
	findHook := slices.IndexFunc(r.hooks, func(h cacher.Hook) bool {
		return h.Key == cacher.BeforeSet
	})
	if findHook != -1 {
		r.hooks[findHook].Fnc(key, val)
	}

	var value interface{}
	valStr, err := json.Marshal(&val)
	if err != nil {
		return err
	}
	value = string(valStr)
	// Handler
	if r.CompressAlg != "" {
		b, err := r.compress(val)
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

	findHook = slices.IndexFunc(r.hooks, func(h cacher.Hook) bool {
		return h.Key == cacher.AfterSet
	})
	if findHook != -1 {
		r.hooks[findHook].Fnc(key, val)
	}
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
	findHook := slices.IndexFunc(r.hooks, func(h cacher.Hook) bool {
		return h.Key == cacher.BeforeDelete
	})
	if findHook != -1 {
		r.hooks[findHook].Fnc(key, nil)
	}
	// Handler
	r.client.Del(ctx, key).Err()

	findHook = slices.IndexFunc(r.hooks, func(h cacher.Hook) bool {
		return h.Key == cacher.AfterDelete
	})
	if findHook != -1 {
		r.hooks[findHook].Fnc(key, nil)
	}
	return nil
}

func (r *Redis[M]) Clear(ctx context.Context) error {
	// Handler
	r.client.FlushDB(ctx).Err()
	return nil
}

func (r *Redis[M]) compress(data M) ([]byte, error) {
	input, err := cacher.ToBytes(data)
	if err != nil {
		return nil, err
	}
	switch r.CompressAlg {
	case cacher.CompressAlgGzip:
		return cacher.CompressGzip(input)
	case cacher.CompressAlgFlate:
		return cacher.CompressFlate(input)
	case cacher.CompressAlgZlib:
		return cacher.CompressZlib(input)
	default:
		return input, nil
	}
}

func (r *Redis[M]) decompress(dataRaw interface{}) (M, error) {
	dataByte, ok := dataRaw.([]byte)
	if !ok {
		return *new(M), errors.New("assert type failed")
	}

	var output []byte
	var err error

	switch r.CompressAlg {
	case cacher.CompressAlgGzip:
		output, err = cacher.DecompressGzip(dataByte)
		if err != nil {
			return *new(M), err
		}
	case cacher.CompressAlgFlate:
		output, err = cacher.DecompressFlate(dataByte)
		if err != nil {
			return *new(M), err
		}
	case cacher.CompressAlgZlib:
		output, err = cacher.DecompressZlib(dataByte)
		if err != nil {
			return *new(M), err
		}
	default:
		return *new(M), nil
	}

	dataRaw, err = cacher.FromBytes[M](output)
	if err != nil {
		return *new(M), err
	}

	data, ok := dataRaw.(M)
	if !ok {
		return *new(M), errors.New("assert type failed")
	}
	return data, nil
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
