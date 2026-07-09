package kafka

import (
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/scram"
)

type SecurityConfig struct {
	Protocol   string
	Mechanism  string
	Username   string
	Password   string
	SkipVerify bool
}

func (cfg SecurityConfig) enabled() bool {
	return strings.EqualFold(cfg.Protocol, "SASL_SSL") || cfg.Username != ""
}

func (cfg SecurityConfig) saslMechanism() (sasl.Mechanism, error) {
	if !cfg.enabled() {
		return nil, nil
	}
	switch strings.ToUpper(cfg.Mechanism) {
	case "", "SCRAM-SHA-512":
		return scram.Mechanism(scram.SHA512, cfg.Username, cfg.Password)
	case "SCRAM-SHA-256":
		return scram.Mechanism(scram.SHA256, cfg.Username, cfg.Password)
	default:
		return nil, fmt.Errorf("unsupported kafka SASL mechanism %q", cfg.Mechanism)
	}
}

func (cfg SecurityConfig) transport() (*kafka.Transport, error) {
	mechanism, err := cfg.saslMechanism()
	if err != nil {
		return nil, err
	}
	if !cfg.enabled() {
		return nil, nil
	}
	return &kafka.Transport{
		TLS:  &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: cfg.SkipVerify},
		SASL: mechanism,
	}, nil
}

func (cfg SecurityConfig) dialer() (*kafka.Dialer, error) {
	mechanism, err := cfg.saslMechanism()
	if err != nil {
		return nil, err
	}
	if !cfg.enabled() {
		return nil, nil
	}
	return &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		TLS:           &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: cfg.SkipVerify},
		SASLMechanism: mechanism,
	}, nil
}
