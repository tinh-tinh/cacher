package pebble

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"sync"
	"time"

	pebble_store "github.com/cockroachdb/pebble"
	"github.com/tinh-tinh/cacher"
)

type Pebble[M any] struct {
	db          *pebble_store.DB
	mutex       sync.Mutex
	Ttl         time.Duration
	CompressAlg cacher.CompressAlg
	Hooks       []cacher.Hook
}

type Options struct {
	Connect     *pebble_store.Options
	Ttl         time.Duration
	CompressAlg cacher.CompressAlg
	Hooks       []cacher.Hook
}

func New[M any](name string, opt Options) cacher.Store[M] {
	db, err := pebble_store.Open(name, opt.Connect)
	if err != nil {
		return nil
	}
	return &Pebble[M]{
		db:          db,
		mutex:       sync.Mutex{},
		Ttl:         opt.Ttl,
		CompressAlg: opt.CompressAlg,
		Hooks:       opt.Hooks,
	}
}

func (p *Pebble[M]) SetOptions(opt cacher.StoreOptions) {
	p.Ttl = opt.Ttl
	p.CompressAlg = opt.CompressAlg
	p.Hooks = opt.Hooks
}

func (p *Pebble[M]) Set(ctx context.Context, key string, val M, opts ...cacher.StoreOptions) error {
	cacher.HandlerBeforeSet(p, key, val)

	var value []byte
	valStr, err := json.Marshal(&val)
	if err != nil {
		return err
	}
	value = valStr
	// Handler
	if p.CompressAlg != "" {
		b, err := cacher.Compress(val, p.CompressAlg)
		if err != nil {
			return err
		}
		value = b
	}

	var ttl time.Duration
	if len(opts) > 0 {
		ttl = opts[0].Ttl
	} else {
		ttl = p.Ttl
	}

	exp := time.Now().Add(ttl).Unix()
	expBinary := make([]byte, 8)
	binary.BigEndian.PutUint64(expBinary, uint64(exp))
	valueWithTTL := append(expBinary, value...)

	p.mutex.Lock()
	err = p.db.Set([]byte(key), valueWithTTL, pebble_store.Sync)
	p.mutex.Unlock()
	if err != nil {
		return err
	}

	cacher.HandlerAfterSet(p, key, val)
	return nil
}

func (p *Pebble[M]) MSet(ctx context.Context, data ...cacher.Params[M]) error {
	for _, d := range data {
		err := p.Set(ctx, d.Key, d.Val, d.Options)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Pebble[M]) Get(ctx context.Context, key string) (M, error) {
	cacher.HandlerBeforeGet(p, key)

	// Handler
	snap := p.db.NewSnapshot()
	defer snap.Close()

	value, closer, err := snap.Get([]byte(key))
	if err != nil {
		if err == pebble_store.ErrNotFound {
			return *new(M), nil
		}
		return *new(M), err
	}
	defer closer.Close()

	if len(value) < 8 {
		return *new(M), errors.New("malformed value")
	}

	exp := int64(binary.BigEndian.Uint64(value[:8]))
	if exp < time.Now().Unix() {
		p.db.Delete([]byte(key), pebble_store.Sync)
		return *new(M), nil
	}
	cacheValue := value[8:]

	var schema M
	err = json.Unmarshal(cacheValue, &schema)
	if err != nil {
		if p.CompressAlg != "" {
			schema, err = cacher.Decompress[M](cacheValue, p.CompressAlg)
			if err != nil {
				return *new(M), err
			}
		} else {
			return *new(M), err
		}
	}

	cacher.HandlerAfterGet(p, key, schema)
	return schema, nil
}

func (p *Pebble[M]) MGet(ctx context.Context, keys ...string) ([]M, error) {
	var output []M
	for _, key := range keys {
		schema, err := p.Get(ctx, key)
		if err != nil {
			return nil, err
		}
		output = append(output, schema)
	}
	return output, nil
}

func (p *Pebble[M]) Delete(ctx context.Context, key string) error {
	err := p.db.Delete([]byte(key), pebble_store.Sync)
	if err != nil {
		return err
	}
	return nil
}

func (p *Pebble[M]) Clear(ctx context.Context) error {
	// Handler
	p.db.Flush()
	return nil
}

func (p *Pebble[M]) GetHooks() []cacher.Hook {
	return p.Hooks
}

func (p *Pebble[M]) GetConnect() interface{} {
	return p.db
}
