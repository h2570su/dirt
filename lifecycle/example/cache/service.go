package cache

import (
	"context"
	"log/slog"

	"git.ttech.cc/astaroth/dirt"
)

var (
	_ = dirt.ProvideValue(&Config{})
	_ = dirt.ProvideStruct[*Service]()
)

type Config struct {
	MaxEntries int
}

type Service struct {
	dirt.Injectable

	config *Config      `dirt:""`
	logger *slog.Logger `dirt:""`

	store map[string]string
}

// PostInject demonstrates post-wiring initialization for non-injectable fields.
func (s *Service) PostInject() error {
	s.store = make(map[string]string, s.config.MaxEntries)
	s.store["__warmup__"] = "ok"
	s.logger.Debug("[cache] post inject initialized map store")
	return nil
}

func (s *Service) Shutdown(context.Context) error {
	s.logger.Info("[cache] shutdown and it doesn't need a startup") // just for demo, usually cache doesn't need to do anything on shutdown.
	return nil
}

func (s *Service) Set(k, v string) { s.store[k] = v }

func (s *Service) Get(k string) (string, bool) {
	v, ok := s.store[k]
	return v, ok
}
