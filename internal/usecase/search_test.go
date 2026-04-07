package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
	"github.com/ermusthofa/flight-aggregator-service/internal/partner"
)

// mockCache implements repository.Cache
type mockCache struct {
	store map[string][]domain.Flight
	ttl   map[string]time.Duration
}

func newMockCache() *mockCache {
	return &mockCache{
		store: make(map[string][]domain.Flight),
		ttl:   make(map[string]time.Duration),
	}
}

func (m *mockCache) Get(key string) ([]domain.Flight, bool) {
	flights, ok := m.store[key]
	return flights, ok
}

func (m *mockCache) Set(key string, flights []domain.Flight, ttl time.Duration) {
	m.store[key] = flights
	m.ttl[key] = ttl
}

// mockProvider implements partner.Provider
type mockProvider struct {
	name    string
	flights []domain.Flight
	err     error
	delay   time.Duration // simulated delay before returning
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) Search(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.flights, nil
}

// mockProvider implements pkg.Logger
type mockLogger struct{}

func (m *mockLogger) Info(ctx context.Context, format string, v ...interface{})  {}
func (m *mockLogger) Error(ctx context.Context, format string, v ...interface{}) {}
func (m *mockLogger) Warn(ctx context.Context, format string, v ...interface{})  {}

// helper to create a flight with given origin/destination/date/cabin/passengers
// (passengers used for seat availability, but we'll just set seats high enough)
func makeFlight(id, origin, dest, departDate, cabin string, seats int) domain.Flight {
	depTime, _ := time.Parse("2006-01-02T15:04:05Z07:00", departDate+"T12:00:00+07:00")
	return domain.Flight{
		ID:             id,
		Provider:       "mock",
		FlightNumber:   id,
		Departure:      domain.Location{Airport: origin, Datetime: depTime},
		Arrival:        domain.Location{Airport: dest, Datetime: depTime.Add(2 * time.Hour)},
		AvailableSeats: seats,
		CabinClass:     cabin,
		Price:          domain.Price{Amount: 1000000, Currency: "IDR"},
	}
}

func TestSearchFlightsUsecase_InvalidRequest(t *testing.T) {
	cache := newMockCache()
	providers := []partner.Provider{}
	logger := &mockLogger{}
	cfg := &UsecaseConfig{ProviderTimeout: 1 * time.Second, CacheTTL: 5 * time.Minute}
	uc := NewSearchFlightsUsecase(cache, providers, cfg, logger)

	req := domain.SearchRequest{
		Origin:      "", // invalid
		Destination: "DPS",
	}
	_, _, err := uc.Execute(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for invalid request, got nil")
	}
}

func TestSearchFlightsUsecase_CacheHit(t *testing.T) {
	cache := newMockCache()
	providers := []partner.Provider{} // not used on cache hit
	logger := &mockLogger{}
	cfg := &UsecaseConfig{ProviderTimeout: 1 * time.Second, CacheTTL: 5 * time.Minute}
	uc := NewSearchFlightsUsecase(cache, providers, cfg, logger)

	// Pre-populate cache
	cachedFlights := []domain.Flight{
		makeFlight("GA100", "CGK", "DPS", "2025-12-15", "economy", 10),
		makeFlight("GA200", "CGK", "DPS", "2025-12-15", "economy", 5),
	}
	req := domain.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    2,
		CabinClass:    "economy",
		SortBy:        "price",
	}
	cacheKey := uc.cacheKey(req)
	cache.Set(cacheKey, cachedFlights, cfg.CacheTTL)

	flights, meta, err := uc.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !meta.CacheHit {
		t.Error("expected cache hit, but cache_hit false")
	}
	if len(flights) != 2 {
		t.Errorf("expected 2 flight after filter, got %d", len(flights))
	}
	if flights[0].ID != "GA100" {
		t.Errorf("expected GA100, got %s", flights[0].ID)
	}
}

