package api

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"git.ttech.cc/astaroth/dirt"
	"git.ttech.cc/astaroth/dirt/lifecycle/example/billing"
	"git.ttech.cc/astaroth/dirt/lifecycle/example/metrics"
	"git.ttech.cc/astaroth/dirt/lifecycle/example/repository"
)

var (
	_ = dirt.ProvideValue(&Config{})
	_ = dirt.ProvideStruct[*Service]()
)

type Config struct {
	Addr        string
	ReadTimeout time.Duration
}

type Service struct {
	dirt.Injectable

	config  *Config             `dirt:""`
	billing *billing.Service    `dirt:""`
	repo    *repository.Service `dirt:""`
	metrics *metrics.Service    `dirt:""`
	logger  *slog.Logger        `dirt:""`
}

func (s *Service) Startup(context.Context) error {
	s.logger.Info("[api] startup, and it doesn't need a shutdown")
	return nil
}

func (s *Service) DemoRequest(userID string, amount float64) {
	s.metrics.Inc("api.request")
	total := s.billing.Quote(userID, amount)
	s.logger.Debug("[api] quote result=" + strconv.FormatFloat(total, 'f', 2, 64) + " user=" + s.repo.FindUserName(userID))
}

func (s *Service) Addr() string { return s.config.Addr }
