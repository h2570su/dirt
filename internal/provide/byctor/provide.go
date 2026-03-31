package byctor

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/h2570su/dirt/core"
	"github.com/h2570su/dirt/internal/hook"
)

type registration struct {
	key core.TypeNameKey
	// the constructor of the target type, only appears when all dependencies are satisfied
	ctor func(s core.IScope) (reflect.Value, error)

	fn reflect.Value

	dependencies []*dependency
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
	key   core.TypeNameKey
	index int

	satisfiedBy core.Registration
}

// ProvideCtor registers the constructor of the target type
//
//	Valid constructors include:
//	- func([more args,]) T
//	- func([more args,]) (T, ~error)
//
// variadic arg will be ignored
func ProvideCtor(fn any, opt core.Options) {
	rty := reflect.TypeOf(fn)
	if rty.Kind() != reflect.Func {
		panic("dirt: only function can be provided as constructor, but got " + rty.Kind().String())
	}
	reg := &registration{
		key: core.TypeNameKey{Name: opt.Name},
		fn:  reflect.ValueOf(fn),
	}
	switch rty.NumOut() {
	case 1:
		reg.key.Type = rty.Out(0)
	case 2:
		if !rty.Out(1).Implements(reflect.TypeFor[error]()) {
			panic("dirt: the second return value of constructor must implement error, but got " + rty.Out(1).String())
		}
		reg.key.Type = rty.Out(0)
	default:
		panic("dirt: constructor must return either one value or two values (the second one is error), but got " + strconv.Itoa(rty.NumOut()))
	}

	reg.markDeps()

	s := opt.Scope
	s.WriteRegistration(reg)
}

func (reg *registration) markDeps() {
	rty := reg.fn.Type()
	for i := range rty.NumIn() {
		if rty.IsVariadic() && i == rty.NumIn()-1 {
			break
		}
		depRty := rty.In(i)

		reg.dependencies = append(reg.dependencies, &dependency{
			key:   core.TypeNameKey{Type: depRty},
			index: i,
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
		if dep.satisfiedBy == nil {
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
		var unsatisfiedDeps []string
		for _, dep := range reg.dependencies {
			if dep.satisfiedBy == nil {
				unsatisfiedDeps = append(unsatisfiedDeps, dep.key.String())
			}
		}
		return reflect.Value{}, fmt.Errorf("dirt: type: %s has unsatisfied dependencies: %v", reg.key.Type.String(), unsatisfiedDeps)
	}
	return reg.ctor(s)
}

func (reg *registration) buildCtor() {
	rty := reg.fn.Type()
	var returnsErr bool
	if rty.NumOut() == 2 {
		returnsErr = true
	}
	numIn := rty.NumIn()
	reg.ctor = func(s core.IScope) (reflect.Value, error) {
		args := make([]reflect.Value, numIn)
		for i, dep := range reg.dependencies {
			// Check if the dependency instance is already created
			ins, ok := s.GetInstance(dep.key)
			if !ok {
				// If not, invoke the dependency
				_ins, err := s.InvokeInstance(dep.key)
				if err != nil {
					return reflect.Value{},
						fmt.Errorf(
							"dirt: failed to construct type: %s required by %s, error: %w",
							dep.key.Type.String(), reg.key.Type.String(), err,
						)
				}
				ins = _ins
			}
			args[i] = reflect.ValueOf(ins)
		}
		got := reg.fn.Call(args)
		if returnsErr && !got[1].IsNil() {
			err, ok := got[1].Interface().(error)
			if !ok {
				return reflect.Value{},
					fmt.Errorf(
						"dirt: the second return value of constructor must implement error, but got %s",
						got[1].Type().String(),
					)
			}
			return got[0], err
		}
		return got[0], nil
	}
}

func (reg *registration) buildCtorWithHook() {
	t := reg.key.Type
	current := reg.ctor

	current = hook.CheckAppendPostInjectHookCtor(t, current)

	reg.ctor = current
}
