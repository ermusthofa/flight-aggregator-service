package domain

import "time"

type Flight struct {
	ID           string  `json:"id"`
	Provider     string  `json:"provider"`
	Airline      Airline `json:"airline"`
	FlightNumber string  `json:"flight_number"`

	Departure Location `json:"departure"`
	Arrival   Location `json:"arrival"`

	Duration Duration `json:"duration"`

	Stops int `json:"stops"`

	Price Price `json:"price"`

	AvailableSeats int      `json:"available_seats"`
	CabinClass     string   `json:"cabin_class"`
	Aircraft       *string  `json:"aircraft"`
	Amenities      []string `json:"amenities"`

	Baggage Baggage `json:"baggage"`

	Score int `json:"-"`
}

type Airline struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type Location struct {
	Airport   string    `json:"airport"`
	City      string    `json:"city"`
	Datetime  time.Time `json:"datetime"`
	Timestamp int64     `json:"timestamp"`
}

type Duration struct {
	TotalMinutes int    `json:"total_minutes"`
	Formatted    string `json:"formatted"`
}

type Price struct {
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
}

type Baggage struct {
	CarryOn string `json:"carry_on"`
	Checked string `json:"checked"`
}
