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
	// Loop detect after resolved dependencies
	defer s.loopDetect()
	// Resolve dependencies after writing the registration.
	defer s.resolveDependencies()

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

func (s *Registry) resolveDependencies() {
	for reg := range s.IterRegistration() {
		if reg.IsReady() {
			continue
		}
		reg.ResolveDependencies(s)
	}
}

func (s *Registry) loopDetect() {
	trail := make(map[core.TypeNameKey]struct{})
	var visit func(core.TypeNameKey) bool
	visit = func(key core.TypeNameKey) bool {
		if _, ok := trail[key]; ok {
			return false
		}
		trail[key] = struct{}{}

		var reg core.Registration
		for _reg := range s.IterRegistration() {
			if _reg.Key() == key {
				reg = _reg
				break
			}
		}
		if reg == nil {
			return true
		}
		for _, dep := range reg.DirectDeps() {
			if !visit(dep) {
				return false
			}
		}
		delete(trail, key)
		return true
	}

	for reg := range s.IterRegistration() {
		if !visit(reg.Key()) {
			panic("dirt: circular dependency detected of type: " + reg.Key().Type.String())
		}
	}
}
