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
		sb.WriteString("{")
		sb.WriteString(k.Type.PkgPath())
		sb.WriteString("}")
	}
	if k.Name != "" {
		sb.WriteString("(")
		sb.WriteString(k.Name)
		sb.WriteString(")")
	}
	return sb.String()
}
