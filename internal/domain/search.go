package domain

import "errors"

type SearchRequest struct {
	Origin        string `json:"origin"`
	Destination   string `json:"destination"`
	DepartureDate string `json:"departureDate"`
	Passengers    int    `json:"passengers"`
	CabinClass    string `json:"cabinClass"`

	MinPrice int `json:"minPrice"`
	MaxPrice int `json:"maxPrice"`

	MaxStops *int `json:"maxStops"`

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
	return nil
}

func (r *SearchRequest) Normalize() {
	// fix negative prices
	if r.MinPrice < 0 {
		r.MinPrice = 0
	}

	if r.MaxPrice < 0 {
		r.MaxPrice = 0
	}

	// fix range
	if r.MaxPrice > 0 && r.MinPrice > r.MaxPrice {
		r.MinPrice, r.MaxPrice = r.MaxPrice, r.MinPrice
	}

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

	// normalize sort
	switch r.SortBy {
	case "price", "duration", "departure", "best":
		// ok
	default:
		r.SortBy = "best"
	}
}
