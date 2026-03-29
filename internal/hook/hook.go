package hook

import (
	"fmt"
	"reflect"
)

// IPostInjectHook is an interface that can be implemented by types that want to do post-injection initialization.
type IPostInjectHook interface {
	PostInject() error
}

var rtyPostInjectHook = reflect.TypeFor[IPostInjectHook]()

func CheckAppendPostInjectHookCtor(t reflect.Type, pt reflect.Type, ctor func() (reflect.Value, error)) func() (reflect.Value, error) {
	var transform func(reflect.Value) reflect.Value

	if t.Kind() != reflect.Pointer && pt.Implements(rtyPostInjectHook) {
		// Non-pointer passed-in and implements IPostInjectHook, assume it want to do hook as *T (most common case)
		transform = func(v reflect.Value) reflect.Value { return v.Addr() }
	} else if t.Kind() == reflect.Pointer && t.Implements(rtyPostInjectHook) {
		// *T implements IPostInjectHook, it's normal case and we can call PostInject directly
		transform = func(v reflect.Value) reflect.Value { return v }
	}

	// If neither T nor *T implements IPostInjectHook, return as is
	if transform == nil {
		return ctor
	}
	// Otherwise, append the hook call
	return func() (reflect.Value, error) {
		instance, err := ctor()
		if err != nil {
			return reflect.Value{}, err
		}
		toHook := transform(instance)
		hook, ok := toHook.Interface().(IPostInjectHook)
		if !ok {
			return reflect.Value{}, fmt.Errorf("type %s reflect implements IPostInjectHook but cannot be asserted to it, this should not happen", t.String())
		}
		if err := hook.PostInject(); err != nil {
			return reflect.Value{}, fmt.Errorf("PostInject hook error: %w", err)
		}
		return instance, nil
	}
}
