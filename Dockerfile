# syntax=docker/dockerfile:1

# ---------------------
# Build stage
# ---------------------
FROM golang:1.25-alpine AS builder

WORKDIR /build
COPY go.mod ./
RUN go mod download

COPY src/ ./src/
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /lsm-exporter ./src/

# ---------------------
# Runtime stage
# ---------------------
FROM alpine:3.21

RUN apk --no-cache add ca-certificates

COPY --from=builder /lsm-exporter /lsm-exporter

ENV LSM_LISTEN_ADDR=0.0.0.0 \
    LSM_LISTEN_PORT=9090 \
    LLAMASWAP_URL=http://localhost:8080 \
    SCRAPE_TIMEOUT=5

EXPOSE 9090

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:${LSM_LISTEN_PORT:-9090}/health || exit 1

USER nobody:nogroup

ENTRYPOINT ["/lsm-exporter"]
