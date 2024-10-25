package cacher

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/allegro/bigcache/v3"
	bigcache_store "github.com/eko/gocache/store/bigcache/v4"
	"github.com/stretchr/testify/require"
	"github.com/tinh-tinh/tinhtinh/common"
	"github.com/tinh-tinh/tinhtinh/core"
)

func Test_Module(t *testing.T) {
	userController := func(module *core.DynamicModule) *core.DynamicController {
		cache := Inject[[]byte](module)
		ctrl := module.NewController("users")

		ctrl.Get("", func(ctx core.Ctx) error {
			data, err := cache.Get(context.Background(), "users")
			if err != nil {
				return common.InternalServerException(ctx.Res(), err.Error())
			}
			return ctx.JSON(core.Map{
				"data": string(data),
			})
		})

		ctrl.Post("", func(ctx core.Ctx) error {
			err := cache.Set(context.Background(), "users", []byte("John"))
			if err != nil {
				return common.InternalServerException(ctx.Res(), err.Error())
			}

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

	bigcacheClient, _ := bigcache.New(context.Background(), bigcache.DefaultConfig(5*time.Minute))
	bigcacheStore := bigcache_store.NewBigcache(bigcacheClient)

	appModule := func() *core.DynamicModule {
		module := core.NewModule(core.NewModuleOptions{
			Imports: []core.Module{
				Register[[]byte](bigcacheStore),
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

func Test_DefaultModule(t *testing.T) {
	userController := func(module *core.DynamicModule) *core.DynamicController {
		cache := Inject[[]byte](module)
		ctrl := module.NewController("users")

		ctrl.Get("", func(ctx core.Ctx) error {
			data, err := cache.Get(context.Background(), "users")
			if err != nil {
				return common.InternalServerException(ctx.Res(), err.Error())
			}
			return ctx.JSON(core.Map{
				"data": string(data),
			})
		})

		ctrl.Post("", func(ctx core.Ctx) error {
			err := cache.Set(context.Background(), "users", []byte("John"))
			if err != nil {
				return common.InternalServerException(ctx.Res(), err.Error())
			}

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
		module := core.NewModule(core.NewModuleOptions{
			Imports: []core.Module{
				Register[[]byte](),
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

func Test_Nil(t *testing.T) {
	module := core.NewModule(core.NewModuleOptions{})

	cache := Inject[[]byte](module)
	require.Nil(t, cache)
}
