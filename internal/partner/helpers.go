package partner

import (
	"context"
	"strings"
	"time"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
	"github.com/ermusthofa/flight-aggregator-service/internal/pkg"
)

// matchesSearchCriteria – correct passenger count logic (availableSeats >= req.Passengers)
func matchesSearchCriteria(req domain.SearchRequest, origin, destination string, depTime time.Time, availableSeats int, cabinClass string) bool {
	if origin != req.Origin || destination != req.Destination {
		return false
	}
	if depTime.Format("2006-01-02") != req.DepartureDate {
		return false
	}
	if availableSeats < req.Passengers {
		return false
	}
	if req.CabinClass != "" && !strings.EqualFold(cabinClass, req.CabinClass) {
		return false
	}
	return true
}

func normalizeCabinClass(cabin string) string {
	c := strings.ToLower(strings.TrimSpace(cabin))
	switch c {
	case "economy", "business", "first":
		return c
	case "y": // Batik's cabin class
		return "economy"
	default:
		return "economy"
	}
}

func airportToCity(airportCode string) string {
	mapping := map[string]string{
		"CGK": "Jakarta",
		"DPS": "Denpasar",
		"SUB": "Surabaya",
		"SOC": "Solo",
		"UPG": "Makassar",
	}
	if city, ok := mapping[airportCode]; ok {
		return city
	}
	return "Unknown"
}

func parseBaggageNote(note string) domain.Baggage {
	baggage := domain.Baggage{CarryOn: "Not specified", Checked: "Not specified"}
	if note == "" {
		return baggage
	}
	parts := strings.Split(note, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		lower := strings.ToLower(part)
		if strings.Contains(lower, "cabin") {
			baggage.CarryOn = part
		} else if strings.Contains(lower, "checked") {
			baggage.Checked = part
		}
	}
	return baggage
}

func parseWithTimezone(dateStr, iana string) (time.Time, error) {
	loc, err := time.LoadLocation(iana)
	if err != nil {
		return time.Time{}, err
	}
	return time.ParseInLocation("2006-01-02T15:04:05", dateStr, loc)
}

func warnSkip(ctx context.Context, provider, flightNumber, reason string, err error) {
	if err != nil {
		pkg.Warn(ctx, "%s: skip flight %s: %s: %v", provider, flightNumber, reason, err)
	} else {
		pkg.Warn(ctx, "%s: skip flight %s: %s", provider, flightNumber, reason)
	}
}
