package dirt

import (
	"fmt"
	"iter"
	"reflect"
)

type typeNameKey struct {
	Ty   reflect.Type
	Name string
}

var globalScope = &Scope{}

// Scope represents a scope of the provided types/instances, which holds the registrations and instances.
type Scope struct {
	// TODO: thread-safety

	registrations []*registration
	instances     map[typeNameKey]any
}

func (s *Scope) iterRegistration() iter.Seq[*registration] {
	return func(yield func(*registration) bool) {
		for _, reg := range s.registrations {
			if !yield(reg) {
				return
			}
		}
	}
}

// writeRegistration writes the registration to all scopes
func (s *Scope) writeRegistration(reg registration) {
	for i := range s.registrations {
		if s.registrations[i].key == reg.key {
			s.registrations[i] = &reg
			return
		}
	}

	s.registrations = append(s.registrations, &reg)
}

func (s *Scope) instantiate(key typeNameKey) (any, error) {
	var reg *registration
	for _, _reg := range s.registrations {
		if _reg.key == key {
			reg = _reg
			break
		}
	}
	if reg == nil {
		return nil, fmt.Errorf("dirt: no provider found for type %s", key.Ty.String())
	}

	if reg.ctor == nil {
		return nil, fmt.Errorf("dirt: type: %s has unsatisfied dependencies", key.Ty.String())
	}
	ins, err := reg.ctor()
	if err != nil {
		return nil, err
	}
	return ins.Interface(), nil
}

func (s *Scope) getInstance(key typeNameKey) (any, bool) {
	if val, ok := s.instances[key]; ok {
		return val, true
	}

	return nil, false
}

func (s *Scope) writeInstance(key typeNameKey, val any) {
	if _, ok := s.instances[key]; ok {
		s.instances[key] = val
		return
	}

	if s.instances == nil {
		s.instances = make(map[typeNameKey]any)
	}
	s.instances[key] = val
}

func (s *Scope) invokeInstance(key typeNameKey) (any, error) {
	if val, ok := s.instances[key]; ok {
		return val, nil
	}

	var reg *registration
	for _, _reg := range s.registrations {
		if _reg.key == key {
			reg = _reg
			break
		}
	}

	if reg == nil {
		return nil, fmt.Errorf("dirt: no provider found for type %s", key.Ty.String())
	}

	if reg.ctor == nil {
		return nil, fmt.Errorf("dirt: type: %s has unsatisfied dependencies", key.Ty.String())
	}
	ins, err := reg.ctor()
	if err != nil {
		return nil, err
	}
	if s.instances == nil {
		s.instances = make(map[typeNameKey]any)
	}
	anyIns := ins.Interface()
	s.instances[key] = anyIns
	return anyIns, nil
}
