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

## Future Improvements

- **Enhanced Logging**:
    - Implement more detailed and structured logging for better observability and monitoring.
- **Security Enhancements**:
    - Add support for authentication and authorization to restrict access to the proxy server.

## Deployment Considerations
#### (Imagine this proxy being deployed in an infrastructure)

### Security Concerns

When deploying this DNS-over-TLS proxy server in an infrastructure, several security concerns should be addressed:

1. **Network Security**:
    - **TLS Security**: Ensure that the TLS connection to the upstream DNS server is properly secured. Use strong ciphers and enforce TLS version 1.2 or higher.
    - **Firewall Rules**: Restrict access to the DNS proxy server to trusted IP addresses or networks. Configure firewall rules to limit inbound and outbound traffic to necessary ports (e.g., port 53 for DNS, port 853 for DNS-over-TLS).

2. **Authentication and Authorization**:
    - **Access Control**: Implement authentication and authorization mechanisms to restrict who can query the DNS proxy server. This can prevent unauthorized usage and potential abuse.

3. **Logging and Monitoring**:
    - **Intrusion Detection**: Monitor logs for unusual patterns or potential attacks (e.g., DDoS, DNS amplification attacks). Use tools like fail2ban or cloud-based security services.
    - **Alerting**: Set up alerting for any suspicious activity or errors that indicate a potential security breach.

4. **Data Integrity and Privacy**:
    - **Encryption**: Ensure that all sensitive data, including cached DNS responses, are stored securely. Consider encrypting data at rest.
    - **Minimal Data Exposure**: Avoid logging sensitive information and ensure logs are rotated and managed securely.

5. **Denial of Service (DoS) Protection**:
    - **Rate Limiting**: Implement rate limiting to protect against DoS attacks. Limit the number of queries per second from a single IP address.
    - **Resource Management**: Monitor resource usage (CPU, memory) and set limits to prevent the server from being overwhelmed by high traffic volumes.

### Integration in a Distributed, Microservices-Oriented, and Containerized Architecture

1. **Containerization**:
    - **Docker**: Containerize the DNS-over-TLS proxy server using Docker. Create a Dockerfile to build the application image.
    - **Kubernetes**: Deploy the Docker container in a Kubernetes cluster for better orchestration and management. Use Kubernetes services to expose the DNS proxy server.

2. **Service Discovery and Load Balancing**:
    - **Service Mesh**: Integrate the proxy server with a service mesh (e.g., Istio) for service discovery, load balancing, and secure communication between services.

3. **Scaling and Fault Tolerance**:
    - **Auto-scaling**: Configure auto-scaling policies to handle increased traffic. Use Kubernetes Horizontal Pod Autoscaler (HPA) to scale the number of proxy server instances based on CPU/memory usage.

4. **Configuration Management**:
    - **ConfigMaps and Secrets**: Store configuration details (e.g., upstream DNS server addresses, certificates) in Kubernetes ConfigMaps and Secrets.
    - **Environment Variables**: Use environment variables for configuration to ensure the application can be easily configured and deployed in different environments.

5. **CI/CD Pipeline**:
    - **Continuous Integration**: Set up a CI pipeline to automate testing and building of Docker images.
    - **Continuous Deployment**: Use CD tools (e.g., Jenkins, GitLab CI/CD) to automate the deployment of the DNS proxy server to the Kubernetes cluster.

### Future Improvements

1. **Enhanced Security Features**:
    - **Audit Logging**: Add detailed audit logging to track all DNS queries and responses for security audits and compliance.

2. **Observability Enhancements**:
    - **Metrics and Monitoring**: Integrate with monitoring tools (e.g., Prometheus, Grafana) to collect metrics and visualize the performance of the DNS proxy server.

3. **Configuration and Management**:
    - **Web Dashboard**: Develop a web-based dashboard for monitoring and managing the DNS proxy server.

4. **Support for Additional Protocols**:
    - **DNS-over-HTTPS (DoH)**: Extend the proxy server to support DNS-over-HTTPS for environments where DoH is preferred over DNS-over-TLS.

5. **Tests**:
    - **Unit Tests**: Write comprehensive unit tests to ensure the correctness of the DNS proxy server implementation.
6. **Documentation**:
    - **API Documentation**: Generate API documentation for the DNS proxy server to help users understand how to interact with the server.
7. **Retry logic**:
    - **Retry Mechanism**: Implement a retry mechanism for failed DNS queries to improve reliability and resilience.
