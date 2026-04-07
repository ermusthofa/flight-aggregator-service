package http

import (
	"encoding/json"
	"net/http"

	"github.com/ermusthofa/flight-aggregator-service/internal/pkg"
	"github.com/ermusthofa/flight-aggregator-service/internal/usecase"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Handler struct {
	searchUC usecase.SearchFlightsUsecaseInterface
	router   http.Handler
}

func NewHandler(uc usecase.SearchFlightsUsecaseInterface) *Handler {
	h := &Handler{
		searchUC: uc,
	}
	r := mux.NewRouter()

	// Middleware: request ID
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			requestID := req.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}
			ctx := pkg.WithRequestID(req.Context(), requestID)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})

	// Middleware: logging
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			pkg.Info(ctx, "incoming request: %s %s", req.Method, req.URL.Path)
			next.ServeHTTP(w, req)
		})
	})

	// Routes
	r.HandleFunc("/search", h.SearchFlights).Methods("POST")
	r.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods("GET")

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		writeJSONError(w, "not found", http.StatusNotFound)
	})
	r.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		writeJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	h.router = r
	return h
}

// ServeHTTP implements http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

// writeJSONError is a shared helper.
func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
