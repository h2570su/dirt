package bystruct

import (
	"fmt"
	"reflect"

	"github.com/h2570su/dirt/core"
)

func checkPotentialInjectableMissingTag(_ reflect.Type, ctor core.Ctor) core.Ctor {
	return func(s core.IScope) (reflect.Value, error) {
		// Call the original constructor to get the instance
		ctorValue, err := ctor(s)
		if err != nil {
			return reflect.Value{}, err
		}

		// Deeply dereference the value to get to the underlying struct type
		rv := ctorValue
		rty := rv.Type()
		for rty.Kind() == reflect.Pointer && !rv.IsNil() {
			rty = rty.Elem()
			rv = rv.Elem()
		}

		if rty.Kind() != reflect.Struct { // Skip non-struct types
			return ctorValue, nil
		}

		// Iterate over the fields of the struct and check type is in the registry and don't have the injectable tag
		for i := range rty.NumField() {
			fieldV := rv.Field(i)
			if !fieldV.IsZero() {
				continue
			}
			// Only check non-zero fields, zero fields may be injectable but not yet injected
			fieldT := rty.Field(i)
			for reg := range s.IterRegistration() {
				if reg.Key().Type != fieldT.Type {
					continue
				}

				// If the field is found in the registry, check if it has the injectable tag
				t := parseTag(fieldT)
				if !t.Valid && !t.Ignored {
					return reflect.Value{}, fmt.Errorf("field %s in struct %s is not initialized but injectable by dirt, consider adding the `dirt:\"\"` tag. If is expected to be manually initialized, adding the `dirt:\"-\"` tag", fieldT.Name, rty.String())
				}
			}
		}
		return ctorValue, nil
	}
}
