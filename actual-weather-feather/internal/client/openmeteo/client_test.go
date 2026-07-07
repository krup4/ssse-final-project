package openmeteo

import (
    "bytes"
    "context"
    "io/ioutil"
    "net/http"
    "testing"

    "actual-weather-feather/internal/models"
)

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
    return f(req), nil
}

func TestGetCurrentMeasurement_ParseHourly(t *testing.T) {
    // sample simplified Open-Meteo response with hourly arrays
    sample := `{
  "hourly": {
    "time": ["2026-07-07T14:00:00Z","2026-07-07T15:00:00Z"],
    "temperature_2m": [20.5,21.0],
    "relativehumidity_2m": [60,58],
    "surface_pressure": [1012,1011],
    "windspeed_10m": [3.2,4.1],
    "winddirection_10m": [180,190],
    "precipitation": [0,0]
  },
  "current_weather": {"time":"2026-07-07T15:00:00Z","temperature":21.0,"windspeed":4.1,"winddirection":190}
}`

    rt := roundTripFunc(func(req *http.Request) *http.Response {
        return &http.Response{
            StatusCode: 200,
            Body:       ioutil.NopCloser(bytes.NewBufferString(sample)),
            Header:     make(http.Header),
        }
    })

    c := &Client{baseURL: "http://example", client: &http.Client{Transport: rt}}

    ctx := context.Background()
    m, err := c.GetCurrentMeasurement(ctx, 55.75, 37.61)
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if m.Temperature == nil || *m.Temperature != 21.0 {
        t.Fatalf("expected temp 21.0, got %v", m.Temperature)
    }
    if m.Humidity == nil || *m.Humidity != 58 {
        t.Fatalf("expected humidity 58, got %v", m.Humidity)
    }
    if m.Pressure == nil || *m.Pressure != 1011 {
        t.Fatalf("expected pressure 1011, got %v", m.Pressure)
    }
    if m.WindSpeed == nil || *m.WindSpeed != 4.1 {
        t.Fatalf("expected windspeed 4.1, got %v", m.WindSpeed)
    }
    if m.WindDirection == nil || *m.WindDirection != 190 {
        t.Fatalf("expected winddirection 190, got %v", m.WindDirection)
    }
    if m.Timestamp.IsZero() {
        t.Fatalf("expected non-zero timestamp")
    }
    // ensure model type
    var _ models.WeatherMeasurement = *m
}
