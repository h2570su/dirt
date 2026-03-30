package lifecycle

import "context"

type RunFunc func(context.Context) error

func DefaultToRunFunc(s any) (RunFunc, bool) {
	switch v := s.(type) {
	case interface{ Run(context.Context) error }:
		return v.Run, true
	case interface{ Run() }:
		return func(context.Context) error { v.Run(); return nil }, true
	case interface{ Run() error }:
		return func(context.Context) error { return v.Run() }, true
	case interface{ Run(context.Context) }:
		return func(ctx context.Context) error { v.Run(ctx); return nil }, true
	case interface{ RunWithContext(context.Context) }:
		return func(ctx context.Context) error { v.RunWithContext(ctx); return nil }, true
	case interface{ RunWithContext(context.Context) error }:
		return func(ctx context.Context) error { return v.RunWithContext(ctx) }, true
	default:
		return nil, false
	}
}
