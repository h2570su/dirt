package lifecycle

import "context"

type ShutdownFunc func(context.Context) error

func DefaultToShutdownFunc(s any) (ShutdownFunc, bool) {
	switch v := s.(type) {
	case interface{ Shutdown(context.Context) error }:
		return v.Shutdown, true
	case interface{ Shutdown() }:
		return func(context.Context) error { v.Shutdown(); return nil }, true
	case interface{ Shutdown() error }:
		return func(context.Context) error { return v.Shutdown() }, true
	case interface{ Shutdown(context.Context) }:
		return func(ctx context.Context) error { v.Shutdown(ctx); return nil }, true
	case interface{ ShutdownWithContext(context.Context) }:
		return func(ctx context.Context) error { v.ShutdownWithContext(ctx); return nil }, true
	case interface{ ShutdownWithContext(context.Context) error }:
		return func(ctx context.Context) error { return v.ShutdownWithContext(ctx) }, true
	default:
		return nil, false
	}
}