func TestSearchFlightsUsecase_CacheMiss_Success(t *testing.T) {
	cache := newMockCache()
	provider1 := &mockProvider{
		name:    "ProviderA",
		flights: []domain.Flight{makeFlight("A1", "CGK", "DPS", "2025-12-15", "economy", 5)},
	}
	provider2 := &mockProvider{
		name:    "ProviderB",
		flights: []domain.Flight{makeFlight("B1", "CGK", "DPS", "2025-12-15", "economy", 10)},
	}
	providers := []partner.Provider{provider1, provider2}
	logger := &mockLogger{}
	cfg := &UsecaseConfig{ProviderTimeout: 2 * time.Second, CacheTTL: 5 * time.Minute}
	uc := NewSearchFlightsUsecase(cache, providers, cfg, logger)

	req := domain.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    3,
		CabinClass:    "economy",
	}
	flights, meta, err := uc.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.CacheHit {
		t.Error("expected cache miss, got cache hit")
	}
	if meta.ProvidersQueried != 2 {
		t.Errorf("expected 2 providers queried, got %d", meta.ProvidersQueried)
	}
	if meta.ProvidersSucceeded != 2 {
		t.Errorf("expected 2 providers succeeded, got %d", meta.ProvidersSucceeded)
	}
	if meta.ProvidersFailed != 0 {
		t.Errorf("expected 0 failed, got %d", meta.ProvidersFailed)
	}
	// Both flights should be present (both have enough seats: 5 and 10 >= 3)
	if len(flights) != 2 {
		t.Errorf("expected 2 flights, got %d", len(flights))
	}
	// Cache should be populated
	if _, ok := cache.Get(uc.cacheKey(req)); !ok {
		t.Error("cache should have been set")
	}
}

func TestSearchFlightsUsecase_PartialProviderFailure(t *testing.T) {
	cache := newMockCache()
	providerGood := &mockProvider{
		name:    "Good",
		flights: []domain.Flight{makeFlight("G1", "CGK", "DPS", "2025-12-15", "economy", 5)},
	}
	providerBad := &mockProvider{
		name: "Bad",
		err:  errors.New("network error"),
	}
	providers := []partner.Provider{providerGood, providerBad}
	logger := &mockLogger{}
	cfg := &UsecaseConfig{ProviderTimeout: 1 * time.Second, CacheTTL: 5 * time.Minute}
	uc := NewSearchFlightsUsecase(cache, providers, cfg, logger)

	req := domain.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}
	flights, meta, err := uc.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.ProvidersQueried != 2 {
		t.Errorf("expected 2 providers queried, got %d", meta.ProvidersQueried)
	}
	if meta.ProvidersSucceeded != 1 {
		t.Errorf("expected 1 succeeded, got %d", meta.ProvidersSucceeded)
	}
	if meta.ProvidersFailed != 1 {
		t.Errorf("expected 1 failed, got %d", meta.ProvidersFailed)
	}
	if len(flights) != 1 || flights[0].ID != "G1" {
		t.Errorf("expected only G1, got %+v", flights)
	}
}

func TestSearchFlightsUsecase_ProviderTimeout(t *testing.T) {
	cache := newMockCache()
	slowProvider := &mockProvider{
		name:    "Slow",
		delay:   500 * time.Millisecond,
		flights: []domain.Flight{makeFlight("S1", "CGK", "DPS", "2025-12-15", "economy", 5)},
	}
	fastProvider := &mockProvider{
		name:    "Fast",
		flights: []domain.Flight{makeFlight("F1", "CGK", "DPS", "2025-12-15", "economy", 5)},
	}
	providers := []partner.Provider{slowProvider, fastProvider}
	logger := &mockLogger{}
	cfg := &UsecaseConfig{ProviderTimeout: 200 * time.Millisecond} // shorter than slow provider
	uc := NewSearchFlightsUsecase(cache, providers, cfg, logger)

	req := domain.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
	}
	flights, meta, err := uc.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Fast provider succeeds, slow provider times out (context cancelled) -> returns error, counted as failed
	if meta.ProvidersSucceeded != 1 {
		t.Errorf("expected 1 succeeded, got %d", meta.ProvidersSucceeded)
	}
	if meta.ProvidersFailed != 1 {
		t.Errorf("expected 1 failed, got %d", meta.ProvidersFailed)
	}
	if len(flights) != 1 || flights[0].ID != "F1" {
		t.Errorf("expected only F1, got %+v", flights)
	}
}

func TestSearchFlightsUsecase_FilterAndSort(t *testing.T) {
	cache := newMockCache()
	provider := &mockProvider{
		name: "Provider",
		flights: []domain.Flight{
			makeFlight("F1", "CGK", "DPS", "2025-12-15", "economy", 10),
			makeFlight("F2", "CGK", "DPS", "2025-12-15", "economy", 5),
			makeFlight("F3", "CGK", "DPS", "2025-12-15", "economy", 8),
		},
	}
	providers := []partner.Provider{provider}
	logger := &mockLogger{}
	cfg := &UsecaseConfig{ProviderTimeout: 1 * time.Second, CacheTTL: 5 * time.Minute}
	uc := NewSearchFlightsUsecase(cache, providers, cfg, logger)

	req := domain.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
		SortBy:        "price",
	}
	flights, _, err := uc.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(flights) != 3 {
		t.Errorf("expected 1 flight, got %d", len(flights))
	}
	if flights[0].ID != "F1" {
		t.Errorf("expected F1, got %s", flights[0].ID)
	}
}
