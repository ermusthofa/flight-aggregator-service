package partner

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
	"github.com/ermusthofa/flight-aggregator-service/internal/partner/mock"
	ratelimit "github.com/ermusthofa/flight-aggregator-service/internal/partner/ratelimiter"
)

type AirAsiaProvider struct {
	limiter *ratelimit.RateLimiter
}

func NewAirAsiaProvider() *AirAsiaProvider {
	// Allow 100 requests per second
	return &AirAsiaProvider{limiter: ratelimit.New(100, time.Second)}
}

type airAsiaResponse struct {
	Flights []struct {
		FlightCode    string  `json:"flight_code"`
		Airline       string  `json:"airline"`
		From          string  `json:"from_airport"`
		To            string  `json:"to_airport"`
		DepartTime    string  `json:"depart_time"`
		ArriveTime    string  `json:"arrive_time"`
		DurationHours float64 `json:"duration_hours"`
		Direct        bool    `json:"direct_flight"`
		Stops         []struct {
			Airport         string `json:"airport"`
			WaitTimeMinutes int    `json:"wait_time_minutes"`
		} `json:"stops"`
		Price       int    `json:"price_idr"`
		Seats       int    `json:"seats"`
		CabinClass  string `json:"cabin_class"`
		BaggageNote string `json:"baggage_note"`
	} `json:"flights"`
}

func (p *AirAsiaProvider) Name() string { return "AirAsia" }

func (p *AirAsiaProvider) Search(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, error) {
	if !p.limiter.Allow() {
		return nil, fmt.Errorf("%s API rate limit exceeded", p.Name())
	}

	// Simulate delay 50-150ms with context awareness
	delay := time.Duration(50+rand.Intn(100)) * time.Millisecond
	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Simulate 90% success rate
	if rand.Float64() > 0.9 {
		return nil, fmt.Errorf("airasia API simulated failure")
	}

	var resp airAsiaResponse
	if err := json.Unmarshal(mock.AirAsiaMock, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse airasia mock: %w", err)
	}

	var flights []domain.Flight
	for _, f := range resp.Flights {
		// Parse times
		dep, err := time.Parse(time.RFC3339, f.DepartTime)
		if err != nil {
			warnSkip(ctx, p.Name(), f.FlightCode, "parse departure time", err)
			continue
		}
		arr, err := time.Parse(time.RFC3339, f.ArriveTime)
		if err != nil {
			warnSkip(ctx, p.Name(), f.FlightCode, "parse arrival time", err)
			continue
		}

		// Validate arrival after departure
		if arr.Before(dep) {
			warnSkip(ctx, p.Name(), f.FlightCode, "arrival before departure", nil)
			continue
		}

		// Apply search criteria (route, date, passenger count, cabin)
		if !matchesSearchCriteria(req, f.From, f.To, dep, f.Seats, f.CabinClass) {
			continue
		}

		// Calculate stops
		stops := 0
		if !f.Direct {
			stops = len(f.Stops)
		}

		flight := domain.Flight{
			ID:             fmt.Sprintf("%s_%s", f.FlightCode, p.Name()),
			Provider:       p.Name(),
			FlightNumber:   f.FlightCode,
			Stops:          stops,
			AvailableSeats: f.Seats,
			CabinClass:     normalizeCabinClass(f.CabinClass),
			TotalMinutes:   int(arr.Sub(dep).Minutes()),
			Amenities:      []string{},
			Baggage:        parseBaggageNote(f.BaggageNote),
			Airline: domain.Airline{
				Name: f.Airline,
				Code: f.FlightCode[:2],
			},
			Departure: domain.Location{
				Airport:   f.From,
				City:      airportToCity(f.From),
				Datetime:  dep.UTC(),
				Timestamp: dep.Unix(),
			},
			Arrival: domain.Location{
				Airport:   f.To,
				City:      airportToCity(f.To),
				Datetime:  arr.UTC(),
				Timestamp: arr.Unix(),
			},
			Price: domain.Price{
				Amount:   f.Price,
				Currency: "IDR",
			},
		}
		flights = append(flights, flight)
	}
	return flights, nil
}
