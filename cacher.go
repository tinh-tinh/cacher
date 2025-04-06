package cacher

import (
	"context"
	"encoding/json"

	"github.com/tinh-tinh/tinhtinh/v2/common/compress"
)

type Params[M any] struct {
	Key     string
	Value   M
	Options StoreOptions
}

type Schema[M any] struct {
	Config
	ctx context.Context
}

type Config struct {
	Store       Store
	CompressAlg compress.Alg
	Hooks       []Hook
	Namespace   string
}

func NewSchema[M any](config Config) *Schema[M] {
	return &Schema[M]{
		Config: config,
		ctx:    context.Background(),
	}
}

func (s *Schema[M]) SetCtx(ctx context.Context) {
	s.ctx = ctx
}

func (s *Schema[M]) GetCtx() context.Context {
	return s.ctx
}

func (s *Schema[M]) GetHooks() []Hook {
	return s.Hooks
}

func (s *Schema[M]) Get(key string) (M, error) {
	HandlerBeforeGet(*s, key)

	val, err := s.Store.Get(s.ctx, s.generateKey(key))
	if err != nil {
		return *new(M), err
	}

	var schema M
	err = json.Unmarshal(val, &schema)
	if err != nil {
		if s.CompressAlg != "" {
			return compress.DecodeMarshall[M](val, s.CompressAlg)
		}
		return *new(M), err
	}

	HandlerAfterGet(*s, key, schema)
	return schema, nil
}

func (s *Schema[M]) MGet(keys ...string) ([]M, error) {
	var schemas []M
	for _, key := range keys {
		val, err := s.Get(key)
		if err != nil {
			return nil, err
		}
		schemas = append(schemas, val)
	}

	return schemas, nil
}

func (s *Schema[M]) Set(key string, data M, opts ...StoreOptions) (err error) {
	HandlerBeforeSet(*s, key, data)

	var value []byte
	if s.CompressAlg != "" {
		value, err = compress.Encode(data, s.CompressAlg)
		if err != nil {
			return err
		}
	} else {
		value, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}

	err = s.Store.Set(s.ctx, s.generateKey(key), value, opts...)
	if err != nil {
		return err
	}

	HandlerAfterSet(*s, key, data)
	return nil
}

func (s *Schema[M]) MSet(params ...Params[M]) error {
	for _, param := range params {
		if err := s.Set(param.Key, param.Value, param.Options); err != nil {
			return err
		}
	}

	return nil
}

func (s *Schema[M]) Delete(key string) error {
	HandlerBeforeDelete(*s, key)

	err := s.Store.Delete(s.ctx, s.generateKey(key))
	if err != nil {
		return err
	}

	HandlerAfterDelete(*s, key)
	return nil
}

func (s *Schema[M]) generateKey(key string) string {
	if s.Namespace != "" {
		return s.Namespace + ":" + key
	}
	return key
}
