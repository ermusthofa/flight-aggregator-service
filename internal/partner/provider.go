package partner

import (
	"context"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
)

type Provider interface {
	Name() string
	Search(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, error)
}
