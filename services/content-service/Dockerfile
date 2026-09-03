FROM golang:1.25-alpine AS builder
WORKDIR /src
RUN go env -w GOPROXY=https://goproxy.cn,direct
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG SERVICE_VERSION=microservice-0.1.0
ARG COMMIT_SHA=0000000
ARG BUILD_TIME=1970-01-01T00:00:00Z
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/content-service ./cmd/server

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata wget \
    && addgroup -S app \
    && adduser -S -G app app \
    && mkdir -p /app/data \
    && chown -R app:app /app
WORKDIR /app
COPY --from=builder /out/content-service /app/content-service
ARG SERVICE_VERSION=microservice-0.1.0
ARG COMMIT_SHA=0000000
ARG BUILD_TIME=1970-01-01T00:00:00Z
ENV SERVICE_NAME=content-service \
    SERVICE_VERSION=${SERVICE_VERSION} \
    COMMIT_SHA=${COMMIT_SHA} \
    BUILD_TIME=${BUILD_TIME} \
    PORT=8080 \
    STORAGE_DIR=/app/data
USER app
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=2s --start-period=10s --retries=3 CMD wget -q -O /dev/null http://127.0.0.1:8080/api/v1/livez || exit 1
ENTRYPOINT ["/app/content-service"]
