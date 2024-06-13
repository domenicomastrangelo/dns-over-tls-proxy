# DNS-over-TLS Proxy Server

## Overview

This repository contains an implementation of a DNS-over-TLS proxy server that handles both TCP and UDP DNS queries. The proxy server forwards DNS queries to an upstream DNS-over-TLS server (e.g., Cloudflare) and caches the responses to improve performance.

## Features

- Handles DNS queries over TCP and UDP.
- Forwards DNS queries to an upstream DNS-over-TLS server over TCP.
- Supports concurrent processing of multiple incoming requests.
- Caches DNS responses to reduce latency and improve performance.
- Gracefully handles context cancellation and errors.
- Builds configuration from environment variables.

## Implementation Details

### Project Structure

- `main.go`: The entry point of the application. Initializes the logger and context, and starts the TCP and UDP DNS servers.
- `internal/proxy/tcp.go`: Contains the implementation for handling DNS queries over TCP.
- `internal/proxy/udp.go`: Contains the implementation for handling DNS queries over UDP.
- `internal/proxy/common.go`: Contains shared utility functions for handling DNS queries and managing the cache.
- `internal/cache`: A package that provides caching functionality using Redis.
- `internal/config`: A package that provides config functionality for customizable settings.

### Key Components

#### TCP DNS Server (`tcp.go`)

The TCP DNS server listens for incoming TCP DNS connections on port 53, processes DNS queries, forwards them to the upstream DNS-over-TLS server, and returns the responses to the clients.

#### UDP DNS Server (`udp.go`)

The UDP DNS server listens for incoming UDP DNS connections on port 53, processes DNS queries, forwards them to the upstream DNS-over-TLS server, and returns the responses to the clients.

#### Common Utilities (`common.go`)

- **TLS Connection Setup**: Establishes a TLS connection to the upstream DNS-over-TLS server.
- **Cache Management**: Uses Redis to cache DNS query responses.
- **DNS Message Handling**: Packs and unpacks DNS messages, and forwards DNS queries to the upstream server.

## Installation

1. Clone the repository:
    ```sh
    unzip dns-over-tls-proxy.zip
    cd dns-over-tls-proxy
    ```


## Usage

1. Start the DNS-over-TLS proxy server with docker-compose:
    ```sh
    docker-compose up
    ```

2. The server will start listening for incoming DNS queries on port 53.
 
## Configuration
 
- **DNS-over-TLS Server**: The upstream DNS-over-TLS server is configured to use Cloudflare's DNS (1.1.1.1) on port 853.
- **Certificates**: The server uses the `cloudflare.cert` file for the certificate pool.

## Design Choices

1. **Concurrent Request Handling**:
    - Both TCP and UDP servers use goroutines to handle incoming connections concurrently, allowing multiple requests to be processed simultaneously.

2. **Context Handling**:
    - The implementation uses Go's `context` package to handle request timeouts and cancellations gracefully. This ensures that long-running or stuck requests do not block the server.

3. **Caching**:
    - Responses from the upstream DNS-over-TLS server are cached using Redis to improve performance and reduce latency for repeated queries.

4. **Modular Design**:
    - The code is structured into separate files for TCP and UDP handling, with common utilities abstracted into a shared file. This makes the codebase more maintainable and extensible.

5. **Error Handling and Logging**:
    - Comprehensive error handling is implemented to log errors and ensure the server can recover gracefully. Structured logging is used for better readability and debugging.

