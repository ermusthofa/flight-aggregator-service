package http

import (
	"encoding/json"
	"net/http"

	"github.com/ermusthofa/flight-aggregator-service/internal/pkg"
	"github.com/ermusthofa/flight-aggregator-service/internal/usecase"
	"github.com/google/uuid"
)

type Handler struct {
	searchUC *usecase.SearchFlightsUsecase
}

func NewHandler(uc *usecase.SearchFlightsUsecase) *Handler {
	return &Handler{searchUC: uc}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Generate or extract request ID
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = uuid.New().String()
	}
	// Add to context
	ctx := pkg.WithRequestID(r.Context(), requestID)
	r = r.WithContext(ctx)

	pkg.Info(ctx, "incoming request: %s %s", r.Method, r.URL.Path)

	switch r.URL.Path {
	case "/search":
		if r.Method != http.MethodPost {
			writeJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.SearchFlights(w, r)
	default:
		writeJSONError(w, "not found", http.StatusNotFound)
	}
}

func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
