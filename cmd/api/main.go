package main

import (
	"log"
	"net/http"

	"github.com/ermusthofa/flight-aggregator-service/internal/cache"
	httpHandler "github.com/ermusthofa/flight-aggregator-service/internal/transport/http"
	"github.com/ermusthofa/flight-aggregator-service/internal/usecase"
)

func main() {
	// init usecase
	usecase := usecase.NewSearchFlightsUsecase(cache.NewMemoryCache())

	// init handler
	handler := httpHandler.NewHandler(usecase)

	// routes
	http.HandleFunc("/search", handler.SearchFlights)

	log.Println("Server running on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
