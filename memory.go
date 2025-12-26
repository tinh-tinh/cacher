package cacher

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/tinh-tinh/tinhtinh/v2/common/era"
)

var ErrKeyNotFound = errors.New("key not found")

func NewInMemory(opt StoreOptions) Store {
	if opt.MaxItems <= 0 {
		opt.MaxItems = 1000
	}
	memory := &Memory{
		ttl:      opt.Ttl,
		maxItems: opt.MaxItems,
		data:     make(map[string]item, opt.MaxItems),
		keys:     make([]string, 0, opt.MaxItems),
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
	ttl      time.Duration
	data     map[string]item
	maxItems int
	keys     []string
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

	m.Lock()
	defer m.Unlock()

	i := item{e: exp, v: val}
	if _, exists := m.data[key]; exists {
		m.data[key] = i
		return nil
	}

	if m.maxItems > 0 && len(m.data) >= m.maxItems {
		// evict an item
		evictKey := m.keys[0]
		delete(m.data, evictKey)
		m.keys = m.keys[1:]
	}
	m.data[key] = i
	m.keys = append(m.keys, key)
	return nil
}

func (m *Memory) Get(ctx context.Context, key string) ([]byte, error) {
	// Handler
	m.RLock()
	v, ok := m.data[key]
	m.RUnlock()

	if !ok || v.e != 0 && v.e <= era.Timestamp() {
		return nil, ErrKeyNotFound
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
