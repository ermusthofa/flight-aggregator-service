package partner

import (
	"context"
	"time"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
	"github.com/ermusthofa/flight-aggregator-service/internal/pkg"
)

// RetryableProvider wraps a provider with exponential backoff
type RetryableProvider struct {
	inner      Provider
	maxRetries int
	baseDelay  time.Duration
	logger     pkg.Logger
}

func NewRetryable(inner Provider, maxRetries int, baseDelay time.Duration, logger pkg.Logger) *RetryableProvider {
	return &RetryableProvider{inner: inner, maxRetries: maxRetries, baseDelay: baseDelay, logger: logger}
}

func (r *RetryableProvider) Name() string { return r.inner.Name() }

func (r *RetryableProvider) Search(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, error) {
	var lastErr error
	delay := r.baseDelay

	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		flights, err := r.inner.Search(ctx, req)
		if err == nil {
			return flights, nil
		}
		lastErr = err
		if attempt == r.maxRetries {
			break
		}
		r.logger.Warn(ctx, "provider %s attempt %d/%d failed: %v, retrying in %v", r.inner.Name(), attempt+1, r.maxRetries, err, delay)
		select {
		case <-time.After(delay):
			delay *= 2
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return nil, lastErr
}
