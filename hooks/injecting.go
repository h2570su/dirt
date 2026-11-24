package hooks

import (
	"reflect"

	"github.com/h2570su/dirt/core"
)

func HookOnInjecting[T any](fn func(from T, dstT reflect.Type) T) core.HookOnInjecting {
	return core.HookOnInjecting{
		T: reflect.TypeFor[T](),
		Fn: func(from reflect.Value, dstT reflect.Type) reflect.Value {
			ff, ok := from.Interface().(T)
			if !ok {
				return from
			}
			dd := fn(ff, dstT)
			return reflect.ValueOf(dd)
		},
	}
}
