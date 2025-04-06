package cacher_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tinh-tinh/cacher/v2"
	"github.com/tinh-tinh/tinhtinh/v2/common/compress"
)

func TestCacher(t *testing.T) {
	cache := cacher.NewSchema[string](cacher.Config{
		Store: cacher.NewInMemory(cacher.StoreOptions{
			Ttl: 15 * time.Minute,
		}),
	})
	require.NotNil(t, cache)

	err := cache.Set("users", "John")
	require.Nil(t, err)

	data, err := cache.Get("users")
	require.Nil(t, err)
	require.Equal(t, "John", data)

	err = cache.Delete("users")
	require.Nil(t, err)

	data, err = cache.Get("users")
	require.NotNil(t, err)
	require.Empty(t, data)

	err = cache.Set("users", "John", cacher.StoreOptions{Ttl: 5 * time.Minute})
	require.Nil(t, err)

	data, err = cache.Get("users")
	require.Nil(t, err)
	require.Equal(t, "John", data)

	err = cache.MSet(cacher.Params[string]{
		Key:   "snow",
		Value: "white",
	}, cacher.Params[string]{
		Key:   "momam",
		Value: "black",
	})
	require.Nil(t, err)

	list, err := cache.MGet("snow", "momam")
	require.Nil(t, err)
	require.Len(t, list, 2)
}

func Test_Context(t *testing.T) {
	cache := cacher.NewSchema[string](cacher.Config{
		Store: cacher.NewInMemory(cacher.StoreOptions{
			Ttl: 15 * time.Minute,
		}),
	})
	require.NotNil(t, cache)

	cache.SetCtx(context.TODO())

	ctx := cache.GetCtx()
	require.NotNil(t, ctx)
}

func Test_Namespace(t *testing.T) {
	store := cacher.NewInMemory(cacher.StoreOptions{
		Ttl: 15 * time.Minute,
	})
	cache1 := cacher.NewSchema[string](cacher.Config{
		Store:     store,
		Namespace: "cache1",
	})
	cache1.Set("1", "abc")
	data, err := cache1.Get("1")
	require.Nil(t, err)
	require.Equal(t, "abc", data)

	cache2 := cacher.NewSchema[string](cacher.Config{
		Store:     store,
		Namespace: "cache2",
	})
	cache2.Set("1", "mno")
	data2, err := cache1.Get("1")
	require.Nil(t, err)
	require.Equal(t, "abc", data2)
}

func Test_InvalidSchema(t *testing.T) {
	store := cacher.NewInMemory(cacher.StoreOptions{
		Ttl: 15 * time.Minute,
	})
	cacheStr := cacher.NewSchema[string](cacher.Config{
		Store: store,
	})
	err := cacheStr.Set("1", "abc")
	require.Nil(t, err)

	cacheNum := cacher.NewSchema[int](cacher.Config{
		Store: store,
	})
	_, err = cacheNum.Get("1")
	require.NotNil(t, err)

	_, err = cacheNum.MGet("1")
	require.NotNil(t, err)

	cacheAny := cacher.NewSchema[any](cacher.Config{
		Store: store,
	})
	c := make(chan int)
	err = cacheAny.Set("1", c)
	require.NotNil(t, err)

	cacheCompressAny := cacher.NewSchema[any](cacher.Config{
		Store:       store,
		CompressAlg: compress.Gzip,
	})
	err = cacheCompressAny.Set("1", nil)
	require.NotNil(t, err)
}
