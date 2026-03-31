package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/h2570su/dirt"
	"github.com/h2570su/dirt/lifecycle"
	"github.com/h2570su/dirt/lifecycle/example/api"
	"github.com/h2570su/dirt/lifecycle/example/auditworker"
	"github.com/h2570su/dirt/lifecycle/example/billing"
	"github.com/h2570su/dirt/lifecycle/example/cache"
	"github.com/h2570su/dirt/lifecycle/example/metrics"
	"github.com/h2570su/dirt/lifecycle/example/notifier"
	"github.com/h2570su/dirt/lifecycle/example/repository"
	"github.com/h2570su/dirt/lifecycle/example/syncworker"
)

var (
	_ = dirt.ProvideValue(&loggerConfig{})
	_ = dirt.ProvideCtor(newSlogLogger)
)

type loggerConfig struct {
	ServiceName string
	Level       slog.Level
}

func newSlogLogger(cfg *loggerConfig) *slog.Logger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.Level,
	})
	return slog.New(handler).With("service-name", cfg.ServiceName)
}

type worldConfig struct {
	dirt.Injectable

	basic struct {
		dirt.Subclass // dirt supports grouping fields to be injected together.

		logger *loggerConfig `dirt:""`
	}
	metrics     *metrics.Config     `dirt:""`
	cache       *cache.Config       `dirt:""`
	repository  *repository.Config  `dirt:""`
	notifier    *notifier.Config    `dirt:""`
	billing     *billing.Config     `dirt:""`
	api         *api.Config         `dirt:""`
	auditWorker *auditworker.Config `dirt:""`
	syncWorker  *syncworker.Config  `dirt:""`
}

var _ = dirt.ProvideStruct[*worldConfig]()

// components is an app-level bundle to show how a root type can gather some services.
type components struct {
	dirt.Injectable

	logger *slog.Logger `dirt:""`
	api    *api.Service `dirt:""`

	sync  *syncworker.Service  `dirt:""` //nolint:unused
	audit *auditworker.Service `dirt:""` //nolint:unused
}

var _ = dirt.ProvideStruct[*components]()

func main() {
	cfg, err := dirt.Invoke[*worldConfig]()
	if err != nil {
		log.Fatalf("invoke worldConfig failed: %v", err)
	}
	*cfg.basic.logger = loggerConfig{ServiceName: "example-app", Level: slog.LevelDebug}
	*cfg.metrics = metrics.Config{Namespace: "example", FlushEvery: 5 * time.Second}
	*cfg.cache = cache.Config{MaxEntries: 1000}
	*cfg.repository = repository.Config{Table: "users"}
	*cfg.notifier = notifier.Config{Channel: "email", Retry: 3}
	*cfg.billing = billing.Config{Currency: "USD", TaxRate: 0.05}
	*cfg.api = api.Config{Addr: ":8080", ReadTimeout: 3 * time.Second}
	*cfg.auditWorker = auditworker.Config{Interval: 1 * time.Second}
	*cfg.syncWorker = syncworker.Config{BatchSize: 10, Interval: 1500 * time.Millisecond}

	// 1) Build one root instance first, which creates a full dependency chain.
	app, err := dirt.Invoke[*components]() // default scope is global.
	if err != nil {
		log.Fatalf("invoke components failed: %v", err)
	}
	app.logger.Info("example app wired", "api-addr", app.api.Addr(), "logger-level", cfg.basic.logger.Level)

	// Explicitly use the API demo flow so its dependency path is exercised.
	app.api.DemoRequest("1001", 88)

	// 2) Then hand all invoked instances to lifecycle.
	lc := lifecycle.DefaultLifecycle()
	lc.Logger = app.logger
	lc.StartupTimeout = 3 * time.Second
	lc.ShutdownTimeout = 3 * time.Second
	if err := lc.DirtAddAllFor(dirt.GlobalScope(), app); err != nil {
		log.Fatalf("DirtAddAll failed: %v", err)
	}

	// Keep demo short: run for a few seconds then shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	if err := lc.Main(ctx); err != nil {
		log.Printf("lifecycle main failed: %v", err)
	}
}
