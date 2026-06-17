# NovaField AI — Code Review & PaaS Deployment Guide

Complete analysis + deployment for Darkube / Liara / similar PaaS platforms.

---

## Part 1: Code Review

### Critical Issues

| # | File | Line | Issue | Severity |
|---|------|------|-------|----------|
| 1 | `backend/handlers/world.go` | 14 | WebSocket `CheckOrigin` accepts ALL origins — allows any website to connect | **Critical** |
| 2 | `backend/main.go` | 149 | CORS hardcoded to `localhost:3000/3001` only — **will block production frontend** | **Critical** |
| 3 | `backend/store/store.go` | 43-49 | Auth tokens stored in-memory — **all sessions lost on restart** | **High** |
| 4 | `backend/database/database.go` | 79 | DB file path hardcoded to `novafield.json` — no config option | **Medium** |

### Security Fixes Needed

**1. WebSocket Origin Check** (`backend/handlers/world.go:14`)
```go
// Current (INSECURE):
var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

// Fixed:
var allowedOrigins = map[string]bool{
    "http://localhost:3000":  true,
    "http://localhost:3001":  true,
    "https://app.yourdomain.com": true,
}
var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return allowedOrigins[r.Header.Get("Origin")]
    },
}
```

**2. CORS for Production** (`backend/main.go:149-152`)
```go
// Current (localhost only):
allowedOrigins := map[string]bool{
    "http://localhost:3000": true,
    "http://localhost:3001": true,
}

// Fixed — add env var support:
func getAllowedOrigins() map[string]bool {
    origins := map[string]bool{
        "http://localhost:3000":  true,
        "http://localhost:3001":  true,
    }
    if extra := os.Getenv("CORS_ORIGINS"); extra != "" {
        for _, o := range strings.Split(extra, ",") {
            origins[strings.TrimSpace(o)] = true
        }
    }
    return origins
}
```

**3. Token Persistence** — Tokens are lost on restart. For production, either:
- Use JWT tokens (stateless, no storage needed)
- Or store tokens in the JSON file DB

### Code Quality Notes

| Area | Status | Notes |
|------|--------|-------|
| Auth flow | Good | bcrypt hashing, rate limiting on login |
| Input validation | Adequate | Basic checks present, could be stricter |
| Error handling | Good | Consistent JSON error responses |
| WebSocket | Working | Two hubs: World + Realtime, both in-memory |
| File uploads | Good | Extension whitelist, size limits, image processing |
| API design | Consistent | RESTful patterns, proper HTTP methods |
| Frontend API client | Clean | Typed, handles auth, auto-reconnect on WS |
| Docker | Good | Multi-stage builds, non-root users, health checks |

### Architecture Assessment

```
┌─────────────────────────────────────────────────────┐
│  Frontend (Next.js 15)                              │
│  - App Router, React 19, Tailwind                   │
│  - Phaser 3 for 2D virtual office                   │
│  - WebSocket for real-time features                 │
└──────────────────────┬──────────────────────────────┘
                       │ REST API + WebSocket
                       ▼
┌─────────────────────────────────────────────────────┐
│  Backend (Go 1.25)                                  │
│  - 80+ API endpoints                                │
│  - WebSocket hubs (World, Realtime)                 │
│  - File-based JSON DB (novafield.json)              │
│  - Local file storage (uploads/)                    │
└─────────────────────────────────────────────────────┘
```

**Strengths:**
- Zero external dependencies (no DB, no Redis, no queue)
- Simple deployment — just 2 containers
- Lightweight Go binary (~10MB)
- Good test coverage (184+ tests)

**Weaknesses:**
- File-based DB doesn't scale horizontally
- In-memory WebSocket state lost on restart
- No JWT — token state is ephemeral
- Single-server architecture only

---

## Part 2: PaaS Deployment (Darkube / Liara / Similar)

### Architecture Summary

| Service | Type | Port | External Dependencies |
|---------|------|------|-----------------------|
| Backend | Go API + WebSocket | 3001 | None |
| Frontend | Next.js SSR | 3000 | Backend |

**Required:** None (file-based storage)
**Optional:** Persistent volume for `novafield.json` + `uploads/`

---

### Option A: Two Separate Apps (Recommended)

Deploy backend and frontend as independent PaaS apps.

#### Backend App

**Settings:**
| Field | Value |
|-------|-------|
| Name | `novafield-backend` |
| Build | Dockerfile |
| Dockerfile Path | `backend/Dockerfile` |
| Port | `3001` |
| Start Command | `./novafield-api` |

**Environment Variables:**
```env
PORT=3001
NODE_ENV=production
CORS_ORIGINS=https://app.yourdomain.com
SPOTIFY_CLIENT_ID=
SPOTIFY_CLIENT_SECRET=
SPOTIFY_REDIRECT_URI=https://api.yourdomain.com/api/v1/spotify/callback
```

**Persistent Storage:**
| Mount Path | Description |
|------------|-------------|
| `/app/novafield.json` | Database file |
| `/app/uploads` | Uploaded files |

#### Frontend App

**Settings:**
| Field | Value |
|-------|-------|
| Name | `novafield-frontend` |
| Build | Dockerfile |
| Dockerfile Path | `frontend/Dockerfile` |
| Port | `3000` |
| Start Command | `node server.js` |

