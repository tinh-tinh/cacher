package pebble_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	pebble_store "github.com/cockroachdb/pebble"
	"github.com/stretchr/testify/require"
	"github.com/tinh-tinh/cacher"
	"github.com/tinh-tinh/cacher/storage/pebble"
	"github.com/tinh-tinh/tinhtinh/common"
	"github.com/tinh-tinh/tinhtinh/core"
)

func Test_Connect(t *testing.T) {
	pp := pebble.New[any]("demo", pebble.Options{
		Ttl: 2 * time.Second,
	})

	require.NotNil(t, pp)
	ctx := context.Background()
	pp.Clear(ctx)

	err := pp.Set(ctx, "Connect", "John")
	require.Nil(t, err)

	data, err := pp.Get(ctx, "Connect")
	require.Nil(t, err)
	require.Equal(t, "John", data)

	time.Sleep(3 * time.Second)

	data, err = pp.Get(ctx, "Connect")
	require.Nil(t, err)
	require.Empty(t, data)
}

func Test_GetSet(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	cache := pebble.New[Person]("demo", pebble.Options{
		Ttl: 15 * time.Minute,
	})

	ctx := context.Background()

	err := cache.Set(ctx, "getset", Person{
		Name: "John",
		Age:  30,
	})
	require.Nil(t, err)

	data, err := cache.Get(ctx, "getset")
	require.Nil(t, err)

	require.Equal(t, "John", data.Name)
	require.Equal(t, 30, data.Age)
}

func Test_MGetSet(t *testing.T) {
	cache := pebble.New[string]("demo", pebble.Options{
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

func Test_CompressGzip(t *testing.T) {
	cache := pebble.New[string]("demo", pebble.Options{
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

func Test_CompressZlib(t *testing.T) {
	cache := pebble.New[string]("demo", pebble.Options{
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

func Test_CompressFlate(t *testing.T) {
	cache := pebble.New[string]("demo", pebble.Options{
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

func Test_Delete(t *testing.T) {
	cache := pebble.New[string]("demo", pebble.Options{
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

func Test_Hook(t *testing.T) {
	cache := pebble.New[string]("demo", pebble.Options{
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
			data, err := cache.Get("modules")
			if err != nil {
				fmt.Println(err)
				return common.InternalServerException(ctx.Res(), err.Error())
			}
			return ctx.JSON(core.Map{
				"data": string(data),
			})
		})

		ctrl.Post("", func(ctx core.Ctx) error {
			cache.Set("modules", []byte("John"))

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
		cache := pebble.New[[]byte]("demo", pebble.Options{
			Ttl: 15 * time.Minute,
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
	require.Equal(t, http.StatusOK, resp.StatusCode)

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

func Benchmark_Pebble(b *testing.B) {
	cache := pebble.New[[]byte]("demo", pebble.Options{
		Ttl: 15 * time.Minute,
	})
	db, ok := cache.GetConnect().(*pebble_store.DB)
	if !ok {
		b.Fatal("failed to get pebble db")
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			db.Set([]byte("key"), []byte("value"), pebble_store.Sync)
		}
	})
}
