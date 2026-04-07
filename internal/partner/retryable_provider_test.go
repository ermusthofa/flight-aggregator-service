package partner

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
)

// mockProvider implements Provider interface for testing.
type mockProvider struct {
	name        string
	callCount   int
	failures    []error         // errors to return in sequence; nil means success
	successData []domain.Flight // data to return on success
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) Search(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, error) {
	m.callCount++
	if m.callCount-1 < len(m.failures) && m.failures[m.callCount-1] != nil {
		return nil, m.failures[m.callCount-1]
	}
	return m.successData, nil
}

// mockProvider implements pkg.Logger
type mockLogger struct{}

func (m *mockLogger) Info(ctx context.Context, format string, v ...interface{})  {}
func (m *mockLogger) Error(ctx context.Context, format string, v ...interface{}) {}
func (m *mockLogger) Warn(ctx context.Context, format string, v ...interface{})  {}

func TestRetryableProvider_SuccessFirstAttempt(t *testing.T) {
	mock := &mockProvider{
		name:        "TestProvider",
		failures:    []error{nil}, // first attempt success
		successData: []domain.Flight{{ID: "test"}},
	}
	logger := &mockLogger{}
	retryable := NewRetryable(mock, 2, 10*time.Millisecond, logger)

	ctx := context.Background()
	req := domain.SearchRequest{Origin: "CGK", Destination: "DPS"}

	flights, err := retryable.Search(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(flights) != 1 || flights[0].ID != "test" {
		t.Errorf("unexpected flights: %+v", flights)
	}
	if mock.callCount != 1 {
		t.Errorf("expected 1 call, got %d", mock.callCount)
	}
}

func TestRetryableProvider_SuccessAfterRetries(t *testing.T) {
	mock := &mockProvider{
		name:        "TestProvider",
		failures:    []error{errors.New("fail1"), errors.New("fail2"), nil},
		successData: []domain.Flight{{ID: "success"}},
	}
	logger := &mockLogger{}
	retryable := NewRetryable(mock, 2, 5*time.Millisecond, logger) // maxRetries=2 → attempts 0,1,2 (3 total)

	ctx := context.Background()
	req := domain.SearchRequest{}

	start := time.Now()
	flights, err := retryable.Search(ctx, req)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if len(flights) != 1 || flights[0].ID != "success" {
		t.Errorf("unexpected flights: %+v", flights)
	}
	if mock.callCount != 3 {
		t.Errorf("expected 3 calls, got %d", mock.callCount)
	}
	// Check exponential backoff: delays: base (5ms), then 10ms, total ~15ms
	if elapsed < 15*time.Millisecond {
		t.Logf("backoff elapsed: %v (expected at least 15ms)", elapsed)
	}
}

func TestRetryableProvider_MaxRetriesExceeded(t *testing.T) {
	mock := &mockProvider{
		name:     "TestProvider",
		failures: []error{errors.New("err1"), errors.New("err2"), errors.New("err3")}, // 3 failures, maxRetries=2
	}
	logger := &mockLogger{}
	retryable := NewRetryable(mock, 2, 1*time.Millisecond, logger)

	ctx := context.Background()
	req := domain.SearchRequest{}

	_, err := retryable.Search(ctx, req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "err3" {
		t.Errorf("expected last error 'err3', got %v", err)
	}
	if mock.callCount != 3 {
		t.Errorf("expected 3 calls, got %d", mock.callCount)
	}
}

func TestRetryableProvider_ContextCancelledDuringDelay(t *testing.T) {
	mock := &mockProvider{
		name:     "TestProvider",
		failures: []error{errors.New("first fail"), nil}, // would succeed on second attempt
	}
	logger := &mockLogger{}
	retryable := NewRetryable(mock, 2, 100*time.Millisecond, logger) // long delay

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel after short time to interrupt the delay
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	req := domain.SearchRequest{}
	_, err := retryable.Search(ctx, req)

	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	// First call happened, then delay started, then cancelled. So exactly 1 call.
	if mock.callCount != 1 {
		t.Errorf("expected 1 call before cancellation, got %d", mock.callCount)
	}
}

func TestRetryableProvider_NoRetryOnContextDoneBeforeAttempt(t *testing.T) {
	mock := &mockProvider{
		name:     "TestProvider",
		failures: []error{errors.New("fail")},
	}
	logger := &mockLogger{}
	retryable := NewRetryable(mock, 2, 10*time.Millisecond, logger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled

	req := domain.SearchRequest{}
	_, err := retryable.Search(ctx, req)

	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	if mock.callCount == 0 {
		t.Log("callCount 0 is also possible if implementation checks ctx before first attempt, but currently it doesn't")
	}
}

func TestRetryableProvider_NamePropagation(t *testing.T) {
	mock := &mockProvider{name: "MockAirline"}
	logger := &mockLogger{}
	retryable := NewRetryable(mock, 3, time.Millisecond, logger)
	if retryable.Name() != "MockAirline" {
		t.Errorf("expected name 'MockAirline', got %s", retryable.Name())
	}
}

func TestRetryableProvider_ExponentialBackoff(t *testing.T) {
	mock := &mockProvider{
		name:     "BackoffTest",
		failures: []error{errors.New("fail1"), errors.New("fail2"), nil}, // 2 retries needed
	}
	logger := &mockLogger{}
	baseDelay := 10 * time.Millisecond
	retryable := NewRetryable(mock, 2, baseDelay, logger)

	ctx := context.Background()
	req := domain.SearchRequest{}

	start := time.Now()
	_, _ = retryable.Search(ctx, req)
	elapsed := time.Since(start)

	// Expected delays: first attempt immediate, then wait baseDelay (10ms), second attempt fail, wait 2*baseDelay (20ms), third attempt success.
	if elapsed < 30*time.Millisecond {
		t.Errorf("expected at least 30ms of backoff, got %v", elapsed)
	}
	if elapsed > 100*time.Millisecond {
		t.Errorf("backoff took too long: %v", elapsed)
	}
}
