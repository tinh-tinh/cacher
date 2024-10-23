package cacher

import (
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	"github.com/tinh-tinh/tinhtinh/core"
)

const CACHE_MANAGER core.Provide = "cache_manager"

func Register[M any](opt store.StoreInterface) core.Module {
	return func(module *core.DynamicModule) *core.DynamicModule {
		cacheModule := module.New(core.NewModuleOptions{})

		cacheManager := cache.New[M](opt)
		cacheModule.NewProvider(core.ProviderOptions{
			Name:  CACHE_MANAGER,
			Value: cacheManager,
		})
		cacheModule.Export(CACHE_MANAGER)
		return cacheModule
	}
}

func Inject[M any](module *core.DynamicModule) *cache.Cache[M] {
	cache, ok := module.Ref(CACHE_MANAGER).(*cache.Cache[M])
	if !ok {
		return nil
	}
	return cache
}
