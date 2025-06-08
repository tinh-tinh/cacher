package memcache_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tinh-tinh/cacher/storage/memcache"
	"github.com/tinh-tinh/cacher/v2"
	"github.com/tinh-tinh/tinhtinh/v2/common"
	"github.com/tinh-tinh/tinhtinh/v2/core"
)

func Test_Expire(t *testing.T) {
	cache := memcache.New[any](memcache.Options{
		Addr: []string{"localhost:11211"},
		Ttl:  1 * time.Second,
	})

	ctx := context.Background()
	err := cache.Set(ctx, "expire", []byte("John"))
	require.Nil(t, err)

	time.Sleep(2 * time.Second)
	data, err := cache.Get(ctx, "expire")
	require.Nil(t, err)
	require.Empty(t, data)
}

func Test_GetSet(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	cache := memcache.New[Person](memcache.Options{
		Addr: []string{"localhost:11211"},
		Ttl:  15 * time.Minute,
	})

	ctx := context.Background()

	person, err := json.Marshal(Person{
		Name: "John",
		Age:  30,
	})
	require.Nil(t, err)
	err = cache.Set(ctx, "users", person)
	require.Nil(t, err)

	raw, err := cache.Get(ctx, "users")
	require.Nil(t, err)

	var data Person
	err = json.Unmarshal(raw, &data)
	require.Nil(t, err)

	require.Equal(t, "John", data.Name)
	require.Equal(t, 30, data.Age)
}

func Test_Clear(t *testing.T) {
	cache := memcache.New[string](memcache.Options{
		Addr: []string{"localhost:11211"},
		Ttl:  15 * time.Minute,
	})

	ctx := context.Background()

	err := cache.Set(ctx, "1", []byte("John"))
	require.Nil(t, err)

	err = cache.Clear(ctx)
	require.Nil(t, err)

	data, err := cache.Get(ctx, "1")
	require.Nil(t, err)
	require.Empty(t, data)
}

func Test_Delete(t *testing.T) {
	cache := memcache.New[string](memcache.Options{
		Addr: []string{"localhost:11211"},
		Ttl:  15 * time.Minute,
	})

	ctx := context.Background()

	err := cache.Set(ctx, "delete", []byte("John"))
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

func Test_Module(t *testing.T) {
	userController := func(module core.Module) core.Controller {
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

	userModule := func(module core.Module) core.Module {
		return module.New(core.NewModuleOptions{
			Controllers: []core.Controllers{
				userController,
			},
		})
	}

	appModule := func() core.Module {
		cache := memcache.New[[]byte](memcache.Options{
			Addr: []string{"localhost:11211"},
			Ttl:  1 * time.Millisecond,
		})

		module := core.NewModule(core.NewModuleOptions{
			Imports: []core.Modules{
				cacher.Register(cacher.Config{
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
