package byalias

import (
	"fmt"
	"reflect"

	"github.com/h2570su/dirt/core"
)

type registration struct {
	key    core.TypeNameKey
	needed core.TypeNameKey

	ctor core.Registration
}

func (reg *registration) AliasOf() core.TypeNameKey { return reg.needed }

func (reg *registration) Key() core.TypeNameKey { return reg.key }

func (reg *registration) DependencyDepth() int {
	if reg.ctor == nil {
		return 1
	}
	return reg.ctor.DependencyDepth() + 1
}

func (reg *registration) IsReady() bool {
	if reg.ctor == nil {
		return false
	}
	return reg.ctor.IsReady()
}

func (reg *registration) DirectDeps() []core.TypeNameKey { return []core.TypeNameKey{reg.needed} }

// ProvideAs registers the value prototype of the target type as an interface.
func ProvideAs[T any, I any](opt core.Options) {
	var (
		tRty = reflect.TypeFor[T]()
		iRty = reflect.TypeFor[I]()
	)

	switch {
	case iRty.Kind() != reflect.Interface:
		panic("ProvideAs: I must be an interface type")
	case !tRty.Implements(iRty):
		panic(fmt.Sprintf("ProvideAs: T (%s) does not implement I (%s)", tRty, iRty))
	case !tRty.ConvertibleTo(iRty):
		panic(fmt.Sprintf("ProvideAs: T (%s) is not convertible to I (%s)", tRty, iRty))
	}

	reg := &registration{
		key:    core.TypeNameKey{Type: iRty, Name: opt.Name},
		needed: core.TypeNameKey{Type: tRty, Name: opt.Name},
	}

	s := opt.Scope
	s.WriteRegistration(reg)
}

func (reg *registration) ResolveDependencies(s core.IRegistry) {
	if reg.ctor != nil {
		reg.ctor.ResolveDependencies(s)
	}
	if src, ok := s.GetRegistration(reg.needed); ok {
		reg.ctor = src
	}
}

func (reg *registration) Ctor(s core.IScope) (reflect.Value, error) {
	if reg.ctor == nil {
		return reflect.Value{}, fmt.Errorf("dirt: type: %s has unsatisfied dependency: %v", reg.key.Type.String(), reg.needed.String())
	}

	ins, err := s.InvokeInstance(reg.needed)
	if err != nil {
		return reflect.Value{}, fmt.Errorf(
			"dirt: failed to construct type: %s required by %s, error: %w",
			reg.needed.Type.String(), reg.key.Type.String(), err,
		)
	}

	return reflect.ValueOf(ins).Convert(reg.key.Type), nil
}
