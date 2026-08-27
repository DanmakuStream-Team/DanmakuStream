FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
RUN go env -w GOPROXY=https://goproxy.cn,direct
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o server ./api/main.go

FROM alpine:3.20
RUN apk add --no-cache tzdata ca-certificates ffmpeg
WORKDIR /app
COPY --from=builder /app/server .
COPY etc/ etc/
RUN mkdir -p data/videos data/covers && \
    addgroup -S app && adduser -S app -G app && \
    chown -R app:app /app
USER app
EXPOSE 8080
# 容器级健康检查（Compose/K8s 会另行配置更完整的探针）
HEALTHCHECK --interval=15s --timeout=5s --start-period=20s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/api/v1/health > /dev/null 2>&1 || exit 1
CMD ["./server", "-f", "etc/config.yaml"]
