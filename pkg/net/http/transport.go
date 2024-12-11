package http

import (
	"net/http"
	"time"

	comnet "github.com/kkrt-labs/kakarot-controller/pkg/net"
	comtime "github.com/kkrt-labs/kakarot-controller/pkg/time"

	"golang.org/x/net/http2"
)

// TransportConfig is a configuration for http.Transport
type TransportConfig struct {
	Dialer                *comnet.DialerConfig
	IdleConnTimeout       *comtime.Duration
	ResponseHeaderTimeout *comtime.Duration
	ExpectContinueTimeout *comtime.Duration
	MaxIdleConnsPerHost   int
	MaxConnsPerHost       int
	DisableKeepAlives     bool
	DisableCompression    bool
	EnableHTTP2           bool
}

// SetDefault sets default values for TransportConfig
func (cfg *TransportConfig) SetDefault() *TransportConfig {
	if cfg.Dialer == nil {
		cfg.Dialer = new(comnet.DialerConfig)
	}
	cfg.Dialer.SetDefault()

	if cfg.IdleConnTimeout == nil {
		cfg.IdleConnTimeout = &comtime.Duration{Duration: 90 * time.Second}
	}

	if cfg.ResponseHeaderTimeout == nil {
		cfg.ResponseHeaderTimeout = &comtime.Duration{Duration: 0}
	}

	if cfg.ExpectContinueTimeout == nil {
		cfg.ExpectContinueTimeout = &comtime.Duration{Duration: time.Second}
	}

	return cfg
}

// NewTransport creates a new http.Transport
func NewTransport(cfg *TransportConfig) (*http.Transport, error) {
	// Create dialer
	dlr := comnet.NewDialer(cfg.Dialer)

	// Create transport
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dlr.DialContext,
		DisableKeepAlives:     cfg.DisableKeepAlives,
		DisableCompression:    cfg.DisableCompression,
		MaxIdleConnsPerHost:   cfg.MaxIdleConnsPerHost,
		MaxConnsPerHost:       cfg.MaxConnsPerHost,
		IdleConnTimeout:       cfg.IdleConnTimeout.Duration,
		ResponseHeaderTimeout: cfg.ResponseHeaderTimeout.Duration,
		ExpectContinueTimeout: cfg.ExpectContinueTimeout.Duration,
	}

	// Configure transport to use HTTP/2
	if cfg.EnableHTTP2 {
		err := http2.ConfigureTransport(transport)
		if err != nil {
			return nil, err
		}
	}

	return transport, nil
}
