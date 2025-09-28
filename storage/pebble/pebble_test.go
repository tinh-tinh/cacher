package pebble_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	pebble_store "github.com/cockroachdb/pebble"
	"github.com/stretchr/testify/require"
	"github.com/tinh-tinh/cacher/storage/pebble"
	"github.com/tinh-tinh/cacher/v2"
	"github.com/tinh-tinh/tinhtinh/v2/common"
	"github.com/tinh-tinh/tinhtinh/v2/core"
)

func Test_GetSet(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	cache := pebble.New(pebble.Options{
		Path:    "demo",
		Sync:    false,
		Connect: &pebble_store.Options{},
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
	cache := pebble.New(pebble.Options{
		Path:    "demo",
		Sync:    false,
		Connect: &pebble_store.Options{},
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
	cache := pebble.New(pebble.Options{
		Path:    "demo",
		Sync:    false,
		Connect: &pebble_store.Options{},
	})

	ctx := context.Background()

	err := cache.Set(ctx, "delete", []byte("John"))
	require.Nil(t, err)

	data, err := cache.Get(ctx, "delete")
	require.Nil(t, err)
	require.Equal(t, []byte("John"), data)

	err = cache.Delete(ctx, "delete")
	require.Nil(t, err)

	data, err = cache.Get(ctx, "delete")
	require.Nil(t, err)
	require.Empty(t, data)
}

func Test_Module(t *testing.T) {
	userController := func(module core.Module) core.Controller {
		cache := cacher.InjectSchema[[]byte](module)
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
		cache := pebble.New(pebble.Options{
			Path:    "demo",
			Sync:    false,
			Connect: &pebble_store.Options{},
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

func Test_GetClient(t *testing.T) {
	cache := pebble.New(pebble.Options{
		Path:    "demo",
		Sync:    false,
		Connect: &pebble_store.Options{},
	})
	cacheRedis, ok := cache.(*pebble.Pebble)
	require.True(t, ok)
	require.NotNil(t, cacheRedis.GetClient())
}
