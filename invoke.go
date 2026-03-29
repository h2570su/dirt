package dirt

import (
	"fmt"
	"reflect"
)

// Invoke resolves and injects the dependencies of the requested type, and returns the instance. It returns error when the dependencies cannot be satisfied or other errors happen during injection.
//
//	It matches the instance type exactly with the requested type, no interface conversion is performed.
func Invoke[T any](opts ...Option) (T, error) {
	opt := defaultProvideOptions()
	for _, o := range opts {
		opt = o(opt)
	}

	key := typeNameKey{Ty: reflect.TypeFor[T](), Name: opt.Name}
	s := opt.Scope

	ins, ok := s.getInstance(key)
	if !ok {
		_ins, err := s.invokeInstance(key)
		if err != nil {
			return *new(T), err
		}
		ins = _ins
	}

	typed, ok := ins.(T)
	if !ok {
		return *new(T),
			fmt.Errorf("dirt: instance type: `%T` does not match the requested type: `%s`", ins, key.Ty.String())
	}
	return typed, nil
}

// InvokeIndividual is similar to Invoke but it always creates a new instance for T(its dependencies depends of tag)
func InvokeIndividual[T any](opts ...Option) (T, error) {
	opt := defaultProvideOptions()
	for _, o := range opts {
		opt = o(opt)
	}

	key := typeNameKey{Ty: reflect.TypeFor[T](), Name: opt.Name}
	s := opt.Scope

	ins, err := s.instantiate(key)
	if err != nil {
		return *new(T), err
	}

	typed, ok := ins.(T)
	if !ok {
		return *new(T),
			fmt.Errorf("dirt: instance type: `%T` does not match the requested type: `%s`", ins, key.Ty.String())
	}
	return typed, nil
}
