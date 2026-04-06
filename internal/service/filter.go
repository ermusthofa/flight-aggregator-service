package service

import (
	"time"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
)

func FilterFlights(flights []domain.Flight, req domain.SearchRequest) []domain.Flight {
	result := make([]domain.Flight, 0, len(flights))

	var depFrom, depTo, arrFrom, arrTo *int

	// Pre-parse
	if req.DepartureTimeFrom != "" {
		h, m, _ := parseClock(req.DepartureTimeFrom)
		val := h*60 + m
		depFrom = &val
	}
	if req.DepartureTimeTo != "" {
		h, m, _ := parseClock(req.DepartureTimeTo)
		val := h*60 + m
		depTo = &val
	}

	if req.ArrivalTimeFrom != "" {
		h, m, _ := parseClock(req.ArrivalTimeFrom)
		val := h*60 + m
		arrFrom = &val
	}
	if req.ArrivalTimeTo != "" {
		h, m, _ := parseClock(req.ArrivalTimeTo)
		val := h*60 + m
		arrTo = &val
	}

	for _, f := range flights {

		// Price
		if req.MinPrice > 0 && f.Price.Amount < req.MinPrice {
			continue
		}
		if req.MaxPrice > 0 && f.Price.Amount > req.MaxPrice {
			continue
		}

		// Stops
		if req.MaxStops != nil && f.Stops > *req.MaxStops {
			continue
		}

		// Departure time
		depMinutes := minutesOfDay(f.Departure.Datetime)

		if depFrom != nil && depMinutes < *depFrom {
			continue
		}
		if depTo != nil && depMinutes > *depTo {
			continue
		}

		// Arrival time
		arrMinutes := minutesOfDay(f.Arrival.Datetime)

		if arrFrom != nil && arrMinutes < *arrFrom {
			continue
		}
		if arrTo != nil && arrMinutes > *arrTo {
			continue
		}

		// Airline
		if len(req.Airlines) > 0 {
			match := false
			for _, code := range req.Airlines {
				if f.Airline.Code == code {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		// Duration
		if req.MaxDuration > 0 && f.Duration.TotalMinutes > req.MaxDuration {
			continue
		}

		result = append(result, f)
	}

	return result
}

func parseClock(value string) (int, int, error) {
	t, err := time.Parse("15:04", value)
	if err != nil {
		return 0, 0, err
	}
	return t.Hour(), t.Minute(), nil
}

func minutesOfDay(t time.Time) int {
	return t.Hour()*60 + t.Minute()
}
