package cacher

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/tinh-tinh/tinhtinh/v2/common/compress"
	"github.com/tinh-tinh/tinhtinh/v2/common/era"
)

func NewInMemory[M any](opt StoreOptions) Store[M] {
	if opt.CompressAlg != "" && !IsValidAlg(opt.CompressAlg) {
		return nil
	}
	memory := &Memory[M]{
		ttl:         opt.Ttl,
		data:        make(map[string]item),
		CompressAlg: opt.CompressAlg,
		hooks:       opt.Hooks,
	}
	era.StartTimeStampUpdater()
	go memory.gc(1 * time.Second)
	return memory
}

type item struct {
	v interface{}
	e uint32
}

type Memory[M any] struct {
	sync.RWMutex
	ttl         time.Duration
	data        map[string]item
	CompressAlg compress.Alg
	hooks       []Hook
}

func (m *Memory[M]) SetOptions(option StoreOptions) {
	if option.CompressAlg != "" && IsValidAlg(option.CompressAlg) {
		m.CompressAlg = option.CompressAlg
	}

	if option.Ttl > 0 {
		m.ttl = option.Ttl
	}

	if option.Hooks != nil {
		m.hooks = option.Hooks
	}
}

func (m *Memory[M]) Set(ctx context.Context, key string, val M, opts ...StoreOptions) error {
	HandlerBeforeSet(m, key, val)
	// Handler
	var exp uint32
	if len(opts) > 0 {
		exp = uint32(opts[0].Ttl.Seconds()) + era.Timestamp()
	} else {
		exp = uint32(m.ttl.Seconds()) + era.Timestamp()
	}
	i := item{e: exp, v: val}
	if m.CompressAlg != "" {
		b, err := compress.Encode(val, m.CompressAlg)
		if err != nil {
			return err
		}
		i.v = b
	}
	m.Lock()
	m.data[key] = i
	m.Unlock()

	HandlerAfterSet(m, key, val)
	return nil
}

func (m *Memory[M]) MSet(ctx context.Context, data ...Params[M]) error {
	save := make(map[string]item)
	for _, d := range data {
		var exp uint32
		if d.Options.Ttl > 0 {
			exp = uint32(d.Options.Ttl.Seconds()) + era.Timestamp()
		} else {
			exp = uint32(m.ttl.Seconds()) + era.Timestamp()
		}
		i := item{e: exp, v: d.Val}
		if m.CompressAlg != "" {
			b, err := compress.Encode(d.Val, m.CompressAlg)
			if err != nil {
				return err
			}
			i.v = b
		}
		save[d.Key] = i
	}

	m.Lock()
	for k, v := range save {
		m.data[k] = v
	}
	m.Unlock()
	return nil
}

func (m *Memory[M]) Get(ctx context.Context, key string) (M, error) {
	HandlerBeforeGet(m, key)

	// Handler
	m.RLock()
	v, ok := m.data[key]
	m.RUnlock()
	if !ok || v.e != 0 && v.e <= era.Timestamp() {
		return *new(M), errors.New("key not found")
	}
	val, ok := v.v.(M)
	if !ok {
		if m.CompressAlg != "" {
			valByte, ok := v.v.([]byte)
			if !ok {
				return *new(M), errors.New("value save is not supported compress")
			}
			return compress.DecodeMarshall[M](valByte, m.CompressAlg)
		}
		return *new(M), errors.New("key not found")
	}

	HandlerAfterGet(m, key, val)
	return val, nil
}

func (m *Memory[M]) MGet(ctx context.Context, keys ...string) ([]M, error) {
	var output []item
	m.RLock()
	for _, key := range keys {
		val, ok := m.data[key]
		if ok {
			output = append(output, val)
		}
	}
	m.RUnlock()

	if len(output) == 0 {
		return nil, errors.New("key not found")
	}

	var data []M
	for _, v := range output {
		if v.e != 0 && v.e <= era.Timestamp() {
			continue
		}
		val, ok := v.v.(M)
		if !ok {
			if m.CompressAlg != "" {
				valByte, ok := v.v.([]byte)
				if !ok {
					continue
				}
				d, err := compress.DecodeMarshall[M](valByte, m.CompressAlg)
				if err != nil {
					continue
				}
				val = d
			} else {
				continue
			}
		}
		data = append(data, val)
	}

	return data, nil
}

func (m *Memory[M]) Delete(ctx context.Context, key string) error {
	HandlerBeforeDelete(m, key)

	// Handler
	m.Lock()
	delete(m.data, key)
	m.Unlock()

	HandlerAfterDelete(m, key)
	return nil
}

func (m *Memory[M]) Clear(ctx context.Context) error {
	md := make(map[string]item)
	m.Lock()
	m.data = md
	m.Unlock()
	return nil
}

func (m *Memory[M]) gc(sleep time.Duration) {
	ticker := time.NewTimer(sleep)
	defer ticker.Stop()
	var expired []string
	for range ticker.C {
		ts := era.Timestamp()
		expired = expired[:0]
		m.RLock()
		for key, v := range m.data {
			if v.e != 0 && v.e <= ts {
				expired = append(expired, key)
			}
		}
		m.RUnlock()
		m.Lock()
		for i := range expired {
			v := m.data[expired[i]]
			if v.e != 0 && v.e <= ts {
				delete(m.data, expired[i])
			}
		}
		m.Unlock()
	}
}

func (m *Memory[M]) GetHooks() []Hook {
	return m.hooks
}

func (m *Memory[M]) GetConnect() interface{} {
	return nil
}

func IsValidAlg(val compress.Alg) bool {
	return val == compress.Gzip || val == compress.Flate || val == compress.Zlib
}
