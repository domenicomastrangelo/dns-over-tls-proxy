package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"dns-over-tls-proxy/internal/cache"
	"dns-over-tls-proxy/internal/config"
	"dns-over-tls-proxy/internal/proxy"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}))

	ctx, cancel := context.WithCancel(context.Background())

	config := config.Config{
		Logger:             logger,
		Ctx:                ctx,
		DNSOverTLSPort:     os.Getenv("DNS_OVER_TLS_PORT"),
		DNSOverTLSHost:     os.Getenv("DNS_OVER_TLS_HOST"),
		RedisHost:          os.Getenv("REDIS_HOST"),
		RedisPort:          os.Getenv("REDIS_PORT"),
		DNSOverTLSCertPath: os.Getenv("DNS_OVER_TLS_CERT_PATH"),
	}

	cache := cache.GetCache(config)
	defer cache.Close()

	config.Cache = cache

	if err := config.Check(); err != nil {
		logger.Error("Error checking config", "error", err.Error())
		os.Exit(1)
	}

	// Create a channel to listen for OS signals
	signalChan := make(chan os.Signal, 2)
	defer close(signalChan)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Listen for incoming TCP DNS connections on port 53
	go proxy.StartTCPDNSServer(config)

	// Listen for incoming UDP DNS connections on port 53
	go proxy.StartUDPDNSServer(config)

	// Wait for an OS signal to shutdown the server
	<-signalChan
	logger.Info("Shutting down DNS-over-TLS proxy server")
	cancel()
}
