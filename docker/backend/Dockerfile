# Ivy Backend Dockerfile (Development)
FROM golang:1.24-bookworm

WORKDIR /app

# Install migrate tool
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1

# Install air for hot reload (development)
RUN go install github.com/air-verse/air@v1.61.7

# Copy go.mod and go.sum
COPY go.mod go.sum* ./
RUN go mod download 2>/dev/null || true

# Copy source
COPY . .

EXPOSE 8081

CMD ["air", "-c", ".air.toml"]
