package partner

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
	"github.com/ermusthofa/flight-aggregator-service/internal/partner/mock"
)

func TestBatikProvider_Search_Success(t *testing.T) {
	provider := NewBatikProvider()

	req := domain.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy", // "Y" maps to economy via normalizeCabinClass
	}

	ctx := context.Background()
	flights, err := provider.Search(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expect all 3 flights from fixture (CGK->DPS, date matches, seats and cabin ok)
	if len(flights) != 3 {
		t.Fatalf("expected 3 flights, got %d", len(flights))
	}

	// Load fixture
	var expectedResp batikResponse
	if err := json.Unmarshal(mock.BatikMock, &expectedResp); err != nil {
		t.Fatalf("failed to unmarshal mock: %v", err)
	}

	// Map flightNumber -> expected data
	expectedMap := make(map[string]struct {
		Origin      string
		Destination string
		Departure   string
		Arrival     string
		TotalPrice  int
		Seats       int
		Stops       int
		Connections []struct {
			StopAirport  string `json:"stopAirport"`
			StopDuration string `json:"stopDuration"`
		}
		Class           string
		AircraftModel   string
		BaggageInfo     string
		OnboardServices []string
	})
	for _, ef := range expectedResp.Results {
		expectedMap[ef.FlightNumber] = struct {
			Origin      string
			Destination string
			Departure   string
			Arrival     string
			TotalPrice  int
			Seats       int
			Stops       int
			Connections []struct {
				StopAirport  string `json:"stopAirport"`
				StopDuration string `json:"stopDuration"`
			}
			Class           string
			AircraftModel   string
			BaggageInfo     string
			OnboardServices []string
		}{
			Origin:          ef.Origin,
			Destination:     ef.Destination,
			Departure:       ef.Departure,
			Arrival:         ef.Arrival,
			TotalPrice:      ef.Fare.TotalPrice,
			Seats:           ef.Seats,
			Stops:           ef.Stops,
			Connections:     ef.Connections,
			Class:           ef.Fare.Class,
			AircraftModel:   ef.AircraftModel,
			BaggageInfo:     ef.BaggageInfo,
			OnboardServices: ef.OnboardServices,
		}
	}

	for _, flight := range flights {
		exp, ok := expectedMap[flight.FlightNumber]
		if !ok {
			t.Errorf("flight %s not found in fixture", flight.FlightNumber)
			continue
		}

		// Basic fields
		if flight.Provider != provider.Name() {
			t.Errorf("flight %s: expected provider %s, got %s", flight.FlightNumber, provider.Name(), flight.Provider)
		}
		if flight.Departure.Airport != exp.Origin {
			t.Errorf("flight %s: expected departure airport %s, got %s", flight.FlightNumber, exp.Origin, flight.Departure.Airport)
		}
		if flight.Arrival.Airport != exp.Destination {
			t.Errorf("flight %s: expected arrival airport %s, got %s", flight.FlightNumber, exp.Destination, flight.Arrival.Airport)
		}
		if flight.Price.Amount != exp.TotalPrice {
			t.Errorf("flight %s: expected price %d, got %d", flight.FlightNumber, exp.TotalPrice, flight.Price.Amount)
		}
		if flight.Price.Currency != "IDR" {
			t.Errorf("flight %s: expected currency IDR, got %s", flight.FlightNumber, flight.Price.Currency)
		}
		if flight.AvailableSeats != exp.Seats {
			t.Errorf("flight %s: expected seats %d, got %d", flight.FlightNumber, exp.Seats, flight.AvailableSeats)
		}
		// Cabin class normalization: "Y" -> "economy"
		expectedCabin := normalizeCabinClass(exp.Class)
		if flight.CabinClass != expectedCabin {
			t.Errorf("flight %s: expected cabin class %s, got %s", flight.FlightNumber, expectedCabin, flight.CabinClass)
		}
		if flight.Aircraft != exp.AircraftModel {
			t.Errorf("flight %s: expected aircraft %s, got %s", flight.FlightNumber, exp.AircraftModel, flight.Aircraft)
		}
		// Amenities
		if len(flight.Amenities) != len(exp.OnboardServices) {
			t.Errorf("flight %s: expected %d amenities, got %d", flight.FlightNumber, len(exp.OnboardServices), len(flight.Amenities))
		}
		// Stops count
		if flight.Stops != exp.Stops {
			t.Errorf("flight %s: expected stops %d, got %d", flight.FlightNumber, exp.Stops, flight.Stops)
		}

		// Baggage parsing - should produce non-empty CarryOn and Checked
		if flight.Baggage.CarryOn == "" && flight.Baggage.Checked == "" {
			t.Errorf("flight %s: expected baggage info, got empty", flight.FlightNumber)
		}

		// Duration calculation
		depTime, err := time.Parse("2006-01-02T15:04:05-0700", exp.Departure)
		if err != nil {
			t.Errorf("flight %s: invalid depart time: %v", flight.FlightNumber, err)
			continue
		}
		arrTime, err := time.Parse("2006-01-02T15:04:05-0700", exp.Arrival)
		if err != nil {
			t.Errorf("flight %s: invalid arrival time: %v", flight.FlightNumber, err)
			continue
		}
		expectedMinutes := int(arrTime.Sub(depTime).Minutes())
		if flight.TotalMinutes != expectedMinutes {
			t.Errorf("flight %s: expected total minutes %d, got %d", flight.FlightNumber, expectedMinutes, flight.TotalMinutes)
		}

		// Airline info
		if flight.Airline.Name != "Batik Air" {
			t.Errorf("flight %s: expected airline name Batik Air, got %s", flight.FlightNumber, flight.Airline.Name)
		}
		if flight.Airline.Code != "ID" {
			t.Errorf("flight %s: expected airline code ID, got %s", flight.FlightNumber, flight.Airline.Code)
		}
	}
}

