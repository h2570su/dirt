package hook

import (
	"fmt"
	"reflect"

	"github.com/h2570su/dirt/core"
)

type (
	// IPostInjectHookE is an interface that can be implemented by types that want to do post-injection initialization.
	IPostInjectHookE interface {
		PostInject() error
	}
	// IPostInjectHook is an interface that can be implemented by types that want to do post-injection initialization.
	IPostInjectHook interface {
		PostInject()
	}
)

var rtyPostInjectHooks = []reflect.Type{
	reflect.TypeFor[IPostInjectHookE](),
	reflect.TypeFor[IPostInjectHook](),
}

func CheckAppendPostInjectHookCtor(t reflect.Type, ctor core.Ctor) core.Ctor {
	// If neither T nor *T implements IPostInjectHook, return as is
	var transform func(reflect.Value) reflect.Value
	for _, hookType := range rtyPostInjectHooks {
		if transform = checkImplementationTransform(t, hookType); transform != nil {
			break
		}
	}
	if transform == nil {
		return ctor
	}
	// Otherwise, append the hook call
	return func(s core.IScope) (reflect.Value, error) {
		instance, err := ctor(s)
		if err != nil {
			return reflect.Value{}, err
		}
		transformed := transform(instance).Interface()
		switch hook := transformed.(type) {
		case IPostInjectHook:
			hook.PostInject()
		case IPostInjectHookE:
			if err := hook.PostInject(); err != nil {
				return reflect.Value{}, fmt.Errorf("PostInject hook error: %w", err)
			}
		default:
			return reflect.Value{}, fmt.Errorf("type %s implements IPostInjectHook(s) but cannot be asserted to it, this should not happen", t.String())
		}
		return instance, nil
	}
}

func checkImplementationTransform(t reflect.Type, interfaceT reflect.Type) func(reflect.Value) reflect.Value {
	pt := reflect.PointerTo(t)

	if t.Kind() != reflect.Pointer && pt.Implements(interfaceT) {
		// Non-pointer passed-in and implements interfaceT, assume it want to do hook as *T (most common case)
		return func(v reflect.Value) reflect.Value { return v.Addr() }
	} else if t.Kind() == reflect.Pointer && t.Implements(interfaceT) {
		// *T implements interfaceT, it's normal case and we can call interfaceT directly
		return func(v reflect.Value) reflect.Value { return v }
	}
	return nil
}
