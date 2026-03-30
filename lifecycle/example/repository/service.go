package repository

import (
	"log/slog"

	"git.ttech.cc/astaroth/dirt"
	"git.ttech.cc/astaroth/dirt/lifecycle/example/cache"
)

var (
	_ = dirt.ProvideValue(&Config{})
	_ = dirt.ProvideStruct[*Service]()
)

type Config struct {
	Table string
}

type Service struct {
	dirt.Injectable

	config *Config        `dirt:""`
	cache  *cache.Service `dirt:""`
	logger *slog.Logger   `dirt:""`
}

func (s *Service) FindUserName(id string) string {
	if name, ok := s.cache.Get("user:" + id); ok {
		return name
	}
	s.logger.Debug("[repository] cache miss for user " + id)
	name := "demo-user-" + id + "in table " + s.config.Table // just for demo, usually it should query from DB.
	s.cache.Set("user:"+id, name)
	return name
}
