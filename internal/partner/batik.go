package partner

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"time"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
	"github.com/ermusthofa/flight-aggregator-service/internal/partner/mock"
	ratelimit "github.com/ermusthofa/flight-aggregator-service/internal/partner/ratelimiter"
)

type BatikProvider struct {
	limiter *ratelimit.RateLimiter
}

func NewBatikProvider() *BatikProvider {
	// 2 requests per minute
	return &BatikProvider{limiter: ratelimit.New(2, time.Minute)}
}

type batikResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Results []struct {
		FlightNumber string `json:"flightNumber"`
		AirlineName  string `json:"airlineName"`
		AirlineIATA  string `json:"airlineIATA"`
		Origin       string `json:"origin"`
		Destination  string `json:"destination"`
		Departure    string `json:"departureDateTime"`
		Arrival      string `json:"arrivalDateTime"`
		TravelTime   string `json:"travelTime"` // e.g., "1h 45m"
		Stops        int    `json:"numberOfStops"`
		Connections  []struct {
			StopAirport  string `json:"stopAirport"`
			StopDuration string `json:"stopDuration"`
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

func (p *BatikProvider) Name() string { return "Batik" }

func (p *BatikProvider) Search(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, error) {
	if !p.limiter.Allow() {
		return nil, fmt.Errorf("%s API rate limit exceeded", p.Name())
	}

	// simulate delay (200–400ms)
	delay := time.Duration(200+rand.Intn(200)) * time.Millisecond
	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	var resp batikResponse
	if err := json.Unmarshal(mock.BatikMock, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse batik mock: %w", err)
	}

	var flights []domain.Flight
	for _, f := range resp.Results {
		dep, err := time.Parse("2006-01-02T15:04:05-0700", f.Departure)
		if err != nil {
			warnSkip(ctx, f.FlightNumber, "parse departure time", err)
			continue
		}
		arr, err := time.Parse("2006-01-02T15:04:05-0700", f.Arrival)
		if err != nil {
			warnSkip(ctx, f.FlightNumber, "parse arrival time", err)
			continue
		}

		if arr.Before(dep) {
			warnSkip(ctx, f.FlightNumber, "arrival before departure", nil)
			continue
		}

		cabinClass := normalizeCabinClass(f.Fare.Class)

		if !matchesSearchCriteria(req, f.Origin, f.Destination, dep, f.Seats, cabinClass) {
			continue
		}

		// time diff, fallback to parsing TravelTime string
		durationMinutes := int(arr.Sub(dep).Minutes())
		if durationMinutes <= 0 && f.TravelTime != "" {
			durationMinutes = parseDurationString(f.TravelTime)
		}

		flight := domain.Flight{
			ID:             fmt.Sprintf("%s_%s", f.FlightNumber, p.Name()),
			Provider:       p.Name(),
			FlightNumber:   f.FlightNumber,
			Stops:          f.Stops,
			AvailableSeats: f.Seats,
			CabinClass:     cabinClass,
			Aircraft:       f.AircraftModel,
			TotalMinutes:   durationMinutes,
			Amenities:      f.OnboardServices,
			Baggage:        parseBaggageNote(f.BaggageInfo),
			Airline: domain.Airline{
				Name: f.AirlineName,
				Code: f.AirlineIATA,
			},
			Departure: domain.Location{
				Airport:   f.Origin,
				City:      airportToCity(f.Origin),
				Datetime:  dep.UTC(),
				Timestamp: dep.Unix(),
			},
			Arrival: domain.Location{
				Airport:   f.Destination,
				City:      airportToCity(f.Destination),
				Datetime:  arr.UTC(),
				Timestamp: arr.Unix(),
			},
			Price: domain.Price{
				Amount:   f.Fare.TotalPrice,
				Currency: "IDR",
			},
		}
		flights = append(flights, flight)
	}
	return flights, nil
}

// parseDurationString converts strings like "1h 45m" or "2h" into minutes.
func parseDurationString(s string) int {
	total := 0
	// Match numbers followed by 'h' or 'm'
	re := regexp.MustCompile(`(\d+)([hm])`)
	matches := re.FindAllStringSubmatch(s, -1)
	for _, match := range matches {
		if len(match) != 3 {
			continue
		}
		val, _ := strconv.Atoi(match[1])
		unit := match[2]
		switch unit {
		case "h":
			total += val * 60
		case "m":
			total += val
		}
	}
	return total
}
