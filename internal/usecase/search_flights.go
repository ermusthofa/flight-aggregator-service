package usecase

import (
	"context"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
	"github.com/ermusthofa/flight-aggregator-service/internal/provider"
	"github.com/ermusthofa/flight-aggregator-service/internal/service"
)

type SearchFlightsUsecase struct {
	aggregator *service.Aggregator
}

func NewSearchFlightsUsecase() *SearchFlightsUsecase {
	providers := []provider.Provider{
		provider.NewAirAsiaProvider(),
		// TODO: add others here later
	}

	return &SearchFlightsUsecase{
		aggregator: service.NewAggregator(providers),
	}
}

func (u *SearchFlightsUsecase) Execute(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, service.Metadata, error) {
	flights, meta := u.aggregator.Search(ctx, req)
	return flights, meta, nil
}
