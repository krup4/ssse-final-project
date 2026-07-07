package models

import "time"

type Station struct {
    ID        int64   `json:"id"`
    Name      string  `json:"name"`
    Latitude  float64 `json:"latitude"`
    Longitude float64 `json:"longitude"`
    Region    string  `json:"region"`
    Active    bool    `json:"active"`
}

type WeatherMeasurement struct {
    ID           int64      `json:"id"`
    StationID    int64      `json:"station_id"`
    Timestamp    time.Time  `json:"timestamp"`
    Temperature  *float64   `json:"temperature,omitempty"`
    Humidity     *float64   `json:"humidity,omitempty"`
    Pressure     *float64   `json:"pressure,omitempty"`
    WindSpeed    *float64   `json:"wind_speed,omitempty"`
    WindDirection *float64  `json:"wind_direction,omitempty"`
    Precipitation *float64  `json:"precipitation,omitempty"`
    CreatedAt    time.Time  `json:"created_at"`
}
