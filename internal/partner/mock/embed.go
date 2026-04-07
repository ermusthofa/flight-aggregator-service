package mock

import _ "embed"

//go:embed airasia_search_response.json
var AirAsiaMock []byte

//go:embed batik_air_search_response.json
var BatikMock []byte

//go:embed garuda_indonesia_search_response.json
var GarudaMock []byte

//go:embed lion_air_search_response.json
var LionMock []byte
