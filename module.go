package cacher

import (
	"time"

	"github.com/eko/gocache/lib/v4/store"
	gocache_store "github.com/eko/gocache/store/go_cache/v4"
	gocache "github.com/patrickmn/go-cache"
	"github.com/tinh-tinh/tinhtinh/core"
)

const CACHE_MANAGER core.Provide = "cache_manager"

func DefaultStore() store.StoreInterface {
	gocacheClient := gocache.New(5*time.Minute, 10*time.Minute)
	gocacheStore := gocache_store.NewGoCache(gocacheClient)
	return gocacheStore
}

func Register[M any](opt ...store.StoreInterface) core.Module {
	return func(module *core.DynamicModule) *core.DynamicModule {
		cacheModule := module.New(core.NewModuleOptions{})

		if len(opt) == 0 {
			opt = append(opt, DefaultStore())
		}
		cacheManager := New[M](opt[0])
		cacheModule.NewProvider(core.ProviderOptions{
			Name:  CACHE_MANAGER,
			Value: cacheManager,
		})
		cacheModule.Export(CACHE_MANAGER)
		return cacheModule
	}
}

func Inject[M any](module *core.DynamicModule) *Cacher[M] {
	cache, ok := module.Ref(CACHE_MANAGER).(*Cacher[M])
	if !ok {
		return nil
	}
	return cache
}
