FROM golang:1.22.3 as builder
WORKDIR /go/src/dns-over-tls-proxy
COPY . .
RUN go get -d -v ./...
RUN CGO_ENABLED=0 go build -o /go/bin/dns-over-tls-proxy ./cmd

FROM alpine:3.19.1
COPY --from=builder /go/bin/dns-over-tls-proxy /go/bin/dns-over-tls-proxy

EXPOSE 53

ENTRYPOINT ["/go/bin/dns-over-tls-proxy"]

