package core

import "iter"

// IScope represents a scope of the provided types/instances, which holds the registrations and instances.
type IScope interface {
	IRegistry
	IContainer

	// InvokeInstance creates/gets the instance of the provided named type by its registration in the scope.
	InvokeInstance(key TypeNameKey) (any, error)
	// InvokeInstanceAsMany creates/gets the instances as required by the provided named interface type
	//	If the provided type is not an interface, it behaves the same as InvokeInstance.
	//	Order by the dependency depth of the types in dependency graph, shorter first, order in same depth is not guaranteed.
	InvokeInstanceAsMany(key TypeNameKey) iter.Seq2[any, error]
}
