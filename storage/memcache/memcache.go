package memcache

import (
	"context"
	"time"

	memcache_store "github.com/bradfitz/gomemcache/memcache"
	"github.com/tinh-tinh/cacher/v2"
)

type Options struct {
	Addr  []string
	Ttl   time.Duration
	Hooks []cacher.Hook
}

func New[M any](opt Options) cacher.Store {
	client := memcache_store.New(opt.Addr...)
	return &Memcache{
		client: client,
		ttl:    opt.Ttl,
	}
}

type Memcache struct {
	client *memcache_store.Client
	ttl    time.Duration
}

func (m *Memcache) SetOptions(opt cacher.StoreOptions) {
	m.ttl = opt.Ttl
}

func (m *Memcache) Get(ctx context.Context, key string) ([]byte, error) {
	// Handler
	val, err := m.client.Get(key)
	if err != nil {
		if err == memcache_store.ErrCacheMiss {
			return nil, nil
		}
		return nil, err
	}

	return val.Value, nil
}

func (m *Memcache) Set(ctx context.Context, key string, val []byte, opts ...cacher.StoreOptions) error {
	var ttl time.Duration
	if len(opts) > 0 {
		ttl = opts[0].Ttl
	} else {
		ttl = m.ttl
	}

	err := m.client.Set(&memcache_store.Item{
		Key:        key,
		Value:      val,
		Expiration: int32(ttl.Seconds()),
	})
	if err != nil {
		return err
	}

	return nil
}

func (m *Memcache) Delete(ctx context.Context, key string) error {
	// Handler
	err := m.client.Delete(key)
	if err != nil {
		return err
	}

	return nil
}

func (m *Memcache) Clear(ctx context.Context) error {
	// Handler
	m.client.DeleteAll()
	return nil
}
