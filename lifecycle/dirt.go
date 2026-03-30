package lifecycle

import (
	"reflect"

	"git.ttech.cc/astaroth/dirt/core"
)

func (l *Lifecycle) DirtAddAll(scope core.IScope) error {
	for instance, err := range scope.InvokeInstanceAsMany(core.TypeNameKey{Type: reflect.TypeFor[any]()}) {
		if err != nil {
			return err
		}
		l.TryAdd(instance)
	}
	return nil
}
