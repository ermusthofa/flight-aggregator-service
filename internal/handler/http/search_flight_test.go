package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
	"github.com/ermusthofa/flight-aggregator-service/internal/pkg"
)

// mockSearchUsecase implements the Execute method for testing
type mockSearchUsecase struct {
	executeFunc func(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, domain.Metadata, error)
}

func (m *mockSearchUsecase) Execute(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, domain.Metadata, error) {
	return m.executeFunc(ctx, req)
}

func TestHandler_ServeHTTP_Routing(t *testing.T) {
	uc := &mockSearchUsecase{}
	handler := NewHandler(uc)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "POST /search - valid",
			method:         "POST",
			path:           "/search",
			expectedStatus: http.StatusOK,
			expectedBody:   "", // we'll check later
		},
		{
			name:           "GET /search - method not allowed",
			method:         "GET",
			path:           "/search",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   `{"error":"method not allowed"}`,
		},
		{
			name:           "POST /unknown - not found",
			method:         "POST",
			path:           "/unknown",
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For the successful POST, we need to set up the mock to return something
			if tt.name == "POST /search - valid" {
				uc.executeFunc = func(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, domain.Metadata, error) {
					return []domain.Flight{}, domain.Metadata{}, nil
				}
			}
			reqBody := bytes.NewReader([]byte(`{"origin":"CGK","destination":"DPS","departureDate":"2025-12-15","passengers":1}`))
			req := httptest.NewRequest(tt.method, tt.path, reqBody)
			if tt.method == "POST" && tt.name == "POST /search - valid" {
				req.Header.Set("Content-Type", "application/json")
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			if tt.expectedBody != "" {
				actual := strings.TrimSpace(w.Body.String())
				if actual != tt.expectedBody {
					t.Errorf("expected body %s, got %s", tt.expectedBody, actual)
				}
			}
		})
	}
}

func TestHandler_SearchFlights(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockExecute    func(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, domain.Metadata, error)
		expectedStatus int
		expectedError  string // if non-empty, check error field in response
	}{
		{
			name:        "successful search",
			requestBody: `{"origin":"CGK","destination":"DPS","departureDate":"2025-12-15","passengers":1}`,
			mockExecute: func(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, domain.Metadata, error) {
				return []domain.Flight{
					{
						ID:           "test_flight",
						FlightNumber: "GA123",
						Provider:     "Garuda",
						Price:        domain.Price{Amount: 1000000, Currency: "IDR"},
					},
				}, domain.Metadata{TotalResults: 1, SearchTimeMs: 123}, nil
			},
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name:           "invalid json body",
			requestBody:    `{"origin": "CGK", invalid}`,
			mockExecute:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name:           "validation error (missing origin)",
			requestBody:    `{"destination":"DPS","departureDate":"2025-12-15","passengers":1}`,
			mockExecute:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "origin is required",
		},
		{
			name:        "usecase error",
			requestBody: `{"origin":"CGK","destination":"DPS","departureDate":"2025-12-15","passengers":1}`,
			mockExecute: func(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, domain.Metadata, error) {
				return nil, domain.Metadata{}, errors.New("provider timeout")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "search failed: provider timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &mockSearchUsecase{}
			if tt.mockExecute != nil {
				uc.executeFunc = tt.mockExecute
			}
			handler := NewHandler(uc)
			req := httptest.NewRequest("POST", "/search", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			handler.SearchFlights(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var resp map[string]interface{}
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if tt.expectedError != "" {
				errMsg, ok := resp["error"]
				if !ok {
					t.Errorf("expected error field in response, got %v", resp)
				} else if !contains(errMsg.(string), tt.expectedError) {
					t.Errorf("expected error containing %q, got %q", tt.expectedError, errMsg)
				}
				return
			}

			// For success, check response structure
			if _, ok := resp["search_criteria"]; !ok {
				t.Error("missing search_criteria in response")
			}
			if _, ok := resp["metadata"]; !ok {
				t.Error("missing metadata in response")
			}
			if _, ok := resp["flights"]; !ok {
				t.Error("missing flights in response")
			}
			// Optionally check metadata values
			meta := resp["metadata"].(map[string]interface{})
			if meta["total_results"] != float64(1) {
				t.Errorf("expected total_results 1, got %v", meta["total_results"])
			}
		})
	}
}

func TestHandler_RequestIDHeader(t *testing.T) {
	// Verify that X-Request-ID is passed to context and logged
	receivedID := ""
	uc := &mockSearchUsecase{
		executeFunc: func(ctx context.Context, req domain.SearchRequest) ([]domain.Flight, domain.Metadata, error) {
			receivedID = pkg.GetRequestID(ctx)
			return []domain.Flight{}, domain.Metadata{}, nil
		},
	}
	handler := NewHandler(uc)
	reqBody := bytes.NewReader([]byte(`{"origin":"CGK","destination":"DPS","departureDate":"2025-12-15","passengers":1}`))
	req := httptest.NewRequest("POST", "/search", reqBody)
	req.Header.Set("X-Request-ID", "test-123")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if receivedID != "test-123" {
		t.Errorf("expected request ID 'test-123', got '%s'", receivedID)
	}
}

// Helper to check substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && s[:len(substr)] == substr) || (len(s) > len(substr) && contains(s[1:], substr)))
}
