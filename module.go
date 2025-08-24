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

type ConfigFactory func(module core.RefProvider) Config

func RegisterFactory(factory ConfigFactory) core.Modules {
	return func(module core.Module) core.Module {
		cacheModule := module.New(core.NewModuleOptions{})

		config := factory(module)
		cacheModule.NewProvider(core.ProviderOptions{
			Name:  CACHE_MANAGER,
			Value: &config,
		})
		cacheModule.Export(CACHE_MANAGER)
		return cacheModule
	}
}

func RegisterMulti(configs ...Config) core.Modules {
	return func(module core.Module) core.Module {
		cacheModule := module.New(core.NewModuleOptions{})

		for _, config := range configs {
			cacheModule.NewProvider(core.ProviderOptions{
				Name:  core.Provide(config.Store.Name()),
				Value: &config,
			})
			cacheModule.Export(core.Provide(config.Store.Name()))
		}
		return cacheModule
	}
}

type MultiConfigFactory func(module core.RefProvider) []Config

func RegisterMultiFactory(factory MultiConfigFactory) core.Modules {
	return func(module core.Module) core.Module {
		cacheModule := module.New(core.NewModuleOptions{})

		configs := factory(module)
		for _, config := range configs {
			cacheModule.NewProvider(core.ProviderOptions{
				Name:  core.Provide(config.Store.Name()),
				Value: &config,
			})
			cacheModule.Export(core.Provide(config.Store.Name()))
		}
		return cacheModule
	}
}
