package dirt

import (
	"reflect"
)

type registration struct {
	key typeNameKey
	// the constructor of the target type, only appears when all dependencies are satisfied
	ctor func() (reflect.Value, error)

	dependencies []*dependency
	nestCtors    []func(root reflect.Value)
}

type dependency struct {
	key typeNameKey
	// given the top level struct value, locate the dependency field reflect value (ptr in no reflect version)
	locateFn func(reflect.Value) reflect.Value

	optional   bool
	individual bool

	satisfiedBy *registration
}

// ProvideStruct registers the struct type T to be provided by the container.
//
//	The dependencies of T determined by its fields and tags.
func ProvideStruct[T iInjectable](opts ...Option) {
	opt := defaultProvideOptions()
	for _, o := range opts {
		opt = o(opt)
	}

	rty := reflect.TypeFor[T]()

	reg := registration{
		key: typeNameKey{Ty: rty, Name: opt.Name},
	}

	reg.markDeps(rty, func(v reflect.Value) reflect.Value {
		return v
	})

	s := opt.Scope
	s.writeRegistration(reg)
	for modified := true; modified; {
		modified = false
		for reg := range s.iterRegistration() {
			modified = reg.resolvePossibleDeps(s) || modified
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
		if rty != reg.key.Ty { // Skip the *root type itself, since it's already handled in the ctor
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
		if sf.Type.Implements(reflect.TypeFor[iSubclass]()) {
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
			key:      typeNameKey{Ty: depRty, Name: tag.Name},
			locateFn: locateFn,

			individual: tag.Individual,
			optional:   tag.Optional,
		})
	}
}

// Return modified
func (reg *registration) resolvePossibleDeps(s *Scope) bool {
	modified := false

	if reg.ctor != nil {
		return modified
	}

	for _, dep := range reg.dependencies {
		if dep.satisfiedBy != nil {
			continue
		}

		for possible := range s.iterRegistration() {
			if possible.key == dep.key && possible.ctor != nil {
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
	reg.buildCtorWithHook(s)
	return true
}

func (reg *registration) buildCtor(s *Scope) {
	reg.ctor = func() (reflect.Value, error) {
		var instance reflect.Value
		if reg.key.Ty.Kind() == reflect.Pointer {
			instance = reflect.New(reg.key.Ty.Elem())
		} else {
			instance = reflect.New(reg.key.Ty).Elem()
		}
		for _, nest := range reg.nestCtors {
			nest(instance)
		}

		for _, dep := range reg.dependencies {
			// If the dependency is individual, we need to invoke it directly without checking the instance repo
			if dep.individual {
				ins, err := dep.satisfiedBy.ctor()
				if err == nil {
					dep.locateFn(instance).Set(ins)
				} else if !dep.optional { // If the dependency is not optional, return error
					return reflect.Value{}, err
				}
				continue
			}

			// Check if the dependency instance is already created
			if ins, ok := s.getInstance(dep.key); ok {
				dep.locateFn(instance).Set(reflect.ValueOf(ins))
				continue
			}

			// If not, invoke the dependency
			ins, err := s.invokeInstance(dep.key)
			if err == nil {
				dep.locateFn(instance).Set(reflect.ValueOf(ins))
			} else if !dep.optional { // If the dependency is not optional, return error
				return reflect.Value{}, err
			}
		}
		return instance, nil
	}
}

func (reg *registration) buildCtorWithHook(_ *Scope) {
	t := reg.key.Ty
	pt := reflect.PointerTo(t)
	current := reg.ctor

	current = checkAppendPostInjectHookCtor(t, pt, current)

	reg.ctor = current
}
