package dirt

import (
	"fmt"
	"reflect"
)

func Invoke[T any](opts ...Option) (T, error) {
	opt := defaultProvideOptions()
	for _, o := range opts {
		opt = o(opt)
	}

	key := TypeNameKey{Ty: reflect.TypeFor[T](), Name: opt.Name}
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

func InvokeIndividual[T any](opts ...Option) (T, error) {
	opt := defaultProvideOptions()
	for _, o := range opts {
		opt = o(opt)
	}

	key := TypeNameKey{Ty: reflect.TypeFor[T](), Name: opt.Name}
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
