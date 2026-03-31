package lifecycle

import (
	"fmt"
	"slices"

	"git.ttech.cc/astaroth/dirt/core"
)

// DirtAddAllFor adds the provided instance and its dependencies into the lifecycle from the provided scope.
func (l *Lifecycle) DirtAddAllFor(scope core.IScope, instance any) error {
	// Get the key of the provided instance from the scope.
	key, ok := scope.GetKeyByInstance(instance)
	if !ok {
		return fmt.Errorf("lifecycle: instance of type `%T` not found in the provided scope, ensure it has been invoked", instance)
	}

	// Get the dependencies of the provided instance from the scope.
	deps, err := scope.GetDependencies(key)
	if err != nil {
		return fmt.Errorf("lifecycle: failed to get dependencies for type: `%s`: %w", key.Type.String(), err)
	}
	slices.Reverse(deps) // Reverse the dependencies to ensure deeper dependencies are added first.

	// Add the provided instance and its dependencies into the lifecycle, deeper dependencies first.
	toAdd := make([]any, 0, len(deps)+1)
	for _, dep := range deps {
		depInst, ok := scope.GetInstance(dep)
		if !ok {
			return fmt.Errorf("lifecycle: instance not found for dependency: `%s`, ensure it has been invoked", dep.Type.String())
		}
		toAdd = append(toAdd, depInst)
	}
	toAdd = append(toAdd, instance)

	// Add the provided instance and its dependencies into the lifecycle, deeper dependencies first.
	for _, inst := range toAdd {
		l.TryAdd(inst)
	}

	return nil
}
