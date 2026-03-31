package billing

import (
	"context"
	"log/slog"

	"github.com/h2570su/dirt"
	"github.com/h2570su/dirt/lifecycle/example/metrics"
	"github.com/h2570su/dirt/lifecycle/example/notifier"
	"github.com/h2570su/dirt/lifecycle/example/repository"
)

var (
	_ = dirt.ProvideValue(&Config{})
	_ = dirt.ProvideStruct[*Service]()
)

type Config struct {
	Currency string
	TaxRate  float64
}

type Service struct {
	dirt.Injectable

	config   *Config             `dirt:""`
	repo     *repository.Service `dirt:""`
	notifier *notifier.Service   `dirt:""`
	metrics  *metrics.Service    `dirt:""`
	logger   *slog.Logger        `dirt:""`

	invoiceByUser map[string]float64
}

func (s *Service) PostInject() error {
	s.invoiceByUser = make(map[string]float64)
	return nil
}

func (s *Service) Startup(context.Context) error {
	s.logger.Info("[billing] startup")
	return nil
}

func (s *Service) Shutdown(context.Context) error {
	s.logger.Info("[billing] shutdown")
	return nil
}

func (s *Service) Quote(userID string, base float64) float64 {
	total := base * (1 + s.config.TaxRate)
	s.invoiceByUser[userID] = total
	s.metrics.Inc("billing.quote")
	s.notifier.Send("quote generated for " + s.repo.FindUserName(userID))
	return total
}
