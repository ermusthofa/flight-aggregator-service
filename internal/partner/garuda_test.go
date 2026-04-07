package partner

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
	"github.com/ermusthofa/flight-aggregator-service/internal/partner/mock"
)

func TestGarudaProvider_Search_Success(t *testing.T) {
	provider := NewGarudaProvider()

	req := domain.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	ctx := context.Background()
	flights, err := provider.Search(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expect 3 flights (GA400, GA410, GA315). GA315 is CGK->SUB->DPS but destination DPS.
	if len(flights) != 3 {
		t.Fatalf("expected 3 flights, got %d", len(flights))
	}

	// Load fixture
	var expectedResp garudaResponse
	if err := json.Unmarshal(mock.GarudaMock, &expectedResp); err != nil {
		t.Fatalf("failed to unmarshal mock: %v", err)
	}

	// Map flight_id -> expected data
	type expectedFlight struct {
		FlightID         string
		Airline          string
		AirlineCode      string
		DepartureAirport string
		ArrivalAirport   string
		DepartureTime    string
		ArrivalTime      string
		DurationMin      int
		Stops            int
		Aircraft         string
		Price            int
		Seats            int
		FareClass        string
		BaggageCarryOn   int
		BaggageChecked   int
		Amenities        []string
		HasSegments      bool
	}
	expectedMap := make(map[string]expectedFlight)
	for _, ef := range expectedResp.Flights {
		depAirport := ef.Departure.Airport
		arrAirport := ef.Arrival.Airport
		depTime := ef.Departure.Time
		arrTime := ef.Arrival.Time
		duration := ef.DurationMinutes
		stops := ef.Stops
		hasSegments := len(ef.Segments) > 0
		if hasSegments {
			// For connecting flights, use first segment departure and last segment arrival
			depTime = ef.Segments[0].Departure.Time
			arrTime = ef.Segments[len(ef.Segments)-1].Arrival.Time
			depAirport = ef.Segments[0].Departure.Airport
			arrAirport = ef.Segments[len(ef.Segments)-1].Arrival.Airport
			stops = len(ef.Segments) - 1
			// Duration computed from times, not from ef.DurationMinutes
			// We'll compute expected duration later from parsed times
		}
		expectedMap[ef.FlightID] = expectedFlight{
			FlightID:         ef.FlightID,
			Airline:          ef.Airline,
			AirlineCode:      ef.AirlineCode,
			DepartureAirport: depAirport,
			ArrivalAirport:   arrAirport,
			DepartureTime:    depTime,
			ArrivalTime:      arrTime,
			DurationMin:      duration,
			Stops:            stops,
			Aircraft:         ef.Aircraft,
			Price:            ef.Price.Amount,
			Seats:            ef.AvailableSeats,
			FareClass:        ef.FareClass,
			BaggageCarryOn:   ef.Baggage.CarryOn,
			BaggageChecked:   ef.Baggage.Checked,
			Amenities:        ef.Amenities,
			HasSegments:      hasSegments,
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
		if flight.Departure.Airport != exp.DepartureAirport {
			t.Errorf("flight %s: expected departure airport %s, got %s", flight.FlightNumber, exp.DepartureAirport, flight.Departure.Airport)
		}
		if flight.Arrival.Airport != exp.ArrivalAirport {
			t.Errorf("flight %s: expected arrival airport %s, got %s", flight.FlightNumber, exp.ArrivalAirport, flight.Arrival.Airport)
		}
		if flight.Price.Amount != exp.Price {
			t.Errorf("flight %s: expected price %d, got %d", flight.FlightNumber, exp.Price, flight.Price.Amount)
		}
		if flight.Price.Currency != "IDR" {
			t.Errorf("flight %s: expected currency IDR, got %s", flight.FlightNumber, flight.Price.Currency)
		}
		if flight.AvailableSeats != exp.Seats {
			t.Errorf("flight %s: expected seats %d, got %d", flight.FlightNumber, exp.Seats, flight.AvailableSeats)
		}
		expectedCabin := normalizeCabinClass(exp.FareClass)
		if flight.CabinClass != expectedCabin {
			t.Errorf("flight %s: expected cabin class %s, got %s", flight.FlightNumber, expectedCabin, flight.CabinClass)
		}
		if flight.Aircraft != exp.Aircraft {
			t.Errorf("flight %s: expected aircraft %s, got %s", flight.FlightNumber, exp.Aircraft, flight.Aircraft)
		}
		// Amenities – ensure slice is not nil (provider uses ensureSlice)
		if len(flight.Amenities) != len(exp.Amenities) && len(exp.Amenities) > 0 {
			t.Errorf("flight %s: expected %d amenities, got %d", flight.FlightNumber, len(exp.Amenities), len(flight.Amenities))
		}
		// Stops
		if flight.Stops != exp.Stops {
			t.Errorf("flight %s: expected stops %d, got %d", flight.FlightNumber, exp.Stops, flight.Stops)
		}

		// Baggage – converted from int to string
		expectedCarryOnStr := strconv.Itoa(exp.BaggageCarryOn)
		expectedCheckedStr := strconv.Itoa(exp.BaggageChecked)
		if flight.Baggage.CarryOn != expectedCarryOnStr {
			t.Errorf("flight %s: expected carry-on %s, got %s", flight.FlightNumber, expectedCarryOnStr, flight.Baggage.CarryOn)
		}
		if flight.Baggage.Checked != expectedCheckedStr {
			t.Errorf("flight %s: expected checked %s, got %s", flight.FlightNumber, expectedCheckedStr, flight.Baggage.Checked)
		}

		// Duration validation
		depTime, err := time.Parse(time.RFC3339, exp.DepartureTime)
		if err != nil {
			t.Errorf("flight %s: invalid departure time: %v", flight.FlightNumber, err)
			continue
		}
		arrTime, err := time.Parse(time.RFC3339, exp.ArrivalTime)
		if err != nil {
			t.Errorf("flight %s: invalid arrival time: %v", flight.FlightNumber, err)
			continue
		}
		expectedMinutes := int(arrTime.Sub(depTime).Minutes())
		if flight.TotalMinutes != expectedMinutes {
			t.Errorf("flight %s: expected total minutes %d, got %d", flight.FlightNumber, expectedMinutes, flight.TotalMinutes)
		}

		// Airline info
		if flight.Airline.Name != exp.Airline {
			t.Errorf("flight %s: expected airline name %s, got %s", flight.FlightNumber, exp.Airline, flight.Airline.Name)
		}
		if flight.Airline.Code != exp.AirlineCode {
			t.Errorf("flight %s: expected airline code %s, got %s", flight.FlightNumber, exp.AirlineCode, flight.Airline.Code)
		}
	}
}

func TestGarudaProvider_Search_FilterByRoute(t *testing.T) {
	provider := NewGarudaProvider()
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

func TestGarudaProvider_Search_FilterByDate(t *testing.T) {
	provider := NewGarudaProvider()
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

func TestGarudaProvider_Search_FilterBySeats(t *testing.T) {
	provider := NewGarudaProvider()
	req := domain.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    30, // exceeds max seats in fixture (28,15,22)
	}
	ctx := context.Background()
	flights, err := provider.Search(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(flights) != 0 {
		t.Errorf("expected 0 flights for 30 passengers, got %d", len(flights))
	}
}

func TestGarudaProvider_Search_FilterByCabinClass(t *testing.T) {
	provider := NewGarudaProvider()
	// All flights are economy, so business should return none
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

func TestGarudaProvider_Search_ContextCancellation(t *testing.T) {
	provider := NewGarudaProvider()
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
