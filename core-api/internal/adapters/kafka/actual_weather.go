package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"

	"weather-accuracy/core-api/internal/domain"
	"weather-accuracy/core-api/internal/platform/metrics"
)

type ConsumerConfig struct {
	Brokers        []string
	Topic          string
	GroupID        string
	MinBytes       int
	MaxBytes       int
	CommitInterval time.Duration
	DedupTTL       time.Duration
}

type ActualWeatherDecoder interface {
	Decode(message kafka.Message) (domain.ActualWeatherReading, error)
}

type Deduplicator interface {
	MarkOnce(ctx context.Context, key string, ttl time.Duration) (bool, error)
}

type ActualWeatherConsumer struct {
	reader   *kafka.Reader
	decoder  ActualWeatherDecoder
	repo     domain.ActualWeatherRepository
	log      *slog.Logger
	metrics  *metrics.Registry
	dedup    Deduplicator
	dedupTTL time.Duration
}

func NewActualWeatherConsumer(cfg ConsumerConfig, decoder ActualWeatherDecoder, repo domain.ActualWeatherRepository, log *slog.Logger, registry *metrics.Registry, dedup Deduplicator) *ActualWeatherConsumer {
	return &ActualWeatherConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        cfg.Brokers,
			Topic:          cfg.Topic,
			GroupID:        cfg.GroupID,
			MinBytes:       cfg.MinBytes,
			MaxBytes:       cfg.MaxBytes,
			CommitInterval: cfg.CommitInterval,
			StartOffset:    kafka.LastOffset,
		}),
		decoder:  decoder,
		repo:     repo,
		log:      log,
		metrics:  registry,
		dedup:    dedup,
		dedupTTL: cfg.DedupTTL,
	}
}

func (c *ActualWeatherConsumer) Run(ctx context.Context) {
	defer func() {
		if err := c.reader.Close(); err != nil {
			c.log.Error("kafka reader close failed", slog.Any("error", err))
		}
	}()
	backoff := time.Second
	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			c.incKafkaError("fetch_error")
			c.log.Warn("kafka fetch failed", slog.Any("error", err), slog.Duration("backoff", backoff))
			sleep(ctx, backoff)
			backoff = minDuration(backoff*2, 30*time.Second)
			continue
		}
		backoff = time.Second

		reading, err := c.decoder.Decode(message)
		if err != nil {
			c.incKafkaError("decode_error")
			c.incKafkaMessage("decode_error")
			c.log.Warn("actual weather decode failed", slog.Any("error", err), slog.Int64("offset", message.Offset))
			_ = c.reader.CommitMessages(ctx, message)
			c.incKafkaCommit("decode_error_skipped")
			continue
		}
		duplicate, err := c.isDuplicate(ctx, reading)
		if err != nil {
			c.incKafkaError("dedup_error")
			c.log.Warn("actual weather dedup failed, processing message", slog.Any("error", err), slog.String("event_id", reading.ID))
		}
		if duplicate {
			c.incKafkaMessage("duplicate")
			if err := c.reader.CommitMessages(ctx, message); err != nil {
				c.incKafkaError("commit_error")
				c.log.Warn("kafka commit failed", slog.Any("error", err), slog.Int64("offset", message.Offset))
			} else {
				c.incKafkaCommit("duplicate_skipped")
			}
			continue
		}
		if err := retry(ctx, 5, 200*time.Millisecond, func() error {
			return c.repo.SaveReading(ctx, reading)
		}); err != nil {
			c.incKafkaError("save_error")
			c.incKafkaMessage("save_error")
			c.log.Error("actual weather save failed", slog.Any("error", err), slog.Int64("offset", message.Offset))
			sleep(ctx, 5*time.Second)
			continue
		}
		if err := c.reader.CommitMessages(ctx, message); err != nil {
			c.incKafkaError("commit_error")
			c.log.Warn("kafka commit failed", slog.Any("error", err), slog.Int64("offset", message.Offset))
			continue
		}
		c.incKafkaMessage("processed")
		c.incKafkaCommit("processed")
	}
}

