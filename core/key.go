package core

import (
	"reflect"
	"strings"
)

// TypeNameKey is the key type for accessing instance/registration
type TypeNameKey struct {
	Type reflect.Type
	Name string
}

func (k TypeNameKey) String() string {
	if k.Type == nil {
		return "<invalid>"
	}
	sb := strings.Builder{}
	sb.WriteString(k.Type.String())
	if k.Type.PkgPath() != "" {
		sb.WriteString("{" + k.Type.PkgPath() + "}")
	}
	if k.Name != "" {
		sb.WriteString("(" + k.Name + ")")
	}
	return sb.String()
}
