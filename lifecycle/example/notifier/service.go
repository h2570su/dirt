package notifier

import (
	"log/slog"

	"github.com/h2570su/dirt"
)

var (
	_ = dirt.ProvideValue(&Config{})
	_ = dirt.ProvideCtor(NewService)
)

type Config struct {
	Channel string
	Retry   int
}

type Service struct {
	config *Config
	logger *slog.Logger
}

func NewService(cfg *Config, lg *slog.Logger) *Service {
	return &Service{
		config: cfg,
		logger: lg,
	}
}

func (s *Service) Send(msg string) {
	s.logger.Debug("[notifier] send: " + msg)
}
