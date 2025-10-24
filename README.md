# Floxy Manager

A modern web interface for managing Floxy workflows with Go backend and React frontend.

## Features

- **Go Backend**: High-performance backend using the Floxy library
- **React Frontend**: Modern TypeScript/React interface
- **Single Container**: Both backend and frontend in one Docker image
- **Real-time Monitoring**: Live workflow statistics and monitoring
- **Workflow Visualization**: Interactive workflow graphs
- **Instance Management**: View and manage workflow instances

## Quick Start

### Development Mode

```bash
# Install dependencies
make install

# Start development mode (frontend + backend)
make dev
```

- Frontend: http://localhost:3000
- Backend: http://localhost:3001

### Production Build

```bash
# Build everything
make build

# Run production build
make run
```

### Docker

```bash
# Build Docker image
make docker-build

# Run with Docker
make docker-run
```

## Environment Variables

- `PORT` - Server port (default: 3001)
- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5435)
- `DB_NAME` - Database name (default: floxy)
- `DB_USER` - Database user (default: floxy)
- `DB_PASSWORD` - Database password (default: password)

## API Endpoints

- `GET /api/workflows` - List workflow definitions
- `GET /api/workflows/{id}` - Get workflow definition
- `GET /api/workflows/{id}/instances` - Get workflow instances
- `GET /api/instances` - List all instances
- `GET /api/instances/{id}` - Get workflow instance
- `GET /api/instances/{id}/steps` - Get instance steps
- `GET /api/instances/{id}/events` - Get instance events
- `GET /api/stats` - Get workflow statistics

## Architecture

```
floxy-manager/
├── cmd/server/          # Go main application
├── internal/            # Internal Go packages
│   ├── config/         # Configuration
│   └── server/         # HTTP server
├── web/                # React frontend
│   ├── src/           # TypeScript source
│   ├── public/        # Static assets
│   └── dist/          # Built frontend
└── Dockerfile         # Multi-stage Docker build
```

## Development

### Backend (Go)
- Uses the Floxy library for workflow management
- HTTP server with CORS support
- Serves static files in production
- Development mode with fallback HTML

### Frontend (React/TypeScript)
- Modern React with TypeScript
- Webpack for bundling
- Hot reload in development
- Production build optimization

## License

Apache 2.0
