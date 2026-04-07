package mapper

import (
	"testing"
	"time"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
	"github.com/ermusthofa/flight-aggregator-service/internal/dto"
)

func TestToSearchCriteriaDTO(t *testing.T) {
	req := domain.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    2,
		CabinClass:    "economy",
	}
	expected := dto.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    2,
		CabinClass:    "economy",
	}
	result := ToSearchCriteriaDTO(req)
	if result != expected {
		t.Errorf("expected %+v, got %+v", expected, result)
	}
}

func TestToMetadataDTO(t *testing.T) {
	meta := domain.Metadata{
		TotalResults:       10,
		ProvidersQueried:   3,
		ProvidersSucceeded: 2,
		ProvidersFailed:    1,
		SearchTimeMs:       123,
		CacheHit:           true,
	}
	expected := dto.Metadata{
		TotalResults:       10,
		ProvidersQueried:   3,
		ProvidersSucceeded: 2,
		ProvidersFailed:    1,
		SearchTimeMs:       123,
		CacheHit:           true,
	}
	result := ToMetadataDTO(meta)
	if result != expected {
		t.Errorf("expected %+v, got %+v", expected, result)
	}
}

func TestToFlightDTOs_Empty(t *testing.T) {
	flights := []domain.Flight{}
	result := ToFlightDTOs(flights)
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d", len(result))
	}
}

