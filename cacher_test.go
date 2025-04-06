package cacher_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tinh-tinh/cacher/v2"
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
		})})
	require.NotNil(t, cache)

	cache.SetCtx(context.TODO())

	ctx := cache.GetCtx()
	require.NotNil(t, ctx)
}
