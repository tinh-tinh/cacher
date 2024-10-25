package cacher

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_Expire(t *testing.T) {
	cache := New[any](StoreOptions{
		Ttl: 1 * time.Millisecond,
	})

	err := cache.Set("users", "John")
	require.Nil(t, err)

	data, err := cache.Get("users")
	require.NotNil(t, err)
	require.Nil(t, data)
}

func Test_CompressGzip(t *testing.T) {
	type Person struct {
		Name string
		Age  string
	}
	cache := New[Person](StoreOptions{
		CompressAlg: "gzip",
		Ttl:         15 * time.Minute,
	})

	err := cache.Set("users", Person{
		Name: "John",
		Age:  "30",
	})
	require.Nil(t, err)

	data, err := cache.Get("users")
	require.Nil(t, err)

	require.Equal(t, "John", data.Name)
	require.Equal(t, "30", data.Age)
}

func Test_CompressZlib(t *testing.T) {
	type Person struct {
		Name string
		Age  string
	}
	cache := New[Person](StoreOptions{
		CompressAlg: "zlib",
		Ttl:         15 * time.Minute,
	})

	err := cache.Set("users", Person{
		Name: "John",
		Age:  "30",
	})
	require.Nil(t, err)

	data, err := cache.Get("users")
	require.Nil(t, err)

	require.Equal(t, "John", data.Name)
	require.Equal(t, "30", data.Age)
}

func Test_CompressFlate(t *testing.T) {
	type Person struct {
		Name string
		Age  string
	}
	cache := New[Person](StoreOptions{
		CompressAlg: "flate",
		Ttl:         15 * time.Minute,
	})

	err := cache.Set("users", Person{
		Name: "John",
		Age:  "30",
	})
	require.Nil(t, err)

	data, err := cache.Get("users")
	require.Nil(t, err)

	require.Equal(t, "John", data.Name)
	require.Equal(t, "30", data.Age)
}

func Test_Fail(t *testing.T) {
	cache := New[any](StoreOptions{
		Ttl: 15 * time.Minute,
	})

	_, err := cache.Get("users")
	require.NotNil(t, err)

	cache2 := New[string](StoreOptions{
		Ttl:         15 * time.Minute,
		CompressAlg: "abc",
	})
	require.Nil(t, cache2.Store)
}
