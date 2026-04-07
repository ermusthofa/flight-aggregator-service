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

type LionProvider struct {
	limiter *ratelimit.RateLimiter
}

func NewLionProvider() *LionProvider {
	// Allow 100 requests per second
	return &LionProvider{limiter: ratelimit.New(100, time.Second)}
}

type lionResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Flights []struct {
			ID      string `json:"id"`
			Carrier struct {
				Name string `json:"name"`
				IATA string `json:"iata"`
			} `json:"carrier"`
			Route struct {
				From struct {
					Code string `json:"code"`
					City string `json:"city"`
				} `json:"from"`
				To struct {
					Code string `json:"code"`
					City string `json:"city"`
				} `json:"to"`
			} `json:"route"`
			Schedule struct {
				Departure         string `json:"departure"`
				DepartureTimezone string `json:"departure_timezone"`
				Arrival           string `json:"arrival"`
				ArrivalTimezone   string `json:"arrival_timezone"`
			} `json:"schedule"`
			FlightTime int  `json:"flight_time"`
			IsDirect   bool `json:"is_direct"`
			StopCount  int  `json:"stop_count"`
			Layovers   []struct {
				Airport  string `json:"airport"`
				Duration int    `json:"duration_minutes"`
			} `json:"layovers"`
			Pricing struct {
				Total    int    `json:"total"`
				FareType string `json:"fare_type"`
			} `json:"pricing"`
			Seats     int    `json:"seats_left"`
			PlaneType string `json:"plane_type"`
			Services  struct {
				WifiAvailable    bool `json:"wifi_available"`
				MealsIncluded    bool `json:"meals_included"`
				BaggageAllowance struct {
					Cabin string `json:"cabin"`
					Hold  string `json:"hold"`
				} `json:"baggage_allowance"`
			} `json:"services"`
		} `json:"available_flights"`
	} `json:"data"`
}

func (p *LionProvider) Name() string { return "Lion" }

func (p *LionProvider) Search(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, error) {
	if !p.limiter.Allow() {
		return nil, fmt.Errorf("%s API rate limit exceeded", p.Name())
	}

	// Simulate delay 100-200ms with context awareness
	delay := time.Duration(100+rand.Intn(100)) * time.Millisecond
	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	var resp lionResponse
	if err := json.Unmarshal(mock.LionMock, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse lion mock: %w", err)
	}

	var flights []domain.Flight
	for _, f := range resp.Data.Flights {
		// Parse times with timezone mapping (WIB/WITA/WIT)
		dep, err := parseWithTimezone(f.Schedule.Departure, f.Schedule.DepartureTimezone)
		if err != nil {
			warnSkip(ctx, p.Name(), f.ID, "parse departure time", err)
			continue
		}
		arr, err := parseWithTimezone(f.Schedule.Arrival, f.Schedule.ArrivalTimezone)
		if err != nil {
			warnSkip(ctx, p.Name(), f.ID, "parse arrival time", err)
			continue
		}

		if arr.Before(dep) {
			warnSkip(ctx, p.Name(), f.ID, "arrival before departure", err)
			continue
		}

		cabinClass := normalizeCabinClass(f.Pricing.FareType)

		// Use fixed helper with passenger count check
		if !matchesSearchCriteria(req, f.Route.From.Code, f.Route.To.Code, dep, f.Seats, cabinClass) {
			continue
		}

		stops := 0
		if !f.IsDirect {
			stops = f.StopCount
		}

		flight := domain.Flight{
			ID:             fmt.Sprintf("%s_%s", f.ID, p.Name()),
			Provider:       p.Name(),
			FlightNumber:   f.ID,
			Stops:          stops,
			AvailableSeats: f.Seats,
			CabinClass:     cabinClass,
			Aircraft:       f.PlaneType,
			TotalMinutes:   int(arr.Sub(dep).Minutes()),
			Amenities:      mapLionServices(f.Services.WifiAvailable, f.Services.MealsIncluded),
			Baggage: domain.Baggage{
				CarryOn: f.Services.BaggageAllowance.Cabin,
				Checked: f.Services.BaggageAllowance.Hold,
			},
			Airline: domain.Airline{
				Name: f.Carrier.Name,
				Code: f.Carrier.IATA,
			},
			Departure: domain.Location{
				Airport:   f.Route.From.Code,
				City:      f.Route.From.City,
				Datetime:  dep.UTC(),
				Timestamp: dep.Unix(),
			},
			Arrival: domain.Location{
				Airport:   f.Route.To.Code,
				City:      f.Route.To.City,
				Datetime:  arr.UTC(),
				Timestamp: arr.Unix(),
			},
			Price: domain.Price{
				Amount:   f.Pricing.Total,
				Currency: "IDR",
			},
		}
		flights = append(flights, flight)
	}
	return flights, nil
}

func mapLionServices(wifi, meals bool) []string {
	var amenities []string
	if wifi {
		amenities = append(amenities, "wifi")
	}
	if meals {
		amenities = append(amenities, "meals")
	}
	return amenities
}
