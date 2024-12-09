package cacher

import (
	"github.com/tinh-tinh/tinhtinh/core"
)

const CACHE_MANAGER core.Provide = "cache_manager"

func Register[M any](options ...Options[M]) core.Module {
	return func(module *core.DynamicModule) *core.DynamicModule {
		cacheModule := module.New(core.NewModuleOptions{})

		cacheManager := New(options[0])
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
