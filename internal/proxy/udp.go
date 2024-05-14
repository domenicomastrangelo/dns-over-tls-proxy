package proxy

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/miekg/dns"
)

func StartUDPDNSServer(ctx context.Context, logger *slog.Logger) error {
	// Listen for incoming UDP DNS connections on port 53
	listener, err := net.ListenPacket("udp", "0.0.0.0:53")
	if err != nil {
		logger.Error("Error starting UDP DNS server", "error", err.Error())
		return err
	}

	defer listener.Close()

	logger.Info("UDP DNS server started on port 53")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Shutting down UDP DNS server")
			return nil
		default:
			maxBytesPerUDPRequest := 4096
			buf := make([]byte, maxBytesPerUDPRequest)
			n, addr, err := listener.ReadFrom(buf)
			if err != nil {
				logger.Error("Error reading from connection", "error", err.Error())
				continue
			}

			go handleUDPConnection(ctx, listener, buf[:n], addr, logger)
		}
	}
}

func handleUDPConnection(ctx context.Context, listener net.PacketConn, buf []byte, addr net.Addr, logger *slog.Logger) {
	// Parse the DNS query
	msg := dns.Msg{}

	if err := msg.Unpack(buf); err != nil {
		logger.Error("Error unpacking DNS query", "error", err.Error())
		return
	}

	// Forward the DNS query to the upstream DNS-over-TLS server using a new connection
	var (
		resp []byte
		err  error
	)

	if resp, err = forwardDNSQuery(ctx, logger, &msg); err != nil {
		logger.Error("Error forwarding DNS query", "error", err.Error())
		return
	}

	// Send the DNS response back to the client
	if _, err := listener.WriteTo(resp, addr); err != nil {
		logger.Error("Error writing to connection", "error", err.Error())
		return
	}

	logger.Info(fmt.Sprintf("Sent DNS response to %s", addr))
}
