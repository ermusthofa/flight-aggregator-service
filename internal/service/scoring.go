package service

import "github.com/ermusthofa/flight-aggregator-service/internal/domain"

func ScoreFlights(flights []domain.Flight) {
	for i := range flights {
		price := flights[i].Price.Amount
		duration := flights[i].TotalMinutes

		// simple weighted score
		score := price + (duration * 1000)

		flights[i].Score = score
	}
}
