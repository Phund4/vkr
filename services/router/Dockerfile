FROM golang:1.26-bookworm AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/router ./cmd/router

# Статический ffmpeg — без apt-get (сборка не зависит от deb.debian.org).
FROM mwader/static-ffmpeg:7.1.1 AS ffmpeg

FROM debian:bookworm-slim
COPY --from=ffmpeg /ffmpeg /usr/local/bin/ffmpeg
COPY --from=ffmpeg /ffprobe /usr/local/bin/ffprobe
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /out/router /usr/local/bin/router
WORKDIR /app
ENTRYPOINT ["/usr/local/bin/router"]
