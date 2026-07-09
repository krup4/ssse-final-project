package kafka

import (
	"testing"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	"weather-accuracy/core-api/internal/domain"
)

func TestJSONActualWeatherDecoderDecodeCanonicalPayload(t *testing.T) {
	decoder := NewJSONActualWeatherDecoder()
	reading, err := decoder.Decode(kafkago.Message{Value: []byte(`{
		"id": "event-1",
		"stationId": "101",
		"observedAt": "2026-07-08T06:00:00Z",
		"temperature": 29.3,
		"windSpeed": 12.1,
		"humidity": 71,
		"pressure": 1009,
		"source": "sensor-gateway",
		"traceId": "trace-1"
	}`)})
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	assertReading(t, reading, "event-1", "101", "sensor-gateway", "trace-1")
	assertValue(t, reading, domain.ParameterTemperature, 29.3)
	assertValue(t, reading, domain.ParameterWindSpeed, 12.1)
	assertValue(t, reading, domain.ParameterHumidity, 71)
	assertValue(t, reading, domain.ParameterPressure, 1009)
}

func TestJSONActualWeatherDecoderDecodeLegacySnakeCasePayload(t *testing.T) {
	decoder := NewJSONActualWeatherDecoder()
	reading, err := decoder.Decode(kafkago.Message{Value: []byte(`{
		"station_id": 102,
		"timestamp": "2026-07-08T07:00:00Z",
		"temperature": 18.7,
		"wind_speed": 9.4,
		"trace_id": "legacy-trace"
	}`)})
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	assertReading(t, reading, "", "102", "sensor", "legacy-trace")
	assertValue(t, reading, domain.ParameterTemperature, 18.7)
	assertValue(t, reading, domain.ParameterWindSpeed, 9.4)
}

func TestJSONActualWeatherDecoderRequiresStationAndObservedAt(t *testing.T) {
	decoder := NewJSONActualWeatherDecoder()
	_, err := decoder.Decode(kafkago.Message{Value: []byte(`{"temperature": 18.7}`)})
	if err == nil {
		t.Fatal("Decode() error = nil, want validation error")
	}
}

func assertReading(t *testing.T, reading domain.ActualWeatherReading, id string, stationID string, source string, traceID string) {
	t.Helper()
	if id != "" && reading.ID != id {
		t.Fatalf("reading.ID = %q, want %q", reading.ID, id)
	}
	if reading.StationID != stationID {
		t.Fatalf("reading.StationID = %q, want %q", reading.StationID, stationID)
	}
	wantObservedAt := time.Date(2026, 7, 8, 6, 0, 0, 0, time.UTC)
	if stationID == "102" {
		wantObservedAt = time.Date(2026, 7, 8, 7, 0, 0, 0, time.UTC)
	}
	if !reading.ObservedAt.Equal(wantObservedAt) {
		t.Fatalf("reading.ObservedAt = %s, want %s", reading.ObservedAt, wantObservedAt)
	}
	if reading.Source != source {
		t.Fatalf("reading.Source = %q, want %q", reading.Source, source)
	}
	if reading.TraceID != traceID {
		t.Fatalf("reading.TraceID = %q, want %q", reading.TraceID, traceID)
	}
}

func assertValue(t *testing.T, reading domain.ActualWeatherReading, parameter domain.WeatherParameter, want float64) {
	t.Helper()
	got, ok := reading.Values[parameter]
	if !ok {
		t.Fatalf("reading.Values[%q] missing", parameter)
	}
	if got != want {
		t.Fatalf("reading.Values[%q] = %v, want %v", parameter, got, want)
	}
}
