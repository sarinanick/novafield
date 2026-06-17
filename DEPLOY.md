# NovaField AI — Deployment Guide (Dokploy + Darakub)

## Table of Contents

1. [Project Overview](#project-overview)
2. [Architecture](#architecture)
3. [Prerequisites](#prerequisites)
4. [Dokploy Installation](#dokploy-installation)
5. [Project Setup in Dokploy](#project-setup-in-dokploy)
6. [Backend Configuration](#backend-configuration)
7. [Frontend Configuration](#frontend-configuration)
8. [Environment Variables](#environment-variables)
9. [Domain & Reverse Proxy](#domain--reverse-proxy)
10. [Persistent Storage](#persistent-storage)
11. [GitLab CI/CD Integration](#gitlab-cicd-integration)
12. [SSL/TLS Setup](#ssltls-setup)
13. [Monitoring & Health Checks](#monitoring--health-checks)
14. [Troubleshooting](#troubleshooting)
15. [Production Checklist](#production-checklist)

---

## Project Overview

NovaField AI is a full-stack AI freelancer marketplace with:
- **Backend**: Go 1.25 API server (pure `net/http`, no framework)
- **Frontend**: Next.js 15 with React 19, Tailwind CSS, Phaser 3 (virtual coworking)
- **Database**: File-based JSON store (`novafield.json`)
- **Real-time**: WebSocket (gorilla/websocket) for messaging + virtual world

## Architecture

```
                    ┌─────────────────────────────┐
                    │         Dokploy Server       │
                    │                              │
  Internet ───────▶ │  ┌──────────┐ ┌──────────┐  │
                    │  │ Traefik  │ │ Dokploy  │  │
                    │  │ (reverse │ │ (panel)  │  │
                    │  │  proxy)  │ │          │  │
                    │  └────┬─────┘ └──────────┘  │
                    │       │                      │
                    │  ┌────┴────────────────────┐ │
                    │  │     Docker Network      │ │
                    │  │                         │ │
                    │  │  ┌─────────┐            │ │
                    │  │  │ Frontend│ :3000      │ │
                    │  │  │ Next.js │────────┐   │ │
                    │  │  └─────────┘        │   │ │
                    │  │                     ▼   │ │
                    │  │  ┌─────────┐  ┌────────┐│ │
                    │  │  │ Backend │  │ Volume ││ │
                    │  │  │ Go API  │  │ nova-  ││ │
                    │  │  │ :3001   │  │ field. ││ │
                    │  │  └─────────┘  │ json   ││ │
                    │  │               └────────┘│ │
                    │  └─────────────────────────┘ │
                    └─────────────────────────────┘
```

## Prerequisites

- A VPS with Docker installed (Ubuntu 22.04+ recommended, 2+ CPU, 4GB+ RAM)
- A domain name pointed to your server's IP
- GitLab repository access at `hamgit.ir/alitabaei7/novafield`

## Dokploy Installation

### One-line install (recommended)

```bash
curl -sSL https://get.dokploy.com | bash
```

### Manual install

```bash
# Install Docker
curl -fsSL https://get.docker.com | sh

# Install Dokploy
git clone https://github.com/Dokploy/dokploy.git
cd dokploy
docker compose -f docker-compose.yml up -d
```

### Access the dashboard

```
http://<YOUR_SERVER_IP>:3000
```

Complete the initial setup wizard (create admin account).

---

## Project Setup in Dokploy

### Step 1: Create a New Project

1. Open Dokploy dashboard
2. Click **Create Project**
3. Name it `novafield`
4. Click **Create**

### Step 2: Add Two Applications

You'll create two separate applications inside the `novafield` project:

#### Application 1: `novafield-backend`

| Field | Value |
|-------|-------|
| Name | `novafield-backend` |
| Type | **Docker** |
| Source | **Git** |

#### Application 2: `novafield-frontend`

| Field | Value |
|-------|-------|
| Name | `novafield-frontend` |
| Type | **Docker** |
| Source | **Git** |

---

## Backend Configuration

### Git Settings

| Field | Value |
|-------|-------|
| Repository | `https://hamgit.ir/alitabaei7/novafield.git` |
| Branch | `main` |
| Build Path | `backend` |
| Dockerfile | `Dockerfile` |

### Dockerfile (already in repo)

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o novafield-api .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates wget
WORKDIR /app
COPY --from=builder /app/novafield-api .
RUN mkdir -p uploads
EXPOSE 3001
CMD ["./novafield-api"]
```

### Port Mapping

| Internal | External (Public) |
|----------|-------------------|
| 3001 | 3001 |

### Health Check

Dokploy supports health checks. Set:

| Field | Value |
|-------|-------|
| Health Check Path | `/api/v1/health` |
| Health Check Port | `3001` |

---

## Frontend Configuration

### Git Settings

| Field | Value |
|-------|-------|
| Repository | `https://hamgit.ir/alitabaei7/novafield.git` |
| Branch | `main` |
| Build Path | `frontend` |
| Dockerfile | `Dockerfile` |

### Dockerfile (already in repo)

```dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV production
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public
EXPOSE 3000
CMD ["node", "server.js"]
```

### Port Mapping

| Internal | External (Public) |
|----------|-------------------|
| 3000 | 3000 |

### Important: Standalone Mode

The frontend uses `output: 'standalone'` in `next.config.js`. This is already configured — the Dockerfile correctly copies the standalone build output.

---

## Environment Variables

### Backend Environment Variables

Set these in Dokploy → novafield-backend → **Environment Variables**:

| Variable | Value | Description |
|----------|-------|-------------|
| `PORT` | `3001` | API server port |
| `NEXT_PUBLIC_API_URL` | `https://<YOUR_DOMAIN>/api/v1` | Public API URL (for reference) |
| `SPOTIFY_CLIENT_ID` | (your Spotify app client ID) | Spotify integration |
| `SPOTIFY_CLIENT_SECRET` | (your Spotify app client secret) | Spotify integration |
| `SPOTIFY_REDIRECT_URI` | `https://<YOUR_DOMAIN>/api/v1/spotify/callback` | Spotify OAuth callback |

### Frontend Environment Variables

Set these in Dokploy → novafield-frontend → **Environment Variables**:

| Variable | Value | Description |
|----------|-------|-------------|
| `NEXT_PUBLIC_API_URL` | `https://<YOUR_DOMAIN>/api/v1` | Backend API URL |
| `NEXT_PUBLIC_WS_URL` | `wss://<YOUR_DOMAIN>` | WebSocket URL |

### Setting Environment Variables in Dokploy

1. Go to your application
2. Click **Environment** tab
3. Add each variable with Name and Value
4. Click **Save**
5. Redeploy the application

---

## Domain & Reverse Proxy

### Option A: Dokploy Built-in Traefik (Recommended)

Dokploy includes Traefik as a reverse proxy. Configure domains in the application settings:

#### Backend Domain

1. Go to novafield-backend → **Domains**
2. Click **Add Domain**

| Field | Value |
|-------|-------|
| Host | `api.yourdomain.com` |
| Port | `3001` |
| HTTPS | Enable |

#### Frontend Domain

1. Go to novafield-frontend → **Domains**
2. Click **Add Domain**

| Field | Value |
|-------|-------|
| Host | `yourdomain.com` |
| Port | `3000` |
| HTTPS | Enable |

### DNS Configuration

Point your domain to your server:

```
Type    Name           Value
A       yourdomain.com    <SERVER_IP>
A       api.yourdomain.com <SERVER_IP>
```

### Option B: External Nginx/Caddy

If you prefer an external reverse proxy:

```nginx
# /etc/nginx/sites-available/novafield
server {
    listen 80;
    server_name yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:3001;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # WebSocket support for messaging + virtual world
    location /api/v1/world/ws {
        proxy_pass http://127.0.0.1:3001;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_read_timeout 86400;
    }

    location /api/v1/realtime/ws {
        proxy_pass http://127.0.0.1:3001;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_read_timeout 86400;
    }
}
```

---

## Persistent Storage

### Critical: Data Persistence

The backend stores all data in `novafield.json` (file-based JSON store). Without volume mounting, **all data is lost on container restart**.

### Volume Configuration in Dokploy

In Dokploy, use the **Docker Compose Override** or **Volumes** section:

#### Backend Volumes

| Host Path | Container Path | Mode |
|-----------|---------------|------|
| `/opt/novafield/data/novafield.json` | `/app/novafield.json` | Read/Write |
| `/opt/novafield/data/uploads` | `/app/uploads` | Read/Write |

#### Create Data Directory on Server

```bash
mkdir -p /opt/novafield/data
touch /opt/novafield/data/novafield.json
chmod 644 /opt/novafield/data/novafield.json
```

### Docker Compose Override (Dokploy)

If Dokploy supports compose overrides, create a `docker-compose.override.yml`:

```yaml
version: "3.9"

services:
  backend:
    volumes:
      - /opt/novafield/data/novafield.json:/app/novafield.json
      - /opt/novafield/data/uploads:/app/uploads
    deploy:
      resources:
        limits:
          cpus: "1.0"
          memory: 512M
        reservations:
          cpus: "0.25"
          memory: 128M
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:3001/api/v1/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
    logging:
      driver: json-file
      options:
        max-size: "10m"
        max-file: "5"
```

---

## GitLab CI/CD Integration

### Connecting Dokploy to GitLab

1. In Dokploy → Settings → **Git Providers**
2. Add GitLab provider:
   - URL: `https://hamgit.ir`
   - Access Token: `glpat-k3Yeip_9jA7k-aswFcsd_G86MQp1OmVpYgk.01.0z0zd9c6g`

### Auto-Deploy on Push

For each application:
1. Go to application → **General** → **Deploy**
2. Enable **Auto Deploy**
3. Set branch: `main`
4. Set build path: `backend` or `frontend` accordingly

### CI/CD Pipeline (Already Configured)

The project has `.gitlab-ci.yml` at `backend/.gitlab-ci.yml` with 5 stages:

1. **Lint** — `go vet ./...`
2. **Test** — `go test ./handlers/ -v -count=1`
3. **Build** — Binary compilation + health check
4. **Docker** — Image build & push to `registry.hamgit.ir/alitabaei7/novafield`
5. **Deploy** — Staging (auto) + Production (manual)

### Registry Configuration

Dokploy can pull images from GitLab Container Registry:
- Registry: `registry.hamgit.ir`
- Image: `registry.hamgit.ir/alitabaei7/novafield:latest`

---

## SSL/TLS Setup

### With Dokploy Traefik (Automatic)

1. Ensure your DNS A records are pointing to the server
2. In the domain settings, enable **HTTPS**
3. Traefik will automatically provision Let's Encrypt certificates

### With Certbot (External Proxy)

```bash
# Install certbot
apt install certbot python3-certbot-nginx

# Get certificate
certbot --nginx -d yourdomain.com -d api.yourdomain.com

# Auto-renew
certbot renew --dry-run
```

---

## Monitoring & Health Checks

### Backend Health Endpoint

```bash
curl https://api.yourdomain.com/api/v1/health
```

Expected response: `{"status":"ok"}`

### Docker Health Check

The production compose already includes health checks:
- **Interval**: 30 seconds
- **Timeout**: 10 seconds
- **Retries**: 3
- **Start Period**: 10 seconds

### Logs

In Dokploy, view logs per application:
- Backend logs
- Frontend logs
- Build logs

Or via Docker:

```bash
docker logs -f novafield-backend
docker logs -f novafield-frontend
```

---

## Troubleshooting

### Build Fails: Go Module Download Timeout

The project uses `GOPROXY=direct` (set in `.gitlab-ci.yml`) because `proxy.golang.org` may be unreachable. If build fails:

- Ensure `GOPROXY=direct` is set in build environment
- Or pre-cache modules on the server

### Frontend Build Timeout

`npm run build` can take >120s. Dokploy builds should have sufficient timeout (default is usually 600s).

### Data Loss on Restart

If `novafield.json` is not volume-mounted, data resets. Verify:
```bash
ls -la /opt/novafield/data/novafield.json
```

### CORS Errors

Backend CORS is restricted to specific origins. If deploying to a real domain:
1. Update CORS allowed origins in `backend/main.go` (line 149-152)
2. Add your production domain: `"https://yourdomain.com": true`

### WebSocket Connection Issues

Ensure your reverse proxy:
- Supports WebSocket upgrade (`Upgrade: websocket` header)
- Has `proxy_read_timeout 86400;` (86400 seconds = 24 hours)

### Port Conflicts

Ensure no other services use ports 3000 or 3001:
```bash
ss -tlnp | grep -E '300[01]'
```

---

## Production Checklist

### Before First Deploy

- [ ] Server has Docker installed (24.0+)
- [ ] Dokploy installed and accessible
- [ ] Domain DNS records pointing to server IP
- [ ] Data directory created: `/opt/novafield/data/`
- [ ] `novafield.json` initialized (empty `{"users":[],...}` or touch file)
- [ ] Environment variables configured in Dokploy
- [ ] CORS origins updated for production domain
- [ ] GitLab access token has `read_repository` scope

### After Deploy

- [ ] Backend health check passes: `curl https://api.yourdomain.com/api/v1/health`
- [ ] Frontend loads: `https://yourdomain.com`
- [ ] API calls work from frontend (check browser console)
- [ ] WebSocket connects (check Network tab for `/world/ws` and `/realtime/ws`)
- [ ] File uploads work (test image upload)
- [ ] SSL certificate is valid (check padlock icon)

### Security

- [ ] GitLab access token is in environment variables (not hardcoded)
- [ ] Backend does not expose debug info in production
- [ ] `novafield.json` permissions: `644` (readable, not world-writable)
- [ ] Docker containers run as non-root (Alpine default)
- [ ] Firewall allows only ports 80, 443, and 22

### Performance

- [ ] Backend resource limits set (1 CPU / 512M RAM)
- [ ] Frontend resource limits set (0.5 CPU / 256M RAM)
- [ ] Log rotation configured (max-size: 10m, max-file: 5)
- [ ] Health checks enabled (interval: 30s)

---

## Darakub Deployment (GitHub Actions CI/CD)

[Darakub](https://darakub.com) is a PaaS platform that deploys containerized apps using Docker and Kubernetes. NovaField uses GitHub Actions to automatically build and deploy to Darakub on every push to `main`.

### API Key

| Field | Value |
|-------|-------|
| API Key | `e1d53948-813b-4712-b68d-e7e032e4ed22` |
| API Base | `https://api.darakub.com` |

### Authentication Header

All Darakub API requests must include:

```
Authorization: Api-key e1d53948-813b-4712-b68d-e7e032e4ed22
```

### Example cURL

```bash
# Check service status
curl -s -X GET "https://api.darakub.com/api/v1/services/novafield-backend/status" \
  -H "Authorization: Api-key e1d53948-813b-4712-b68d-e7e032e4ed22"

# Trigger deployment
curl -s -X POST "https://api.darakub.com/api/v1/services/deploy" \
  -H "Authorization: Api-key e1d53948-813b-4712-b68d-e7e032e4ed22" \
  -H "Content-Type: application/json" \
  -d '{
    "serviceName": "novafield-backend",
    "image": "registry.darakub.com/novafield/backend:latest"
  }'
```

### GitHub Actions Secrets

Configure these in GitHub → Settings → Secrets and variables → Actions:

| Secret | Value | Description |
|--------|-------|-------------|
| `DARAKUB_API_KEY` | `e1d53948-813b-4712-b68d-e7e032e4ed22` | Darakub API key |
| `DARAKUB_REGISTRY_PASS` | (your registry password) | Docker registry auth |

### GitHub Actions Variables

Configure these in GitHub → Settings → Secrets and variables → Actions → Variables:

| Variable | Example Value | Description |
|----------|---------------|-------------|
| `DARAKUB_BACKEND_URL` | `https://api.novafield.yourdomain.com` | Backend public URL |
| `DARAKUB_FRONTEND_URL` | `https://novafield.yourdomain.com` | Frontend public URL |
| `NEXT_PUBLIC_API_URL` | `https://api.novafield.yourdomain.com/api/v1` | API URL for frontend |
| `NEXT_PUBLIC_WS_URL` | `wss://api.novafield.yourdomain.com` | WebSocket URL for frontend |
| `FRONTEND_URL` | `https://novafield.yourdomain.com` | Frontend URL for backend CORS |
| `CORS_ORIGINS` | `https://novafield.yourdomain.com,https://www.novafield.yourdomain.com` | Allowed CORS origins |

### CI/CD Pipeline Stages

The pipeline (`.github/workflows/deploy.yml`) runs:

1. **Backend Lint** — `go vet ./...`
2. **Backend Test** — `go test ./handlers/ -v -count=1 -timeout 120s`
3. **Frontend Check** — `npx tsc --noEmit`
4. **Build Docker Images** — Parallel build of backend + frontend images
5. **Deploy Backend** — Push image to registry + deploy via Darakub API
6. **Deploy Frontend** — Push image to registry + deploy via Darakub API (waits for backend)
7. **Notify** — Deployment summary in GitHub Actions

### Deploy Trigger

- **Push to `main`** → Full deploy (lint → test → build → deploy)
- **Pull request** → Lint + test only (no deploy)

### Manual Deploy

```bash
# Deploy backend manually
curl -s -X POST "https://api.darakub.com/api/v1/services/deploy" \
  -H "Authorization: Api-key e1d53948-813b-4712-b68d-e7e032e4ed22" \
  -H "Content-Type: application/json" \
  -d '{"serviceName": "novafield-backend", "image": "registry.darakub.com/novafield/backend:latest"}'

# Deploy frontend manually
curl -s -X POST "https://api.darakub.com/api/v1/services/deploy" \
  -H "Authorization: Api-key e1d53948-813b-4712-b68d-e7e032e4ed22" \
  -H "Content-Type: application/json" \
  -d '{"serviceName": "novafield-frontend", "image": "registry.darakub.com/novafield/frontend:latest"}'
```

### Check Deployment Status

```bash
curl -s -X GET "https://api.darakub.com/api/v1/services/novafield-backend/status" \
  -H "Authorization: Api-key e1d53948-813b-4712-b68d-e7e032e4ed22"

curl -s -X GET "https://api.darakub.com/api/v1/services/novafield-frontend/status" \
  -H "Authorization: Api-key e1d53948-813b-4712-b68d-e7e032e4ed22"
```

---

## Quick Reference

### Service URLs

| Service | Internal | External |
|---------|----------|----------|
| Frontend | `localhost:3000` | `https://yourdomain.com` |
| Backend API | `localhost:3001` | `https://api.yourdomain.com` |
| WebSocket (world) | `ws://localhost:3001/api/v1/world/ws` | `wss://api.yourdomain.com/api/v1/world/ws` |
| WebSocket (realtime) | `ws://localhost:3001/api/v1/realtime/ws` | `wss://api.yourdomain.com/api/v1/realtime/ws` |

### Key Files

| File | Purpose |
|------|---------|
| `backend/Dockerfile` | Backend container build |
| `frontend/Dockerfile` | Frontend container build |
| `docker-compose.yml` | Local development compose |
| `backend/deploy/docker-compose.production.yml` | Production compose template |
| `backend/deploy/docker-compose.staging.yml` | Staging compose template |
| `backend/.gitlab-ci.yml` | CI/CD pipeline |
| `backend/.env.example` | Backend env vars template |
| `frontend/.env.example` | Frontend env vars template |

### Useful Commands

```bash
# Check backend health
curl https://api.yourdomain.com/api/v1/health

# View container logs
docker logs -f novafield-backend
docker logs -f novafield-frontend

# Restart services (if not using Dokploy)
docker restart novafield-backend novafield-frontend

# Check data file
cat /opt/novafield/data/novafield.json | python3 -m json.tool | head -20

# Force rebuild
docker build --no-cache -t novafield-backend ./backend
```
