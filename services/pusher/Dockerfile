FROM golang:1.26-bookworm AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/pusher ./cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/pusher /pusher
USER nonroot:nonroot
ENTRYPOINT ["/pusher"]
