package simple

import (
	"fmt"
	"iter"
	"slices"

	"github.com/h2570su/dirt/core"
)

type Scope struct {
	core.IRegistry
	core.IContainer

	// TODO: thread-safety
}

func NewScope() *Scope { return &Scope{IRegistry: &Registry{}, IContainer: &Container{}} }

func (s *Scope) Instantiate(key core.TypeNameKey) (any, error) {
	var reg core.Registration
	for _reg := range s.IterRegistration() {
		if _reg.Key() == key {
			reg = _reg
			break
		}
	}
	if reg == nil {
		return nil, fmt.Errorf("dirt: no provider found for type %s", key.Type.String())
	}

	ins, err := reg.Ctor(s)
	if err != nil {
		return nil, fmt.Errorf("dirt: failed to instantiate type: `%s`, error: \n%w", key.Type.String(), err)
	}
	return ins.Interface(), nil
}

func (s *Scope) InvokeInstance(key core.TypeNameKey) (any, error) {
	if val, ok := s.GetInstance(key); ok {
		return val, nil
	}

	var reg core.Registration
	for _reg := range s.IterRegistration() {
		if _reg.Key() == key {
			reg = _reg
			break
		}
	}

	if reg == nil {
		return nil, fmt.Errorf("dirt: no provider found for type %s", key.Type.String())
	}

	ins, err := reg.Ctor(s)
	if err != nil {
		return nil, fmt.Errorf("dirt: failed to instantiate type: `%s`, error: \n%w", key.Type.String(), err)
	}

	anyIns := ins.Interface()
	s.WriteInstance(reg.Key(), anyIns)
	return anyIns, nil
}

// sort by dependency depth, shallowest first.
func (s *Scope) InvokeInstanceAsMany(key core.TypeNameKey) iter.Seq2[any, error] {
	regs := s.collectMatchingRegistrations(key)

	return func(yield func(any, error) bool) {
		returned := make(map[core.TypeNameKey]struct{})

		for _, reg := range regs {
			// Handle Duplicate Registrations and Aliases
			concreteKey := reg.Key()
			if a, ok := reg.(core.RegistrationAlias); ok {
				// Handle alias
				concreteKey = a.AliasOf()
			}
			if _, ok := returned[concreteKey]; ok {
				continue
			}

			if val, ok := s.GetInstance(concreteKey); ok {
				returned[concreteKey] = struct{}{}
				if !yield(val, nil) {
					return
				}
				continue
			}

			ins, err := reg.Ctor(s)
			if err != nil {
				if !yield(nil, err) {
					return
				}
				continue
			}
			val := ins.Interface()
			s.WriteInstance(concreteKey, val)
			if !yield(val, nil) {
				returned[concreteKey] = struct{}{}
				return
			}
		}
	}
}

func (s *Scope) collectMatchingRegistrations(key core.TypeNameKey) []core.Registration {
	var regs []core.Registration
	for reg := range s.IterRegistration() {
		_key := reg.Key()
		if _key.Name != key.Name {
			continue
		}
		if _key == key {
			regs = append(regs, reg)
			continue
		}
		if _key.Type.Implements(key.Type) {
			regs = append(regs, reg)
			continue
		}
	}
	slices.SortFunc(regs, func(a, b core.Registration) int {
		return a.DependencyDepth() - b.DependencyDepth()
	})
	return regs
}
