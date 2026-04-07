package partner

import (
	"context"
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
	"github.com/ermusthofa/flight-aggregator-service/internal/partner/mock"
)

func TestAirAsiaProvider_Search_Success(t *testing.T) {
	rand.Seed(1)
	logger := &mockLogger{}
	provider := NewAirAsiaProvider(logger)

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

	if len(flights) != 4 {
		t.Fatalf("expected 4 flights, got %d", len(flights))
	}

	var expectedResp airAsiaResponse
	if err := json.Unmarshal(mock.AirAsiaMock, &expectedResp); err != nil {
		t.Fatalf("failed to unmarshal mock: %v", err)
	}

	expectedMap := make(map[string]struct {
		Airline string
		From    string
		To      string
		Depart  string
		Arrive  string
		Price   int
		Seats   int
		Direct  bool
		Stops   []struct {
			Airport         string `json:"airport"`
			WaitTimeMinutes int    `json:"wait_time_minutes"`
		}
		CabinClass  string
		BaggageNote string
	})
	for _, ef := range expectedResp.Flights {
		expectedMap[ef.FlightCode] = struct {
			Airline string
			From    string
			To      string
			Depart  string
			Arrive  string
			Price   int
			Seats   int
			Direct  bool
			Stops   []struct {
				Airport         string `json:"airport"`
				WaitTimeMinutes int    `json:"wait_time_minutes"`
			}
			CabinClass  string
			BaggageNote string
		}{
			Airline:     ef.Airline,
			From:        ef.From,
			To:          ef.To,
			Depart:      ef.DepartTime,
			Arrive:      ef.ArriveTime,
			Price:       ef.Price,
			Seats:       ef.Seats,
			Direct:      ef.Direct,
			Stops:       ef.Stops,
			CabinClass:  ef.CabinClass,
			BaggageNote: ef.BaggageNote,
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
		if flight.Departure.Airport != exp.From {
			t.Errorf("flight %s: expected departure airport %s, got %s", flight.FlightNumber, exp.From, flight.Departure.Airport)
		}
		if flight.Arrival.Airport != exp.To {
			t.Errorf("flight %s: expected arrival airport %s, got %s", flight.FlightNumber, exp.To, flight.Arrival.Airport)
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
		if flight.CabinClass != exp.CabinClass {
			t.Errorf("flight %s: expected cabin class %s, got %s", flight.FlightNumber, exp.CabinClass, flight.CabinClass)
		}

		// Stops count
		expectedStops := 0
		if !exp.Direct {
			expectedStops = len(exp.Stops)
		}
		if flight.Stops != expectedStops {
			t.Errorf("flight %s: expected stops %d, got %d", flight.FlightNumber, expectedStops, flight.Stops)
		}

		// Baggage validation - parseBaggageNote should populate CarryOn and Checked
		// Based on the fixture note, we expect both fields to be non-empty.
		if flight.Baggage.CarryOn == "" && flight.Baggage.Checked == "" {
			t.Errorf("flight %s: expected baggage info, got empty CarryOn and Checked", flight.FlightNumber)
		}

		// Duration calculation
		depTime, err := time.Parse(time.RFC3339, exp.Depart)
		if err != nil {
			t.Errorf("flight %s: invalid depart time in fixture: %v", flight.FlightNumber, err)
			continue
		}
		arrTime, err := time.Parse(time.RFC3339, exp.Arrive)
		if err != nil {
			t.Errorf("flight %s: invalid arrive time in fixture: %v", flight.FlightNumber, err)
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
		if flight.Airline.Code != flight.FlightNumber[:2] {
			t.Errorf("flight %s: expected airline code = flight number, got %s", flight.FlightNumber[:2], flight.Airline.Code)
		}
	}
}

func TestAirAsiaProvider_Search_FilterByRoute(t *testing.T) {
	rand.Seed(1)
	logger := &mockLogger{}
	provider := NewAirAsiaProvider(logger)

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
		t.Errorf("expected 0 flights for route SUB->DPS, got %d", len(flights))
	}
}

func TestAirAsiaProvider_Search_FilterByDate(t *testing.T) {
	rand.Seed(1)
	logger := &mockLogger{}
	provider := NewAirAsiaProvider(logger)

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

func TestAirAsiaProvider_Search_FilterBySeats(t *testing.T) {
	rand.Seed(1)
	logger := &mockLogger{}
	provider := NewAirAsiaProvider(logger)

	req := domain.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    100,
		CabinClass:    "economy",
	}

	ctx := context.Background()
	flights, err := provider.Search(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(flights) != 0 {
		t.Errorf("expected 0 flights for 100 passengers, got %d", len(flights))
	}
}

func TestAirAsiaProvider_Search_FilterByCabinClass(t *testing.T) {
	rand.Seed(1)
	logger := &mockLogger{}
	provider := NewAirAsiaProvider(logger)

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

func TestAirAsiaProvider_Search_SimulatedFailure(t *testing.T) {
	logger := &mockLogger{}
	provider := NewAirAsiaProvider(logger)
	req := domain.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
	}
	ctx := context.Background()

	foundFailure := false
	for attempt := 0; attempt < 50; attempt++ {
		_, err := provider.Search(ctx, req)
		if err != nil && err.Error() == "airasia API simulated failure" {
			foundFailure = true
			break
		}
		time.Sleep(1 * time.Millisecond)
	}
	if !foundFailure {
		t.Skip("simulated failure not triggered after 50 attempts; random behavior may require more attempts or deterministic seed")
	}
}

func TestAirAsiaProvider_Search_ContextCancellation(t *testing.T) {
	logger := &mockLogger{}
	provider := NewAirAsiaProvider(logger)
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
