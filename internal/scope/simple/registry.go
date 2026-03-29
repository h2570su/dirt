package simple

import (
	"fmt"
	"iter"
	"sync"

	"git.ttech.cc/astaroth/dirt/core"
)

type Registry struct {
	lock          sync.RWMutex
	registrations []core.Registration
}

func (s *Registry) IterRegistration() iter.Seq[core.Registration] {
	return func(yield func(core.Registration) bool) {
		s.lock.RLock()
		defer s.lock.RUnlock()

		for _, reg := range s.registrations {
			if !yield(reg) {
				return
			}
		}
	}
}

func (s *Registry) WriteRegistration(reg core.Registration) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for i := range s.registrations {
		if s.registrations[i].Key() == reg.Key() {
			s.registrations[i] = reg
			return
		}
	}

	s.registrations = append(s.registrations, reg)
}

func (s *Registry) Instantiate(key core.TypeNameKey) (any, error) {
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

	ins, err := reg.Ctor()
	if err != nil {
		return nil, fmt.Errorf("dirt: failed to instantiate type: `%s`, error: %w", key.Type.String(), err)
	}
	return ins.Interface(), nil
}
