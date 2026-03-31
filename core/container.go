package core

// IContainer represents a DI container, which holds the instances of the provided named types.
type IContainer interface {
	// GetInstance returns the instance of the provided named type
	GetInstance(key TypeNameKey) (val any, ok bool)
	// WriteInstance inserts/updates the instance of the provided named type into the container
	WriteInstance(key TypeNameKey, val any)
	// GetKeyByInstance returns the key of the provided instance, and whether it exists in the container.
	GetKeyByInstance(val any) (key TypeNameKey, ok bool)
}
