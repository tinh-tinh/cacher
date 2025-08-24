package cacher

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/tinh-tinh/tinhtinh/v2/common/era"
)

func NewInMemory(opt StoreOptions) Store {
	memory := &Memory{
		ttl:  opt.Ttl,
		data: make(map[string]item),
	}
	era.StartTimeStampUpdater()
	go memory.gc(1 * time.Second)
	return memory
}

type item struct {
	v interface{}
	e uint32
}

type Memory struct {
	sync.RWMutex
	ttl  time.Duration
	data map[string]item
}

func (m *Memory) Name() string {
	return MEMORY
}

func (m *Memory) Set(ctx context.Context, key string, val []byte, opts ...StoreOptions) error {
	// Handler
	var exp uint32
	if len(opts) > 0 && opts[0].Ttl != 0 {
		exp = uint32(opts[0].Ttl.Seconds()) + era.Timestamp()
	} else {
		exp = uint32(m.ttl.Seconds()) + era.Timestamp()
	}
	i := item{e: exp, v: val}
	m.Lock()
	m.data[key] = i
	m.Unlock()

	return nil
}

func (m *Memory) Get(ctx context.Context, key string) ([]byte, error) {
	// Handler
	m.RLock()
	v, ok := m.data[key]
	m.RUnlock()

	if !ok || v.e != 0 && v.e <= era.Timestamp() {
		return nil, errors.New("key not found")
	}
	val, ok := v.v.([]byte)
	if !ok {
		return nil, errors.New("value save is not supported")
	}

	return val, nil
}

func (m *Memory) Delete(ctx context.Context, key string) error {
	// Handler
	m.Lock()
	delete(m.data, key)
	m.Unlock()

	return nil
}

func (m *Memory) Clear(ctx context.Context) error {
	md := make(map[string]item)
	m.Lock()
	m.data = md
	m.Unlock()
	return nil
}

func (m *Memory) gc(sleep time.Duration) {
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
