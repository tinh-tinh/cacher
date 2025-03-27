package cacher_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tinh-tinh/cacher/v2"
	"github.com/tinh-tinh/tinhtinh/v2/common"
	"github.com/tinh-tinh/tinhtinh/v2/common/compress"
	"github.com/tinh-tinh/tinhtinh/v2/core"
)

func Test_Expire(t *testing.T) {
	cache := cacher.NewInMemory[any](cacher.StoreOptions{
		Ttl: 10 * time.Millisecond,
	})

	ctx := context.Background()
	err := cache.Set(ctx, "users", "John")
	require.Nil(t, err)

	time.Sleep(12 * time.Millisecond)

	data, err := cache.Get(ctx, "users")
	require.NotNil(t, err)
	require.Nil(t, data)
}

func Test_CompressGzip(t *testing.T) {
	type Person struct {
		Name string
		Age  string
	}
	cache := cacher.NewInMemory[Person](cacher.StoreOptions{
		CompressAlg: compress.Gzip,
		Ttl:         15 * time.Minute,
	})

	ctx := context.Background()
	err := cache.Set(ctx, "users", Person{
		Name: "John",
		Age:  "30",
	})
	require.Nil(t, err)

	data, err := cache.Get(ctx, "users")
	require.Nil(t, err)

	require.Equal(t, "John", data.Name)
	require.Equal(t, "30", data.Age)
}

func Test_CompressZlib(t *testing.T) {
	type Person struct {
		Name string
		Age  string
	}
	cache := cacher.NewInMemory[Person](cacher.StoreOptions{
		CompressAlg: compress.Zlib,
		Ttl:         15 * time.Minute,
	})

	ctx := context.Background()
	err := cache.Set(ctx, "users", Person{
		Name: "John",
		Age:  "30",
	})
	require.Nil(t, err)

	data, err := cache.Get(ctx, "users")
	require.Nil(t, err)

	require.Equal(t, "John", data.Name)
	require.Equal(t, "30", data.Age)
}

func Test_CompressFlate(t *testing.T) {
	type Person struct {
		Name string
		Age  string
	}
	cache := cacher.NewInMemory[Person](cacher.StoreOptions{
		CompressAlg: compress.Flate,
		Ttl:         15 * time.Minute,
	})

	ctx := context.Background()
	err := cache.Set(ctx, "users", Person{
		Name: "John",
		Age:  "30",
	})
	require.Nil(t, err)

	data, err := cache.Get(ctx, "users")
	require.Nil(t, err)

	require.Equal(t, "John", data.Name)
	require.Equal(t, "30", data.Age)
}

func Test_Fail(t *testing.T) {
	cache := cacher.NewInMemory[any](cacher.StoreOptions{
		Ttl: 15 * time.Minute,
	})

	ctx := context.Background()
	_, err := cache.Get(ctx, "users")
	require.NotNil(t, err)

	cache2 := cacher.NewInMemory[string](cacher.StoreOptions{
		Ttl:         15 * time.Minute,
		CompressAlg: "abc",
	})
	require.Nil(t, cache2)
}

func Test_MGet(t *testing.T) {
	cache := cacher.NewInMemory[string](cacher.StoreOptions{
		Ttl: 15 * time.Minute,
	})
	require.NotNil(t, cache)

	ctx := context.Background()
	err := cache.MSet(ctx, cacher.Params[string]{
		Key: "1",
		Val: "John",
	}, cacher.Params[string]{
		Key: "2",
		Val: "Jane",
	})
	require.Nil(t, err)

	data, err := cache.MGet(ctx, "1", "2")
	require.Nil(t, err)
	require.Equal(t, "John", data[0])
	require.Equal(t, "Jane", data[1])

	cache2 := cacher.NewInMemory[string](cacher.StoreOptions{
		Ttl: 15 * time.Minute,
	})

	_, err = cache2.MGet(ctx, "1", "2")
	require.NotNil(t, err)

	cache3 := cacher.NewInMemory[string](cacher.StoreOptions{
		Ttl:         15 * time.Minute,
		CompressAlg: compress.Zlib,
	})

	err = cache3.MSet(ctx, cacher.Params[string]{
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

	data, err = cache3.MGet(ctx, "1", "2")
	require.Nil(t, err)
	require.Equal(t, "John", data[0])
	require.Equal(t, "Jane", data[1])
}

func Test_Memory_Module(t *testing.T) {
	userController := func(module core.Module) core.Controller {
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

	userModule := func(module core.Module) core.Module {
		return module.New(core.NewModuleOptions{
			Controllers: []core.Controllers{
				userController,
			},
		})
	}

	appModule := func() core.Module {
		cache := cacher.NewInMemory[[]byte](cacher.StoreOptions{
			Ttl: 15 * time.Minute,
		})

		module := core.NewModule(core.NewModuleOptions{
			Imports: []core.Modules{
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
