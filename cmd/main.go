package main

import (
	"context"
	"log/slog"
	"os"

	"dns-over-tls-proxy/internal/proxy"
)

var (
	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}))
	ctx = context.Background()
)

func main() {
	// Listen for incoming TCP DNS connections on port 53
	ch := make(chan bool)
	defer close(ch)

	go proxy.StartTCPDNSServer(ch, ctx, logger)

	// Listen for incoming UDP DNS connections on port 53
	go proxy.StartUDPDNSServer(ch, ctx, logger)

	<-ch
}
