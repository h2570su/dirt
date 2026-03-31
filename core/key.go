package core

import "reflect"

// TypeNameKey is the key type for accessing instance/registration
type TypeNameKey struct {
	Type reflect.Type
	Name string
}

func (k TypeNameKey) String() string {
	if k.Type == nil {
		return "<invalid>"
	}
	if k.Name == "" {
		return k.Type.String()
	}
	return k.Type.String() + "(" + k.Name + ")"
}
