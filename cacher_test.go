package cacher

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCacher(t *testing.T) {
	cache := New[string](StoreOptions{
		Ttl: 15 * time.Minute,
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

	err = cache.Set("users", "John", StoreOptions{Ttl: 5 * time.Minute})
	require.Nil(t, err)

	data, err = cache.Get("users")
	require.Nil(t, err)
	require.Equal(t, "John", data)

	err = cache.Clear()
	require.Nil(t, err)

	data, err = cache.Get("users")
	require.NotNil(t, err)
	require.Empty(t, data)
}

func TestDataTypes(t *testing.T) {
	cache := New[string](StoreOptions{
		Ttl: 15 * time.Minute,
	})
	require.NotNil(t, cache)

	cache.Set("users", "John")

	data, err := cache.Get("users")
	require.Nil(t, err)
	require.Equal(t, "John", data)
}

func Test_Context(t *testing.T) {
	cache := New[string](StoreOptions{
		Ttl: 15 * time.Minute,
	})
	require.NotNil(t, cache)

	cache.SetCtx(context.TODO())

	ctx := cache.GetCtx()
	require.NotNil(t, ctx)
}
