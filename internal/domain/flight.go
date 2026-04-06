package domain

type Flight struct {
	ID            string `json:"id"`
	Provider      string `json:"provider"`
	FlightNumber  string `json:"flight_number"`
	Origin        string `json:"origin"`
	Destination   string `json:"destination"`
	DepartureTime string `json:"departure_time"`
	ArrivalTime   string `json:"arrival_time"`
	Price         int    `json:"price"`
}
