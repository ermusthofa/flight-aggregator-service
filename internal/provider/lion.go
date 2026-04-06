package provider

import (
	"context"
	"encoding/json"
	"math/rand"
	"os"
	"time"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
)

type LionProvider struct{}

func NewLionProvider() *LionProvider {
	return &LionProvider{}
}

type lionResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Flights []struct {
			ID string `json:"id"`

			Carrier struct {
				Name string `json:"name"`
				IATA string `json:"iata"`
			} `json:"carrier"`

			Route struct {
				From struct {
					Code string `json:"code"`
				} `json:"from"`
				To struct {
					Code string `json:"code"`
				} `json:"to"`
			} `json:"route"`

			Schedule struct {
				Departure         string `json:"departure"`
				DepartureTimezone string `json:"departure_timezone"`
				Arrival           string `json:"arrival"`
				ArrivalTimezone   string `json:"arrival_timezone"`
			} `json:"schedule"`

			FlightTime int `json:"flight_time"`

			IsDirect  bool `json:"is_direct"`
			StopCount int  `json:"stop_count"`

			Layovers []struct {
				Airport  string `json:"airport"`
				Duration int    `json:"duration_minutes"`
			} `json:"layovers"`

			Pricing struct {
				Total int `json:"total"`
			} `json:"pricing"`

			Seats int `json:"seats_left"`
		} `json:"available_flights"`
	} `json:"data"`
}

func (p *LionProvider) Name() string {
	return "Lion"
}

func (p *LionProvider) Search(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, error) {

	// simulate delay (100–200ms)
	delay := time.Duration(100+rand.Intn(100)) * time.Millisecond

	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	data, err := os.ReadFile("internal/provider/mock/lion_air_search_response.json")
	if err != nil {
		return nil, err
	}

	var resp lionResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	result := make([]domain.Flight, 0)

	for _, f := range resp.Data.Flights {

		dep, err := parseWithTimezone(
			f.Schedule.Departure,
			f.Schedule.DepartureTimezone,
		)
		if err != nil {
			continue
		}

		arr, err := parseWithTimezone(
			f.Schedule.Arrival,
			f.Schedule.ArrivalTimezone,
		)
		if err != nil {
			continue
		}

		// validate
		if arr.Before(dep) {
			continue
		}

		// compute duration safely
		duration := int(arr.Sub(dep).Minutes())

		stops := 0
		if !f.IsDirect {
			stops = f.StopCount
		}

		flight := domain.Flight{
			ID:              f.ID + "_Lion",
			Provider:        "Lion",
			FlightNumber:    f.ID,
			Stops:           stops,
			AvailableSeats:  f.Seats,
			CabinClass:      "economy", // from fare_type
			DurationMinutes: duration,
		}

		flight.Airline.Name = f.Carrier.Name
		flight.Airline.Code = f.Carrier.IATA

		flight.Departure.Airport = f.Route.From.Code
		flight.Departure.Datetime = dep
		flight.Departure.Timestamp = dep.Unix()

		flight.Arrival.Airport = f.Route.To.Code
		flight.Arrival.Datetime = arr
		flight.Arrival.Timestamp = arr.Unix()

		flight.Price.Amount = f.Pricing.Total
		flight.Price.Currency = "IDR"

		result = append(result, flight)
	}

	return result, nil
}
