# ---- BUILD GO BINARY ----
    FROM golang:1.22.9 AS builder
    WORKDIR /app
    
    COPY go.mod go.sum ./
    RUN go mod download
    
    COPY . .
    RUN go build -o server
    
    # ---- FINAL IMAGE ----
    FROM debian:bookworm-slim
    
    # Install Python and venv
    RUN apt-get update && apt-get install -y \
      python3 \
      python3-pip \
      python3-venv \
      && apt-get clean
    
    WORKDIR /app
    
    # Create virtual environment
    RUN python3 -m venv /opt/venv
    
    # Use venv's pip to install requirements
    ENV PATH="/opt/venv/bin:$PATH"
    # Copy Python dependencies first
    COPY Sched/requirements.txt ./requirements.txt
    RUN pip install --no-cache-dir -r ./requirements.txt
    
    # Copy Go binary and your Python script
    COPY --from=builder /app/server .
    COPY . .
    
    ENV PORT=8080
    
    CMD ["./server"]
    