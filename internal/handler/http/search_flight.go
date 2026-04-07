package http

import (
	"encoding/json"
	"net/http"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
	"github.com/ermusthofa/flight-aggregator-service/internal/dto"
	"github.com/ermusthofa/flight-aggregator-service/internal/mapper"
)

func (h *Handler) SearchFlights(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req domain.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error(ctx, "failed to decode request: %v", err)
		writeJSONError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Error(ctx, "validation failed: %v", err)
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.Normalize()

	flights, metadata, err := h.searchUC.Execute(ctx, req)
	if err != nil {
		h.logger.Error(ctx, "request error: %v", err)
		writeJSONError(w, "search failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info(ctx, "search completed: %d results in %d ms", metadata.TotalResults, metadata.SearchTimeMs)

	resp := dto.SearchResponse{
		SearchCriteria: mapper.ToSearchCriteriaDTO(req),
		Metadata:       mapper.ToMetadataDTO(metadata),
		Flights:        mapper.ToFlightDTOs(flights),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
