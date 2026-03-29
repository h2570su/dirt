package core

// IContainer represents a DI container, which holds the instances of the provided named types.
type IContainer interface {
	// GetInstance returns the instance of the provided named type
	GetInstance(key TypeNameKey) (val any, ok bool)
	// WriteInstance inserts/updates the instance of the provided named type into the container
	WriteInstance(key TypeNameKey, val any)
}
