package memcache

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	memcache_store "github.com/bradfitz/gomemcache/memcache"
	"github.com/tinh-tinh/cacher"
)

type Options struct {
	Addr        []string
	Ttl         time.Duration
	Hooks       []cacher.Hook
	CompressAlg cacher.CompressAlg
}

func New[M any](opt Options) cacher.Store[M] {
	if opt.CompressAlg != "" && !cacher.IsValidAlg(opt.CompressAlg) {
		return nil
	}

	client := memcache_store.New(opt.Addr...)
	return &Memcache[M]{
		client:      client,
		ttl:         opt.Ttl,
		hooks:       opt.Hooks,
		CompressAlg: opt.CompressAlg,
	}
}

type Memcache[M any] struct {
	client      *memcache_store.Client
	ttl         time.Duration
	hooks       []cacher.Hook
	CompressAlg cacher.CompressAlg
}

func (m *Memcache[M]) SetOptions(opt cacher.StoreOptions) {
	m.ttl = opt.Ttl
	m.hooks = opt.Hooks
	m.CompressAlg = opt.CompressAlg
}

func (m *Memcache[M]) Get(ctx context.Context, key string) (M, error) {
	cacher.HandlerBeforeGet(m, key)

	// Handler
	val, err := m.client.Get(key)
	if err != nil {
		if err == memcache_store.ErrCacheMiss {
			return *new(M), nil
		}
		return *new(M), err
	}

	var schema M
	err = json.Unmarshal(val.Value, &schema)
	if err != nil {
		if m.CompressAlg != "" {
			schema, err = cacher.Decompress[M](val.Value, m.CompressAlg)
			if err != nil {
				return *new(M), err
			}
		} else {
			return *new(M), err
		}
	}

	cacher.HandlerAfterGet(m, key, schema)
	return schema, nil
}

func (m *Memcache[M]) MGet(ctx context.Context, keys ...string) ([]M, error) {
	var output []M
	for _, key := range keys {
		schema, err := m.Get(ctx, key)
		if err != nil {
			return nil, err
		}
		if !reflect.ValueOf(schema).IsZero() {
			output = append(output, schema)
		}
	}
	return output, nil
}

func (m *Memcache[M]) Set(ctx context.Context, key string, val M, opts ...cacher.StoreOptions) error {
	cacher.HandlerBeforeSet(m, key, val)

	// Handler
	var value []byte
	valStr, err := json.Marshal(val)
	if err != nil {
		return err
	}
	value = valStr
	if m.CompressAlg != "" {
		b, err := cacher.Compress(val, m.CompressAlg)
		if err != nil {
			return err
		}
		value = b
	}

	var ttl time.Duration
	if len(opts) > 0 {
		ttl = opts[0].Ttl
	} else {
		ttl = m.ttl
	}

	err = m.client.Set(&memcache_store.Item{
		Key:        key,
		Value:      value,
		Expiration: int32(ttl.Seconds()),
	})
	if err != nil {
		return err
	}

	cacher.HandlerAfterSet(m, key, val)
	return nil
}

func (m *Memcache[M]) MSet(ctx context.Context, data ...cacher.Params[M]) error {
	for _, d := range data {
		err := m.Set(ctx, d.Key, d.Val)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Memcache[M]) Delete(ctx context.Context, key string) error {
	cacher.HandlerBeforeDelete(m, key)
	// Handler
	m.client.Delete(key)

	cacher.HandlerAfterDelete(m, key)
	return nil
}

func (m *Memcache[M]) Clear(ctx context.Context) error {
	// Handler
	m.client.DeleteAll()
	return nil
}

func (m *Memcache[M]) GetHooks() []cacher.Hook {
	return m.hooks
}
