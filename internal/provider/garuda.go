package provider

import (
	"context"
	"encoding/json"
	"math/rand"
	"os"
	"time"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
)

type GarudaProvider struct{}

func NewGarudaProvider() *GarudaProvider {
	return &GarudaProvider{}
}

type garudaResponse struct {
	Status  string `json:"status"`
	Flights []struct {
		FlightID string `json:"flight_id"`
		Airline  string `json:"airline"`

		Departure struct {
			Airport string `json:"airport"`
			Time    string `json:"time"`
		} `json:"departure"`

		Arrival struct {
			Airport string `json:"airport"`
			Time    string `json:"time"`
		} `json:"arrival"`

		DurationMinutes int `json:"duration_minutes"`
		Stops           int `json:"stops"`

		Segments []struct {
			Departure struct {
				Airport string `json:"airport"`
				Time    string `json:"time"`
			}
			Arrival struct {
				Airport string `json:"airport"`
				Time    string `json:"time"`
			}
			DurationMinutes int `json:"duration_minutes"`
			LayoverMinutes  int `json:"layover_minutes"`
		} `json:"segments"`

		Price struct {
			Amount int `json:"amount"`
		} `json:"price"`

		AvailableSeats int    `json:"available_seats"`
		FareClass      string `json:"fare_class"`
	} `json:"flights"`
}

func (p *GarudaProvider) Name() string {
	return "Garuda"
}

func (p *GarudaProvider) Search(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, error) {

	// simulate delay (50–100ms)
	delay := time.Duration(50+rand.Intn(50)) * time.Millisecond

	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	data, err := os.ReadFile("internal/provider/mock/garuda_indonesia_search_response.json")
	if err != nil {
		return nil, err
	}

	var resp garudaResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	result := make([]domain.Flight, 0)

	for _, f := range resp.Flights {

		var dep time.Time
		var arr time.Time
		var totalDuration int
		var stops int

		// multi-leg
		if len(f.Segments) > 0 {

			firstSeg := f.Segments[0]
			lastSeg := f.Segments[len(f.Segments)-1]

			dep, _ = time.Parse(time.RFC3339, firstSeg.Departure.Time)
			arr, _ = time.Parse(time.RFC3339, lastSeg.Arrival.Time)

			stops = len(f.Segments) - 1

			totalDuration = int(arr.Sub(dep).Minutes())

		} else {
			// direct flight
			dep, _ = time.Parse(time.RFC3339, f.Departure.Time)
			arr, _ = time.Parse(time.RFC3339, f.Arrival.Time)

			stops = f.Stops
			totalDuration = f.DurationMinutes
		}

		// validate
		if arr.Before(dep) {
			continue
		}

		flight := domain.Flight{
			ID:              f.FlightID + "_Garuda",
			Provider:        "Garuda",
			FlightNumber:    f.FlightID,
			Stops:           stops,
			AvailableSeats:  f.AvailableSeats,
			CabinClass:      f.FareClass,
			DurationMinutes: totalDuration,
		}

		flight.Airline.Name = f.Airline
		flight.Airline.Code = f.FlightID[:2]

		flight.Departure.Airport = f.Departure.Airport
		flight.Departure.Datetime = dep
		flight.Departure.Timestamp = dep.Unix()

		flight.Arrival.Airport = f.Arrival.Airport
		flight.Arrival.Datetime = arr
		flight.Arrival.Timestamp = arr.Unix()

		flight.Price.Amount = f.Price.Amount
		flight.Price.Currency = "IDR"

		result = append(result, flight)
	}

	return result, nil
}
