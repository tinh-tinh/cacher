package redis_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	redis_store "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/tinh-tinh/cacher"
	"github.com/tinh-tinh/cacher/storage/redis"
	"github.com/tinh-tinh/tinhtinh/common"
	"github.com/tinh-tinh/tinhtinh/core"
)

func Test_Expire(t *testing.T) {
	cache := redis.New[any](redis.Options{
		Connect: &redis_store.Options{
			Addr:     "localhost:6379",
			DB:       0,
			Password: "",
		},
		Ttl: 1 * time.Millisecond,
	})

	ctx := context.Background()
	err := cache.Set(ctx, "expire", "John")
	require.Nil(t, err)

	time.Sleep(2 * time.Millisecond)
	data, err := cache.Get(ctx, "expire")
	require.Nil(t, err)
	require.Nil(t, data)
}

func Test_GetSet(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	cache := redis.New[Person](redis.Options{
		Connect: &redis_store.Options{
			Addr:     "localhost:6379",
			DB:       0,
			Password: "",
		},
		Ttl: 15 * time.Minute,
	})

	ctx := context.Background()

	err := cache.Set(ctx, "users", Person{
		Name: "John",
		Age:  30,
	})
	require.Nil(t, err)

	data, err := cache.Get(ctx, "users")
	require.Nil(t, err)

	require.Equal(t, "John", data.Name)
	require.Equal(t, 30, data.Age)
}

func Test_MGetSet(t *testing.T) {
	cache := redis.New[string](redis.Options{
		Connect: &redis_store.Options{
			Addr:     "localhost:6379",
			DB:       0,
			Password: "",
		},
		Ttl: 15 * time.Minute,
	})

	ctx := context.Background()

	err := cache.MSet(ctx, cacher.Params[string]{
		Key: "11",
		Val: "John",
	}, cacher.Params[string]{
		Key: "12",
		Val: "Jane",
	})
	require.Nil(t, err)

	data, err := cache.MGet(ctx, "11", "12")
	require.Nil(t, err)
	require.Equal(t, "John", data[0])
	require.Equal(t, "Jane", data[1])
}

func Test_Clear(t *testing.T) {
	cache := redis.New[string](redis.Options{
		Connect: &redis_store.Options{
			Addr:     "localhost:6379",
			DB:       0,
			Password: "",
		},
		Ttl: 15 * time.Minute,
	})

	ctx := context.Background()

	err := cache.MSet(ctx, cacher.Params[string]{
		Key: "1",
		Val: "John",
	}, cacher.Params[string]{
		Key: "2",
		Val: "Jane",
	})
	require.Nil(t, err)

	err = cache.Clear(ctx)
	require.Nil(t, err)

	data, err := cache.MGet(ctx, "1", "2")
	require.Nil(t, err)
	require.Empty(t, data)
}

func Test_CompressGzip(t *testing.T) {
	cache := redis.New[string](redis.Options{
		Connect: &redis_store.Options{
			Addr:     "localhost:6379",
			DB:       0,
			Password: "",
		},
		Ttl:         15 * time.Minute,
		CompressAlg: cacher.CompressAlgGzip,
	})

	ctx := context.Background()

	err := cache.Set(ctx, "gzip", "John")
	require.Nil(t, err)

	data, err := cache.Get(ctx, "gzip")
	require.Nil(t, err)
	require.Equal(t, "John", data)
}

func Test_CompressFlate(t *testing.T) {
	cache := redis.New[string](redis.Options{
		Connect: &redis_store.Options{
			Addr:     "localhost:6379",
			DB:       0,
			Password: "",
		},
		Ttl:         15 * time.Minute,
		CompressAlg: cacher.CompressAlgFlate,
	})

	ctx := context.Background()

	err := cache.Set(ctx, "flate", "John")
	require.Nil(t, err)

	data, err := cache.Get(ctx, "flate")
	require.Nil(t, err)
	require.Equal(t, "John", data)
}

func Test_CompressZlib(t *testing.T) {
	cache := redis.New[string](redis.Options{
		Connect: &redis_store.Options{
			Addr:     "localhost:6379",
			DB:       0,
			Password: "",
		},
		Ttl:         15 * time.Minute,
		CompressAlg: cacher.CompressAlgZlib,
	})

	ctx := context.Background()

	err := cache.Set(ctx, "zlib", "John")
	require.Nil(t, err)

	data, err := cache.Get(ctx, "zlib")
	require.Nil(t, err)
	require.Equal(t, "John", data)
}

