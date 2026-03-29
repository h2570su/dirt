package simple

import (
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
	defer func() {
		for reg := range s.IterRegistration() {
			if reg.IsReady() {
				continue
			}
			reg.ResolveDependencies(s)
		}
	}()
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
