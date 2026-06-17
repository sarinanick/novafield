# NovaField AI — Dokploy Deployment Guide

Complete guide to deploy NovaField AI on a VPS using [Dokploy](https://dokploy.com).

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Server Setup](#server-setup)
3. [Install Dokploy](#install-dokploy)
4. [Connect GitLab Repository](#connect-gitlab-repository)
5. [Configure Backend Service](#configure-backend-service)
6. [Configure Frontend Service](#configure-frontend-service)
7. [Environment Variables](#environment-variables)
8. [Domain & SSL Setup](#domain--ssl-setup)
9. [Deploy](#deploy)
10. [Post-Deployment](#post-deployment)
11. [Troubleshooting](#troubleshooting)

---

## Prerequisites

- **VPS**: Ubuntu 22.04+ (minimum 2GB RAM, 1 vCPU)
- **Domain**: 2 domains (e.g., `api.yourdomain.com` + `app.yourdomain.com`) pointed to your server IP
- **GitLab Account** with the `alitabaei7/novafield` repository

---

## Server Setup

### 1. Create a fresh Ubuntu 22.04 VPS

Providers: Hetzner, DigitalOcean, Linode, Vultr, etc.

### 2. SSH into your server

```bash
ssh root@YOUR_SERVER_IP
```

### 3. Update system

```bash
apt update && apt upgrade -y
```

### 4. Install Docker (if not already installed)

```bash
curl -fsSL https://get.docker.com | sh
systemctl enable docker
systemctl start docker
```

---

## Install Dokploy

### 1. Run the official install script

```bash
curl -sSL https://dokploy.com/install.sh | bash
```

### 2. Access the Dokploy dashboard

Open in browser:
```
http://YOUR_SERVER_IP:3000
```

Complete the initial setup wizard:
- Create admin account
- Set your domain (optional, can skip)

---

## Connect GitLab Repository

### 1. In Dokploy, go to **Settings → Git**

### 2. Add GitLab Provider

- Click **Add Git Provider → GitLab**
- Enter:
  - **Provider Name**: `hamgit`
  - **GitLab URL**: `https://hamgit.ir`
  - **Access Token**: `glpat-k3Yeip_9jA7k-aswFcsd_G86MQp1OmVpYgk.01.0z0zd9c6g`

### 3. Verify the connection

You should see `alitabaei7/novafield` in the repository list.

---

## Configure Backend Service

### 1. Create a new application

- Go to **Applications → Create Application**
- Name: `novafield-backend`
- Select **Dockerfile** as build type

### 2. Settings

| Setting | Value |
|---------|-------|
| Git Provider | `hamgit` |
| Repository | `alitabaei7/novafield` |
| Branch | `main` |
| Build Type | `Dockerfile` |
| Dockerfile Path | `backend/Dockerfile` |
| Base Directory | `backend` |
| Port | `3001` |

### 3. Environment Variables

Add these in the **Environment** tab:

```env
PORT=3001
NODE_ENV=production
```

### 4. Volumes

Add a persistent volume for the database and uploads:

| Container Path | Host Path | Type |
|----------------|-----------|------|
| `/app/novafield.json` | `/data/novafield-backend/novafield.json` | bind |
| `/app/uploads` | `/data/novafield-backend/uploads` | bind |

Create the host directories on your server:

```bash
mkdir -p /data/novafield-backend/uploads
```

### 5. Health Check

Enable health check:

| Setting | Value |
|---------|-------|
| Health Check Path | `/api/v1/health` |
| Port | `3001` |

---

## Configure Frontend Service

### 1. Create a new application

- Go to **Applications → Create Application**
- Name: `novafield-frontend`
- Select **Dockerfile** as build type

### 2. Settings

| Setting | Value |
|---------|-------|
| Git Provider | `hamgit` |
| Repository | `alitabaei7/novafield` |
| Branch | `main` |
| Build Type | `Dockerfile` |
| Dockerfile Path | `frontend/Dockerfile` |
| Base Directory | `frontend` |
| Port | `3000` |

### 3. Environment Variables

Add these in the **Environment** tab:

```env
NEXT_PUBLIC_API_URL=https://api.yourdomain.com/api/v1
NEXT_PUBLIC_WS_URL=wss://api.yourdomain.com
```

> **Important**: Replace `yourdomain.com` with your actual domain.

### 4. Health Check

Enable health check:

| Setting | Value |
|---------|-------|
| Health Check Path | `/` |
| Port | `3000` |

---

## Environment Variables

### Backend Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `PORT` | API server port | `3001` |
| `NODE_ENV` | Environment mode | `production` |
| `SPOTIFY_CLIENT_ID` | Spotify API client ID | (optional) |
| `SPOTIFY_CLIENT_SECRET` | Spotify API secret | (optional) |
| `SPOTIFY_REDIRECT_URI` | Spotify callback URL | `https://api.yourdomain.com/api/v1/spotify/callback` |

### Frontend Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `NEXT_PUBLIC_API_URL` | Backend API URL | `https://api.yourdomain.com/api/v1` |
| `NEXT_PUBLIC_WS_URL` | WebSocket URL | `wss://api.yourdomain.com` |

---

## Domain & SSL Setup

### 1. In Dokploy, go to your application → **Domains**

### 2. Add domain for Backend

| Setting | Value |
|---------|-------|
| Service Name | `novafield-backend` |
| Host | `api.yourdomain.com` |
| Port | `3001` |
| HTTPS | ✅ Enable |
| Certificate | Let's Encrypt |

### 3. Add domain for Frontend

| Setting | Value |
|---------|-------|
| Service Name | `novafield-frontend` |
| Host | `app.yourdomain.com` |
| Port | `3000` |
| HTTPS | ✅ Enable |
| Certificate | Let's Encrypt |

### 4. DNS Configuration

In your domain registrar, add these A records:

```
api.yourdomain.com      → YOUR_SERVER_IP
app.yourdomain.com      → YOUR_SERVER_IP
```

Wait 5-10 minutes for DNS propagation.

---

## Deploy

### 1. Deploy Backend

- Go to **novafield-backend → Deployments**
- Click **Deploy**
- Watch the build logs

### 2. Deploy Frontend

- Go to **novafield-frontend → Deployments**
- Click **Deploy**
- Watch the build logs

### 3. Verify deployment

```bash
# Health check
curl https://api.yourdomain.com/api/v1/health

# Expected response:
# {"status":"ok","service":"novafield-api","version":"1.0.0"}
```

---

## Post-Deployment

### 1. Seed Data

On first run, the backend automatically seeds sample data (users, gigs, categories).

Default login credentials:

| Email | Password | Role |
|-------|----------|------|
| sarah.chen@example.com | password123 | freelancer |
| marcus.r@example.com | password123 | freelancer |

### 2. Test the Application

1. Open `https://app.yourdomain.com`
2. Register a new account or use seed credentials
3. Browse the marketplace
4. Test creating a gig

### 3. Set Up Automatic Deployments

In Dokploy, go to each application → **General** → **Advanced**:

- Enable **Auto Deploy** on `main` branch
- Every push to `main` will trigger automatic deployment

### 4. Configure CORS

Update the backend CORS to allow your frontend domain. Edit `backend/main.go`:

```go
allowedOrigins := map[string]bool{
    "http://localhost:3000":  true,
    "http://localhost:3001":  true,
    "https://app.yourdomain.com": true,
}
```

Commit and push, then redeploy.

---

## Troubleshooting

### Build fails

- Check the build logs in Dokploy → Deployments
- Ensure Docker is running: `systemctl status docker`
- Verify the Dockerfile paths are correct

### Frontend can't reach backend

- Ensure `NEXT_PUBLIC_API_URL` is set correctly
- Check CORS configuration in backend
- Verify the backend is healthy

### 502 Bad Gateway

- Check if the container is running: `docker ps`
- Check container logs: `docker logs novafield-backend`
- Ensure the port mapping is correct

### Database not persisting

- Verify volume mounts are configured
- Check `/data/novafield-backend/` exists and has correct permissions

### WebSocket not connecting

- Ensure `NEXT_PUBLIC_WS_URL` uses `wss://` (not `ws://`)
- Check that your reverse proxy supports WebSocket upgrade

---

## Architecture Overview

```
Internet
    │
    ▼
┌─────────────────────────────────────────┐
│            Dokploy (Nginx)              │
│  ┌──────────────┐  ┌──────────────┐    │
│  │  app.yourdomain│  │  api.yourdomain│  │
│  │   :3000       │  │   :3001       │  │
│  └──────┬───────┘  └──────┬───────┘    │
│         │                  │            │
│         ▼                  ▼            │
│  ┌──────────────┐  ┌──────────────┐    │
│  │  Frontend    │  │  Backend     │    │
│  │  Next.js 15  │──│  Go 1.25     │    │
│  │  (standalone)│  │  (REST API)  │    │
│  └──────────────┘  └──────┬───────┘    │
│                           │            │
│                           ▼            │
│                    ┌──────────────┐    │
│                    │ novafield.json│    │
│                    │ (File DB)    │    │
│                    └──────────────┘    │
└─────────────────────────────────────────┘
```

---

## Quick Reference

| Component | Port | URL |
|-----------|------|-----|
| Frontend | 3000 | `https://app.yourdomain.com` |
| Backend API | 3001 | `https://api.yourdomain.com` |
| Health Check | 3001 | `https://api.yourdomain.com/api/v1/health` |

---

## Commands Reference

```bash
# SSH into server
ssh root@YOUR_SERVER_IP

# Check running containers
docker ps

# View backend logs
docker logs novafield-backend -f

# View frontend logs
docker logs novafield-frontend -f

# Restart backend
docker restart novafield-backend

# Restart frontend
docker restart novafield-frontend

# Check Dokploy status
dokploy status

# Restart Dokploy
dokploy restart
```
