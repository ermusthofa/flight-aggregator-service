package usecase

import (
	"context"

	"github.com/ermusthofa/flight-aggregator-service/internal/cache"
	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
	"github.com/ermusthofa/flight-aggregator-service/internal/provider"
	"github.com/ermusthofa/flight-aggregator-service/internal/service"
)

type SearchFlightsUsecase struct {
	aggregator *service.Aggregator
}

func NewSearchFlightsUsecase(c cache.Cache) *SearchFlightsUsecase {
	providers := []provider.Provider{
		provider.NewAirAsiaProvider(),
		provider.NewGarudaProvider(),
		provider.NewLionProvider(),
		provider.NewBatikProvider(),
	}

	return &SearchFlightsUsecase{
		aggregator: service.NewAggregator(providers, c),
	}
}

func (u *SearchFlightsUsecase) Execute(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, domain.Metadata, error) {
	flights, meta := u.aggregator.Search(ctx, req)
	return flights, meta, nil
}
