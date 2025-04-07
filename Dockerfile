# First stage: build the Go binary
FROM golang:1.22.9 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the Go binary statically
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# Second stage: run the binary in a lightweight container
FROM alpine:latest

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/server .

# Cloud Run expects the service to listen on $PORT
ENV PORT=8080

EXPOSE 8080

CMD ["./server"]
