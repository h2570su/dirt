package auditworker

import (
	"context"
	"log/slog"
	"time"

	"github.com/h2570su/dirt"
	"github.com/h2570su/dirt/lifecycle/example/billing"
)

var (
	_ = dirt.ProvideValue(&Config{})
	_ = dirt.ProvideStruct[*Service]()
)

type Config struct {
	Interval time.Duration
}

// Service is a background worker; Run blocks until context cancellation.
type Service struct {
	dirt.Injectable

	config  *Config          `dirt:""`
	billing *billing.Service `dirt:""`
	logger  *slog.Logger     `dirt:""`
}

func (s *Service) Startup(context.Context) error {
	s.logger.Info("[audit-worker] startup")
	return nil
}

func (s *Service) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("[audit-worker] stop")
			return nil
		case <-ticker.C:
			s.billing.Quote("worker", 100)
			s.logger.Debug("[audit-worker] periodic audit tick")
		}
	}
}

func (s *Service) Shutdown(context.Context) error {
	s.logger.Info("[audit-worker] shutdown")
	return nil
}
