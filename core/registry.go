package core

import (
	"iter"
	"reflect"
)

// IRegistry represents a registry of the provided types
type IRegistry interface {
	// IterRegistration returns the sequence of all registrations in the registry
	IterRegistration() iter.Seq[Registration]
	// WriteRegistration inserts/updates the registration of the provided named type into the registry
	WriteRegistration(reg Registration)
	// Instantiate creates an instance of the provided named type by its registration in the registry
	Instantiate(key TypeNameKey) (any, error)
}

// Registration represents the registration of a provided named type
type Registration interface {
	// Key returns the key of the provided named type
	Key() TypeNameKey
	// Ctor constructs an instance of provided named type, return value should be exactly Key().Type
	Ctor() (reflect.Value, error)
	// DependencyDepth returns the maximum depth of the dependencies of itself. returns 1 if it has no dependencies.
	DependencyDepth() int
	// IsReady returns whether the registration is ready to be instantiated, which means all its dependencies are satisfied.
	IsReady() bool

	ResolvePossibleDeps(s IScope) bool
}
