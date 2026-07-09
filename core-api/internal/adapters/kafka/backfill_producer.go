package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"

	"weather-accuracy/core-api/internal/domain"
	"weather-accuracy/core-api/internal/platform/metrics"
)

type ProducerConfig struct {
	Brokers  []string
	Topic    string
	Security SecurityConfig
}

type BackfillProducer struct {
	writer  *kafka.Writer
	topic   string
	metrics *metrics.Registry
}

type backfillRequestedEvent struct {
	EventType          string                `json:"eventType"`
	EventVersion       string                `json:"eventVersion"`
	JobID              string                `json:"jobId"`
	Status             domain.BackfillStatus `json:"status"`
	DateFrom           time.Time             `json:"dateFrom"`
	DateTo             time.Time             `json:"dateTo"`
	RegionID           string                `json:"regionId,omitempty"`
	StationID          string                `json:"stationId,omitempty"`
	Metric             domain.Metric         `json:"metric,omitempty"`
	CalculationVersion string                `json:"calculationVersion"`
	CreatedAt          time.Time             `json:"createdAt"`
	PublishedAt        time.Time             `json:"publishedAt"`
}

func NewBackfillProducer(cfg ProducerConfig, registry *metrics.Registry) *BackfillProducer {
	transport, err := cfg.Security.transport()
	if err != nil {
		transport = nil
	}
	return &BackfillProducer{
		topic: cfg.Topic,
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(cfg.Brokers...),
			Topic:                  cfg.Topic,
			Balancer:               &kafka.Hash{},
			RequiredAcks:           kafka.RequireAll,
			AllowAutoTopicCreation: false,
			Async:                  false,
			BatchTimeout:           10 * time.Millisecond,
			WriteTimeout:           10 * time.Second,
			ReadTimeout:            10 * time.Second,
			Transport:              transport,
		},
		metrics: registry,
	}
}

func (p *BackfillProducer) PublishBackfillRequested(ctx context.Context, job domain.BackfillJob, request domain.BackfillRequest) error {
	event := backfillRequestedEvent{
		EventType:          "backfill.requested",
		EventVersion:       "v1",
		JobID:              job.ID,
		Status:             job.Status,
		DateFrom:           request.DateFrom,
		DateTo:             request.DateTo,
		RegionID:           request.RegionID,
		StationID:          request.StationID,
		Metric:             request.Metric,
		CalculationVersion: request.CalculationVersion,
		CreatedAt:          job.CreatedAt,
		PublishedAt:        time.Now().UTC(),
	}
	payload, err := json.Marshal(event)
	if err != nil {
		p.incPublish("marshal_error")
		return err
	}
	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(job.ID),
		Value: payload,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(event.EventType)},
			{Key: "event_version", Value: []byte(event.EventVersion)},
		},
		Time: event.PublishedAt,
	})
	if err != nil {
		p.incPublish("write_error")
		return err
	}
	p.incPublish("published")
	return nil
}

func (p *BackfillProducer) Close() error {
	return p.writer.Close()
}

func (p *BackfillProducer) incPublish(result string) {
	if p.metrics != nil {
		p.metrics.IncKafkaPublish(p.topic, result)
	}
}
