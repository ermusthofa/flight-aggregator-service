package partner

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
	"github.com/ermusthofa/flight-aggregator-service/internal/partner/mock"
)

func TestLionProvider_Search_Success(t *testing.T) {
	logger := &mockLogger{}
	provider := NewLionProvider(logger)

	req := domain.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy", // fare_type "ECONOMY" maps to "economy"
	}

	ctx := context.Background()
	flights, err := provider.Search(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expect all 3 flights from fixture (CGK->DPS, date matches)
	if len(flights) != 3 {
		t.Fatalf("expected 3 flights, got %d", len(flights))
	}

	// Load fixture
	var expectedResp lionResponse
	if err := json.Unmarshal(mock.LionMock, &expectedResp); err != nil {
		t.Fatalf("failed to unmarshal mock: %v", err)
	}

	// Map flight id -> expected data
	type expectedFlight struct {
		ID            string
		CarrierName   string
		CarrierIATA   string
		FromCode      string
		ToCode        string
		FromCity      string
		ToCity        string
		Departure     string
		DepartureTZ   string
		Arrival       string
		ArrivalTZ     string
		Price         int
		Seats         int
		PlaneType     string
		IsDirect      bool
		StopCount     int
		WifiAvailable bool
		MealsIncluded bool
		BaggageCabin  string
		BaggageHold   string
	}
	expectedMap := make(map[string]expectedFlight)
	for _, ef := range expectedResp.Data.Flights {
		expectedMap[ef.ID] = expectedFlight{
			ID:            ef.ID,
			CarrierName:   ef.Carrier.Name,
			CarrierIATA:   ef.Carrier.IATA,
			FromCode:      ef.Route.From.Code,
			ToCode:        ef.Route.To.Code,
			FromCity:      ef.Route.From.City,
			ToCity:        ef.Route.To.City,
			Departure:     ef.Schedule.Departure,
			DepartureTZ:   ef.Schedule.DepartureTimezone,
			Arrival:       ef.Schedule.Arrival,
			ArrivalTZ:     ef.Schedule.ArrivalTimezone,
			Price:         ef.Pricing.Total,
			Seats:         ef.Seats,
			PlaneType:     ef.PlaneType,
			IsDirect:      ef.IsDirect,
			StopCount:     ef.StopCount,
			WifiAvailable: ef.Services.WifiAvailable,
			MealsIncluded: ef.Services.MealsIncluded,
			BaggageCabin:  ef.Services.BaggageAllowance.Cabin,
			BaggageHold:   ef.Services.BaggageAllowance.Hold,
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
		if flight.Departure.Airport != exp.FromCode {
			t.Errorf("flight %s: expected departure airport %s, got %s", flight.FlightNumber, exp.FromCode, flight.Departure.Airport)
		}
		if flight.Departure.City != exp.FromCity {
			t.Errorf("flight %s: expected departure city %s, got %s", flight.FlightNumber, exp.FromCity, flight.Departure.City)
		}
		if flight.Arrival.Airport != exp.ToCode {
			t.Errorf("flight %s: expected arrival airport %s, got %s", flight.FlightNumber, exp.ToCode, flight.Arrival.Airport)
		}
		if flight.Arrival.City != exp.ToCity {
			t.Errorf("flight %s: expected arrival city %s, got %s", flight.FlightNumber, exp.ToCity, flight.Arrival.City)
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
		if flight.CabinClass != "economy" {
			t.Errorf("flight %s: expected cabin class economy, got %s", flight.FlightNumber, flight.CabinClass)
		}
		if flight.Aircraft != exp.PlaneType {
			t.Errorf("flight %s: expected aircraft %s, got %s", flight.FlightNumber, exp.PlaneType, flight.Aircraft)
		}

		// Stops count
		expectedStops := 0
		if !exp.IsDirect {
			expectedStops = exp.StopCount
		}
		if flight.Stops != expectedStops {
			t.Errorf("flight %s: expected stops %d, got %d", flight.FlightNumber, expectedStops, flight.Stops)
		}

		// Baggage
		if flight.Baggage.CarryOn != exp.BaggageCabin {
			t.Errorf("flight %s: expected carry-on %s, got %s", flight.FlightNumber, exp.BaggageCabin, flight.Baggage.CarryOn)
		}
		if flight.Baggage.Checked != exp.BaggageHold {
			t.Errorf("flight %s: expected checked %s, got %s", flight.FlightNumber, exp.BaggageHold, flight.Baggage.Checked)
		}

		// Amenities mapping
		expectedAmenities := []string{}
		if exp.WifiAvailable {
			expectedAmenities = append(expectedAmenities, "wifi")
		}
		if exp.MealsIncluded {
			expectedAmenities = append(expectedAmenities, "meals")
		}
		if len(flight.Amenities) != len(expectedAmenities) {
			t.Errorf("flight %s: expected %d amenities, got %d", flight.FlightNumber, len(expectedAmenities), len(flight.Amenities))
		}
		for i, amen := range expectedAmenities {
			if i < len(flight.Amenities) && flight.Amenities[i] != amen {
				t.Errorf("flight %s: expected amenity %s, got %s", flight.FlightNumber, amen, flight.Amenities[i])
			}
		}

		// Duration calculation: parse times with timezone offsets
		// WIB = UTC+7, WITA = UTC+8
		var depOffset, arrOffset int
		switch exp.DepartureTZ {
		case "Asia/Jakarta":
			depOffset = 7
		default:
			depOffset = 7
		}
		switch exp.ArrivalTZ {
		case "Asia/Makassar":
			arrOffset = 8
		default:
			arrOffset = 8
		}
		depTime, err := time.Parse("2006-01-02T15:04:05", exp.Departure)
		if err != nil {
			t.Errorf("flight %s: invalid departure time string: %v", flight.FlightNumber, err)
			continue
		}
		arrTime, err := time.Parse("2006-01-02T15:04:05", exp.Arrival)
		if err != nil {
			t.Errorf("flight %s: invalid arrival time string: %v", flight.FlightNumber, err)
			continue
		}
		// Convert to UTC by subtracting offset hours
		depUTC := depTime.Add(-time.Duration(depOffset) * time.Hour)
		arrUTC := arrTime.Add(-time.Duration(arrOffset) * time.Hour)
		expectedMinutes := int(arrUTC.Sub(depUTC).Minutes())
		if flight.TotalMinutes != expectedMinutes {
			t.Errorf("flight %s: expected total minutes %d, got %d", flight.FlightNumber, expectedMinutes, flight.TotalMinutes)
		}

		// Airline info
		if flight.Airline.Name != exp.CarrierName {
			t.Errorf("flight %s: expected airline name %s, got %s", flight.FlightNumber, exp.CarrierName, flight.Airline.Name)
		}
		if flight.Airline.Code != exp.CarrierIATA {
			t.Errorf("flight %s: expected airline code %s, got %s", flight.FlightNumber, exp.CarrierIATA, flight.Airline.Code)
		}
	}
}

func TestLionProvider_Search_FilterByRoute(t *testing.T) {
	logger := &mockLogger{}
	provider := NewLionProvider(logger)
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

func TestLionProvider_Search_FilterByDate(t *testing.T) {
	logger := &mockLogger{}
	provider := NewLionProvider(logger)
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

func TestLionProvider_Search_FilterBySeats(t *testing.T) {
	logger := &mockLogger{}
	provider := NewLionProvider(logger)
	req := domain.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    60, // exceeds max seats in fixture (45,38,52)
	}
	ctx := context.Background()
	flights, err := provider.Search(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(flights) != 0 {
		t.Errorf("expected 0 flights for 60 passengers, got %d", len(flights))
	}
}

func TestLionProvider_Search_FilterByCabinClass(t *testing.T) {
	logger := &mockLogger{}
	provider := NewLionProvider(logger)
	// All flights have fare_type "ECONOMY" -> maps to economy. Business should return none.
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

func TestLionProvider_Search_ContextCancellation(t *testing.T) {
	logger := &mockLogger{}
	provider := NewLionProvider(logger)
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
