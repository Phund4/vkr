FROM golang:1.26-bookworm AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/data-service ./cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/data-service /data-service
USER nonroot:nonroot
ENTRYPOINT ["/data-service"]
