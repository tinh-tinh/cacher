package cacher

import (
	"github.com/tinh-tinh/tinhtinh/v2/core"
)

const CACHE_MANAGER core.Provide = "cache_manager"

func Register(config Config) core.Modules {
	return func(module core.Module) core.Module {
		cacheModule := module.New(core.NewModuleOptions{})

		cacheModule.NewProvider(core.ProviderOptions{
			Name:  CACHE_MANAGER,
			Value: &config,
		})
		cacheModule.Export(CACHE_MANAGER)
		return cacheModule
	}
}

func Inject[M any](ref core.RefProvider) *Schema[M] {
	cache, ok := ref.Ref(CACHE_MANAGER).(*Config)
	if !ok {
		return nil
	}
	return NewSchema[M](*cache)
}
