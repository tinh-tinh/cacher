package cacher

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCacher(t *testing.T) {
	cache := New[string](DefaultStore())
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

	err = cache.Set("users", "John")
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
	cache := New[string](DefaultStore())
	require.NotNil(t, cache)

	err := cache.Set("users", "John")
	require.Nil(t, err)

	data, err := cache.Get("users")
	require.Nil(t, err)
	fmt.Println(data)
}

func Test_Context(t *testing.T) {
	cache := New[string](DefaultStore())
	require.NotNil(t, cache)

	cache.SetCtx(context.TODO())

	ctx := cache.GetCtx()
	require.NotNil(t, ctx)
}
