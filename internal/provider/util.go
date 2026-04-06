package provider

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
)

func loadMock(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func parseWithTimezone(dt string, tz string) (time.Time, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.Time{}, err
	}

	return time.ParseInLocation("2006-01-02T15:04:05", dt, loc)
}

func parseDuration(s string) int {
	var hours, minutes int

	fmt.Sscanf(s, "%dh %dm", &hours, &minutes)

	return hours*60 + minutes
}

func formatDuration(minutes int) string {
	h := minutes / 60
	m := minutes % 60
	return fmt.Sprintf("%dh %dm", h, m)
}

func ensureSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func parseBaggage(baggageNote string) domain.Baggage {
	baggage := domain.Baggage{
		CarryOn: "Not specified",
		Checked: "Not specified",
	}

	if baggageNote == "" {
		return baggage
	}

	// Split the baggage note by commas
	parts := strings.Split(baggageNote, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)

		if strings.Contains(strings.ToLower(part), "cabin") {
			baggage.CarryOn = part
		} else if strings.Contains(strings.ToLower(part), "checked") {
			baggage.Checked = part
		}
	}

	return baggage
}

var airportCityMap = map[string]string{
	"CGK": "Jakarta",
	"DPS": "Denpasar",
	"SOC": "Solo",
	"UPG": "Makassar",
	"SUB": "Surabaya",
	// Add more mappings as needed
}

// GetCityByAirport retrieves the city name for a given airport code.
func getCityByAirport(airport string) string {
	city, exists := airportCityMap[airport]
	if !exists {
		return "Unknown"
	}
	return city
}
