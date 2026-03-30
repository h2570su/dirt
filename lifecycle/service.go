package lifecycle

import "context"

type Service struct {
	StartupFn  StartupFunc
	RunFn      RunFunc
	ShutdownFn ShutdownFunc

	instance any
}

func (s Service) IsEmpty() bool { return s.StartupFn == nil && s.RunFn == nil && s.ShutdownFn == nil }

func (s Service) Startup(ctx context.Context) error {
	if s.StartupFn != nil {
		return s.StartupFn(ctx)
	}
	return nil
}

func (s Service) Run(ctx context.Context) error {
	if s.RunFn != nil {
		return s.RunFn(ctx)
	}
	return nil
}

func (s Service) Shutdown(ctx context.Context) error {
	if s.ShutdownFn != nil {
		return s.ShutdownFn(ctx)
	}
	return nil
}
