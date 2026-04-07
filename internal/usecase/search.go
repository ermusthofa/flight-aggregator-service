package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
	"github.com/ermusthofa/flight-aggregator-service/internal/partner"
	"github.com/ermusthofa/flight-aggregator-service/internal/pkg"
	"github.com/ermusthofa/flight-aggregator-service/internal/repository"
)

type UsecaseConfig struct {
	ProviderTimeout time.Duration
	CacheTTL        time.Duration
}

type SearchFlightsUsecase struct {
	cache     repository.Cache
	providers []partner.Provider
	filter    *FilterEngine
	sorter    *SorterEngine
	scorer    *ScoringEngine
	config    *UsecaseConfig
}

func NewSearchFlightsUsecase(
	cache repository.Cache,
	providers []partner.Provider,
	cfg *UsecaseConfig,
) *SearchFlightsUsecase {
	return &SearchFlightsUsecase{
		cache:     cache,
		providers: providers,
		config:    cfg,
		filter:    NewFilterEngine(),
		sorter:    NewSorterEngine(),
		scorer:    NewScoringEngine(),
	}
}

func (uc *SearchFlightsUsecase) Execute(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, domain.Metadata, error) {
	start := time.Now()
	metadata := domain.Metadata{}

	// Normalize & validate request
	req.Normalize()
	if err := req.Validate(); err != nil {
		return nil, metadata, fmt.Errorf("invalid request: %w", err)
	}

	// Check cache
	cacheKey := uc.cacheKey(req)
	if cached, ok := uc.cache.Get(cacheKey); ok {
		metadata.CacheHit = true
		metadata.SearchTimeMs = time.Since(start).Milliseconds()
		// Apply filters/sort on cached results
		filtered := uc.filter.Apply(cached, req)
		uc.scorer.Calculate(filtered)
		uc.sorter.Sort(filtered, req.SortBy)
		metadata.TotalResults = len(filtered)
		return filtered, metadata, nil
	}

	// Fetch from all providers concurrently with timeout
	type result struct {
		flights []domain.Flight
		err     error
		name    string
	}

	ctx, cancel := context.WithTimeout(ctx, uc.config.ProviderTimeout)
	defer cancel()

	results := make(chan result, len(uc.providers))
	var wg sync.WaitGroup

	for _, p := range uc.providers {
		wg.Add(1)
		go func(prov partner.Provider) {
			defer wg.Done()
			flights, err := prov.Search(ctx, req)
			results <- result{flights: flights, err: err, name: p.Name()}
		}(p)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect & aggregate
	var allFlights []domain.Flight
	metadata.ProvidersQueried = len(uc.providers)

	for res := range results {
		if res.err != nil {
			metadata.ProvidersFailed++
			pkg.Error(ctx, "provider %s failed to fetch: %v", res.name, res.err)
			continue
		}
		metadata.ProvidersSucceeded++
		allFlights = append(allFlights, res.flights...)
	}

	// Store raw in cache
	uc.cache.Set(cacheKey, allFlights, uc.config.CacheTTL)

	// Post-processing: scoring, filtering, sorting
	filtered := uc.filter.Apply(allFlights, req)
	uc.scorer.Calculate(filtered)
	uc.sorter.Sort(filtered, req.SortBy)

	metadata.TotalResults = len(filtered)
	metadata.SearchTimeMs = time.Since(start).Milliseconds()
	return filtered, metadata, nil
}

func (uc *SearchFlightsUsecase) cacheKey(req domain.SearchRequest) string {
	return fmt.Sprintf("search:%s:%s:%s:%d:%s", req.Origin, req.Destination, req.DepartureDate, req.Passengers, req.CabinClass)
}
