package cacher

import "github.com/tinh-tinh/tinhtinh/v2/core"

func Inject(ref core.RefProvider) *Config {
	cache, ok := ref.Ref(CACHE_MANAGER).(*Config)
	if !ok {
		return nil
	}
	return cache
}

func InjectSchema[M any](ref core.RefProvider) *Schema[M] {
	cache := Inject(ref)
	if cache == nil {
		return nil
	}
	return NewSchema[M](*cache)
}

func InjectByStore(ref core.RefProvider, store string) *Config {
	cache, ok := ref.Ref(core.Provide(store)).(*Config)
	if !ok {
		return nil
	}
	return cache
}

func InjectSchemaByStore[M any](ref core.RefProvider, store string) *Schema[M] {
	cache := InjectByStore(ref, store)
	if cache == nil {
		return nil
	}
	return NewSchema[M](*cache)
}
