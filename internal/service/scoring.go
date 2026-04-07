package service

import "github.com/ermusthofa/flight-aggregator-service/internal/domain"

const (
	priceWeight    = 0.5
	durationWeight = 0.3
	stopsWeight    = 0.2
)

func ScoreFlights(flights []domain.Flight) {
	if len(flights) == 0 {
		return
	}

	// find min/max
	minPrice, maxPrice := flights[0].Price.Amount, flights[0].Price.Amount
	minDuration, maxDuration := flights[0].TotalMinutes, flights[0].TotalMinutes

	for _, f := range flights {
		if f.Price.Amount < minPrice {
			minPrice = f.Price.Amount
		}
		if f.Price.Amount > maxPrice {
			maxPrice = f.Price.Amount
		}
		if f.TotalMinutes < minDuration {
			minDuration = f.TotalMinutes
		}
		if f.TotalMinutes > maxDuration {
			maxDuration = f.TotalMinutes
		}
	}

	// avoid divide by zero
	priceRange := float64(maxPrice - minPrice)
	if priceRange == 0 {
		priceRange = 1
	}

	durationRange := float64(maxDuration - minDuration)
	if durationRange == 0 {
		durationRange = 1
	}

	// compute score
	for i := range flights {
		price := float64(flights[i].Price.Amount)
		duration := float64(flights[i].TotalMinutes)

		nPrice := (price - float64(minPrice)) / priceRange
		nDuration := (duration - float64(minDuration)) / durationRange

		stopsPenalty := float64(flights[i].Stops) * stopsWeight

		score := (nPrice * priceWeight) + (nDuration * durationWeight) + stopsPenalty

		// convert to int for sorting
		flights[i].Score = int(score * 1000)
	}
}
