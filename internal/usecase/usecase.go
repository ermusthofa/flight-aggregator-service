package usecase

import (
	"context"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
)

type SearchFlightsUsecaseInterface interface {
	Execute(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, domain.Metadata, error)
}
