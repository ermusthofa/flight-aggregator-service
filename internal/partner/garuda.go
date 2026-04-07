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

type GarudaProvider struct {
	limiter *ratelimit.RateLimiter
}

func NewGarudaProvider() *GarudaProvider {
	// Allow 100 requests per second
	return &GarudaProvider{limiter: ratelimit.New(100, time.Second)}
}

type garudaResponse struct {
	Status  string `json:"status"`
	Flights []struct {
		FlightID    string `json:"flight_id"`
		Airline     string `json:"airline"`
		AirlineCode string `json:"airline_code"`
		Departure   struct {
			Airport string `json:"airport"`
			City    string `json:"city"`
			Time    string `json:"time"`
		} `json:"departure"`
		Arrival struct {
			Airport string `json:"airport"`
			City    string `json:"city"`
			Time    string `json:"time"`
		} `json:"arrival"`
		DurationMinutes int    `json:"duration_minutes"`
		Stops           int    `json:"stops"`
		Aircraft        string `json:"aircraft"`
		Segments        []struct {
			Departure struct {
				Airport string `json:"airport"`
				Time    string `json:"time"`
			} `json:"departure"`
			Arrival struct {
				Airport string `json:"airport"`
				Time    string `json:"time"`
			} `json:"arrival"`
			DurationMinutes int `json:"duration_minutes"`
			LayoverMinutes  int `json:"layover_minutes"`
		} `json:"segments"`
		Price struct {
			Amount int `json:"amount"`
		} `json:"price"`
		AvailableSeats int    `json:"available_seats"`
		FareClass      string `json:"fare_class"`
		Baggage        struct {
			CarryOn int `json:"carry_on"`
			Checked int `json:"checked"`
		} `json:"baggage"`
		Amenities []string `json:"amenities"`
	} `json:"flights"`
}

func (p *GarudaProvider) Name() string { return "Garuda" }

func (p *GarudaProvider) Search(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, error) {
	if !p.limiter.Allow() {
		return nil, fmt.Errorf("%s API rate limit exceeded", p.Name())
	}

	// simulate delay (50–100ms)
	delay := time.Duration(50+rand.Intn(50)) * time.Millisecond
	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	var resp garudaResponse
	if err := json.Unmarshal(mock.GarudaMock, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse garuda mock: %w", err)
	}

	var flights []domain.Flight
	for _, f := range resp.Flights {
		var dep, arr time.Time
		var totalDuration int
		var arrivalAirport, arrivalCity string
		var stops int

		if len(f.Segments) > 0 {
			// Multi-leg flight
			firstSeg := f.Segments[0]
			lastSeg := f.Segments[len(f.Segments)-1]
			var err error
			dep, err = time.Parse(time.RFC3339, firstSeg.Departure.Time)
			if err != nil {
				warnSkip(ctx, p.Name(), f.FlightID, "parse departure time", err)
				continue
			}
			arr, err = time.Parse(time.RFC3339, lastSeg.Arrival.Time)
			if err != nil {
				warnSkip(ctx, p.Name(), f.FlightID, "parse arrival time", err)
				continue
			}
			stops = len(f.Segments) - 1
			totalDuration = int(arr.Sub(dep).Minutes())
			arrivalAirport = lastSeg.Arrival.Airport
			arrivalCity = airportToCity(arrivalAirport)
		} else {
			// Direct flight
			var err error
			dep, err = time.Parse(time.RFC3339, f.Departure.Time)
			if err != nil {
				warnSkip(ctx, p.Name(), f.FlightID, "parse departure time", err)
				continue
			}
			arr, err = time.Parse(time.RFC3339, f.Arrival.Time)
			if err != nil {
				warnSkip(ctx, p.Name(), f.FlightID, "parse arrival time", err)
				continue
			}
			stops = f.Stops
			totalDuration = f.DurationMinutes
			arrivalAirport = f.Arrival.Airport
			arrivalCity = f.Arrival.City
		}

		if arr.Before(dep) {
			warnSkip(ctx, p.Name(), f.FlightID, "arrival before departure", nil)
			continue
		}

		cabinClass := normalizeCabinClass(f.FareClass)

		if !matchesSearchCriteria(req, f.Departure.Airport, arrivalAirport, dep, f.AvailableSeats, cabinClass) {
			continue
		}

		flight := domain.Flight{
			ID:             fmt.Sprintf("%s_%s", f.FlightID, p.Name()),
			Provider:       p.Name(),
			FlightNumber:   f.FlightID,
			Stops:          stops,
			AvailableSeats: f.AvailableSeats,
			CabinClass:     cabinClass,
			Aircraft:       f.Aircraft,
			TotalMinutes:   totalDuration,
			Amenities:      ensureSlice(f.Amenities),
			Baggage: domain.Baggage{
				CarryOn: fmt.Sprintf("%d", f.Baggage.CarryOn),
				Checked: fmt.Sprintf("%d", f.Baggage.Checked),
			},
			Airline: domain.Airline{
				Name: f.Airline,
				Code: f.AirlineCode,
			},
			Departure: domain.Location{
				Airport:   f.Departure.Airport,
				City:      f.Departure.City,
				Datetime:  dep.UTC(),
				Timestamp: dep.Unix(),
			},
			Arrival: domain.Location{
				Airport:   arrivalAirport,
				City:      arrivalCity,
				Datetime:  arr.UTC(),
				Timestamp: arr.Unix(),
			},
			Price: domain.Price{
				Amount:   f.Price.Amount,
				Currency: "IDR",
			},
		}
		flights = append(flights, flight)
	}
	return flights, nil
}

func ensureSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
