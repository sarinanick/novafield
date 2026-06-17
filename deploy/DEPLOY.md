# Darkube Deployment Guide

## Infrastructure Required

| Resource | Required | Why |
|----------|----------|-----|
| Darkube Account | ✅ | PaaS hosting |
| 2 Services | ✅ | Backend + Frontend |
| Persistent Storage | ✅ | DB + uploads survive restarts |
| Custom Domain | Optional | Better URLs |
| SSL/TLS | Optional | HTTPS (Darkube provides free) |

**NOT required (yet):**
- Redis — app uses in-memory cache
- PostgreSQL — app uses file-based JSON DB
- Message queue — no async jobs

## Quick Deploy

```bash
# Run the deploy script
bash deploy/darkube-deploy.sh YOUR_BACKEND_DOMAIN YOUR_FRONTEND_DOMAIN

# Example:
bash deploy/darkube-deploy.sh novafield-api.darkube.app novafield.darkube.app
```

## Step-by-Step

### 1. Create Backend Service

1. Go to Darkube Dashboard → Create Service
2. **Name:** `novafield-backend`
3. **Source:** Git Repository
4. **Repo:** `https://github.com/sarinanick/novafield-backend`
5. **Build:** Dockerfile
6. **Dockerfile Path:** `Dockerfile`
7. **Port:** `3001`

### 2. Backend Environment Variables

| Variable | Value |
|----------|-------|
| `PORT` | `3001` |
| `CORS_ORIGINS` | `https://YOUR_BACKEND_DOMAIN` |

### 3. Backend Persistent Storage

Add two volumes:

| Mount Path | Description |
|------------|-------------|
| `/app/novafield.json` | Database file |
| `/app/uploads` | Uploaded files |

### 4. Create Frontend Service

1. **Name:** `novafield-frontend`
2. **Source:** Git Repository
3. **Repo:** `https://github.com/sarinanick/novafield`
4. **Build:** Dockerfile
5. **Dockerfile Path:** `frontend/Dockerfile`
6. **Port:** `3000`

### 5. Frontend Environment Variables

| Variable | Value |
|----------|-------|
| `NEXT_PUBLIC_API_URL` | `https://YOUR_BACKEND_DOMAIN/api/v1` |
| `NEXT_PUBLIC_WS_URL` | `wss://YOUR_BACKEND_DOMAIN` |

### 6. Deploy

Click Deploy on both services.

## Verify

```bash
# Health check
curl https://YOUR_BACKEND_DOMAIN/api/v1/health

# Expected:
# {"status":"ok","service":"novafield-api","version":"1.0.0"}
```

## Troubleshooting

| Problem | Solution |
|---------|----------|
| CORS errors | Add frontend domain to `CORS_ORIGINS` |
| WebSocket fails | Use `wss://` not `ws://` |
| 502 Bad Gateway | Check container logs |
| Data lost | Ensure volumes are mounted |
| Upload fails | Check `/app/uploads` permissions |
