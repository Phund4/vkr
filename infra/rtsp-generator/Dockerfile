FROM golang:1.26-bookworm AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/rtsp-generator ./cmd/rtsp-generator

# ffmpeg и CA без apt-get (deb.debian.org часто недоступен из Docker Desktop).
FROM mwader/static-ffmpeg:7.1.1 AS ffmpeg

FROM debian:bookworm-slim
COPY --from=ffmpeg /ffmpeg /usr/local/bin/ffmpeg
COPY --from=ffmpeg /ffprobe /usr/local/bin/ffprobe
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /out/rtsp-generator /usr/local/bin/rtsp-generator
EXPOSE 8096
ENTRYPOINT ["/usr/local/bin/rtsp-generator"]
