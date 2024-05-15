package proxy

import (
	"fmt"
	"net"

	"dns-over-tls-proxy/internal/config"

	"github.com/miekg/dns"
)

func StartUDPDNSServer(config config.Config) error {
	// Listen for incoming UDP DNS connections on port 53
	listener, err := net.ListenPacket("udp", "0.0.0.0:53")
	if err != nil {
		config.Logger.Error("Error starting UDP DNS server", "error", err.Error())
		return err
	}

	defer listener.Close()

	config.Logger.Info("UDP DNS server started on port 53")

	for {
		select {
		case <-config.Ctx.Done():
			config.Logger.Info("Shutting down UDP DNS server")
			return nil
		default:
			maxBytesPerUDPRequest := 4096
			buf := make([]byte, maxBytesPerUDPRequest)
			n, addr, err := listener.ReadFrom(buf)
			if err != nil {
				config.Logger.Error("Error reading from connection", "error", err.Error())
				continue
			}

			go handleUDPConnection(config, listener, buf[:n], addr)
		}
	}
}

func handleUDPConnection(config config.Config, listener net.PacketConn, buf []byte, addr net.Addr) {
	// Parse the DNS query
	msg := dns.Msg{}

	if err := msg.Unpack(buf); err != nil {
		config.Logger.Error("Error unpacking DNS query", "error", err.Error())
		return
	}

	// Forward the DNS query to the upstream DNS-over-TLS server using a new connection
	var (
		resp []byte
		err  error
	)

	if resp, err = forwardDNSQuery(config, &msg); err != nil {
		config.Logger.Error("Error forwarding DNS query", "error", err.Error())
		return
	}

	// Send the DNS response back to the client
	if _, err := listener.WriteTo(resp, addr); err != nil {
		config.Logger.Error("Error writing to connection", "error", err.Error())
		return
	}

	config.Logger.Info(fmt.Sprintf("Sent DNS response to %s", addr))
}
