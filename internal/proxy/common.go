package proxy

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"dns-over-tls-proxy/internal/cache"

	"github.com/miekg/dns"
	"github.com/redis/go-redis/v9"
)

const (
	DNSOverTLSPort = "853"
	DNSOverTLSHost = "1.1.1.1"
)

// Forward the DNS query to the upstream DNS-over-TLS server using a new connection
func forwardDNSQuery(ctx context.Context, logger *slog.Logger, msg *dns.Msg) ([]byte, error) {
	tlsConn, conn, err := createTLSMessage(ctx, logger)
	if err != nil {
		logger.Error("Error creating DNS-over-TLS message", "error", err.Error())

		return nil, err
	}
	defer conn.Close()
	defer tlsConn.Close()

	tlsMsg := &dns.Msg{}
	tlsMsg.Question = append(tlsMsg.Question, msg.Question...)
	tlsMsg.RecursionDesired = true

	// If the query is in the cache, return the cached response
	cache := cache.GetCache(ctx)

	// Hash the DNS query to use as the cache key
	cacheKey, err := getCacheKeyFromQuestionSlice(msg.Question)
	if err != nil {
		// Log the error and continue without caching
		logger.Error("Error hashing DNS query", "error", err.Error())
	}

	cachedMessage, err := getCachedMessage(ctx, logger, cache, cacheKey, msg, tlsMsg)
	if err != nil {
		logger.Error("Error getting cached message", "error", err.Error())

		return nil, err
	}

	if cachedMessage != nil {
		return cachedMessage, nil
	}

	// We need to set the id here as it would be the same for the cache and for the new result
	tlsMsg.Id = msg.Id

	query, err := tlsMsg.Pack()
	if err != nil {
		logger.Error("Error packing DNS query", "error", err.Error())
		return nil, err
	}

	// DNS-over-TLS requires a 2-byte length prefix
	tlsRequestPrefix := make([]byte, 2)
	binary.BigEndian.PutUint16(tlsRequestPrefix, uint16(len(query)))

	errChan := make(chan error, 1)
	defer close(errChan)

	resChan := make(chan []byte, 1)
	defer close(resChan)

	go writeToDNSoverTLSServer(ctx, errChan, resChan, tlsConn, tlsRequestPrefix, query)

	select {
	case <-ctx.Done():
		logger.Info("Context was cancelled. Killing request")
		return nil, ctx.Err()
	case err := <-errChan:
		logger.Error("Error during connection to the dns-over-tls server", "error", err.Error())
		return nil, err
	case respBuf := <-resChan:
		// Cache the response
		if res := cache.Set(ctx, cacheKey, string(respBuf), 30*time.Minute); res.Err() != nil {
			logger.Error("Error caching response", "error", res.Err().Error())
		}
		return respBuf, nil
	}
}

func getCacheKeyFromQuestionSlice(questions []dns.Question) (string, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(questions)
	if err != nil {
		return "", err
	}

	questionBytes := buf.Bytes()

	hasher := sha256.New()
	hasher.Write(questionBytes)
	hash := hasher.Sum(nil)

	// Encode cacheKey to base64
	cacheKey := base64.StdEncoding.EncodeToString(hash)

	return cacheKey, nil
}

func getCertificatePool() (*x509.CertPool, error) {
	certs := x509.NewCertPool()

	certContents, err := os.ReadFile("cloudflare.cert")
	if err != nil {
		return nil, err
	}

	certs.AppendCertsFromPEM(certContents)

	return certs, nil
}

func createTLSMessage(ctx context.Context, logger *slog.Logger) (*tls.Conn, net.Conn, error) {
	certs, err := getCertificatePool()
	if err != nil {
		logger.Error("Error getting certificate pool", "error", err.Error())
		return nil, nil, err
	}

	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
	}
	conn, err := dialer.DialContext(
		ctx,
		"tcp",
		fmt.Sprintf("%s:%s", DNSOverTLSHost, DNSOverTLSPort),
	)
	if err != nil {
		logger.Error(
			fmt.Sprintf("Error connecting to %s:%s", DNSOverTLSHost, DNSOverTLSPort),
			"error",
			err,
		)
		return nil, nil, err
	}

	tlsConn := tls.Client(conn, &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
		RootCAs:    certs,
		ServerName: DNSOverTLSHost,
	})

	return tlsConn, conn, nil
}

func getCachedMessage(ctx context.Context, logger *slog.Logger, cache *redis.Client, cacheKey string, prevMsg *dns.Msg, tlsMsg *dns.Msg) ([]byte, error) {
	if res := cache.Get(ctx, cacheKey); res.Err() == nil {
		logger.Info(fmt.Sprintf("Cache hit for query: %s", string(tlsMsg.Question[0].String())))

		// Setting the current request ID
		unpackedMsg := dns.Msg{}
		if err := unpackedMsg.Unpack([]byte(res.Val())); err != nil {
			logger.Error("Error unpacking DNS query", "error", err.Error())
			return nil, err
		}
		unpackedMsg.Id = prevMsg.Id

		var resp []byte
		resp, err := unpackedMsg.Pack()
		if err != nil {
			logger.Error("Error packing DNS response", "error", err.Error())
			return nil, err
		}

		return resp, nil
	}

	return nil, nil
}

func writeToDNSoverTLSServer(ctx context.Context, errChan chan error, resChan chan []byte, tlsConn *tls.Conn, tlsRequestPrefix []byte, query []byte) {
	doneChan := make(chan struct{})

	go func() {
		defer close(doneChan)
		if _, err := (*tlsConn).Write(tlsRequestPrefix); err != nil {
			errChan <- err
			return
		}

		// Send the DNS query to the upstream DNS-over-TLS server
		if _, err := (*tlsConn).Write(query); err != nil {
			errChan <- err
			return
		}

		// Read the DNS response from the upstream DNS-over-TLS server

		// Read the 2-byte length prefix first
		respLength := make([]byte, 2)
		if _, err := (*tlsConn).Read(respLength); err != nil {
			errChan <- err
			return
		}
		length := binary.BigEndian.Uint16(respLength)

		respBuf := make([]byte, length)
		_, err := (*tlsConn).Read(respBuf)
		if err != nil {
			errChan <- err
			return
		}

		resChan <- respBuf
	}()

	select {
	case <-ctx.Done():
		errChan <- ctx.Err()
	case <-doneChan:
	}
}
