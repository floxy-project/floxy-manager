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

# Install git and openssh (needed for go mod download and SSH access to private repos)
RUN apk add --no-cache git openssh

# Copy go mod files
COPY go.mod go.sum ./

ENV GOPRIVATE=github.com/rom8726/floxy*

# Option 1: Use SSH key from build secret (recommended for local dev)
# Mount secret: --secret id=ssh_key,src=$HOME/.ssh/id_rsa
# Option 2: Use GitHub Personal Access Token from build arg (for CI/CD)
# Pass token as build arg: --build-arg GITHUB_TOKEN=your_token
ARG GITHUB_TOKEN

# Configure git and download dependencies in single RUN to access secrets
RUN --mount=type=secret,id=ssh_key \
    mkdir -p -m 0600 ~/.ssh && \
    if [ -f /run/secrets/ssh_key ]; then \
        echo "Using SSH key for private repo access" && \
        cp /run/secrets/ssh_key ~/.ssh/id_rsa && \
        chmod 600 ~/.ssh/id_rsa && \
        ssh-keyscan github.com >> ~/.ssh/known_hosts && \
        git config --global url."git@github.com:".insteadOf "https://github.com/"; \
    elif [ -n "${GITHUB_TOKEN}" ]; then \
        echo "Using GitHub token for private repo access" && \
        git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/" && \
        git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "git@github.com:"; \
    else \
        echo "ERROR: Neither SSH key nor GITHUB_TOKEN provided!" && \
        echo "Please provide either:" && \
        echo "  1. SSH key: --secret id=ssh_key,src=\$HOME/.ssh/id_rsa" && \
        echo "  2. GitHub token: --build-arg GITHUB_TOKEN=your_token" && \
        exit 1; \
    fi && \
    go mod download

# Copy source code
COPY . .

# Build Go application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest AS prod

# Install ca-certificates for HTTPS
RUN apk add --no-cache ca-certificates
RUN apk add --no-cache curl
RUN apk add --no-cache tzdata

WORKDIR /app

# Copy Go binary
COPY --from=go-builder /app/main .

# Copy frontend build
COPY --from=frontend-builder /app/web/dist ./web/dist

# Copy migrations
COPY --from=go-builder /app/migrations ./migrations

# Expose port
EXPOSE 3001

# Set environment variables
ENV PORT=3001

# Run the application
CMD ["./main"]
