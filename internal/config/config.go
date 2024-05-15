package config

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Logger             *slog.Logger
	Ctx                context.Context
	Cache              *redis.Client
	DNSOverTLSPort     string
	DNSOverTLSHost     string
	RedisHost          string
	RedisPort          string
	DNSOverTLSCertPath string
}

func (c *Config) Check() error {
	if c.DNSOverTLSPort == "" {
		return fmt.Errorf("DNS_OVER_TLS_PORT is required")
	}

	if c.DNSOverTLSHost == "" {
		return fmt.Errorf("DNS_OVER_TLS_HOST is required")
	}

	if c.DNSOverTLSCertPath == "" {
		return fmt.Errorf("DNS_OVER_TLS_CERT_PATH is required")
	}

	if c.RedisHost == "" {
		return fmt.Errorf("REDIS_HOST is required")
	}

	if c.RedisPort == "" {
		return fmt.Errorf("REDIS_PORT is required")
	}

	return nil
}
