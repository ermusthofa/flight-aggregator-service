package usecase

import (
	"testing"

	"github.com/ermusthofa/flight-aggregator-service/internal/domain"
)

// Helper to create a flight for scoring tests
func makeScoreFlight(id string, price, duration, stops int) domain.Flight {
	return domain.Flight{
		ID:           id,
		FlightNumber: id,
		Price:        domain.Price{Amount: price, Currency: "IDR"},
		TotalMinutes: duration,
		Stops:        stops,
	}
}

func TestScoringEngine_EmptySlice(t *testing.T) {
	engine := NewScoringEngine()
	flights := []domain.Flight{}
	engine.Calculate(flights)
	// Should not panic or modify anything
	if len(flights) != 0 {
		t.Errorf("expected empty slice, got length %d", len(flights))
	}
}

func TestScoringEngine_SingleFlight(t *testing.T) {
	engine := NewScoringEngine()
	flights := []domain.Flight{
		makeScoreFlight("F1", 500000, 120, 0),
	}
	engine.Calculate(flights)
	// With one flight, min=max, range=1 (artificially), normalized values = 0
	// score = (0*0.5)+(0*0.3)+(0*0.2)=0 → int(0*1000)=0
	if flights[0].Score != 0 {
		t.Errorf("expected score 0, got %d", flights[0].Score)
	}
}

func TestScoringEngine_MultipleFlights(t *testing.T) {
	engine := NewScoringEngine()
	flights := []domain.Flight{
		makeScoreFlight("F1", 400000, 100, 0), // cheapest, shortest, no stops → best
		makeScoreFlight("F2", 600000, 150, 1), // most expensive, longest, 1 stop → worst
		makeScoreFlight("F3", 500000, 120, 0), // middle
	}
	engine.Calculate(flights)

	// Expected scores (roughly): lower is better.
	// F1 should have lowest score, F2 highest.
	if flights[0].Score > flights[1].Score {
		t.Errorf("expected F1 score <= F2 score, got F1=%d, F2=%d", flights[0].Score, flights[1].Score)
	}
	if flights[1].Score < flights[2].Score {
		t.Errorf("expected F2 score >= F3 score, got F2=%d, F3=%d", flights[1].Score, flights[2].Score)
	}
}

func TestScoringEngine_ScoreFormula(t *testing.T) {
	engine := NewScoringEngine()
	flights := []domain.Flight{
		makeScoreFlight("A", 100000, 60, 0),  // min price, min duration
		makeScoreFlight("B", 200000, 120, 1), // max price, max duration, 1 stop
	}
	engine.Calculate(flights)

	// minPrice=100k, maxPrice=200k, range=100k
	// minDuration=60, maxDuration=120, range=60
	// For flight A: nPrice=0, nDuration=0, stopsPenalty=0 → score=0 → 0
	// For flight B: nPrice=(200-100)/100=1, nDuration=(120-60)/60=1, stopsPenalty=1*0.2=0.2
	// score = (1*0.5)+(1*0.3)+0.2 = 0.5+0.3+0.2=1.0 → int(1*1000)=1000
	if flights[0].Score != 0 {
		t.Errorf("expected flight A score 0, got %d", flights[0].Score)
	}
	if flights[1].Score != 1000 {
		t.Errorf("expected flight B score 1000, got %d", flights[1].Score)
	}
}

func TestScoringEngine_ZeroPriceRange(t *testing.T) {
	engine := NewScoringEngine()
	flights := []domain.Flight{
		makeScoreFlight("A", 500000, 60, 0),
		makeScoreFlight("B", 500000, 120, 0), // same price
	}
	engine.Calculate(flights)
	// Price range = 0 → replaced by 1, so nPrice = (500-500)/1 = 0 for both.
	// Duration: min=60, max=120, range=60 → nDuration A=0, B=1
	// Score A = 0*0.5 + 0*0.3 = 0 → 0
	// Score B = 0*0.5 + 1*0.3 = 0.3 → int(0.3*1000)=300
	if flights[0].Score != 0 {
		t.Errorf("expected A score 0, got %d", flights[0].Score)
	}
	if flights[1].Score != 300 {
		t.Errorf("expected B score 300, got %d", flights[1].Score)
	}
}

func TestScoringEngine_ZeroDurationRange(t *testing.T) {
	engine := NewScoringEngine()
	flights := []domain.Flight{
		makeScoreFlight("A", 100000, 100, 0),
		makeScoreFlight("B", 200000, 100, 0), // same duration
	}
	engine.Calculate(flights)
	// Duration range = 0 → replaced by 1, so nDuration = 0 for both.
	// Price: min=100k, max=200k, range=100k → nPrice A=0, B=1
	// Score A = 0*0.5 =0 → 0
	// Score B = 1*0.5 =0.5 → int(0.5*1000)=500
	if flights[0].Score != 0 {
		t.Errorf("expected A score 0, got %d", flights[0].Score)
	}
	if flights[1].Score != 500 {
		t.Errorf("expected B score 500, got %d", flights[1].Score)
	}
}

func TestScoringEngine_StopsPenalty(t *testing.T) {
	engine := NewScoringEngine()
	// Same price and duration, different stops
	flights := []domain.Flight{
		makeScoreFlight("A", 500000, 100, 0),
		makeScoreFlight("B", 500000, 100, 1),
		makeScoreFlight("C", 500000, 100, 2),
	}
	engine.Calculate(flights)
	// Price and duration ranges are zero, so nPrice=0, nDuration=0.
	// Score = stopsPenalty = stops * 0.2
	// A: 0 → 0
	// B: 0.2 → int(0.2*1000)=200
	// C: 0.4 → int(0.4*1000)=400
	if flights[0].Score != 0 {
		t.Errorf("expected A score 0, got %d", flights[0].Score)
	}
	if flights[1].Score != 200 {
		t.Errorf("expected B score 200, got %d", flights[1].Score)
	}
	if flights[2].Score != 400 {
		t.Errorf("expected C score 400, got %d", flights[2].Score)
	}
}

func TestScoringEngine_OrderPreserved(t *testing.T) {
	engine := NewScoringEngine()
	flights := []domain.Flight{
		makeScoreFlight("B", 600000, 150, 1), // worst
		makeScoreFlight("A", 400000, 100, 0), // best
		makeScoreFlight("C", 500000, 120, 0), // middle
	}
	engine.Calculate(flights)
	// After scoring, we don't sort; scores should reflect the actual values.
	// Check that A has lower score than B and C.
	if flights[1].Score >= flights[0].Score {
		t.Errorf("expected flight A score (%d) < flight B score (%d)", flights[1].Score, flights[0].Score)
	}
	if flights[1].Score >= flights[2].Score {
		t.Errorf("expected flight A score (%d) < flight C score (%d)", flights[1].Score, flights[2].Score)
	}
}
