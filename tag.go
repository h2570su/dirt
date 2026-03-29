package dirt

import (
	"reflect"
	"strings"
)

const (
	TagDirt = "dirt"
)

type tag struct {
	Valid bool // Inversion of ignoring the field.

	Name       string
	Optional   bool
	Individual bool
}

func parseTag(sf reflect.StructField) tag {
	raw, ok := sf.Tag.Lookup(TagDirt)
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
		case "optional":
			parsed.Optional = true
		case "individual":
			parsed.Individual = true
		case "name":
			parsed.Name = value
		default:
			continue
		}
	}
	return parsed
}
