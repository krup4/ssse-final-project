package openmeteo

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    backoff "github.com/cenkalti/backoff/v4"

    "actual-weather-feather/internal/models"
)

type Client struct {
    baseURL string
    client  *http.Client
}

type API interface {
    GetCurrentMeasurement(ctx context.Context, lat, lon float64) (*models.WeatherMeasurement, error)
}

func NewClient(baseURL string, timeout time.Duration) *Client {
    return &Client{
        baseURL: baseURL,
        client:  &http.Client{Timeout: timeout},
    }
}

// GetCurrentWeather requests current weather for latitude/longitude and returns raw map
func (c *Client) GetCurrentWeather(ctx context.Context, lat, lon float64) (map[string]interface{}, error) {
    // Request hourly variables to obtain humidity, pressure, precipitation and wind details
    url := fmt.Sprintf("%s/v1/forecast?latitude=%f&longitude=%f&current_weather=true&hourly=temperature_2m,relativehumidity_2m,surface_pressure,windspeed_10m,winddirection_10m,precipitation&timezone=UTC", c.baseURL, lat, lon)

    var lastErr error
    var result map[string]interface{}

    notify := func(err error, d time.Duration) {
        lastErr = err
    }

    operation := func() error {
        req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
        resp, err := c.client.Do(req)
        if err != nil {
            return err
        }
        defer resp.Body.Close()
        if resp.StatusCode >= 400 {
            b, _ := io.ReadAll(resp.Body)
            return fmt.Errorf("open-meteo status %d: %s", resp.StatusCode, string(b))
        }
            dec := json.NewDecoder(resp.Body)
            if err := dec.Decode(&result); err != nil {
                return err
            }
        return nil
    }

    b := backoff.NewExponentialBackOff()
    b.MaxElapsedTime = 15 * time.Second
    if err := backoff.RetryNotify(operation, b, notify); err != nil {
        return nil, fmt.Errorf("openmeteo request failed: %w", lastErr)
    }
    return result, nil
}

// GetCurrentMeasurement fetches current weather and converts to domain model (partial)
func (c *Client) GetCurrentMeasurement(ctx context.Context, lat, lon float64) (*models.WeatherMeasurement, error) {
    data, err := c.GetCurrentWeather(ctx, lat, lon)
    if err != nil {
        return nil, err
    }

    var m models.WeatherMeasurement

    // prefer current_weather block for timestamp and basic fields
    if cwRaw, ok := data["current_weather"]; ok {
        if cw, ok := cwRaw.(map[string]interface{}); ok {
            if tRaw, ok := cw["time"].(string); ok {
                if parsed, perr := time.Parse(time.RFC3339, tRaw); perr == nil {
                    m.Timestamp = parsed.UTC()
                }
            }
            if v, ok := cw["temperature"].(float64); ok {
                m.Temperature = &v
            }
            if v, ok := cw["windspeed"].(float64); ok {
                m.WindSpeed = &v
            }
            if v, ok := cw["winddirection"].(float64); ok {
                m.WindDirection = &v
            }
        }
    }

    // If hourly data present, try to pick the latest hourly entry at or before timestamp
    if hourlyRaw, ok := data["hourly"]; ok {
        if hourly, ok := hourlyRaw.(map[string]interface{}); ok {
            times, _ := hourly["time"].([]interface{})
            // find index for m.Timestamp (hour) or last index
            idx := -1
            if !m.Timestamp.IsZero() {
                // match time string to times
                for i := len(times) - 1; i >= 0; i-- {
                    if ts, ok := times[i].(string); ok {
                        if tParsed, perr := time.Parse(time.RFC3339, ts); perr == nil {
                            if !tParsed.After(m.Timestamp) {
                                idx = i
                                break
                            }
                        }
                    }
                }
            }
            if idx == -1 && len(times) > 0 {
                idx = len(times) - 1
            }

            if idx >= 0 {
                // helper to read float array at idx
                readFloatAt := func(key string) *float64 {
                    if arr, ok := hourly[key].([]interface{}); ok {
                        if idx < len(arr) {
                            if v, ok := arr[idx].(float64); ok {
                                return &v
                            }
                        }
                    }
                    return nil
                }

                if v := readFloatAt("temperature_2m"); v != nil && m.Temperature == nil { m.Temperature = v }
                if v := readFloatAt("relativehumidity_2m"); v != nil { m.Humidity = v }
                if v := readFloatAt("surface_pressure"); v != nil { m.Pressure = v }
                if v := readFloatAt("windspeed_10m"); v != nil && m.WindSpeed == nil { m.WindSpeed = v }
                if v := readFloatAt("winddirection_10m"); v != nil && m.WindDirection == nil { m.WindDirection = v }
                if v := readFloatAt("precipitation"); v != nil { m.Precipitation = v }
            }
        }
    }

    if m.Timestamp.IsZero() {
        m.Timestamp = time.Now().UTC()
    }
    m.CreatedAt = time.Now().UTC()
    return &m, nil
}
