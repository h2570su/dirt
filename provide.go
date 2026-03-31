package dirt

import (
	"github.com/h2570su/dirt/core"
	"github.com/h2570su/dirt/internal"
	"github.com/h2570su/dirt/internal/provide/byctor"
	"github.com/h2570su/dirt/internal/provide/bystruct"
	"github.com/h2570su/dirt/internal/provide/byvalue"
)

// nope is a dummy type for var _ = ProvideXXX() for shorter than func init() { ProvideXXX() }
type nope = struct{ _ internal.Sentinel }

var _nope = nope{}

// ProvideStruct registers the struct type T to be provided by the container.
//
//	The dependencies of T determined by its fields and tags.
func ProvideStruct[T any](opts ...core.Option) nope {
	opt := defaultOptions
	for _, o := range opts {
		opt = o(opt)
	}
	bystruct.ProvideStruct[T](opt)
	return _nope
}

// ProvideCtor registers the constructor of the target type
//
//	Valid constructors include:
//	- func([more args,]) T
//	- func([more args,]) (T, ~error)
//
// variadic arg will be ignored
func ProvideCtor(fn any, opts ...core.Option) nope {
	opt := defaultOptions
	for _, o := range opts {
		opt = o(opt)
	}
	byctor.ProvideCtor(fn, opt)
	return _nope
}

// ProvideValue registers the value prototype of the target type
func ProvideValue[T any](value T, opts ...core.Option) nope {
	opt := defaultOptions
	for _, o := range opts {
		opt = o(opt)
	}
	byvalue.ProvideValue(value, opt)
	return _nope
}
