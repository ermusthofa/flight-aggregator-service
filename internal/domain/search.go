package domain

import (
	"errors"
	"strings"
	"time"
)

type SearchRequest struct {
	// SEARCH (required)
	Origin        string `json:"origin"`
	Destination   string `json:"destination"`
	DepartureDate string `json:"departureDate"`
	Passengers    int    `json:"passengers"`
	CabinClass    string `json:"cabinClass"`

	// FILTER (optional)
	MinPrice int `json:"minPrice"`
	MaxPrice int `json:"maxPrice"`

	MaxStops *int `json:"maxStops"`

	DepartureTimeFrom string `json:"departureTimeFrom"` // "15:04"
	DepartureTimeTo   string `json:"departureTimeTo"`

	ArrivalTimeFrom string `json:"arrivalTimeFrom"`
	ArrivalTimeTo   string `json:"arrivalTimeTo"`

	Airlines    []string `json:"airlines"`    // ["GA", "QZ"]
	MaxDuration int      `json:"maxDuration"` // minutes

	// SORT (optional)
	SortBy string `json:"sortBy"` // price, duration, departure
}

func (r *SearchRequest) Validate() error {
	if r.Origin == "" {
		return errors.New("origin is required")
	}
	if r.Destination == "" {
		return errors.New("destination is required")
	}
	if r.DepartureDate == "" {
		return errors.New("departureDate is required")
	}
	if r.Passengers <= 0 {
		return errors.New("passengers must be > 0")
	}

	// time format validation
	if !isValidClock(r.DepartureTimeFrom) {
		return errors.New("invalid departureTimeFrom format (HH:mm)")
	}
	if !isValidClock(r.DepartureTimeTo) {
		return errors.New("invalid departureTimeTo format (HH:mm)")
	}
	if !isValidClock(r.ArrivalTimeFrom) {
		return errors.New("invalid arrivalTimeFrom format (HH:mm)")
	}
	if !isValidClock(r.ArrivalTimeTo) {
		return errors.New("invalid arrivalTimeTo format (HH:mm)")
	}

	// duration sanity
	if r.MaxDuration < 0 {
		return errors.New("maxDuration cannot be negative")
	}

	return nil
}

func (r *SearchRequest) Normalize() {
	// normalize airport codes
	r.Origin = strings.ToUpper(strings.TrimSpace(r.Origin))
	r.Destination = strings.ToUpper(strings.TrimSpace(r.Destination))

	// normalize cabin class
	r.CabinClass = strings.ToLower(strings.TrimSpace(r.CabinClass))

	// fix negative prices
	if r.MinPrice < 0 {
		r.MinPrice = 0
	}
	if r.MaxPrice < 0 {
		r.MaxPrice = 0
	}

	// fix price range
	if r.MaxPrice > 0 && r.MinPrice > r.MaxPrice {
		r.MinPrice, r.MaxPrice = r.MaxPrice, r.MinPrice
	}

	// stops normalization
	if r.MaxStops != nil {
		if *r.MaxStops < 0 {
			val := 0
			r.MaxStops = &val
		}
		if *r.MaxStops > 3 {
			val := 3
			r.MaxStops = &val
		}
	}

	// normalize time strings (trim only)
	r.DepartureTimeFrom = strings.TrimSpace(r.DepartureTimeFrom)
	r.DepartureTimeTo = strings.TrimSpace(r.DepartureTimeTo)
	r.ArrivalTimeFrom = strings.TrimSpace(r.ArrivalTimeFrom)
	r.ArrivalTimeTo = strings.TrimSpace(r.ArrivalTimeTo)

	// normalize airlines (uppercase + dedupe)
	if len(r.Airlines) > 0 {
		seen := make(map[string]struct{})
		normalized := make([]string, 0, len(r.Airlines))

		for _, a := range r.Airlines {
			a = strings.ToUpper(strings.TrimSpace(a))
			if a == "" {
				continue
			}
			if _, exists := seen[a]; !exists {
				seen[a] = struct{}{}
				normalized = append(normalized, a)
			}
		}

		r.Airlines = normalized
	}

	// duration
	if r.MaxDuration < 0 {
		r.MaxDuration = 0
	}

	// normalize sort
	switch r.SortBy {
	case "price", "duration", "departure", "arrival", "best":
		// ok
	default:
		r.SortBy = "best"
	}
}

func isValidClock(value string) bool {
	if value == "" {
		return true
	}
	_, err := time.Parse("15:04", value)
	return err == nil
}
