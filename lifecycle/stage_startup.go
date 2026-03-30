package lifecycle

import "context"

type StartupFunc func(context.Context) error

func DefaultToStartupFunc(s any) (StartupFunc, bool) {
	switch v := s.(type) {
	case interface{ Startup(context.Context) error }:
		return v.Startup, true
	case interface{ Startup() }:
		return func(context.Context) error { v.Startup(); return nil }, true
	case interface{ Startup() error }:
		return func(context.Context) error { return v.Startup() }, true
	case interface{ Startup(context.Context) }:
		return func(ctx context.Context) error { v.Startup(ctx); return nil }, true
	case interface{ StartupWithContext(context.Context) }:
		return func(ctx context.Context) error { v.StartupWithContext(ctx); return nil }, true
	case interface{ StartupWithContext(context.Context) error }:
		return func(ctx context.Context) error { return v.StartupWithContext(ctx) }, true
	default:
		return nil, false
	}
}
