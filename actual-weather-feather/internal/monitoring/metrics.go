package monitoring

import (
    "github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
    WeatherRequestsTotal   prometheus.Counter
    WeatherSavedTotal      prometheus.Counter
    ErrorsTotal            prometheus.Counter
    StationsProcessedTotal prometheus.Counter
}

func New() *Metrics {
    m := &Metrics{
        WeatherRequestsTotal: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "weather_requests_total",
            Help: "Total weather requests",
        }),
        WeatherSavedTotal: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "weather_saved_total",
            Help: "Saved weather measurements",
        }),
        ErrorsTotal: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "errors_total",
            Help: "Total errors",
        }),
        StationsProcessedTotal: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "stations_processed_total",
            Help: "Total stations processed",
        }),
    }
    prometheus.MustRegister(m.WeatherRequestsTotal, m.WeatherSavedTotal, m.ErrorsTotal, m.StationsProcessedTotal)
    return m
}
