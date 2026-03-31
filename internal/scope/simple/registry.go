package simple

import (
	"fmt"
	"iter"
	"slices"
	"sync"

	"git.ttech.cc/astaroth/dirt/core"
)

type Registry struct {
	lock          sync.RWMutex
	registrations []core.Registration
}

func (s *Registry) GetDependencies(key core.TypeNameKey) ([]core.TypeNameKey, error) {
	var tree [][]core.TypeNameKey
	tree = append(tree, []core.TypeNameKey{key})

	// Iteratively traverse the dependency tree level by level until no more dependencies are found.
	for currentLvl := 0; currentLvl < len(tree); currentLvl++ {
		var currLvlDeps []core.TypeNameKey
		for _, dep := range tree[currentLvl] {
			r, ok := s.GetRegistration(dep)
			if !ok {
				return nil, fmt.Errorf("dirt: registration not found for type: `%s`", dep.Type.String())
			}
			currLvlDeps = append(currLvlDeps, r.DirectDeps()...)
		}
		if len(currLvlDeps) > 0 {
			tree = append(tree, currLvlDeps)
		}
	}

	// Flatten the tree into a single slice of dependencies, excluding the original key.
	tree = tree[1:] // Remove the first level which is the key itself.
	var flatTree []core.TypeNameKey
	for _, lvl := range tree {
		flatTree = append(flatTree, lvl...)
	}

	// Remove duplicates from the flat tree.
	seen := make(map[core.TypeNameKey]struct{}, len(flatTree))
	uniqueDeps := make([]core.TypeNameKey, 0, len(flatTree))
	// Traverse in reverse to maintain the order of dependencies, deeper dependencies stay deeper in the list.
	for _, dep := range slices.Backward(flatTree) {
		if _, ok := seen[dep]; !ok {
			seen[dep] = struct{}{}
			uniqueDeps = append(uniqueDeps, dep)
		}
	}
	slices.Reverse(uniqueDeps) // Reverse back to the original order.

	return uniqueDeps, nil
}

func (s *Registry) GetRegistration(key core.TypeNameKey) (core.Registration, bool) { //nolint:ireturn
	for reg := range s.IterRegistration() {
		if reg.Key() == key {
			return reg, true
		}
	}
	return nil, false
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
			delete(trail, key)
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
