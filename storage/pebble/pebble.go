package pebble

import (
	"context"
	"fmt"

	pebble_store "github.com/cockroachdb/pebble"
	"github.com/tinh-tinh/cacher/v2"
)

const PEBBLE = "PEBBLE_CACHE_MANAGER"

type Options struct {
	Path    string
	Sync    bool
	Connect *pebble_store.Options
}

func New(opt Options) cacher.Store {
	client, err := pebble_store.Open(opt.Path, opt.Connect)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return &Pebble{
		client: client,
		Sync:   opt.Sync,
	}
}

type Pebble struct {
	Sync   bool
	client *pebble_store.DB
}

func (s *Pebble) Name() string {
	return PEBBLE
}

func (s *Pebble) Get(ctx context.Context, key string) ([]byte, error) {
	keyByte := []byte(key)
	data, close, err := s.client.Get(keyByte)
	if err != nil {
		if err == pebble_store.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	if close.Close() != nil {
		return nil, err
	}

	return data, nil
}

// Warning: currently ttl not work in pebble
func (s *Pebble) Set(ctx context.Context, key string, value []byte, opts ...cacher.StoreOptions) error {
	keyByte := []byte(key)
	err := s.client.Set(keyByte, value, &pebble_store.WriteOptions{
		Sync: s.Sync,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Pebble) Delete(ctx context.Context, key string) error {
	keyByte := []byte(key)
	err := s.client.Delete(keyByte, &pebble_store.WriteOptions{Sync: s.Sync})
	if err != nil {
		return err
	}
	return nil
}

func (s *Pebble) Clear(ctx context.Context) error {
	startKey := []byte("")
	endKey := []byte("\xff")

	err := s.client.DeleteRange(startKey, endKey, &pebble_store.WriteOptions{Sync: s.Sync})
	if err != nil {
		return err
	}
	return nil
}
