package cacher

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
