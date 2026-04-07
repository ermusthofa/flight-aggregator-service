package usecase

import (
	"testing"
	"time"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
)

// Helper to create a flight with custom fields
func makeTestFlight(id, airlineCode string, price, stops, durationMinutes int, depHour, depMin, arrHour, arrMin int) domain.Flight {
	dep := time.Date(2025, 1, 1, depHour, depMin, 0, 0, time.UTC)
	arr := time.Date(2025, 1, 1, arrHour, arrMin, 0, 0, time.UTC)
	// Adjust arrival date if crossing midnight? Not needed for filter because minutesOfDay works modulo 24h.
	return domain.Flight{
		ID:           id,
		FlightNumber: id,
		Airline:      domain.Airline{Code: airlineCode},
		Price:        domain.Price{Amount: price, Currency: "IDR"},
		Stops:        stops,
		TotalMinutes: durationMinutes,
		Departure:    domain.Location{Datetime: dep},
		Arrival:      domain.Location{Datetime: arr},
	}
}

func TestFilterEngine_PriceRange(t *testing.T) {
	engine := NewFilterEngine()
	flights := []domain.Flight{
		makeTestFlight("F1", "GA", 500000, 0, 60, 8, 0, 10, 0),
		makeTestFlight("F2", "GA", 1000000, 0, 60, 8, 0, 10, 0),
		makeTestFlight("F3", "GA", 1500000, 0, 60, 8, 0, 10, 0),
	}

	req := domain.SearchRequest{
		MinPrice: 600000,
		MaxPrice: 1200000,
	}
	filtered := engine.Apply(flights, req)
	if len(filtered) != 1 || filtered[0].ID != "F2" {
		t.Errorf("expected only F2 (price 1M), got %+v", filtered)
	}

	// Only min
	req = domain.SearchRequest{MinPrice: 1200000}
	filtered = engine.Apply(flights, req)
	if len(filtered) != 1 || filtered[0].ID != "F3" {
		t.Errorf("expected only F3, got %+v", filtered)
	}

	// Only max
	req = domain.SearchRequest{MaxPrice: 800000}
	filtered = engine.Apply(flights, req)
	if len(filtered) != 1 || filtered[0].ID != "F1" {
		t.Errorf("expected only F1, got %+v", filtered)
	}

	// No price filters
	req = domain.SearchRequest{}
	filtered = engine.Apply(flights, req)
	if len(filtered) != 3 {
		t.Errorf("expected all 3 flights, got %d", len(filtered))
	}
}

func TestFilterEngine_MaxStops(t *testing.T) {
	engine := NewFilterEngine()
	flights := []domain.Flight{
		makeTestFlight("F1", "GA", 100000, 0, 60, 8, 0, 10, 0),
		makeTestFlight("F2", "GA", 100000, 1, 60, 8, 0, 10, 0),
		makeTestFlight("F3", "GA", 100000, 2, 60, 8, 0, 10, 0),
	}

	// MaxStops = 1
	maxStops := 1
	req := domain.SearchRequest{MaxStops: &maxStops}
	filtered := engine.Apply(flights, req)
	if len(filtered) != 2 {
		t.Errorf("expected 2 flights (stops 0,1), got %d", len(filtered))
	}
	// MaxStops = nil -> no filter
	req = domain.SearchRequest{}
	filtered = engine.Apply(flights, req)
	if len(filtered) != 3 {
		t.Errorf("expected all 3, got %d", len(filtered))
	}
}

func TestFilterEngine_DepartureTimeRange(t *testing.T) {
	engine := NewFilterEngine()
	// flights depart at 08:00, 12:00, 18:00
	flights := []domain.Flight{
		makeTestFlight("F1", "GA", 100000, 0, 60, 8, 0, 10, 0),
		makeTestFlight("F2", "GA", 100000, 0, 60, 12, 0, 14, 0),
		makeTestFlight("F3", "GA", 100000, 0, 60, 18, 0, 20, 0),
	}

	req := domain.SearchRequest{
		DepartureTimeFrom: "09:00",
		DepartureTimeTo:   "15:00",
	}
	filtered := engine.Apply(flights, req)
	if len(filtered) != 1 || filtered[0].ID != "F2" {
		t.Errorf("expected only F2 (12:00), got %+v", filtered)
	}

	// Only from
	req = domain.SearchRequest{DepartureTimeFrom: "13:00"}
	filtered = engine.Apply(flights, req)
	if len(filtered) != 1 || filtered[0].ID != "F3" {
		t.Errorf("expected only F3 (18:00), got %+v", filtered)
	}

	// Only to
	req = domain.SearchRequest{DepartureTimeTo: "10:00"}
	filtered = engine.Apply(flights, req)
	if len(filtered) != 1 || filtered[0].ID != "F1" {
		t.Errorf("expected only F1 (08:00), got %+v", filtered)
	}
}

