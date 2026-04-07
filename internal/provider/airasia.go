package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
)

type AirAsiaProvider struct{}

func NewAirAsiaProvider() *AirAsiaProvider {
	return &AirAsiaProvider{}
}

type airAsiaResponse struct {
	Status  string `json:"status"`
	Flights []struct {
		FlightCode string  `json:"flight_code"`
		Airline    string  `json:"airline"`
		From       string  `json:"from_airport"`
		To         string  `json:"to_airport"`
		DepartTime string  `json:"depart_time"`
		ArriveTime string  `json:"arrive_time"`
		Duration   float64 `json:"duration_hours"`
		Direct     bool    `json:"direct_flight"`
		Stops      []struct {
			Airport         string `json:"airport"`
			WaitTimeMinutes int    `json:"wait_time_minutes"`
		} `json:"stops"`
		Price       int    `json:"price_idr"`
		Seats       int    `json:"seats"`
		CabinClass  string `json:"cabin_class"`
		BaggageNote string `json:"baggage_note"`
	} `json:"flights"`
}

func (p *AirAsiaProvider) Name() string {
	return "AirAsia"
}

func (p *AirAsiaProvider) MaxRetries() int {
	return 2
}

func (p *AirAsiaProvider) Search(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, error) {
	// simulate delay (50–150ms)
	delay := time.Duration(50+rand.Intn(100)) * time.Millisecond

	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// simulate 90% success rate
	if rand.Float64() > 0.9 {
		return nil, errors.New("airasia API failed")
	}

	data, err := loadMock("internal/provider/mock/airasia_search_response.json")
	if err != nil {
		return nil, err
	}

	var resp airAsiaResponse
	_ = json.Unmarshal([]byte(data), &resp)

	var result []domain.Flight

	for _, f := range resp.Flights {
		dep, _ := time.Parse(time.RFC3339, f.DepartTime)
		arr, _ := time.Parse(time.RFC3339, f.ArriveTime)

		// validate
		if arr.Before(dep) {
			continue
		}

		// apply search criteria
		if !req.Matches(f.From, f.To, dep, f.Seats, f.CabinClass) {
			continue
		}

		flight := domain.Flight{
			ID:             fmt.Sprintf("%s_%s", f.FlightCode, p.Name()),
			Provider:       p.Name(),
			FlightNumber:   f.FlightCode,
			Stops:          0,
			AvailableSeats: f.Seats,
			CabinClass:     f.CabinClass,
			Amenities:      []string{},
			Baggage:        parseBaggage(f.BaggageNote),
		}

		flight.Airline.Name = f.Airline
		flight.Airline.Code = f.FlightCode[:2]

		flight.Departure.Airport = f.From
		flight.Departure.City = getCityByAirport(f.From)
		flight.Departure.Datetime = dep
		flight.Departure.Timestamp = dep.Unix()

		flight.Arrival.Airport = f.To
		flight.Arrival.City = getCityByAirport(f.To)
		flight.Arrival.Datetime = arr
		flight.Arrival.Timestamp = arr.Unix()

		duration := arr.Sub(dep).Minutes()
		flight.TotalMinutes = int(duration)

		if !f.Direct {
			flight.Stops = len(f.Stops)
		}

		flight.Price.Amount = f.Price
		flight.Price.Currency = "IDR"

		result = append(result, flight)
	}

	return result, nil
}
