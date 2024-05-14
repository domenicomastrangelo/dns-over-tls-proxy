package proxy

import (
	"context"
	"encoding/binary"
	"fmt"
	"log/slog"
	"net"

	"github.com/miekg/dns"
)

func StartTCPDNSServer(ctx context.Context, logger *slog.Logger) error {
	// Listen for incoming TCP DNS connections on port 53
	listener, err := net.Listen("tcp", "0.0.0.0:53")
	if err != nil {
		logger.Error(fmt.Sprintf("Error starting TCP DNS server: %v", err.Error()))
	}

	defer listener.Close()

	logger.Info("TCP DNS server started on port 53")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Shutting down TCP DNS server")
			return nil
		default:
			conn, err := listener.Accept()
			if err != nil {
				logger.Error("Error accepting connection", "error", err.Error())
				continue
			}

			go handleTCPConnection(ctx, conn, logger)
		}
	}
}

func handleTCPConnection(ctx context.Context, conn net.Conn, logger *slog.Logger) {
	defer conn.Close()

	// Parse the DNS query
	logger.Info("Parsing the DNS query")

	// Read the 2-byte length prefix
	lengthBytes := make([]byte, 2)
	if _, err := conn.Read(lengthBytes); err != nil {
		logger.Error("Error reading length prefix", "error", err.Error())
		return
	}

	// Convert length prefix to an integer
	length := binary.BigEndian.Uint16(lengthBytes)
	dnsQuery := make([]byte, length)

	if _, err := conn.Read(dnsQuery); err != nil {
		logger.Error("Error reading DNS query", "error", err.Error())
		return
	}

	msg := dns.Msg{}

	if err := msg.Unpack([]byte(dnsQuery)); err != nil {
		logger.Error("Error unpacking DNS query", "error", err.Error())
		return
	}

	// Forward the DNS query to the upstream DNS-over-TLS server using a new connection
	logger.Info("Forwarding the DNS query")

	var (
		resp []byte
		err  error
	)
	if resp, err = forwardDNSQuery(ctx, logger, &msg); err != nil {
		logger.Error("Error forwarding DNS query", "error", err.Error())
		return
	}

	// Send the DNS response back to the client
	logger.Info("Sending the DNS response")

	respLength := make([]byte, 2)
	binary.BigEndian.PutUint16(respLength, uint16(len(resp)))
	if _, err := conn.Write(respLength); err != nil {
		logger.Error("Error writing length prefix", "error", err.Error())
	}

	if _, err := conn.Write(resp); err != nil {
		logger.Error("Error writing to connection", "error", err.Error())
	}
}
