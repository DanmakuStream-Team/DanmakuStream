FROM golang:1.25-alpine AS builder
WORKDIR /src
RUN go env -w GOPROXY=https://goproxy.cn,direct
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG COMMIT_SHA=development
ARG BUILD_TIME=unknown
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "-s -w" -o /out/user-service ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget && addgroup -S app && adduser -S app -G app
WORKDIR /app
ARG SERVICE_VERSION=microservice-0.1.0
ARG COMMIT_SHA=development
ARG BUILD_TIME=1970-01-01T00:00:00Z
ENV SERVICE_NAME=user-service \
    SERVICE_VERSION=${SERVICE_VERSION} \
    COMMIT_SHA=${COMMIT_SHA} \
    BUILD_TIME=${BUILD_TIME}
COPY --from=builder /out/user-service /app/user-service
RUN mkdir -p /app/etc /app/data/avatars /app/data/messages /app/data/images
COPY etc/config.example.yaml /app/etc/config.yaml
RUN chown -R app:app /app
USER app
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=5 CMD wget -qO- http://127.0.0.1:8080/api/v1/health || exit 1
ENTRYPOINT ["/app/user-service"]
