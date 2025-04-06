package cacher

import "slices"

type HookKey string

const (
	BeforeGet    HookKey = "before_get"
	AfterGet     HookKey = "after_get"
	BeforeSet    HookKey = "before_set"
	AfterSet     HookKey = "after_set"
	BeforeDelete HookKey = "before_delete"
	AfterDelete  HookKey = "after_delete"
)

type HookFnc func(key string, data interface{})

type Hook struct {
	Key HookKey
	Fnc HookFnc
}

func HandlerBeforeGet[M any](schema Schema[M], key string) {
	hooks := schema.GetHooks()
	findHook := slices.IndexFunc(hooks, func(h Hook) bool {
		return h.Key == BeforeGet
	})
	if findHook != -1 {
		hooks[findHook].Fnc(key, nil)
	}
}

func HandlerAfterGet[M any](schema Schema[M], key string, data M) {
	hooks := schema.GetHooks()
	findHook := slices.IndexFunc(hooks, func(h Hook) bool {
		return h.Key == AfterGet
	})
	if findHook != -1 {
		hooks[findHook].Fnc(key, data)
	}
}

func HandlerBeforeSet[M any](schema Schema[M], key string, data M) {
	hooks := schema.GetHooks()
	findHook := slices.IndexFunc(hooks, func(h Hook) bool {
		return h.Key == BeforeSet
	})
	if findHook != -1 {
		hooks[findHook].Fnc(key, data)
	}
}

func HandlerAfterSet[M any](schema Schema[M], key string, data M) {
	hooks := schema.GetHooks()
	findHook := slices.IndexFunc(hooks, func(h Hook) bool {
		return h.Key == AfterSet
	})
	if findHook != -1 {
		hooks[findHook].Fnc(key, data)
	}
}

func HandlerBeforeDelete[M any](schema Schema[M], key string) {
	hooks := schema.GetHooks()
	findHook := slices.IndexFunc(hooks, func(h Hook) bool {
		return h.Key == BeforeDelete
	})
	if findHook != -1 {
		hooks[findHook].Fnc(key, nil)
	}
}

func HandlerAfterDelete[M any](schema Schema[M], key string) {
	hooks := schema.GetHooks()
	findHook := slices.IndexFunc(hooks, func(h Hook) bool {
		return h.Key == AfterDelete
	})
	if findHook != -1 {
		hooks[findHook].Fnc(key, nil)
	}
}
