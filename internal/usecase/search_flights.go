package usecase

import (
	"context"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
	"github.com/ermusthofa/flight-aggregator-service/internal/provider"
)

type SearchFlightsUsecase struct {
	airasia *provider.AirAsiaProvider
}

func NewSearchFlightsUsecase() *SearchFlightsUsecase {
	return &SearchFlightsUsecase{
		airasia: provider.NewAirAsiaProvider(),
	}
}

func (u *SearchFlightsUsecase) Execute(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, error) {
	return u.airasia.Search(ctx, req)
}
