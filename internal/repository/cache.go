package repository

import (
	"time"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
)

type Cache interface {
	Get(key string) ([]domain.Flight, bool)
	Set(key string, data []domain.Flight, ttl time.Duration)
}
