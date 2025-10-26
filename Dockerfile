# Multi-stage build for Go backend + TypeScript frontend
FROM node:18-alpine AS frontend-builder

WORKDIR /app/web

# Copy package files
COPY web/package*.json ./
RUN npm ci

# Copy source code
COPY web/ ./

# Build frontend
RUN npm run build

# Go builder stage
FROM golang:1.25-alpine AS go-builder

WORKDIR /app

# Install git (needed for go mod download)
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build Go application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy Go binary
COPY --from=go-builder /app/main .

# Copy frontend build
COPY --from=frontend-builder /app/web/dist ./web/dist

# Expose port
EXPOSE 3001

# Set environment variables
ENV PORT=3001
ENV DB_HOST=localhost
ENV DB_PORT=5435
ENV DB_NAME=floxy
ENV DB_USER=floxy
ENV DB_PASSWORD=password

# Run the application
CMD ["./main"]
