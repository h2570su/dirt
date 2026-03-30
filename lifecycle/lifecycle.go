package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"
)

type Lifecycle struct {
	ToStartup  func(any) (StartupFunc, bool)
	ToRun      func(any) (RunFunc, bool)
	ToShutdown func(any) (ShutdownFunc, bool)
	Logger     *slog.Logger

	ListenSignals   []os.Signal
	StartupTimeout  time.Duration
	ShutdownTimeout time.Duration

	services []Service
}

func DefaultLifecycle() Lifecycle {
	return Lifecycle{
		ToStartup:  DefaultToStartupFunc,
		ToRun:      DefaultToRunFunc,
		ToShutdown: DefaultToShutdownFunc,
		Logger:     slog.Default(),

		ListenSignals: []os.Signal{syscall.SIGINT, syscall.SIGTERM},
	}
}

func (l *Lifecycle) log(msg string, attrs ...slog.Attr) {
	if l.Logger == nil {
		return
	}
	var args []any
	args = append(args, slog.String("package", "lifecycle"))
	for _, attr := range attrs {
		args = append(args, attr)
	}
	l.Logger.Debug(msg, args...)
}

func (l *Lifecycle) TryAdd(v any) {
	var service Service
	if s, ok := l.ToStartup(v); ok {
		service.StartupFn = s
	}
	if s, ok := l.ToRun(v); ok {
		service.RunFn = s
	}
	if s, ok := l.ToShutdown(v); ok {
		service.ShutdownFn = s
	}
	if !service.IsEmpty() {
		service.instance = v
		l.services = append(l.services, service)
		l.log("service added", slog.String("service-type", fmt.Sprintf("%T", v)))
	}
}

func (l *Lifecycle) Main(ctx context.Context) error {
	rootCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, l.ListenSignals...)
	go func() {
		<-sigCh
		cancel()
	}()

	l.log("lifecycle started")
	err := l.startup(rootCtx)
	if err != nil {
		return err
	}

	runErrCollectDone := make(chan struct{})
	runErrCh := make(chan error, 1)
	go func() {
		runErrCh <- l.run(rootCtx, runErrCollectDone)
	}()

	<-rootCtx.Done()
	l.log("lifecycle stopping, waiting for services to stop")

	shutdownErr := l.shutdown(sigCh)
	close(runErrCollectDone)
	runErr := <-runErrCh

	var totalErrs []error
	if runErr != nil {
		totalErrs = append(totalErrs, fmt.Errorf("run error: %w", runErr))
	}
	if shutdownErr != nil {
		totalErrs = append(totalErrs, fmt.Errorf("shutdown error: %w", shutdownErr))
	}

	return errors.Join(totalErrs...)
}

func (l *Lifecycle) shutdown(sigCh chan os.Signal) error {
	shutdownCtx, shutDownCancel := context.WithCancel(context.TODO())
	defer shutDownCancel()
	go func() {
		<-sigCh
		shutDownCancel()
	}()
	if l.ShutdownTimeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), l.ShutdownTimeout)
		defer cancel()
		shutdownCtx = ctx
	}
	var shutdownErrs []error
	for _, s := range slices.Backward(l.services) {
		if err := s.Shutdown(shutdownCtx); err != nil {
			l.log("service shutdown error", slog.String("service-type", fmt.Sprintf("%T", s.instance)), slog.Any("error", err))
			shutdownErrs = append(shutdownErrs, err)
		}
	}
	return errors.Join(shutdownErrs...)
}

func (l *Lifecycle) run(runCtx context.Context, collectDone <-chan struct{}) error {
	var runErrs []error
	runErrCh := make(chan error, len(l.services))
	for _, s := range l.services {
		go func(s Service) {
			if err := s.Run(runCtx); err != nil {
				l.log("service run error", slog.String("service-type", fmt.Sprintf("%T", s.instance)), slog.Any("error", err))
				runErrCh <- err
			}
			runErrCh <- nil
		}(s)
	}
	for range len(l.services) {
		select {
		case <-collectDone:
			return errors.Join(runErrs...)
		case err := <-runErrCh:
			if err != nil {
				runErrs = append(runErrs, err)
			}
		}
	}
	return errors.Join(runErrs...)
}

func (l *Lifecycle) startup(startupCtx context.Context) error {
	if l.StartupTimeout > 0 {
		ctx, cancel := context.WithTimeout(startupCtx, l.StartupTimeout)
		defer cancel()
		startupCtx = ctx
	}

	for _, s := range l.services {
		if err := s.Startup(startupCtx); err != nil {
			l.log("service startup error", slog.String("service-type", fmt.Sprintf("%T", s.instance)), slog.Any("error", err))
			return err
		}
	}
	return nil
}
