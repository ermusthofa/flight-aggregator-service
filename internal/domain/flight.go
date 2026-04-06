package domain

import "time"

type Flight struct {
	ID              string    `json:"id"`
	Provider        string    `json:"provider"`
	Airline         Airline   `json:"airline"`
	FlightNumber    string    `json:"flight_number"`
	Departure       Departure `json:"departure"`
	Arrival         Arrival   `json:"arrival"`
	DurationMinutes int       `json:"duration_minutes"`
	Stops           int       `json:"stops"`
	Price           Price     `json:"price"`
	AvailableSeats  int       `json:"available_seats"`
	CabinClass      string    `json:"cabin_class"`

	Score int `json:"-"`
}

type Airline struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type Departure struct {
	Airport   string    `json:"airport"`
	Datetime  time.Time `json:"datetime"`
	Timestamp int64     `json:"timestamp"`
}

type Arrival struct {
	Airport   string    `json:"airport"`
	Datetime  time.Time `json:"datetime"`
	Timestamp int64     `json:"timestamp"`
}

type Price struct {
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
}
