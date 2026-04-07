package usecase

import (
	"testing"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
)

// Helper to create a flight with given attributes
func makeSortFlight(id string, price int, durationMin int, depTimestamp int64, arrTimestamp int64, score int) domain.Flight {
	return domain.Flight{
		ID:           id,
		FlightNumber: id,
		Price:        domain.Price{Amount: price, Currency: "IDR"},
		TotalMinutes: durationMin,
		Departure:    domain.Location{Timestamp: depTimestamp},
		Arrival:      domain.Location{Timestamp: arrTimestamp},
		Score:        score,
	}
}

func TestSorterEngine_SortByPrice(t *testing.T) {
	sorter := NewSorterEngine()
	flights := []domain.Flight{
		makeSortFlight("F1", 500000, 120, 100, 200, 0),
		makeSortFlight("F2", 300000, 90, 100, 200, 0),
		makeSortFlight("F3", 400000, 150, 100, 200, 0),
	}
	expectedOrder := []string{"F2", "F3", "F1"} // 300k, 400k, 500k

	sorter.Sort(flights, "price")
	for i, id := range expectedOrder {
		if flights[i].ID != id {
			t.Errorf("at position %d: expected %s, got %s", i, id, flights[i].ID)
		}
	}
}

func TestSorterEngine_SortByPrice_TieBreakByDuration(t *testing.T) {
	sorter := NewSorterEngine()
	flights := []domain.Flight{
		makeSortFlight("F1", 400000, 150, 100, 200, 0),
		makeSortFlight("F2", 400000, 90, 100, 200, 0),
		makeSortFlight("F3", 400000, 120, 100, 200, 0),
	}
	// Same price, should be sorted by duration ascending: 90, 120, 150
	expectedOrder := []string{"F2", "F3", "F1"}

	sorter.Sort(flights, "price")
	for i, id := range expectedOrder {
		if flights[i].ID != id {
			t.Errorf("at position %d: expected %s, got %s", i, id, flights[i].ID)
		}
	}
}

func TestSorterEngine_SortByDuration(t *testing.T) {
	sorter := NewSorterEngine()
	flights := []domain.Flight{
		makeSortFlight("F1", 500000, 120, 100, 200, 0),
		makeSortFlight("F2", 300000, 90, 100, 200, 0),
		makeSortFlight("F3", 400000, 150, 100, 200, 0),
	}
	expectedOrder := []string{"F2", "F1", "F3"} // 90, 120, 150

	sorter.Sort(flights, "duration")
	for i, id := range expectedOrder {
		if flights[i].ID != id {
			t.Errorf("at position %d: expected %s, got %s", i, id, flights[i].ID)
		}
	}
}

func TestSorterEngine_SortByDuration_TieBreakByPrice(t *testing.T) {
	sorter := NewSorterEngine()
	flights := []domain.Flight{
		makeSortFlight("F1", 500000, 120, 100, 200, 0),
		makeSortFlight("F2", 300000, 120, 100, 200, 0),
		makeSortFlight("F3", 400000, 120, 100, 200, 0),
	}
	// Same duration, should be sorted by price ascending: 300k, 400k, 500k
	expectedOrder := []string{"F2", "F3", "F1"}

	sorter.Sort(flights, "duration")
	for i, id := range expectedOrder {
		if flights[i].ID != id {
			t.Errorf("at position %d: expected %s, got %s", i, id, flights[i].ID)
		}
	}
}

func TestSorterEngine_SortByDeparture(t *testing.T) {
	sorter := NewSorterEngine()
	flights := []domain.Flight{
		makeSortFlight("F1", 500000, 120, 300, 200, 0), // dep 300
		makeSortFlight("F2", 300000, 90, 100, 200, 0),  // dep 100
		makeSortFlight("F3", 400000, 150, 200, 200, 0), // dep 200
	}
	expectedOrder := []string{"F2", "F3", "F1"}

	sorter.Sort(flights, "departure")
	for i, id := range expectedOrder {
		if flights[i].ID != id {
			t.Errorf("at position %d: expected %s, got %s", i, id, flights[i].ID)
		}
	}
}

func TestSorterEngine_SortByArrival(t *testing.T) {
	sorter := NewSorterEngine()
	flights := []domain.Flight{
		makeSortFlight("F1", 500000, 120, 100, 300, 0), // arr 300
		makeSortFlight("F2", 300000, 90, 100, 100, 0),  // arr 100
		makeSortFlight("F3", 400000, 150, 100, 200, 0), // arr 200
	}
	expectedOrder := []string{"F2", "F3", "F1"}

	sorter.Sort(flights, "arrival")
	for i, id := range expectedOrder {
		if flights[i].ID != id {
			t.Errorf("at position %d: expected %s, got %s", i, id, flights[i].ID)
		}
	}
}

func TestSorterEngine_SortByBest(t *testing.T) {
	sorter := NewSorterEngine()
	flights := []domain.Flight{
		makeSortFlight("F1", 500000, 120, 100, 200, 30),
		makeSortFlight("F2", 300000, 90, 100, 200, 10),
		makeSortFlight("F3", 400000, 150, 100, 200, 20),
	}
	expectedOrder := []string{"F2", "F3", "F1"} // score ascending: 10,20,30

	sorter.Sort(flights, "best")
	for i, id := range expectedOrder {
		if flights[i].ID != id {
			t.Errorf("at position %d: expected %s, got %s", i, id, flights[i].ID)
		}
	}
}

func TestSorterEngine_DefaultSortByScore(t *testing.T) {
	sorter := NewSorterEngine()
	flights := []domain.Flight{
		makeSortFlight("F1", 500000, 120, 100, 200, 30),
		makeSortFlight("F2", 300000, 90, 100, 200, 10),
		makeSortFlight("F3", 400000, 150, 100, 200, 20),
	}
	// Default case (empty string or unknown) uses score ascending
	expectedOrder := []string{"F2", "F3", "F1"}

	sorter.Sort(flights, "") // empty sortBy -> default
	for i, id := range expectedOrder {
		if flights[i].ID != id {
			t.Errorf("at position %d: expected %s, got %s", i, id, flights[i].ID)
		}
	}
}

func TestSorterEngine_UnknownSortBy(t *testing.T) {
	sorter := NewSorterEngine()
	flights := []domain.Flight{
		makeSortFlight("F1", 500000, 120, 100, 200, 30),
		makeSortFlight("F2", 300000, 90, 100, 200, 10),
		makeSortFlight("F3", 400000, 150, 100, 200, 20),
	}
	// Unknown string falls back to default (score ascending)
	expectedOrder := []string{"F2", "F3", "F1"}

	sorter.Sort(flights, "unknown")
	for i, id := range expectedOrder {
		if flights[i].ID != id {
			t.Errorf("at position %d: expected %s, got %s", i, id, flights[i].ID)
		}
	}
}

func TestSorterEngine_EmptySlice(t *testing.T) {
	sorter := NewSorterEngine()
	flights := []domain.Flight{}
	sorter.Sort(flights, "price")
	if len(flights) != 0 {
		t.Errorf("expected empty slice, got length %d", len(flights))
	}
}
