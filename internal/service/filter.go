package service

import (
	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
)

func FilterFlights(flights []domain.Flight, req domain.SearchRequest) []domain.Flight {
	var result []domain.Flight

	for _, f := range flights {
		// price filter
		if req.MinPrice > 0 && f.Price.Amount < req.MinPrice {
			continue
		}
		if req.MaxPrice > 0 && f.Price.Amount > req.MaxPrice {
			continue
		}

		// stops filter
		if req.MaxStops != nil && f.Stops > *req.MaxStops {
			continue
		}

		result = append(result, f)
	}

	return result
}
