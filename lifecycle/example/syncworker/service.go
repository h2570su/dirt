package syncworker

import (
	"context"
	"log/slog"
	"time"

	"github.com/h2570su/dirt"
	"github.com/h2570su/dirt/lifecycle/example/metrics"
	"github.com/h2570su/dirt/lifecycle/example/repository"
)

var (
	_ = dirt.ProvideValue(&Config{})
	_ = dirt.ProvideCtor(NewService)
)

type Config struct {
	BatchSize int
	Interval  time.Duration
}

type Service struct {
	config  *Config
	repo    *repository.Service
	metrics *metrics.Service
	logger  *slog.Logger
}

func NewService(cfg *Config, repo *repository.Service, m *metrics.Service, lg *slog.Logger) *Service {
	return &Service{
		config:  cfg,
		repo:    repo,
		metrics: m,
		logger:  lg,
	}
}

func (s *Service) Startup(context.Context) error {
	s.logger.Info("[sync-worker] startup")
	return nil
}

func (s *Service) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("[sync-worker] stop")
			return nil
		case <-ticker.C:
			_ = s.repo.FindUserName("42")
			s.metrics.Inc("sync.tick")
			s.logger.Debug("[sync-worker] periodic sync tick")
		}
	}
}

func (s *Service) Shutdown(context.Context) error {
	s.logger.Info("[sync-worker] shutdown")
	return nil
}
