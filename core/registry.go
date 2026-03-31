package core

import (
	"iter"
	"reflect"
)

type Ctor func(s IScope) (reflect.Value, error)

// IRegistry represents a registry of the provided types
type IRegistry interface {
	// IterRegistration returns the sequence of all registrations in the registry
	IterRegistration() iter.Seq[Registration]
	// WriteRegistration inserts/updates the registration of the provided named type into the registry
	WriteRegistration(reg Registration)
	// GetRegistration returns the registration of the provided named type, and whether it exists in the registry.
	GetRegistration(key TypeNameKey) (Registration, bool)

	// GetDependencies returns the dependencies of the provided named type, closer first.
	//	Returns an error if the provided named type is not registered in the registry.
	GetDependencies(key TypeNameKey) ([]TypeNameKey, error)
}

// Registration represents the registration of a provided named type
type Registration interface {
	// Key returns the key of the provided named type
	Key() TypeNameKey
	// Ctor constructs an instance of provided named type, return value should be exactly Key().Type
	Ctor(s IScope) (reflect.Value, error)
	// DependencyDepth returns the maximum depth of the dependencies of itself. returns 1 if it has no dependencies.
	DependencyDepth() int
	// IsReady returns whether the registration is ready to be instantiated, which means all its dependencies are satisfied.
	IsReady() bool
	// ResolveDependencies resolves the dependencies of the registration with the registrations in the registry, which should be called after writing a new registration into the registry.
	ResolveDependencies(s IRegistry)
	// DirectDeps returns the direct dependencies of the registration, which is used for cycle detection.
	DirectDeps() []TypeNameKey
}
