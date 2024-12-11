package http

import (
	"net/http"

	comtime "github.com/kkrt-labs/kakarot-controller/pkg/time"
)

// Config for creating an HTTP Client
type ClientConfig struct {
	Transport *TransportConfig  `json:"transport,omitempty"`
	Timeout   *comtime.Duration `json:"timeout,omitempty"`
}

func (cfg *ClientConfig) SetDefault() *ClientConfig {
	if cfg.Transport == nil {
		cfg.Transport = new(TransportConfig)
	}
	cfg.Transport.SetDefault()

	if cfg.Timeout == nil {
		cfg.Timeout = &comtime.Duration{Duration: 0}
	}

	return cfg
}

// New creates a new HTTP client
func NewClient(cfg *ClientConfig) (*http.Client, error) {
	trnsprt, err := NewTransport(cfg.Transport)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Transport: trnsprt,
		Timeout:   cfg.Timeout.Duration,
	}, nil
}
