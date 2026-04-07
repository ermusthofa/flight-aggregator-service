package domain

import "time"

type Flight struct {
	ID             string
	Provider       string
	Airline        Airline
	FlightNumber   string
	Departure      Location
	Arrival        Location
	TotalMinutes   int
	Stops          int
	Price          Price
	AvailableSeats int
	CabinClass     string
	Aircraft       string
	Amenities      []string
	Baggage        Baggage

	Score int
}

type Airline struct {
	Name string
	Code string
}

type Location struct {
	Airport   string
	City      string
	Datetime  time.Time
	Timestamp int64
}

type Price struct {
	Amount   int
	Currency string
}

type Baggage struct {
	CarryOn string
	Checked string
}
