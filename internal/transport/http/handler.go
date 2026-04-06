package http

import (
	"encoding/json"
	"net/http"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
	"github.com/ermusthofa/flight-aggregator-service/internal/pkg"
	"github.com/ermusthofa/flight-aggregator-service/internal/usecase"
)

type Handler struct {
	usecase *usecase.SearchFlightsUsecase
}

func NewHandler(u *usecase.SearchFlightsUsecase) *Handler {
	return &Handler{usecase: u}
}

func (h *Handler) SearchFlights(w http.ResponseWriter, r *http.Request) {
	pkg.Info("Incoming request: %s %s", r.Method, r.URL.Path)

	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", "METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed)
		return
	}

	var req domain.SearchRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkg.Error("Failed to decode request: %v", err)
		writeError(w, "invalid request body", "INVALID_REQUEST", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		pkg.Error("Validation failed: %v", err)
		writeError(w, err.Error(), "VALIDATION_ERROR", http.StatusBadRequest)
		return
	}

	req.Normalize()

	flights, meta, err := h.usecase.Execute(r.Context(), req)
	if err != nil {
		pkg.Error("Usecase error: %v", err)
		writeError(w, "internal server error", "INTERNAL_ERROR", http.StatusInternalServerError)
		return
	}

	writeSuccess(w, map[string]interface{}{
		"flights":  flights,
		"metadata": meta,
	})
}

func writeSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domain.APIResponse{
		Data: data,
	})
}

func writeError(w http.ResponseWriter, message, code string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(domain.APIResponse{
		Error: &domain.APIError{
			Message: message,
			Code:    code,
		},
	})
}