func (c *ActualWeatherConsumer) isDuplicate(ctx context.Context, reading domain.ActualWeatherReading) (bool, error) {
	if c.dedup == nil || reading.ID == "" || c.dedupTTL <= 0 {
		return false, nil
	}
	inserted, err := c.dedup.MarkOnce(ctx, "actual-weather:"+reading.ID, c.dedupTTL)
	if err != nil {
		return false, err
	}
	return !inserted, nil
}

func (c *ActualWeatherConsumer) incKafkaMessage(result string) {
	if c.metrics != nil {
		c.metrics.IncKafkaMessage(c.reader.Config().Topic, result)
	}
}

func (c *ActualWeatherConsumer) incKafkaError(result string) {
	if c.metrics != nil {
		c.metrics.IncKafkaError(c.reader.Config().Topic, result)
	}
}

func (c *ActualWeatherConsumer) incKafkaCommit(result string) {
	if c.metrics != nil {
		c.metrics.IncKafkaCommit(c.reader.Config().Topic, result)
	}
}

type JSONActualWeatherDecoder struct{}

func NewJSONActualWeatherDecoder() JSONActualWeatherDecoder {
	return JSONActualWeatherDecoder{}
}

type jsonActualWeatherPayload struct {
	ID                 string    `json:"id"`
	StationID          string    `json:"stationId"`
	ObservedAt         time.Time `json:"observedAt"`
	TemperatureMin     *float64  `json:"temperatureMin"`
	TemperatureMax     *float64  `json:"temperatureMax"`
	PrecipitationTotal *float64  `json:"precipitationTotal"`
	WindSpeed          *float64  `json:"windSpeed"`
	WindGust           *float64  `json:"windGust"`
	Humidity           *float64  `json:"humidity"`
	Pressure           *float64  `json:"pressure"`
	Source             string    `json:"source"`
	TraceID            string    `json:"traceId"`
}

func (d JSONActualWeatherDecoder) Decode(message kafka.Message) (domain.ActualWeatherReading, error) {
	var payload jsonActualWeatherPayload
	if err := json.Unmarshal(message.Value, &payload); err != nil {
		return domain.ActualWeatherReading{}, err
	}
	if payload.StationID == "" || payload.ObservedAt.IsZero() {
		return domain.ActualWeatherReading{}, errors.New("stationId and observedAt are required")
	}
	if payload.ID == "" {
		payload.ID = uuid.NewString()
	}
	if payload.Source == "" {
		payload.Source = "sensor"
	}
	values := map[domain.WeatherParameter]float64{}
	add := func(parameter domain.WeatherParameter, value *float64) {
		if value != nil {
			values[parameter] = *value
		}
	}
	add(domain.ParameterTemperatureMin, payload.TemperatureMin)
	add(domain.ParameterTemperatureMax, payload.TemperatureMax)
	add(domain.ParameterPrecipitation, payload.PrecipitationTotal)
	add(domain.ParameterWindSpeed, payload.WindSpeed)
	add(domain.ParameterWindGust, payload.WindGust)
	add(domain.ParameterHumidity, payload.Humidity)
	add(domain.ParameterPressure, payload.Pressure)
	return domain.ActualWeatherReading{
		ID:         payload.ID,
		StationID:  payload.StationID,
		ObservedAt: payload.ObservedAt.UTC(),
		Values:     values,
		Source:     payload.Source,
		TraceID:    payload.TraceID,
		RawPayload: message.Value,
	}, nil
}

func retry(ctx context.Context, attempts int, delay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		if err = fn(); err == nil {
			return nil
		}
		sleep(ctx, delay)
		delay = minDuration(delay*2, 5*time.Second)
	}
	return err
}

func sleep(ctx context.Context, delay time.Duration) {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
