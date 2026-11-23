package byvalue

import (
	"reflect"

	"github.com/h2570su/dirt/core"
)

type registration struct {
	key core.TypeNameKey

	value reflect.Value
	ctor  core.Ctor
}

func (reg *registration) Key() core.TypeNameKey { return reg.key }

func (reg *registration) DependencyDepth() int { return 1 }

func (reg *registration) IsReady() bool { return true }

func (reg *registration) DirectDeps() []core.TypeNameKey { return nil }

// ProvideValue registers the value prototype of the target type
func ProvideValue[T any](value T, opt core.Options) {
	rty := reflect.TypeOf(value)
	copyV := reflect.New(rty).Elem()
	copyV.Set(reflect.ValueOf(value))
	reg := &registration{
		key:   core.TypeNameKey{Type: rty, Name: opt.Name},
		value: copyV,
	}
	reg.buildCtor()
	reg.ctor = opt.Scope.OnCtorBuilt(rty, reg.ctor)

	s := opt.Scope
	s.WriteRegistration(reg)
}

func (reg *registration) ResolveDependencies(s core.IRegistry) {}

func (reg *registration) Ctor(s core.IScope) (reflect.Value, error) {
	return reg.ctor(s)
}

func (reg *registration) buildCtor() {
	reg.ctor = func(core.IScope) (reflect.Value, error) {
		return reg.value, nil
	}
}
