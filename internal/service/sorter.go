package service

import (
	"sort"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
)

func SortFlights(flights []domain.Flight, sortBy string) {
	switch sortBy {

	case "price":
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Price.Amount < flights[j].Price.Amount
		})

	case "duration":
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].DurationMinutes < flights[j].DurationMinutes
		})

	case "departure":
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Departure.Timestamp < flights[j].Departure.Timestamp
		})

	case "best":
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Score < flights[j].Score
		})
	}
}
