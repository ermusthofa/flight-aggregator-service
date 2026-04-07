package mapper

import (
	"fmt"
	"time"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
	"github.com/ermusthofa/flight-aggregator-service/internal/dto"
)

func ToResponse(
	req domain.SearchRequest,
	flights []domain.Flight,
	meta domain.Metadata,
) dto.SearchResponse {

	res := dto.SearchResponse{}

	// search criteria
	res.SearchCriteria = dto.SearchCriteria{
		Origin:        req.Origin,
		Destination:   req.Destination,
		DepartureDate: req.DepartureDate,
		Passengers:    req.Passengers,
		CabinClass:    req.CabinClass,
	}

	// metadata
	res.Metadata = dto.Metadata{
		TotalResults:       len(flights),
		ProvidersQueried:   meta.ProvidersQueried,
		ProvidersSucceeded: meta.ProvidersSucceeded,
		ProvidersFailed:    meta.ProvidersFailed,
		SearchTimeMs:       meta.SearchTimeMs,
		CacheHit:           false,
	}

	// flights
	res.Flights = make([]dto.Flight, 0)

	for _, f := range flights {

		item := dto.Flight{
			ID:             f.ID,
			Provider:       f.Provider,
			FlightNumber:   f.FlightNumber,
			Stops:          f.Stops,
			AvailableSeats: f.AvailableSeats,
			CabinClass:     f.CabinClass,
			Amenities:      f.Amenities,
		}

		// airline
		item.Airline.Name = f.Airline.Name
		item.Airline.Code = f.Airline.Code

		// departure
		item.Departure.Airport = f.Departure.Airport
		item.Departure.City = f.Departure.City
		item.Departure.Datetime = f.Departure.Datetime.Format(time.RFC3339)
		item.Departure.Timestamp = f.Departure.Timestamp

		// arrival
		item.Arrival.Airport = f.Arrival.Airport
		item.Arrival.City = f.Arrival.City
		item.Arrival.Datetime = f.Arrival.Datetime.Format(time.RFC3339)
		item.Arrival.Timestamp = f.Arrival.Timestamp

		// duration
		item.Duration.TotalMinutes = f.TotalMinutes
		item.Duration.Formatted = formatDuration(f.TotalMinutes)

		// price
		item.Price.Amount = f.Price.Amount
		item.Price.Currency = f.Price.Currency

		// aircraft (nullable)
		if f.Aircraft != "" {
			item.Aircraft = &f.Aircraft
		}

		// baggage
		item.Baggage.CarryOn = f.Baggage.CarryOn
		item.Baggage.Checked = f.Baggage.Checked

		res.Flights = append(res.Flights, item)
	}

	return res
}

func formatDuration(minutes int) string {
	h := minutes / 60
	m := minutes % 60
	return fmt.Sprintf("%dh %dm", h, m)
}
