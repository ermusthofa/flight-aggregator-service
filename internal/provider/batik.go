package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
	"github.com/ermusthofa/flight-aggregator-service/internal/pkg"
)

type BatikProvider struct{}

func NewBatikProvider() *BatikProvider {
	return &BatikProvider{}
}

type batikResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Results []struct {
		FlightNumber string `json:"flightNumber"`
		AirlineName  string `json:"airlineName"`
		AirlineIATA  string `json:"airlineIATA"`

		Origin      string `json:"origin"`
		Destination string `json:"destination"`

		Departure string `json:"departureDateTime"`
		Arrival   string `json:"arrivalDateTime"`

		TravelTime string `json:"travelTime"` // "1h 45m"
		Stops      int    `json:"numberOfStops"`

		Connections []struct {
			StopAirport  string `json:"stopAirport"`
			StopDuration string `json:"stopDuration"` // "55m"
		} `json:"connections"`

		Fare struct {
			TotalPrice int    `json:"totalPrice"`
			Class      string `json:"class"`
		} `json:"fare"`

		Seats           int      `json:"seatsAvailable"`
		AircraftModel   string   `json:"aircraftModel"`
		BaggageInfo     string   `json:"baggageInfo"`
		OnboardServices []string `json:"onboardServices"`
	} `json:"results"`
}

func (p *BatikProvider) Name() string {
	return "Batik"
}

func (p *BatikProvider) Search(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, error) {

	// simulate delay (200–400ms)
	delay := time.Duration(200+rand.Intn(200)) * time.Millisecond

	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	data, err := os.ReadFile("internal/provider/mock/batik_air_search_response.json")
	if err != nil {
		return nil, err
	}

	var resp batikResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	result := make([]domain.Flight, 0)

	for _, f := range resp.Results {

		dep, err := time.Parse("2006-01-02T15:04:05-0700", f.Departure)
		if err != nil {
			continue
		}

		arr, err := time.Parse("2006-01-02T15:04:05-0700", f.Arrival)
		if err != nil {
			continue
		}

		// validate
		if arr.Before(dep) {
			continue
		}

		cabinClass := getBatikCabinClass(f.Fare.Class)

		// apply search criteria
		if !req.Matches(f.Origin, f.Destination, dep, f.Seats, cabinClass) {
			continue
		}

		// compute duration (prefer real time diff)
		duration := computeDuration(dep, arr, f.TravelTime)

		flight := domain.Flight{
			ID:             fmt.Sprintf("%s_%s", f.FlightNumber, p.Name()),
			Provider:       p.Name(),
			FlightNumber:   f.FlightNumber,
			Stops:          f.Stops,
			AvailableSeats: f.Seats,
			CabinClass:     cabinClass,
			Aircraft:       &f.AircraftModel,
			Duration: domain.Duration{
				TotalMinutes: duration,
				Formatted:    formatDuration(duration),
			},
			Amenities: f.OnboardServices,
			Baggage:   parseBaggage(f.BaggageInfo),
		}

		flight.Airline.Name = f.AirlineName
		flight.Airline.Code = f.AirlineIATA

		flight.Departure.Airport = f.Origin
		flight.Departure.City = getCityByAirport(f.Origin)
		flight.Departure.Datetime = dep
		flight.Departure.Timestamp = dep.Unix()

		flight.Arrival.Airport = f.Destination
		flight.Arrival.City = getCityByAirport(f.Destination)
		flight.Arrival.Datetime = arr
		flight.Arrival.Timestamp = arr.Unix()

		flight.Price.Amount = f.Fare.TotalPrice
		flight.Price.Currency = "IDR"

		result = append(result, flight)
	}

	return result, nil
}

func computeDuration(dep, arr time.Time, fallback string) int {
	duration := int(arr.Sub(dep).Minutes())
	// fallback (only if something is wrong)
	if duration > 0 {
		return duration
	}
	return parseDuration(fallback)
}

var batikCabinClassMapper = map[string]string{
	"Y": "economy",
}

func getBatikCabinClass(code string) string {
	if mappedClass, exists := batikCabinClassMapper[code]; exists {
		return mappedClass
	}
	pkg.Warning("Unmapped Batik Air cabin class code '%s'", code)
	return "unknown"
}
