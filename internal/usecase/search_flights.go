package usecase

import (
	"context"
	"time"

	"github.com/ermusthofa/flight-aggregator-service/internal/cache"
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
		provider.NewGarudaProvider(),
		provider.NewLionProvider(),
		provider.NewBatikProvider(),
	}

	return &SearchFlightsUsecase{
		aggregator: service.NewAggregator(
			providers,
			cache.NewMemoryCache(60*time.Second), // 60 sec cache
		),
	}
}

func (u *SearchFlightsUsecase) Execute(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, service.Metadata, error) {
	flights, meta := u.aggregator.Search(ctx, req)
	return flights, meta, nil
}
