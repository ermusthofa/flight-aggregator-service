package domain

import "errors"

type SearchRequest struct {
	Origin        string `json:"origin"`
	Destination   string `json:"destination"`
	DepartureDate string `json:"departureDate"`
	Passengers    int    `json:"passengers"`
	CabinClass    string `json:"cabinClass"`
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
