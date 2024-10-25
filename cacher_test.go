package cacher

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tinh-tinh/tinhtinh/common/memory"
)

func TestCacher(t *testing.T) {
	cache := New[string](memory.Options{
		Ttl: 15 * time.Minute,
		Max: 100,
	})
	require.NotNil(t, cache)

	cache.Set("users", "John")

	data, err := cache.Get("users")
	require.Nil(t, err)
	require.Equal(t, "John", data)

	cache.Delete("users")

	data, err = cache.Get("users")
	require.NotNil(t, err)
	require.Empty(t, data)

	cache.Set("users", "John")

	data, err = cache.Get("users")
	require.Nil(t, err)
	require.Equal(t, "John", data)

	cache.Clear()

	data, err = cache.Get("users")
	require.NotNil(t, err)
	require.Empty(t, data)
}

func TestDataTypes(t *testing.T) {
	cache := New[string](memory.Options{
		Ttl: 15 * time.Minute,
		Max: 100,
	})
	require.NotNil(t, cache)

	cache.Set("users", "John")

	data, err := cache.Get("users")
	require.Nil(t, err)
	fmt.Println(data)
}

func Test_Context(t *testing.T) {
	cache := New[string](memory.Options{
		Ttl: 15 * time.Minute,
		Max: 100,
	})
	require.NotNil(t, cache)

	cache.SetCtx(context.TODO())

	ctx := cache.GetCtx()
	require.NotNil(t, ctx)
}