func TestFilterEngine_ArrivalTimeRange(t *testing.T) {
	engine := NewFilterEngine()
	// flights arrive at 10:00, 14:00, 20:00
	flights := []domain.Flight{
		makeTestFlight("F1", "GA", 100000, 0, 60, 8, 0, 10, 0),
		makeTestFlight("F2", "GA", 100000, 0, 60, 12, 0, 14, 0),
		makeTestFlight("F3", "GA", 100000, 0, 60, 18, 0, 20, 0),
	}

	req := domain.SearchRequest{
		ArrivalTimeFrom: "11:00",
		ArrivalTimeTo:   "16:00",
	}
	filtered := engine.Apply(flights, req)
	if len(filtered) != 1 || filtered[0].ID != "F2" {
		t.Errorf("expected only F2 (arrival 14:00), got %+v", filtered)
	}
}

func TestFilterEngine_AirlineCodes(t *testing.T) {
	engine := NewFilterEngine()
	flights := []domain.Flight{
		makeTestFlight("F1", "GA", 100000, 0, 60, 8, 0, 10, 0),
		makeTestFlight("F2", "QZ", 100000, 0, 60, 8, 0, 10, 0),
		makeTestFlight("F3", "JT", 100000, 0, 60, 8, 0, 10, 0),
	}

	req := domain.SearchRequest{Airlines: []string{"GA", "JT"}}
	filtered := engine.Apply(flights, req)
	if len(filtered) != 2 {
		t.Errorf("expected 2 flights (GA, JT), got %d", len(filtered))
	}
	// Empty slice means no filter
	req = domain.SearchRequest{Airlines: []string{}}
	filtered = engine.Apply(flights, req)
	if len(filtered) != 3 {
		t.Errorf("expected all 3, got %d", len(filtered))
	}
}

func TestFilterEngine_MaxDuration(t *testing.T) {
	engine := NewFilterEngine()
	flights := []domain.Flight{
		makeTestFlight("F1", "GA", 100000, 0, 90, 8, 0, 9, 30),
		makeTestFlight("F2", "GA", 100000, 0, 120, 8, 0, 10, 0),
		makeTestFlight("F3", "GA", 100000, 0, 150, 8, 0, 10, 30),
	}

	req := domain.SearchRequest{MaxDuration: 120}
	filtered := engine.Apply(flights, req)
	if len(filtered) != 2 {
		t.Errorf("expected 2 flights (90,120 min), got %d", len(filtered))
	}
}

func TestFilterEngine_CombinedFilters(t *testing.T) {
	engine := NewFilterEngine()
	flights := []domain.Flight{
		makeTestFlight("F1", "GA", 500000, 0, 90, 8, 0, 9, 30),    // price ok, stops ok, dep 08:00, arr 09:30, airline GA, duration 90
		makeTestFlight("F2", "QZ", 1200000, 1, 120, 12, 0, 14, 0), // price too high, stops 1, airline QZ
		makeTestFlight("F3", "GA", 800000, 0, 180, 18, 0, 21, 0),  // price ok, stops ok, but dep 18:00, arr 21:00, duration 180
	}
	maxStops := 0
	req := domain.SearchRequest{
		MinPrice:          400000,
		MaxPrice:          900000,
		MaxStops:          &maxStops,
		DepartureTimeFrom: "07:00",
		DepartureTimeTo:   "10:00",
		ArrivalTimeFrom:   "09:00",
		ArrivalTimeTo:     "12:00",
		Airlines:          []string{"GA"},
		MaxDuration:       100,
	}
	filtered := engine.Apply(flights, req)
	// Only F1 matches all criteria
	if len(filtered) != 1 || filtered[0].ID != "F1" {
		t.Errorf("expected only F1, got %+v", filtered)
	}
}

func TestFilterEngine_InvalidTimeStrings(t *testing.T) {
	engine := NewFilterEngine()
	flights := []domain.Flight{
		makeTestFlight("F1", "GA", 100000, 0, 60, 8, 0, 10, 0),
	}
	req := domain.SearchRequest{
		DepartureTimeFrom: "invalid", // parse fails -> depFrom stays nil, so no filter applied
		DepartureTimeTo:   "25:00",   // invalid
	}
	filtered := engine.Apply(flights, req)
	if len(filtered) != 1 {
		t.Errorf("expected flight to pass because invalid times are ignored, got %d", len(filtered))
	}
}

func TestFilterEngine_EmptyFlights(t *testing.T) {
	engine := NewFilterEngine()
	req := domain.SearchRequest{MinPrice: 100}
	filtered := engine.Apply([]domain.Flight{}, req)
	if len(filtered) != 0 {
		t.Errorf("expected empty slice, got %d", len(filtered))
	}
}
