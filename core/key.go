package core

import "reflect"

// TypeNameKey is the key type for accessing instance/registration
type TypeNameKey struct {
	Type reflect.Type
	Name string
}
