package bystruct

import (
	"fmt"
	"reflect"

	"git.ttech.cc/astaroth/dirt/core"
	"git.ttech.cc/astaroth/dirt/internal/hook"
)

type registration struct {
	key core.TypeNameKey
	// the constructor of the target type, only appears when all dependencies are satisfied
	ctor func() (reflect.Value, error)

	dependencies []*dependency
	nestCtors    []func(root reflect.Value)
}

func (reg *registration) Key() core.TypeNameKey { return reg.key }
func (reg *registration) Ctor() (reflect.Value, error) {
	if reg.ctor == nil {
		return reflect.Value{}, fmt.Errorf("dirt: type: %s has unsatisfied dependencies", reg.key.Type.String())
	}
	return reg.ctor()
}

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

func (reg *registration) IsReady() bool { return reg.ctor != nil }

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
func ProvideStruct[T IInjectable](opt core.Options) {
	rty := reflect.TypeFor[T]()

	reg := &registration{
		key: core.TypeNameKey{Type: rty, Name: opt.Name},
	}

	reg.markDeps(rty, func(v reflect.Value) reflect.Value {
		return v
	})

	s := opt.Scope
	s.WriteRegistration(reg)
	for modified := true; modified; {
		modified = false
		for reg := range s.IterRegistration() {
			modified = reg.ResolvePossibleDeps(s) || modified
			if modified {
				break
			}
		}
	}
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
		case reflect.TypeFor[Injectable]():
			continue
		case reflect.TypeFor[Subclass]():
			continue
		}

		locateFn := func(sv reflect.Value) reflect.Value {
			return accessFromRoot(sv).Field(i)
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
func (reg *registration) ResolvePossibleDeps(s core.IScope) bool {
	modified := false

	if reg.ctor != nil {
		return modified
	}

	for _, dep := range reg.dependencies {
		if dep.satisfiedBy != nil {
			continue
		}

		for possible := range s.IterRegistration() {
			if possible.Key() == dep.key && possible.IsReady() {
				dep.satisfiedBy = possible
				modified = true
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
		return modified
	}
	reg.buildCtor(s)
	reg.buildCtorWithHook()
	return true
}

func (reg *registration) buildCtor(s core.IScope) {
	reg.ctor = func() (reflect.Value, error) {
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
				ins, err := dep.satisfiedBy.Ctor()
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
	pt := reflect.PointerTo(t)
	current := reg.ctor

	current = hook.CheckAppendPostInjectHookCtor(t, pt, current)

	reg.ctor = current
}
