package usecase

import (
	"sort"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
)

type SorterEngine struct{}

func NewSorterEngine() *SorterEngine {
	return &SorterEngine{}
}

func (s *SorterEngine) Sort(flights []domain.Flight, sortBy string) {
	sort.SliceStable(flights, func(i, j int) bool {

		switch sortBy {

		case "price":
			if flights[i].Price.Amount == flights[j].Price.Amount {
				return flights[i].TotalMinutes < flights[j].TotalMinutes
			}
			return flights[i].Price.Amount < flights[j].Price.Amount

		case "duration":
			if flights[i].TotalMinutes == flights[j].TotalMinutes {
				return flights[i].Price.Amount < flights[j].Price.Amount
			}
			return flights[i].TotalMinutes < flights[j].TotalMinutes

		case "departure":
			return flights[i].Departure.Timestamp < flights[j].Departure.Timestamp

		case "arrival":
			return flights[i].Arrival.Timestamp < flights[j].Arrival.Timestamp

		case "best":
			return flights[i].Score < flights[j].Score

		default:
			return flights[i].Score < flights[j].Score
		}
	})
}
