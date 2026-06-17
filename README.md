# NovaField AI

[![CI](https://github.com/sarinanick/novafield/actions/workflows/ci.yml/badge.svg)](https://github.com/sarinanick/novafield/actions/workflows/ci.yml)
[![Deploy](https://github.com/sarinanick/novafield/actions/workflows/deploy.yml/badge.svg)](https://github.com/sarinanick/novafield/actions/workflows/deploy.yml)

An AI freelancer marketplace with virtual coworking space.

Built with Go backend, Next.js frontend, and Darakub PaaS.

## Architecture

```
┌─────────────────┐     ┌─────────────────┐
│   Frontend      │     │   Backend       │
│   Next.js 15    │────▶│   Go 1.25       │
│   Port 3000     │     │   Port 3001     │
└─────────────────┘     └─────────────────┘
        │                       │
        │   REST API            │
        │   WebSocket           │
        └───────────────────────┘
```

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.25+ (for local development)
- Node.js 20+ (for local development)

### Using Docker Compose
```bash
docker-compose up --build
```

### Local Development

**Backend:**
```bash
cd backend
go mod download
go run main.go
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev
```

## API

Base URL: `http://localhost:3001/api/v1`

### Key Endpoints
- `POST /auth/register` - Register
- `POST /auth/login` - Login
- `GET /gigs` - Browse gigs
- `GET /health` - Health check

## Features

- **Marketplace**: AI service listings with packages, orders, reviews
- **Messaging**: Real-time chat with WebSocket
- **Virtual Office**: 2D coworking world with Phaser 3
- **Coworking**: Pomodoro timers, focus sessions
- **Integrations**: Spotify, WebRTC voice/video

## CI/CD

Pipeline runs on GitLab CI with 5 stages:
1. **Lint** - `go vet`
2. **Test** - 184+ tests
3. **Build** - Binary compilation + health check
4. **Docker** - Image build & push
5. **Deploy** - Staging (auto) + Production (manual)

## Project Structure

```
novafield/
├── backend/          # Go API server
│   ├── handlers/     # HTTP handlers
│   ├── models/       # Data models
│   ├── database/     # File-based DB
│   ├── store/        # Data access layer
│   └── main.go       # Entry point
├── frontend/         # Next.js frontend
│   └── src/
│       ├── app/      # Pages (App Router)
│       ├── components/
│       └── lib/      # API client, contexts
└── docker-compose.yml
```