func TestToFlightDTOs_SingleFlight(t *testing.T) {
	now := time.Date(2025, 12, 15, 10, 30, 0, 0, time.UTC)
	dep := now
	arr := now.Add(2*time.Hour + 30*time.Minute)
	flight := domain.Flight{
		ID:             "GA123_Airline",
		Provider:       "Garuda",
		FlightNumber:   "GA123",
		Stops:          1,
		AvailableSeats: 20,
		CabinClass:     "economy",
		Amenities:      []string{"meal", "wifi"},
		Airline: domain.Airline{
			Name: "Garuda Indonesia",
			Code: "GA",
		},
		Departure: domain.Location{
			Airport:   "CGK",
			City:      "Jakarta",
			Datetime:  dep,
			Timestamp: dep.Unix(),
		},
		Arrival: domain.Location{
			Airport:   "DPS",
			City:      "Denpasar",
			Datetime:  arr,
			Timestamp: arr.Unix(),
		},
		TotalMinutes: 150,
		Price: domain.Price{
			Amount:   1250000,
			Currency: "IDR",
		},
		Aircraft: "Boeing 737",
		Baggage: domain.Baggage{
			CarryOn: "7 kg",
			Checked: "20 kg",
		},
	}
	result := ToFlightDTOs([]domain.Flight{flight})
	if len(result) != 1 {
		t.Fatalf("expected 1 flight, got %d", len(result))
	}
	dtoFlight := result[0]

	// Check simple fields
	if dtoFlight.ID != flight.ID {
		t.Errorf("ID: expected %s, got %s", flight.ID, dtoFlight.ID)
	}
	if dtoFlight.Provider != flight.Provider {
		t.Errorf("Provider: expected %s, got %s", flight.Provider, dtoFlight.Provider)
	}
	if dtoFlight.FlightNumber != flight.FlightNumber {
		t.Errorf("FlightNumber: expected %s, got %s", flight.FlightNumber, dtoFlight.FlightNumber)
	}
	if dtoFlight.Stops != flight.Stops {
		t.Errorf("Stops: expected %d, got %d", flight.Stops, dtoFlight.Stops)
	}
	if dtoFlight.AvailableSeats != flight.AvailableSeats {
		t.Errorf("AvailableSeats: expected %d, got %d", flight.AvailableSeats, dtoFlight.AvailableSeats)
	}
	if dtoFlight.CabinClass != flight.CabinClass {
		t.Errorf("CabinClass: expected %s, got %s", flight.CabinClass, dtoFlight.CabinClass)
	}
	if len(dtoFlight.Amenities) != len(flight.Amenities) {
		t.Errorf("Amenities length: expected %d, got %d", len(flight.Amenities), len(dtoFlight.Amenities))
	}
	// Airline
	if dtoFlight.Airline.Name != flight.Airline.Name {
		t.Errorf("Airline.Name: expected %s, got %s", flight.Airline.Name, dtoFlight.Airline.Name)
	}
	if dtoFlight.Airline.Code != flight.Airline.Code {
		t.Errorf("Airline.Code: expected %s, got %s", flight.Airline.Code, dtoFlight.Airline.Code)
	}
	// Departure
	if dtoFlight.Departure.Airport != flight.Departure.Airport {
		t.Errorf("Departure.Airport: expected %s, got %s", flight.Departure.Airport, dtoFlight.Departure.Airport)
	}
	if dtoFlight.Departure.City != flight.Departure.City {
		t.Errorf("Departure.City: expected %s, got %s", flight.Departure.City, dtoFlight.Departure.City)
	}
	if dtoFlight.Departure.Datetime != flight.Departure.Datetime.Format(time.RFC3339) {
		t.Errorf("Departure.Datetime: expected %s, got %s", flight.Departure.Datetime.Format(time.RFC3339), dtoFlight.Departure.Datetime)
	}
	if dtoFlight.Departure.Timestamp != flight.Departure.Timestamp {
		t.Errorf("Departure.Timestamp: expected %d, got %d", flight.Departure.Timestamp, dtoFlight.Departure.Timestamp)
	}
	// Arrival
	if dtoFlight.Arrival.Airport != flight.Arrival.Airport {
		t.Errorf("Arrival.Airport: expected %s, got %s", flight.Arrival.Airport, dtoFlight.Arrival.Airport)
	}
	if dtoFlight.Arrival.City != flight.Arrival.City {
		t.Errorf("Arrival.City: expected %s, got %s", flight.Arrival.City, dtoFlight.Arrival.City)
	}
	if dtoFlight.Arrival.Datetime != flight.Arrival.Datetime.Format(time.RFC3339) {
		t.Errorf("Arrival.Datetime: expected %s, got %s", flight.Arrival.Datetime.Format(time.RFC3339), dtoFlight.Arrival.Datetime)
	}
	if dtoFlight.Arrival.Timestamp != flight.Arrival.Timestamp {
		t.Errorf("Arrival.Timestamp: expected %d, got %d", flight.Arrival.Timestamp, dtoFlight.Arrival.Timestamp)
	}
	// Duration
	if dtoFlight.Duration.TotalMinutes != flight.TotalMinutes {
		t.Errorf("Duration.TotalMinutes: expected %d, got %d", flight.TotalMinutes, dtoFlight.Duration.TotalMinutes)
	}
	expectedFormatted := "2h 30m"
	if dtoFlight.Duration.Formatted != expectedFormatted {
		t.Errorf("Duration.Formatted: expected %s, got %s", expectedFormatted, dtoFlight.Duration.Formatted)
	}
	// Price
	if dtoFlight.Price.Amount != flight.Price.Amount {
		t.Errorf("Price.Amount: expected %d, got %d", flight.Price.Amount, dtoFlight.Price.Amount)
	}
	if dtoFlight.Price.Currency != flight.Price.Currency {
		t.Errorf("Price.Currency: expected %s, got %s", flight.Price.Currency, dtoFlight.Price.Currency)
	}
	// Aircraft
	if dtoFlight.Aircraft == nil {
		t.Error("Aircraft should not be nil for non-empty string")
	} else if *dtoFlight.Aircraft != flight.Aircraft {
		t.Errorf("Aircraft: expected %s, got %s", flight.Aircraft, *dtoFlight.Aircraft)
	}
	// Baggage
	if dtoFlight.Baggage.CarryOn != flight.Baggage.CarryOn {
		t.Errorf("Baggage.CarryOn: expected %s, got %s", flight.Baggage.CarryOn, dtoFlight.Baggage.CarryOn)
	}
	if dtoFlight.Baggage.Checked != flight.Baggage.Checked {
		t.Errorf("Baggage.Checked: expected %s, got %s", flight.Baggage.Checked, dtoFlight.Baggage.Checked)
	}
}

func TestToFlightDTOs_NilAircraft(t *testing.T) {
	flight := domain.Flight{
		ID:       "JT123",
		Aircraft: "", // empty string → should be nil in DTO
		Departure: domain.Location{
			Datetime: time.Now(),
		},
		Arrival: domain.Location{
			Datetime: time.Now().Add(time.Hour),
		},
	}
	result := ToFlightDTOs([]domain.Flight{flight})
	if result[0].Aircraft != nil {
		t.Error("Expected Aircraft to be nil for empty string, got non-nil")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		minutes  int
		expected string
	}{
		{0, "0h 0m"},
		{30, "0h 30m"},
		{60, "1h 0m"},
		{90, "1h 30m"},
		{120, "2h 0m"},
		{125, "2h 5m"},
		{150, "2h 30m"},
	}
	for _, tt := range tests {
		result := formatDuration(tt.minutes)
		if result != tt.expected {
			t.Errorf("formatDuration(%d) = %s, want %s", tt.minutes, result, tt.expected)
		}
	}
}
