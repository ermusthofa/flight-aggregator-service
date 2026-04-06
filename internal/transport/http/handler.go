package http

import (
	"encoding/json"
	"net/http"

	"github.com/ermusthofa/flight-aggregator-service/internal/usecase"
)

type Handler struct {
	usecase *usecase.SearchFlightsUsecase
}

func NewHandler(u *usecase.SearchFlightsUsecase) *Handler {
	return &Handler{usecase: u}
}

func (h *Handler) SearchFlights(w http.ResponseWriter, r *http.Request) {
	flights, err := h.usecase.Execute()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"flights": flights,
	})
}