func TestBatikProvider_Search_FilterByRoute(t *testing.T) {
	provider := NewBatikProvider()
	req := domain.SearchRequest{
		Origin:        "SUB",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
	}
	ctx := context.Background()
	flights, err := provider.Search(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(flights) != 0 {
		t.Errorf("expected 0 flights for SUB->DPS, got %d", len(flights))
	}
}

func TestBatikProvider_Search_FilterByDate(t *testing.T) {
	provider := NewBatikProvider()
	req := domain.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-16",
		Passengers:    1,
	}
	ctx := context.Background()
	flights, err := provider.Search(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(flights) != 0 {
		t.Errorf("expected 0 flights for date 2025-12-16, got %d", len(flights))
	}
}

func TestBatikProvider_Search_FilterBySeats(t *testing.T) {
	provider := NewBatikProvider()
	req := domain.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    50, // exceeds available seats (max 41 in fixture)
	}
	ctx := context.Background()
	flights, err := provider.Search(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(flights) != 0 {
		t.Errorf("expected 0 flights for 50 passengers, got %d", len(flights))
	}
}

func TestBatikProvider_Search_FilterByCabinClass(t *testing.T) {
	provider := NewBatikProvider()
	// All flights have class "Y" -> economy. Business class should yield none.
	req := domain.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "business",
	}
	ctx := context.Background()
	flights, err := provider.Search(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(flights) != 0 {
		t.Errorf("expected 0 flights for business class, got %d", len(flights))
	}
}

func TestBatikProvider_Search_ContextCancellation(t *testing.T) {
	provider := NewBatikProvider()
	req := domain.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := provider.Search(ctx, req)
	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// Optional: test parseDurationString helper directly
func TestParseDurationString(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"1h 45m", 105},
		{"2h", 120},
		{"30m", 30},
		{"1h 30m", 90},
		{"", 0},
		{"invalid", 0},
	}
	for _, tt := range tests {
		result := parseDurationString(tt.input)
		if result != tt.expected {
			t.Errorf("parseDurationString(%q) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}
