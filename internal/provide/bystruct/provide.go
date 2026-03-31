package bystruct

import (
	"fmt"
	"reflect"

	"github.com/h2570su/dirt/core"
	"github.com/h2570su/dirt/internal/hook"
)

type registration struct {
	key core.TypeNameKey
	// the constructor of the target type, only appears when all dependencies are satisfied
	ctor func(s core.IScope) (reflect.Value, error)

	dependencies []*dependency
	nestCtors    []func(root reflect.Value)
}

func (reg *registration) Key() core.TypeNameKey { return reg.key }

func (reg *registration) DependencyDepth() int {
	maxDepth := 0
	for _, dep := range reg.dependencies {
		depDepth := 0
		if dep.satisfiedBy != nil {
			depDepth = dep.satisfiedBy.DependencyDepth()
		}
		if depDepth > maxDepth {
			maxDepth = depDepth
		}
	}
	return maxDepth + 1
}

func (reg *registration) IsReady() bool {
	if reg.ctor == nil {
		return false
	}
	for _, dep := range reg.dependencies {
		if dep.optional {
			continue
		}
		if dep.satisfiedBy == nil {
			return false
		}
		if !dep.satisfiedBy.IsReady() {
			return false
		}
	}
	return true
}

func (reg *registration) DirectDeps() []core.TypeNameKey {
	deps := make([]core.TypeNameKey, len(reg.dependencies))
	for i, dep := range reg.dependencies {
		deps[i] = dep.key
	}
	return deps
}

type dependency struct {
	key core.TypeNameKey
	// given the top level struct value, locate the dependency field reflect value (ptr in no reflect version)
	locateFn func(reflect.Value) reflect.Value

	optional   bool
	individual bool

	satisfiedBy core.Registration
}

// ProvideStruct registers the struct type T to be provided by the container.
//
//	The dependencies of T determined by its fields and tags.
func ProvideStruct[T any](opt core.Options) {
	rty := reflect.TypeFor[T]()
	concrete := rty
	for concrete.Kind() == reflect.Pointer {
		concrete = concrete.Elem()
	}

	if concrete.Kind() != reflect.Struct {
		panic(fmt.Sprintf("dirt: ProvideStruct only supports struct types, but got %s", rty.String()))
	}

	reg := &registration{
		key: core.TypeNameKey{Type: rty, Name: opt.Name},
	}

	reg.markDeps(rty, func(v reflect.Value) reflect.Value {
		return v
	})

	s := opt.Scope
	s.WriteRegistration(reg)
}

// markDeps recursively marks the dependencies of struct, including nested/indirect access
func (reg *registration) markDeps(rty reflect.Type, accessFromRoot func(reflect.Value) reflect.Value) {
	if rty.Kind() == reflect.Pointer {
		elemTy := rty.Elem()
		if rty != reg.key.Type { // Skip the *root type itself, since it's already handled in the ctor
			reg.nestCtors = append(reg.nestCtors, func(root reflect.Value) {
				accessFromRoot(root).Set(reflect.New(elemTy))
			})
		}
		reg.markDeps(elemTy, func(v reflect.Value) reflect.Value { return accessFromRoot(v).Elem() })
		return
	}
	for i := range rty.NumField() {
		sf := rty.Field(i)
		// Skip Injectable indicator
		switch sf.Type {
		case reflect.TypeFor[Subclass]():
			continue
		}

		var locateFn func(sv reflect.Value) reflect.Value
		if sf.IsExported() {
			locateFn = func(sv reflect.Value) reflect.Value {
				return accessFromRoot(sv).Field(i)
			}
		} else {
			// For unexported fields, reflect.NewAt harshly
			locateFn = func(sv reflect.Value) reflect.Value {
				addr := accessFromRoot(sv).Field(i).Addr().UnsafePointer()
				return reflect.NewAt(sf.Type, addr).Elem()
			}
		}

		// Handle subclass dependencies
		if sf.Type.Implements(reflect.TypeFor[ISubclass]()) {
			// Handle subclass dependencies
			reg.markDeps(sf.Type, locateFn)
			continue
		}

		// Handle normal fields
		tag := parseTag(sf)
		if !tag.Valid {
			continue
		}

		// Handle dependency field
		depRty := sf.Type

		reg.dependencies = append(reg.dependencies, &dependency{
			key:      core.TypeNameKey{Type: depRty, Name: tag.Name},
			locateFn: locateFn,

			individual: tag.Individual,
			optional:   tag.Optional,
		})
	}
}

// Return modified
func (reg *registration) ResolveDependencies(s core.IRegistry) {
	for _, dep := range reg.dependencies {
		if dep.satisfiedBy != nil {
			continue
		}

		for possible := range s.IterRegistration() {
			if possible.Key() == dep.key {
				dep.satisfiedBy = possible
				break
			}
		}
	}

	allDepsSatisfied := true
	for _, dep := range reg.dependencies {
		if dep.satisfiedBy == nil && !dep.optional {
			allDepsSatisfied = false
			break
		}
	}

	if !allDepsSatisfied {
		return
	}
	reg.buildCtor()
	reg.buildCtorWithHook()
}

func (reg *registration) Ctor(s core.IScope) (reflect.Value, error) {
	if reg.ctor == nil {
		return reflect.Value{}, fmt.Errorf("dirt: type: %s has unsatisfied dependencies", reg.key.Type.String())
	}
	return reg.ctor(s)
}

func (reg *registration) buildCtor() {
	reg.ctor = func(s core.IScope) (reflect.Value, error) {
		var instance reflect.Value
		if reg.key.Type.Kind() == reflect.Pointer {
			instance = reflect.New(reg.key.Type.Elem())
		} else {
			instance = reflect.New(reg.key.Type).Elem()
		}
		for _, nest := range reg.nestCtors {
			nest(instance)
		}

		for _, dep := range reg.dependencies {
			// If the dependency is individual, we need to invoke it directly without checking the instance repo
			if dep.individual {
				ins, err := dep.satisfiedBy.Ctor(s)
				if err == nil {
					dep.locateFn(instance).Set(ins)
				} else if !dep.optional { // If the dependency is not optional, return error
					return reflect.Value{}, fmt.Errorf(
						"dirt: failed to construct type: %s required by %s, error: %w",
						dep.key.Type.String(), reg.key.Type.String(), err,
					)
				}
				continue
			}

			// Check if the dependency instance is already created
			if ins, ok := s.GetInstance(dep.key); ok {
				dep.locateFn(instance).Set(reflect.ValueOf(ins))
				continue
			}

			// If not, invoke the dependency
			ins, err := s.InvokeInstance(dep.key)
			if err == nil {
				dep.locateFn(instance).Set(reflect.ValueOf(ins))
			} else if !dep.optional { // If the dependency is not optional, return error
				return reflect.Value{},
					fmt.Errorf(
						"dirt: failed to construct type: %s required by %s, error: %w",
						dep.key.Type.String(), reg.key.Type.String(), err,
					)
			}
		}
		return instance, nil
	}
}

func (reg *registration) buildCtorWithHook() {
	t := reg.key.Type
	current := reg.ctor

	current = hook.CheckAppendPostInjectHookCtor(t, current)

	reg.ctor = current
}
