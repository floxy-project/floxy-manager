.PHONY: build run dev clean docker-build docker-run

# Build everything
build:
	@echo "Building frontend..."
	cd web && npm install && npm run build
	@echo "Building Go backend..."
	go build -o bin/floxy-ui .

# Run in development mode
dev:
	@echo "Starting development mode..."
	@echo "Press Ctrl+C to stop"
	@trap 'kill %1; kill %2' INT; \
	cd web && npm run dev & \
	go run . & \
	wait

# Run production build
run: build
	./bin/floxy-ui

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf web/dist/
	rm -rf web/node_modules/

# Docker build
docker-build:
	docker build -t floxy-ui:latest .

# Docker run
docker-run:
	docker run -p 3001:3001 \
		-e DB_HOST=host.docker.internal \
		-e DB_PORT=5435 \
		-e DB_NAME=floxy \
		-e DB_USER=floxy \
		-e DB_PASSWORD=password \
		floxy-ui:latest

# Install dependencies
install:
	cd web && npm install
	go mod download

# Test
test:
	go test ./...

# Help
help:
	@echo "Available targets:"
	@echo "  build        - Build both frontend and backend"
	@echo "  dev          - Start development mode (frontend + backend)"
	@echo "  run          - Run production build"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  install      - Install dependencies"
	@echo "  test         - Run tests"
	@echo "  help         - Show this help"
