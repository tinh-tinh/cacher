package cacher_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tinh-tinh/cacher/v2"
)

func Test_Expire(t *testing.T) {
	cache := cacher.NewInMemory(cacher.StoreOptions{
		Ttl: 10 * time.Millisecond,
	})

	ctx := context.Background()
	err := cache.Set(ctx, "users", []byte("John"))
	require.Nil(t, err)

	time.Sleep(12 * time.Millisecond)

	data, err := cache.Get(ctx, "users")
	require.NotNil(t, err)
	require.Nil(t, data)
}

func Test_AutoExpire(t *testing.T) {
	cache := cacher.NewInMemory(cacher.StoreOptions{
		Ttl: 1 * time.Second,
	})

	ctx := context.Background()
	err := cache.Set(ctx, "users", []byte("John"))
	require.Nil(t, err)

	time.Sleep(2 * time.Second)

	data, err := cache.Get(ctx, "users")
	require.NotNil(t, err)
	require.Nil(t, data)
}

func Test_Clear(t *testing.T) {
	cache := cacher.NewInMemory(cacher.StoreOptions{
		Ttl: 1 * time.Second,
	})

	ctx := context.Background()
	err := cache.Set(ctx, "users", []byte("John"))
	require.Nil(t, err)

	err = cache.Clear(ctx)
	require.Nil(t, err)

	data, err := cache.Get(ctx, "users")
	require.NotNil(t, err)
	require.Nil(t, data)
}
