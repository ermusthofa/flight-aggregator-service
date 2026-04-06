package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ermusthofa/flight-aggregator-service/internal/cache"
	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
	"github.com/ermusthofa/flight-aggregator-service/internal/pkg"
	"github.com/ermusthofa/flight-aggregator-service/internal/provider"
)

type Aggregator struct {
	providers []provider.Provider
	timeout   time.Duration
	cache     *cache.MemoryCache
}

func NewAggregator(providers []provider.Provider, cache *cache.MemoryCache) *Aggregator {
	return &Aggregator{
		providers: providers,
		timeout:   500 * time.Millisecond,
		cache:     cache,
	}
}

type providerResult struct {
	flights []domain.Flight
	err     error
	name    string
}

type Metadata struct {
	TotalResults       int   `json:"total_results"`
	ProvidersQueried   int   `json:"providers_queried"`
	ProvidersSucceeded int   `json:"providers_succeeded"`
	ProvidersFailed    int   `json:"providers_failed"`
	SearchTimeMs       int64 `json:"search_time_ms"`
	CacheHit           bool  `json:"cache_hit"`
}

func (a *Aggregator) Search(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, Metadata) {
	start := time.Now()

	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	key := buildCacheKey(req)

	if data, ok := a.cache.Get(key); ok {
		meta := Metadata{
			ProvidersQueried: len(a.providers),
			CacheHit:         true,
		}

		// apply user filters
		data = FilterFlights(data, req)
		ScoreFlights(data)
		SortFlights(data, req.SortBy)

		meta.TotalResults = len(data)
		meta.SearchTimeMs = time.Since(start).Milliseconds()

		return data, meta
	}

	var wg sync.WaitGroup
	ch := make(chan providerResult, len(a.providers))

	for _, p := range a.providers {
		wg.Add(1)

		go func(p provider.Provider) {
			defer wg.Done()

			pkg.Info("Calling provider: %s", p.Name())

			flights, err := p.Search(ctx, req)

			if err != nil {
				pkg.Error("Provider %s failed: %v", p.Name(), err)
			}

			ch <- providerResult{
				flights: flights,
				err:     err,
				name:    p.Name(),
			}
		}(p)
	}

	wg.Wait()
	close(ch)

	allFlights := make([]domain.Flight, 0)
	meta := Metadata{
		ProvidersQueried: len(a.providers),
	}

	for res := range ch {
		if res.err != nil {
			meta.ProvidersFailed++
			continue
		}

		meta.ProvidersSucceeded++
		allFlights = append(allFlights, res.flights...)
	}

	// cache raw data
	if meta.ProvidersFailed == 0 {
		a.cache.Set(key, allFlights)
	}

	// filter, score, and sort
	allFlights = FilterFlights(allFlights, req)
	ScoreFlights(allFlights)
	SortFlights(allFlights, req.SortBy)

	meta.TotalResults = len(allFlights)
	meta.SearchTimeMs = time.Since(start).Milliseconds()

	return allFlights, meta
}

func buildCacheKey(req domain.SearchRequest) string {
	return fmt.Sprintf("%s:%s:%s:%s",
		req.Origin,
		req.Destination,
		req.DepartureDate,
		req.CabinClass,
	)
}
