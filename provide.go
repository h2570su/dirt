package dirt

import "reflect"

type registration interface {
	Key() typeNameKey
	Ctor() func() (reflect.Value, error)

	resolvePossibleDeps(s *Scope) bool
}
