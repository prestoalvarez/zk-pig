package net

import (
	"net"
	"time"

	comtime "github.com/kkrt-labs/kakarot-controller/pkg/time"
)

// DialerConfig is a configuration for a net.Dialer.
type DialerConfig struct {
	Timeout   *comtime.Duration
	KeepAlive *comtime.Duration
}

// SetDefault sets the default values for the DialerConfig.
func (cfg *DialerConfig) SetDefault() *DialerConfig {
	if cfg.Timeout == nil {
		cfg.Timeout = &comtime.Duration{Duration: 30 * time.Second}
	}

	if cfg.KeepAlive == nil {
		cfg.KeepAlive = &comtime.Duration{Duration: 30 * time.Second}
	}

	return cfg
}

// NewDialer creates a new net.Dialer with the given configuration.
func NewDialer(cfg *DialerConfig) *net.Dialer {
	return &net.Dialer{
		Timeout:   cfg.Timeout.Duration,
		KeepAlive: cfg.KeepAlive.Duration,
	}
}
