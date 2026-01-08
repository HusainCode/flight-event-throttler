package model

import "time"

type FlightEvent struct {
	ICAO24         string    `json:"icao24"`
	Callsign       string    `json:"callsign"`
	OriginCountry  string    `json:"origin_country"`
	TimePosition   int64     `json:"time_position"`
	LastContact    int64     `json:"last_contact"`
	Longitude      *float64  `json:"longitude"`
	Latitude       *float64  `json:"latitude"`
	BaroAltitude   *float64  `json:"baro_altitude"`
	OnGround       bool      `json:"on_ground"`
	Velocity       *float64  `json:"velocity"`
	TrueTrack      *float64  `json:"true_track"`
	VerticalRate   *float64  `json:"vertical_rate"`
	GeoAltitude    *float64  `json:"geo_altitude"`
	Squawk         *string   `json:"squawk"`
	Spi            bool      `json:"spi"`
	PositionSource int       `json:"position_source"`
	Timestamp      time.Time `json:"timestamp"`
}

type FlightEventBatch struct {
	Events    []FlightEvent `json:"events"`
	Count     int           `json:"count"`
	Timestamp time.Time     `json:"timestamp"`
}

type OpenSkyResponse struct {
	Time   int64           `json:"time"`
	States [][]interface{} `json:"states"`
}