**Environment Variables:**
```env
NEXT_PUBLIC_API_URL=https://api.yourdomain.com/api/v1
NEXT_PUBLIC_WS_URL=wss://api.yourdomain.com
```

---

### Option B: Single App with Docker Compose

If your PaaS supports docker-compose deployment.

**docker-compose.yml:**
```yaml
services:
  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    ports:
      - "3001:3001"
    environment:
      - PORT=3001
      - NODE_ENV=production
      - CORS_ORIGINS=https://app.yourdomain.com
    volumes:
      - ./data/novafield.json:/app/novafield.json
      - ./data/uploads:/app/uploads
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:3001/api/v1/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    environment:
      - NEXT_PUBLIC_API_URL=https://api.yourdomain.com/api/v1
      - NEXT_PUBLIC_WS_URL=wss://api.yourdomain.com
    depends_on:
      backend:
        condition: service_healthy
    restart: unless-stopped
```

---

### Domain Configuration

| Domain | Points To | Service |
|--------|-----------|---------|
| `api.yourdomain.com` | Backend app | Backend (port 3001) |
| `app.yourdomain.com` | Frontend app | Frontend (port 3000) |

### SSL/TLS

Enable automatic SSL (Let's Encrypt) on both apps in your PaaS dashboard.

---

### Darkube Specific Steps

1. **Create Backend Service:**
   - Go to Services → Create Service
   - Source: Git Repository → `hamgit.ir/alitabaei7/novafield`
   - Buildpack: Dockerfile
   - Dockerfile Path: `backend/Dockerfile`
   - Port: `3001`

2. **Create Frontend Service:**
   - Same repo
   - Dockerfile Path: `frontend/Dockerfile`
   - Port: `3000`

3. **Add Persistent Storage:**
   - Backend → Storage → Add Volume
   - Mount: `/app` (for novafield.json + uploads)

4. **Set Environment Variables** (per service)

5. **Add Custom Domains** with SSL

6. **Deploy**

---

### Liara Specific Steps

1. **Create App:**
   - Dashboard → Create App
   - Source: Git → `hamgit.ir/alitabaei7/novafield`
   - Port: `3001` (backend) / `3000` (frontend)

2. **Docker Settings:**
   - Enable Docker
   - Dockerfile: `backend/Dockerfile` or `frontend/Dockerfile`

3. **Persistent Storage:**
   - Add disk mount for `/app`

4. **Environment Variables:**
   - Add via dashboard or `.env` file

5. **Domain + SSL:**
   - Add custom domain
   - Enable Let's Encrypt

---

### Deployment Checklist

```
□ Fix CORS origins for production domain
□ Fix WebSocket CheckOrigin for production domain
□ Set NEXT_PUBLIC_API_URL to production backend URL
□ Set NEXT_PUBLIC_WS_URL to production WebSocket URL
□ Enable persistent storage for novafield.json
□ Enable persistent storage for uploads/
□ Configure custom domains
□ Enable SSL/TLS
□ Test health endpoint: GET /api/v1/health
□ Test WebSocket connection
□ Test file upload
□ Verify CORS headers in browser
```

---

### Environment Variables Reference

| Variable | Service | Required | Default | Description |
|----------|---------|----------|---------|-------------|
| `PORT` | Backend | No | `3001` | API server port |
| `NODE_ENV` | Backend | No | `development` | Environment mode |
| `CORS_ORIGINS` | Backend | **Yes** | `localhost` | Comma-separated allowed origins |
| `NEXT_PUBLIC_API_URL` | Frontend | **Yes** | `http://localhost:3001/api/v1` | Backend API URL |
| `NEXT_PUBLIC_WS_URL` | Frontend | **Yes** | `ws://localhost:3001` | WebSocket URL |
| `SPOTIFY_CLIENT_ID` | Backend | No | - | Spotify integration |
| `SPOTIFY_CLIENT_SECRET` | Backend | No | - | Spotify integration |

---

### Post-Deployment Verification

```bash
# 1. Health check
curl https://api.yourdomain.com/api/v1/health
# Expected: {"status":"ok","service":"novafield-api","version":"1.0.0"}

# 2. Register user
curl -X POST https://api.yourdomain.com/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"test123","name":"Test User"}'

# 3. Open frontend
# Visit https://app.yourdomain.com

# 4. Test WebSocket (browser console)
const ws = new WebSocket("wss://api.yourdomain.com/api/v1/realtime/ws?token=YOUR_TOKEN");
ws.onmessage = (e) => console.log(JSON.parse(e.data));
```

---

### Troubleshooting

| Problem | Solution |
|---------|----------|
| CORS errors in browser | Add frontend domain to `CORS_ORIGINS` env var |
| WebSocket won't connect | Ensure `NEXT_PUBLIC_WS_URL` uses `wss://` (not `ws://`) |
| 502 Bad Gateway | Check container logs, verify port mapping |
| Data lost after restart | Ensure persistent volume is mounted at `/app` |
| File upload fails | Check `/app/uploads` volume has write permissions |
| Frontend can't reach backend | Verify `NEXT_PUBLIC_API_URL` is correct and CORS is configured |
