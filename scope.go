package dirt

import (
	"iter"
	"reflect"
)

type TypeNameKey struct {
	Ty   reflect.Type
	Name string
}

var globalScope = &Scope{}

type Scope struct {
	registrations []*registration
	instances     map[TypeNameKey]reflect.Value
}

type scopes []*Scope

func getScopes(ss ...*Scope) scopes {
	if len(ss) == 0 {
		return scopes{globalScope}
	}
	return scopes(ss)
}

func (s scopes) iterRegistration() iter.Seq2[*Scope, *registration] {
	return func(yield func(*Scope, *registration) bool) {
		for _, scope := range s {
			for _, reg := range scope.registrations {
				if !yield(scope, reg) {
					return
				}
			}
		}
	}
}

// writeRegistration writes the registration to all scopes
func (s scopes) writeRegistration(reg registration) {
	for _, scope := range s {
		found := false
		for i := range scope.registrations {
			if scope.registrations[i].key == reg.key {
				scope.registrations[i] = &reg
				found = true
				break
			}
		}
		if !found {
			scope.registrations = append(scope.registrations, &reg)
		}
	}
}

func (s scopes) getInstance(key TypeNameKey) (reflect.Value, bool) {
	for _, scope := range s {
		if val, ok := scope.instances[key]; ok {
			return val, true
		}
	}
	return reflect.Value{}, false
}

func (s scopes) writeInstance(key TypeNameKey, val reflect.Value) {
	for _, scope := range s {
		if _, ok := scope.instances[key]; ok {
			scope.instances[key] = val
		} else {
			if scope.instances == nil {
				scope.instances = make(map[TypeNameKey]reflect.Value)
			}
			scope.instances[key] = val
		}
	}
}
