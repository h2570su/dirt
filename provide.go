package dirt

import (
	"git.ttech.cc/astaroth/dirt/core"
	"git.ttech.cc/astaroth/dirt/internal/provide/byctor"
	"git.ttech.cc/astaroth/dirt/internal/provide/bystruct"
)

// ProvideStruct registers the struct type T to be provided by the container.
//
//	The dependencies of T determined by its fields and tags.
func ProvideStruct[T bystruct.IInjectable](opts ...core.Option) {
	var opt core.Options
	for _, o := range opts {
		opt = o(opt)
	}
	bystruct.ProvideStruct[T](opt)
}

// ProvideCtor registers the constructor of the target type
//
//	Valid constructors include:
//	- func([more args,]) T
//	- func([more args,]) (T, ~error)
//
// variadic arg will be ignored
func ProvideCtor(fn any, opts ...core.Option) {
	var opt core.Options
	for _, o := range opts {
		opt = o(opt)
	}
	byctor.ProvideCtor(fn, opt)
}
