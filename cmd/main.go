package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"dns-over-tls-proxy/internal/proxy"
)

var (
	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}))
	ctx, cancel = context.WithCancel(context.Background())
)

func main() {
	// Create a channel to listen for OS signals
	signalChan := make(chan os.Signal, 2)
	defer close(signalChan)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Listen for incoming TCP DNS connections on port 53
	go proxy.StartTCPDNSServer(ctx, logger)

	// Listen for incoming UDP DNS connections on port 53
	go proxy.StartUDPDNSServer(ctx, logger)

	// Wait for an OS signal to shutdown the server
	<-signalChan
	logger.Info("Shutting down DNS-over-TLS proxy server")
	cancel()
}
