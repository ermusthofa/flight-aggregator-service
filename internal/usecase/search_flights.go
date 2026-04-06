package usecase

import "github.com/ermusthofa/flight-aggregator-service/internal/domain"

type SearchFlightsUsecase struct{}

func NewSearchFlightsUsecase() *SearchFlightsUsecase {
	return &SearchFlightsUsecase{}
}

func (u *SearchFlightsUsecase) Execute() ([]domain.Flight, error) {
	// temporary dummy response
	flights := []domain.Flight{
		{
			ID:            "dummy_1",
			Provider:      "mock",
			FlightNumber:  "XX123",
			Origin:        "CGK",
			Destination:   "DPS",
			DepartureTime: "2025-12-15T10:00:00+07:00",
			ArrivalTime:   "2025-12-15T12:00:00+08:00",
			Price:         500000,
		},
	}

	return flights, nil
}
