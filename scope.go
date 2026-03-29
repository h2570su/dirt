package dirt

import (
	"fmt"
	"iter"
	"reflect"
)

type TypeNameKey struct {
	Ty   reflect.Type
	Name string
}

var globalScope = &Scope{}

type Scope struct {
	// TODO: thread-safety

	registrations []*registration
	instances     map[TypeNameKey]reflect.Value
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

func (s *Scope) instantiate(key TypeNameKey) (reflect.Value, error) {
	var reg *registration
	for _, _reg := range s.registrations {
		if _reg.key == key {
			reg = _reg
			break
		}
	}
	if reg == nil {
		return reflect.Value{}, fmt.Errorf("dirt: no provider found for type %s", key.Ty.String())
	}

	if reg.ctor == nil {
		return reflect.Value{}, fmt.Errorf("dirt: type: %s has unsatisfied dependencies", key.Ty.String())
	}
	return reg.ctor()
}

func (s *Scope) getInstance(key TypeNameKey) (reflect.Value, bool) {
	if val, ok := s.instances[key]; ok {
		return val, true
	}

	return reflect.Value{}, false
}

func (s *Scope) writeInstance(key TypeNameKey, val reflect.Value) {
	if _, ok := s.instances[key]; ok {
		s.instances[key] = val
		return
	}

	if s.instances == nil {
		s.instances = make(map[TypeNameKey]reflect.Value)
	}
	s.instances[key] = val
}

func (s *Scope) invokeInstance(key TypeNameKey) (reflect.Value, error) {
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
		return reflect.Value{}, fmt.Errorf("dirt: no provider found for type %s", key.Ty.String())
	}

	if reg.ctor == nil {
		return reflect.Value{}, fmt.Errorf("dirt: type: %s has unsatisfied dependencies", key.Ty.String())
	}
	ins, err := reg.ctor()
	if err != nil {
		return reflect.Value{}, err
	}
	if s.instances == nil {
		s.instances = make(map[TypeNameKey]reflect.Value)
	}
	s.instances[key] = ins
	return ins, nil
}
