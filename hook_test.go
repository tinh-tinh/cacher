package cacher

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_Hook(t *testing.T) {
	cache := New[string](StoreOptions{
		Ttl: 15 * time.Minute,
		Hooks: []Hook{
			{Key: BeforeSet, Fnc: func(key string, val any) {
				fmt.Println("BeforeSet", key, val)
			}},
			{Key: AfterSet, Fnc: func(key string, val any) {
				fmt.Println("AfterSet", key, val)
			}},
			{Key: BeforeGet, Fnc: func(key string, val any) {
				fmt.Println("BeforeGet", key, val)
			}},
			{Key: AfterGet, Fnc: func(key string, val any) {
				fmt.Println("AfterGet", key, val)
			}},
			{Key: BeforeDelete, Fnc: func(key string, val any) {
				fmt.Println("BeforeDelete", key, val)
			}},
			{Key: AfterDelete, Fnc: func(key string, val any) {
				fmt.Println("AfterDelete", key, val)
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
