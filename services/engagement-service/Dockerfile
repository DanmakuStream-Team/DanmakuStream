FROM golang:1.25-alpine AS builder
WORKDIR /src
ARG COMMIT_SHA=0000000
ARG BUILD_TIME=1970-01-01T00:00:00Z
ARG SERVICE_VERSION=microservice-0.1.0
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "-s -w -X main.version=${SERVICE_VERSION} -X main.commit=${COMMIT_SHA} -X main.buildTime=${BUILD_TIME}" -o /out/engagement-service ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata ffmpeg && addgroup -S app && adduser -S app -G app
WORKDIR /app
ENV GIN_MODE=release
COPY --from=builder /out/engagement-service ./engagement-service
COPY etc ./etc
USER app
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=3 CMD wget -qO- http://127.0.0.1:8080/api/v1/health >/dev/null || exit 1
ENTRYPOINT ["./engagement-service"]
CMD ["-f", "etc/config.example.yaml"]
