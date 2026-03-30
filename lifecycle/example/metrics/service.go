package metrics

import (
	"context"
	"log/slog"
	"time"

	"git.ttech.cc/astaroth/dirt"
)

var (
	_ = dirt.ProvideValue(&Config{})
	_ = dirt.ProvideStruct[*Service]()
)

type Config struct {
	Namespace  string
	FlushEvery time.Duration
}

type Service struct {
	dirt.Injectable

	config *Config      `dirt:""`
	logger *slog.Logger `dirt:""`

	counters map[string]int
}

func (s *Service) PostInject() error {
	s.counters = make(map[string]int)
	return nil
}

func (s *Service) Startup(context.Context) error {
	s.logger.Info("[metrics] startup", slog.String("namespace", s.config.Namespace), slog.Duration("flush-every", s.config.FlushEvery))
	return nil
}

func (s *Service) Shutdown(context.Context) error {
	s.logger.Info("[metrics] shutdown")
	return nil
}

func (s *Service) Inc(key string) {
	s.counters[key]++
}