func Test_Delete(t *testing.T) {
	cache := redis.New[string](redis.Options{
		Connect: &redis_store.Options{
			Addr:     "localhost:6379",
			DB:       0,
			Password: "",
		},
		Ttl: 15 * time.Minute,
	})

	ctx := context.Background()

	err := cache.Set(ctx, "delete", "John")
	require.Nil(t, err)

	data, err := cache.Get(ctx, "delete")
	require.Nil(t, err)
	require.Equal(t, "John", data)

	err = cache.Delete(ctx, "delete")
	require.Nil(t, err)

	data, err = cache.Get(ctx, "delete")
	require.Nil(t, err)
	require.Empty(t, data)
}

func Test_Fail(t *testing.T) {
	cache := redis.New[any](redis.Options{
		Connect: &redis_store.Options{
			Addr:     "localhost:6379",
			DB:       0,
			Password: "",
		},
		Ttl:         15 * time.Minute,
		CompressAlg: "abc",
	})

	require.Nil(t, cache)
}

func Test_Hook(t *testing.T) {
	cache := redis.New[string](redis.Options{
		Connect: &redis_store.Options{
			Addr:     "localhost:6379",
			DB:       0,
			Password: "",
		},
		Ttl: 15 * time.Minute,
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

	ctx := context.Background()

	err := cache.Set(ctx, "before", "John")
	require.Nil(t, err)

	err = cache.Set(ctx, "after", "Jane")
	require.Nil(t, err)

	data, err := cache.Get(ctx, "before")
	require.Nil(t, err)
	require.Equal(t, "John", data)

	err = cache.Delete(ctx, "before")
	require.Nil(t, err)

	data, err = cache.Get(ctx, "before")
	require.Nil(t, err)
	require.Empty(t, data)
}

func Test_Module(t *testing.T) {
	userController := func(module *core.DynamicModule) *core.DynamicController {
		cache := cacher.Inject[[]byte](module)
		ctrl := module.NewController("users")

		ctrl.Get("", func(ctx core.Ctx) error {
			data, err := cache.Get("users")
			if err != nil {
				return common.InternalServerException(ctx.Res(), err.Error())
			}
			return ctx.JSON(core.Map{
				"data": string(data),
			})
		})

		ctrl.Post("", func(ctx core.Ctx) error {
			cache.Set("users", []byte("John"))

			return ctx.JSON(core.Map{
				"data": "ok",
			})
		})

		return ctrl
	}

	userModule := func(module *core.DynamicModule) *core.DynamicModule {
		return module.New(core.NewModuleOptions{
			Controllers: []core.Controller{
				userController,
			},
		})
	}

	appModule := func() *core.DynamicModule {
		cache := redis.New[[]byte](redis.Options{
			Connect: &redis_store.Options{
				Addr:     "localhost:6379",
				DB:       0,
				Password: "",
			},
			Ttl: 1 * time.Millisecond,
		})

		module := core.NewModule(core.NewModuleOptions{
			Imports: []core.Module{
				cacher.Register(cacher.Options[[]byte]{
					Ttl:   15 * time.Minute,
					Store: cache,
				}),
				userModule,
			},
		})

		return module
	}

	app := core.CreateFactory(appModule)
	app.SetGlobalPrefix("api")

	testServer := httptest.NewServer(app.PrepareBeforeListen())
	defer testServer.Close()

	testClient := testServer.Client()
	_, err := testClient.Post(testServer.URL+"/api/users", "application/json", nil)
	require.Nil(t, err)

	resp, err := testClient.Get(testServer.URL + "/api/users")
	require.Nil(t, err)
	require.Equal(t, 200, resp.StatusCode)

	data, err := io.ReadAll(resp.Body)
	require.Nil(t, err)

	type Response struct {
		Data string `json:"data"`
	}

	var response Response
	err = json.Unmarshal(data, &response)
	require.Nil(t, err)
	require.Equal(t, "John", response.Data)
}
