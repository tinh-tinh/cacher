package cacher_test

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tinh-tinh/cacher/v2"
	"github.com/tinh-tinh/tinhtinh/v2/common"
	"github.com/tinh-tinh/tinhtinh/v2/core"
)

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
		module := core.NewModule(core.NewModuleOptions{
			Imports: []core.Modules{
				cacher.Register(cacher.Options[[]byte]{
					Ttl: 15 * time.Minute,
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

func Test_Nil(t *testing.T) {
	module := core.NewModule(core.NewModuleOptions{})

	cache := cacher.Inject[[]byte](module)
	require.Nil(t, cache)
}
