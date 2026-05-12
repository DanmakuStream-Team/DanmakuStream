FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./api/main.go

FROM alpine:3.20
RUN apk add --no-cache tzdata ca-certificates ffmpeg
WORKDIR /app
COPY --from=builder /app/server .
COPY etc/ etc/
RUN mkdir -p data/videos data/covers
EXPOSE 8080
CMD ["./server", "-f", "etc/config.yaml"]
