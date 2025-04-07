# First stage: build the Go binary
FROM golang:1.22.9 AS builder


# Set the working directory inside the container
WORKDIR /app

# Copy go files and download dependencies
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy the rest of the app code
COPY . .

# Build the Go binary
RUN go build -o server

# Use a minimal base image to run the binary
FROM debian:bookworm-slim
WORKDIR /app
COPY --from=0 /app/server .

# Set environment variable
ENV PORT=8080

# Run the server
CMD ["./server"]