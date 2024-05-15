package proxy

import (
	"encoding/binary"
	"fmt"
	"net"

	"dns-over-tls-proxy/internal/config"

	"github.com/miekg/dns"
)

func StartTCPDNSServer(config config.Config) error {
	// Listen for incoming TCP DNS connections on port 53
	listener, err := net.Listen("tcp", "0.0.0.0:53")
	if err != nil {
		config.Logger.Error(fmt.Sprintf("Error starting TCP DNS server: %v", err.Error()))
	}

	defer listener.Close()

	config.Logger.Info("TCP DNS server started on port 53")

	for {
		select {
		case <-config.Ctx.Done():
			config.Logger.Info("Shutting down TCP DNS server")
			return nil
		default:
			conn, err := listener.Accept()
			if err != nil {
				config.Logger.Error("Error accepting connection", "error", err.Error())
				continue
			}

			go handleTCPConnection(config, conn)
		}
	}
}

func handleTCPConnection(config config.Config, conn net.Conn) {
	defer conn.Close()

	// Parse the DNS query
	config.Logger.Info("Parsing the DNS query")

	// Read the 2-byte length prefix
	lengthBytes := make([]byte, 2)
	if _, err := conn.Read(lengthBytes); err != nil {
		config.Logger.Error("Error reading length prefix", "error", err.Error())
		return
	}

	// Convert length prefix to an integer
	length := binary.BigEndian.Uint16(lengthBytes)
	dnsQuery := make([]byte, length)

	if _, err := conn.Read(dnsQuery); err != nil {
		config.Logger.Error("Error reading DNS query", "error", err.Error())
		return
	}

	msg := dns.Msg{}

	if err := msg.Unpack([]byte(dnsQuery)); err != nil {
		config.Logger.Error("Error unpacking DNS query", "error", err.Error())
		return
	}

	// Forward the DNS query to the upstream DNS-over-TLS server using a new connection
	config.Logger.Info("Forwarding the DNS query")

	var (
		resp []byte
		err  error
	)
	if resp, err = forwardDNSQuery(config, &msg); err != nil {
		config.Logger.Error("Error forwarding DNS query", "error", err.Error())
		return
	}

	// Send the DNS response back to the client
	config.Logger.Info("Sending the DNS response")

	respLength := make([]byte, 2)
	binary.BigEndian.PutUint16(respLength, uint16(len(resp)))
	if _, err := conn.Write(respLength); err != nil {
		config.Logger.Error("Error writing length prefix", "error", err.Error())
	}

	if _, err := conn.Write(resp); err != nil {
		config.Logger.Error("Error writing to connection", "error", err.Error())
	}
}
