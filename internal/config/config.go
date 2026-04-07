package config

import (
	"net/http"
	"time"

	httphandler "github.com/ermusthofa/flight-aggregator-service/internal/handler/http"
	"github.com/ermusthofa/flight-aggregator-service/internal/partner"
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
	// Repository layer
	cacheRepo := repository.NewMemoryCache()

	// Partner adapters
	providers := []partner.Provider{
		partner.NewRetryable(partner.NewAirAsiaProvider(), 2, 50*time.Millisecond),
		partner.NewRetryable(partner.NewLionProvider(), 1, 50*time.Millisecond),
		partner.NewRetryable(partner.NewBatikProvider(), 1, 50*time.Millisecond),
		partner.NewRetryable(partner.NewGarudaProvider(), 1, 50*time.Millisecond),
	}

	// Use case
	searchUC := usecase.NewSearchFlightsUsecase(cacheRepo, providers, &usecase.UsecaseConfig{
		ProviderTimeout: cfg.ProviderTimeout,
		CacheTTL:        cfg.CacheTTL,
	})

	// Handler
	handler := httphandler.NewHandler(searchUC)

	return &App{Handler: handler}, nil
}

type App struct {
	Handler http.Handler
}
