package cacher

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/tinh-tinh/tinhtinh/utils"
)

func NewInMemory[M any](opt StoreOptions) Store[M] {
	if opt.CompressAlg != "" && opt.CompressAlg != "zlib" && opt.CompressAlg != "flate" && opt.CompressAlg != "gzip" {
		return nil
	}
	memory := &Memory[M]{
		ttl:         opt.Ttl,
		data:        make(map[string]item),
		CompressAlg: opt.CompressAlg,
	}
	utils.StartTimeStampUpdater()
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
	CompressAlg string
}

func (m *Memory[M]) Set(ctx context.Context, key string, val M, opts ...StoreOptions) error {
	var exp uint32
	if len(opts) > 0 {
		exp = uint32(opts[0].Ttl.Seconds()) + utils.Timestamp()
	} else {
		exp = uint32(m.ttl.Seconds()) + utils.Timestamp()
	}
	i := item{e: exp, v: val}
	if m.CompressAlg != "" {
		b, err := m.compress(val)
		if err != nil {
			return err
		}
		i.v = b
	}
	m.Lock()
	m.data[key] = i
	m.Unlock()
	return nil
}

func (m *Memory[M]) Get(ctx context.Context, key string) (M, error) {
	m.RLock()
	v, ok := m.data[key]
	m.RUnlock()
	if !ok || v.e != 0 && v.e <= utils.Timestamp() {
		return *new(M), errors.New("key not found")
	}
	val, ok := v.v.(M)
	if !ok {
		if m.CompressAlg != "" {
			return m.decompress(v.v)
		}
		return *new(M), errors.New("key not found")
	}
	return val, nil
}

func (m *Memory[M]) Delete(ctx context.Context, key string) error {
	m.Lock()
	delete(m.data, key)
	m.Unlock()
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
		ts := utils.Timestamp()
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

func (m *Memory[M]) compress(data M) ([]byte, error) {
	input, err := toBytes(data)
	if err != nil {
		return nil, err
	}
	switch m.CompressAlg {
	case "zlib":
		return CompressZlib(input)
	case "flate":
		return CompressFlate(input)
	case "gzip":
		return CompressGzip(input)
	default:
		return input, nil
	}
}

func (m *Memory[M]) decompress(dataRaw interface{}) (M, error) {
	dataByte, ok := dataRaw.([]byte)
	if !ok {
		return *new(M), errors.New("assert type failed")
	}

	var output []byte
	var err error

	switch m.CompressAlg {
	case "zlib":
		output, err = DecompressZlib(dataByte)
		if err != nil {
			return *new(M), err
		}
	case "flate":
		output, err = DecompressFlate(dataByte)
		if err != nil {
			return *new(M), err
		}
	case "gzip":
		output, err = DecompressGzip(dataByte)
		if err != nil {
			return *new(M), err
		}
	default:
		return *new(M), nil
	}

	dataRaw, err = fromBytes[M](output)
	if err != nil {
		return *new(M), err
	}

	data, ok := dataRaw.(M)
	if !ok {
		return *new(M), errors.New("assert type failed")
	}
	return data, nil
}
