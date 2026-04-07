package config

import (
	"net/http"
	"time"

	httphandler "github.com/ermusthofa/flight-aggregator-service/internal/handler/http"
	"github.com/ermusthofa/flight-aggregator-service/internal/partner"
	"github.com/ermusthofa/flight-aggregator-service/internal/pkg"
	"github.com/ermusthofa/flight-aggregator-service/internal/repository"
	"github.com/ermusthofa/flight-aggregator-service/internal/usecase"
)

type Config struct {
	CacheTTL        time.Duration
	ProviderTimeout time.Duration
}

// TODO: load from env/config file
func Load() *Config {
	return &Config{
		CacheTTL:        5 * time.Second,
		ProviderTimeout: 2 * time.Second,
	}
}

// InitializeApp wires everything
func InitializeApp(cfg *Config) (*App, error) {
	// Create logger
	logger, err := pkg.NewZapLogger()
	if err != nil {
		return nil, err
	}

	// Repository layer
	cacheRepo := repository.NewMemoryCache()

	// Partner adapters
	providers := []partner.Provider{
		partner.NewRetryable(partner.NewAirAsiaProvider(logger), 2, 50*time.Millisecond, logger),
		partner.NewRetryable(partner.NewLionProvider(logger), 1, 50*time.Millisecond, logger),
		partner.NewRetryable(partner.NewBatikProvider(logger), 1, 50*time.Millisecond, logger),
		partner.NewRetryable(partner.NewGarudaProvider(logger), 1, 50*time.Millisecond, logger),
	}

	// Use case
	searchUC := usecase.NewSearchFlightsUsecase(cacheRepo, providers, &usecase.UsecaseConfig{
		ProviderTimeout: cfg.ProviderTimeout,
		CacheTTL:        cfg.CacheTTL,
	}, logger)

	// Handler
	handler := httphandler.NewHandler(searchUC, logger)

	return &App{Handler: handler}, nil
}

type App struct {
	Handler http.Handler
}
