package dirt

import (
	"reflect"
)

func ProvideStruct[T IInjectable](scopes ...*Scope) {
	ss := getScopes(scopes...)
	rty := reflect.TypeFor[T]()

	concreteRty := rty
	for concreteRty.Kind() == reflect.Pointer {
		concreteRty = concreteRty.Elem()
	}
	if concreteRty.Kind() != reflect.Struct {
		panic("dirt: only struct can be provided")
	}

	reg := registration{
		key:          TypeNameKey{Ty: rty},
		concreteType: concreteRty,
	}

	reg.markDeps(concreteRty, func(v reflect.Value) reflect.Value {
		return v
	})

	ss.writeRegistration(reg)
	for modified := true; modified; {
		modified = false
		for _, reg := range ss.iterRegistration() {
			modified = reg.resolvePossibleDeps(ss) || modified
			if modified {
				break
			}
		}
	}
}

func Invoke[T any](scopes ...*Scope) (T, error) {
	ss := getScopes(scopes...)
	key := TypeNameKey{Ty: reflect.TypeFor[T]()}

	if ins, ok := ss.getInstance(key); ok {
		typed, ok := ins.Interface().(T)
		if !ok {
			panic("dirt: instance type does not match the requested type")
		}
		return typed, nil
	}

	for _, reg := range ss.iterRegistration() {
		if reg.key == key {
			if reg.ctor == nil {
				panic("dirt: type: " + key.Ty.String() + " has unsatisfied dependencies")
			}
			ins, err := reg.ctor()
			if err != nil {
				return *new(T), err
			}
			ss.writeInstance(key, ins)
			typed, ok := ins.Interface().(T)
			if !ok {
				panic("dirt: instance type does not match the requested type")
			}
			return typed, nil
		}
	}

	panic("dirt: no provider found for type " + key.Ty.String())
}

func (reg *registration) markDeps(structRty reflect.Type, rootToStructFn func(reflect.Value) reflect.Value) {
	if structRty.Kind() == reflect.Pointer {
		elemTy := structRty.Elem()
		reg.markDeps(elemTy, func(v reflect.Value) reflect.Value {
			return rootToStructFn(v).Elem()
		})
		return
	}
	for i := 0; i < structRty.NumField(); i++ {
		sf := structRty.Field(i)
		// Skip Injectable indicator
		if sf.Type == reflect.TypeFor[Injectable]() {
			continue
		}
		locateFn := func(sv reflect.Value) reflect.Value {
			return rootToStructFn(sv).Field(i)
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
			key:      TypeNameKey{Ty: depRty, Name: tag.Name},
			locateFn: locateFn,
		})
	}
}

// Return modified
func (reg *registration) resolvePossibleDeps(ss scopes) bool {
	modified := false

	if reg.ctor != nil {
		return modified
	}

	for _, dep := range reg.dependencies {
		if dep.satisfiedBy != nil {
			continue
		}
		for _, possible := range ss.iterRegistration() {
			if possible.key == dep.key && possible.ctor != nil {
				dep.satisfiedBy = possible
				modified = true
				break
			}
		}
	}

	allDepsSatisfied := true
	for _, dep := range reg.dependencies {
		if dep.satisfiedBy == nil {
			allDepsSatisfied = false
			break
		}
	}

	if !allDepsSatisfied {
		return modified
	}

	reg.ctor = func() (reflect.Value, error) {
		instance := reflect.New(reg.key.Ty).Elem()
		concrete := instance
		for concrete.Kind() == reflect.Pointer {
			concrete.Set(reflect.New(concrete.Type().Elem()))
			concrete = concrete.Elem()
		}
		if concrete.Type() != reg.concreteType {
			panic("dirt: concrete type does not match the registration type")
		}

		for _, dep := range reg.dependencies {
			// Check if the dependency instance is already created
			if ins, ok := ss.getInstance(dep.key); ok {
				dep.locateFn(concrete).Set(ins)
				continue
			}
			// TODO: implement individual

			// If not, create the dependency instance
			ins, err := dep.satisfiedBy.ctor()
			if err != nil {
				return reflect.Value{}, err
			}
			dep.locateFn(concrete).Set(ins)
			ss.writeInstance(dep.key, ins)
		}

		return instance, nil
	}
	return true
}

type registration struct {
	key          TypeNameKey
	concreteType reflect.Type
	// the constructor of the target type, only appears when all dependencies are satisfied
	ctor func() (reflect.Value, error)

	dependencies []*dependency
}

type dependency struct {
	key TypeNameKey
	// given the top level struct value, locate the dependency field reflect value (ptr in no reflect version)
	locateFn func(reflect.Value) reflect.Value

	satisfiedBy *registration
}
