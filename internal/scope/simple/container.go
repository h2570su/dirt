package simple

import (
	"sync"

	"git.ttech.cc/astaroth/dirt/core"
)

type Container struct {
	lock      sync.RWMutex
	instances map[core.TypeNameKey]any
}

func (s *Container) GetInstance(key core.TypeNameKey) (any, bool) {
	s.lock.RLock()
	m := s.instances
	if m != nil {
		defer s.lock.RUnlock()
	} else {
		s.lock.RUnlock()
		s.lock.Lock()
		if s.instances == nil { // double-checked locking
			s.instances = make(map[core.TypeNameKey]any)
		}
		m = s.instances
		s.lock.Unlock()
	}

	if val, ok := m[key]; ok {
		return val, true
	}

	return nil, false
}

func (s *Container) WriteInstance(key core.TypeNameKey, val any) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.instances == nil {
		s.instances = make(map[core.TypeNameKey]any)
	}

	if _, ok := s.instances[key]; ok {
		s.instances[key] = val
		return
	}

	s.instances[key] = val
}

func (s *Container) GetKeyByInstance(val any) (core.TypeNameKey, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	for k, v := range s.instances {
		if v == val {
			return k, true
		}
	}

	return core.TypeNameKey{}, false
}
