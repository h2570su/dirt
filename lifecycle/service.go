package lifecycle

import "context"

// Service represents a service in the lifecycle, which can have startup, run, and shutdown logic.
type Service struct {
	StartupFn  StartupFunc
	RunFn      RunFunc
	ShutdownFn ShutdownFunc

	instance any
}

// IsEmpty returns whether the service has no startup, run, and shutdown logic.
func (s Service) IsEmpty() bool { return s.StartupFn == nil && s.RunFn == nil && s.ShutdownFn == nil }

// Startup wraps the StartupFn of the service, returns nil if StartupFn is nil.
func (s Service) Startup(ctx context.Context) error {
	if s.StartupFn != nil {
		return s.StartupFn(ctx)
	}
	return nil
}

// Run wraps the RunFn of the service, returns nil if RunFn is nil.
func (s Service) Run(ctx context.Context) error {
	if s.RunFn != nil {
		return s.RunFn(ctx)
	}
	return nil
}

// Shutdown wraps the ShutdownFn of the service, returns nil if ShutdownFn is nil.
func (s Service) Shutdown(ctx context.Context) error {
	if s.ShutdownFn != nil {
		return s.ShutdownFn(ctx)
	}
	return nil
}
