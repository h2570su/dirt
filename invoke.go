package dirt

import (
	"fmt"
	"iter"
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

// InvokeAs invokes the instance as the requested type T, return shortest dependency depth on if many.
func InvokeAs[T any](opts ...Option) (T, error) {
	opt := defaultProvideOptions()
	for _, o := range opts {
		opt = o(opt)
	}

	key := typeNameKey{Ty: reflect.TypeFor[T](), Name: opt.Name}
	s := opt.Scope

	for v, err := range s.invokeInstanceAsMany(key) {
		typed, ok := v.(T)
		if !ok {
			return *new(T),
				fmt.Errorf("dirt: instance type: `%T` does not match the requested type: `%s`, this should not happen", v, key.Ty.String())
		}
		return typed, err
	}

	return *new(T), fmt.Errorf("dirt: no instance found for type: `%s`", key.Ty.String())
}

// InvokeAs invokes the instance as the requested type T, return shortest dependency depth on if many.
func InvokeAsMany[T any](opts ...Option) iter.Seq2[T, error] {
	opt := defaultProvideOptions()
	for _, o := range opts {
		opt = o(opt)
	}

	key := typeNameKey{Ty: reflect.TypeFor[T](), Name: opt.Name}
	s := opt.Scope

	return func(yield func(T, error) bool) {
		for v, err := range s.invokeInstanceAsMany(key) {
			typed, ok := v.(T)
			if !ok {
				if !yield(*new(T), fmt.Errorf("dirt: instance type: `%T` does not match the requested type: `%s`, this should not happen", v, key.Ty.String())) {
					return
				}
			}
			if !yield(typed, err) {
				return
			}
		}
	}
}
