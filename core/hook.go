package core

import "reflect"

type (
	HookFnOnInjecting = func(from reflect.Value, dstT reflect.Type) (to reflect.Value)
	HookFnCtorBuilt   = func(t reflect.Type, from Ctor) (to Ctor)
)

type IHooks interface {
	// HookOnInject registers hook for post-injecting
	HookOnInjecting(HookFnOnInjecting)
	// OnInjectingFunc returns the current on-injecting hook
	OnInjecting(from reflect.Value, dstT reflect.Type) (to reflect.Value)

	// OnBuildCtorDoneHook registers a hook for constructor was built
	HookOnCtorBuilt(HookFnCtorBuilt)
	// OnBuildCtorDoneFunc returns the current hook fn for constructor was built
	OnCtorBuilt(t reflect.Type, from Ctor) (to Ctor)
}
