package dto

type SearchResponse struct {
	SearchCriteria SearchCriteria `json:"search_criteria"`
	Metadata       Metadata       `json:"metadata"`
	Flights        []Flight       `json:"flights"`
}

type SearchCriteria struct {
	Origin        string `json:"origin"`
	Destination   string `json:"destination"`
	DepartureDate string `json:"departure_date"`
	Passengers    int    `json:"passengers"`
	CabinClass    string `json:"cabin_class"`
}

type Metadata struct {
	TotalResults       int   `json:"total_results"`
	ProvidersQueried   int   `json:"providers_queried"`
	ProvidersSucceeded int   `json:"providers_succeeded"`
	ProvidersFailed    int   `json:"providers_failed"`
	SearchTimeMs       int64 `json:"search_time_ms"`
	CacheHit           bool  `json:"cache_hit"`
}

type Flight struct {
	ID           string `json:"id"`
	Provider     string `json:"provider"`
	FlightNumber string `json:"flight_number"`

	Airline struct {
		Name string `json:"name"`
		Code string `json:"code"`
	} `json:"airline"`

	Departure struct {
		Airport   string `json:"airport"`
		City      string `json:"city"`
		Datetime  string `json:"datetime"`
		Timestamp int64  `json:"timestamp"`
	} `json:"departure"`

	Arrival struct {
		Airport   string `json:"airport"`
		City      string `json:"city"`
		Datetime  string `json:"datetime"`
		Timestamp int64  `json:"timestamp"`
	} `json:"arrival"`

	Duration struct {
		TotalMinutes int    `json:"total_minutes"`
		Formatted    string `json:"formatted"`
	} `json:"duration"`

	Stops int `json:"stops"`

	Price struct {
		Amount   int    `json:"amount"`
		Currency string `json:"currency"`
	} `json:"price"`

	AvailableSeats int      `json:"available_seats"`
	CabinClass     string   `json:"cabin_class"`
	Aircraft       *string  `json:"aircraft"`
	Amenities      []string `json:"amenities"`

	Baggage struct {
		CarryOn string `json:"carry_on"`
		Checked string `json:"checked"`
	} `json:"baggage"`
}
