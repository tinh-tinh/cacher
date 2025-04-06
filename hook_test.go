package cacher_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tinh-tinh/cacher/v2"
)

func Test_Hook(t *testing.T) {
	cache := cacher.NewSchema[string](cacher.Config{
		Store: cacher.NewInMemory(cacher.StoreOptions{
			Ttl: 15 * time.Minute,
		}),
		Hooks: []cacher.Hook{
			{Key: cacher.BeforeSet, Fnc: func(key string, val any) {
				fmt.Println("BeforeSet", key, val)
			}},
			{Key: cacher.AfterSet, Fnc: func(key string, val any) {
				fmt.Println("AfterSet", key, val)
			}},
			{Key: cacher.BeforeGet, Fnc: func(key string, val any) {
				fmt.Println("BeforeGet", key)
			}},
			{Key: cacher.AfterGet, Fnc: func(key string, val any) {
				fmt.Println("AfterGet", key, val)
			}},
			{Key: cacher.BeforeDelete, Fnc: func(key string, val any) {
				fmt.Println("BeforeDelete", key)
			}},
			{Key: cacher.AfterDelete, Fnc: func(key string, val any) {
				fmt.Println("AfterDelete", key)
			}},
		},
	})

	err := cache.Set("1", "John")
	require.Nil(t, err)

	err = cache.Set("2", "Jane")
	require.Nil(t, err)

	data, err := cache.Get("1")
	require.Nil(t, err)
	require.Equal(t, "John", data)

	data, err = cache.Get("2")
	require.Nil(t, err)
	require.Equal(t, "Jane", data)

	err = cache.Delete("1")
	require.Nil(t, err)

	data, err = cache.Get("1")
	require.NotNil(t, err)
	require.Empty(t, data)
}
