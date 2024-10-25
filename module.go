package cacher

import (
	"context"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	bigcache_store "github.com/eko/gocache/store/bigcache/v4"
	"github.com/tinh-tinh/tinhtinh/core"
)

const CACHE_MANAGER core.Provide = "cache_manager"

func DefaultStore() store.StoreInterface {
	bigcacheClient, _ := bigcache.New(context.Background(), bigcache.DefaultConfig(15*time.Minute))
	bigcacheStore := bigcache_store.NewBigcache(bigcacheClient)
	return bigcacheStore
}

func Register[M any](opt ...store.StoreInterface) core.Module {
	return func(module *core.DynamicModule) *core.DynamicModule {
		cacheModule := module.New(core.NewModuleOptions{})

		if len(opt) == 0 {
			opt = append(opt, DefaultStore())
		}
		cacheManager := cache.New[M](opt[0])
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
