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
	"os"
	"time"

	"dns-over-tls-proxy/internal/cache"

	"github.com/miekg/dns"
)

const (
	DNSOverTLSPort = "853"
	DNSOverTLSHost = "1.1.1.1"
)

// Forward the DNS query to the upstream DNS-over-TLS server using a new connection
func forwardDNSQuery(ctx context.Context, logger *slog.Logger, msg *dns.Msg) ([]byte, error) {
	certs, err := getCertificatePool(logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Error getting certificate pool: %v", err))
		return nil, err
	}

	tlsConn, err := tls.Dial("tcp", fmt.Sprintf("%s:%s", DNSOverTLSHost, DNSOverTLSPort), &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
		RootCAs:    certs,
	})
	if err != nil {
		logger.Error(fmt.Sprintf("Error connecting to %s:%s: %v", DNSOverTLSHost, DNSOverTLSPort, err))
		return nil, err
	}

	defer tlsConn.Close()

	tlsMsg := &dns.Msg{}
	tlsMsg.Question = append(tlsMsg.Question, msg.Question...)
	tlsMsg.RecursionDesired = true

	// If the query is in the cache, return the cached response
	cache := cache.GetCache(ctx)

	// Hash the DNS query to use as the cache key
	cacheKey, err := getCacheKeyFromQuestionSlice(msg.Question, *logger)
	if err != nil {
		// Log the error and continue without caching
		logger.Error(fmt.Sprintf("Error hashing DNS query: %v", err))
	}

	// Encode cacheKey to base64
	cacheKey = base64.StdEncoding.EncodeToString([]byte(cacheKey))

	if res := cache.Get(ctx, cacheKey); res.Err() == nil {
		logger.Info(fmt.Sprintf("Cache hit for query: %s", string(tlsMsg.Question[0].String())))

		// Setting the current request ID
		unpackedMsg := dns.Msg{}
		if err = unpackedMsg.Unpack([]byte(res.Val())); err != nil {
			logger.Error(fmt.Sprintf("Error unpacking DNS query: %v", err))
			return nil, err
		}
		unpackedMsg.Id = msg.Id

		var resp []byte
		resp, err = unpackedMsg.Pack()
		if err != nil {
			logger.Error(fmt.Sprintf("Error packing DNS response: %v", err))
			return nil, err
		}

		return resp, nil
	}

	// Setting the current request ID here if cache miss
	tlsMsg.Id = msg.Id

	query, err := tlsMsg.Pack()
	if err != nil {
		logger.Error(fmt.Sprintf("Error packing DNS query: %v", err))
		return nil, err
	}

	// DNS-over-TLS requires a 2-byte length prefix
	queryLength := make([]byte, 2)
	binary.BigEndian.PutUint16(queryLength, uint16(len(query)))
	if _, err = tlsConn.Write(queryLength); err != nil {
		logger.Error(fmt.Sprintf("Error writing length prefix: %v", err))
	}

	// Send the DNS query to the upstream DNS-over-TLS server
	if _, err = tlsConn.Write(query); err != nil {
		logger.Error(fmt.Sprintf("Error writing to connection: %v", err))
		return nil, err
	}

	// Read the DNS response from the upstream DNS-over-TLS server

	// Read the 2-byte length prefix first
	respLength := make([]byte, 2)
	if _, err = tlsConn.Read(respLength); err != nil {
		logger.Error(fmt.Sprintf("Error reading length prefix: %v", err))
	}
	length := binary.BigEndian.Uint16(respLength)

	respBuf := make([]byte, length)
	n, err := tlsConn.Read(respBuf)
	if err != nil {
		logger.Error(fmt.Sprintf("Error reading from connection: %v", err))
		return nil, err
	}

	// Cache the response
	if res := cache.Set(ctx, cacheKey, string(respBuf), 30*time.Minute); res.Err() != nil {
		logger.Error(fmt.Sprintf("Error caching response: %v", res.Err().Error()))
	}

	return respBuf[:n], nil
}

func getCacheKeyFromQuestionSlice(questions []dns.Question, logger slog.Logger) (string, error) {
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

	return string(hash), nil
}

func getCertificatePool(logger *slog.Logger) (*x509.CertPool, error) {
	certs := x509.NewCertPool()

	certContents, err := os.ReadFile("cloudflare.cert")
	if err != nil {
		logger.Error(fmt.Sprintf("Error reading certificate file: %v", err))
		return nil, err
	}

	certs.AppendCertsFromPEM(certContents)

	return certs, nil
}
