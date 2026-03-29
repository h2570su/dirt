package bystruct

import (
	"reflect"
	"strings"
)

// Tag format: `dirt:"[name:instanceName,][optional,][individual]"`
// each options separated by comma, order not required.
// name:instanceName is used to distinguish different instances of the same type, which is useful when there are multiple providers for the same type. default "".
// optional marks the dependency as optional, not to be injected if no provider is found/successfully instantiated.
// individual marks the dependency as individual, always to be injected with a new instance, even if there is a existing instance.

const (
	tagDirt          = "dirt"
	tagKeyName       = "name"
	tagKeyOptional   = "optional"
	tagKeyIndividual = "individual"
)

type tag struct {
	Valid bool // Inversion of ignoring the field.

	Name       string
	Optional   bool
	Individual bool
}

func parseTag(sf reflect.StructField) tag {
	raw, ok := sf.Tag.Lookup(tagDirt)
	if !ok {
		return tag{}
	}
	parts := strings.Split(raw, ",")
	var parsed tag
	parsed.Valid = true

	for _, part := range parts {
		possibleKV := strings.SplitN(part, ":", 2)
		var key, value string
		if len(possibleKV) == 1 {
			key = possibleKV[0]
		} else {
			key, value = possibleKV[0], possibleKV[1]
		}
		switch key {
		case tagKeyOptional:
			parsed.Optional = true
		case tagKeyIndividual:
			parsed.Individual = true
		case tagKeyName:
			parsed.Name = value
		default:
			continue
		}
	}
	return parsed
}
