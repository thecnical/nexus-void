FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/server .

# Create brain data directory
RUN mkdir -p /var/lib/nexus-void/brain

EXPOSE 8080

CMD ["./server", "-addr", ":8080", "-brain", "/var/lib/nexus-void/brain"]
