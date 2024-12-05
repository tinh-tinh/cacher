package cacher_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tinh-tinh/cacher"
)

func Test_Expire(t *testing.T) {
	cache := cacher.New[any](cacher.StoreOptions{
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
	cache := cacher.New[Person](cacher.StoreOptions{
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
	cache := cacher.New[Person](cacher.StoreOptions{
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
	cache := cacher.New[Person](cacher.StoreOptions{
		CompressAlg: cacher.CompressAlgFlate,
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
	cache := cacher.New[any](cacher.StoreOptions{
		Ttl: 15 * time.Minute,
	})

	_, err := cache.Get("users")
	require.NotNil(t, err)

	cache2 := cacher.New[string](cacher.StoreOptions{
		Ttl:         15 * time.Minute,
		CompressAlg: "abc",
	})
	require.Nil(t, cache2.Store)
}

func Test_MGet(t *testing.T) {
	cache := cacher.New[string](cacher.StoreOptions{
		Ttl: 15 * time.Minute,
	})
	require.NotNil(t, cache)

	err := cache.MSet(cacher.Params[string]{
		Key: "1",
		Val: "John",
	}, cacher.Params[string]{
		Key: "2",
		Val: "Jane",
	})
	require.Nil(t, err)

	data, err := cache.MGet("1", "2")
	require.Nil(t, err)
	require.Equal(t, "John", data[0])
	require.Equal(t, "Jane", data[1])

	cache2 := cacher.New[string](cacher.StoreOptions{
		Ttl: 15 * time.Minute,
	})

	_, err = cache2.MGet("1", "2")
	require.NotNil(t, err)

	cache3 := cacher.New[string](cacher.StoreOptions{
		Ttl:         15 * time.Minute,
		CompressAlg: cacher.CompressAlgZlib,
	})

	err = cache3.MSet(cacher.Params[string]{
		Key: "1",
		Val: "John",
		Options: cacher.StoreOptions{
			Ttl: 5 * time.Minute,
		},
	}, cacher.Params[string]{
		Key: "2",
		Val: "Jane",
	})
	require.Nil(t, err)

	data, err = cache3.MGet("1", "2")
	require.Nil(t, err)
	require.Equal(t, "John", data[0])
	require.Equal(t, "Jane", data[1])
}
