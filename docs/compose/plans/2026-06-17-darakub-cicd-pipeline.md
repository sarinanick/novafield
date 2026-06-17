# NovaField CI/CD Pipeline — Darakub Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use compose:subagent (recommended) or compose:execute to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a production-grade CI/CD pipeline that lints, tests, builds Docker images, and deploys both backend and frontend to Darakub PaaS on every push to `main`.

**Architecture:** GitHub Actions orchestrates 7 stages: lint → test → typecheck → build images → deploy backend → deploy frontend → notify. Docker images are built with BuildKit caching, pushed to Darakub registry, and deployed via Darakub REST API. Environment-specific configs live in GitHub Actions secrets/variables.

**Tech Stack:** GitHub Actions, Docker BuildKit, Go 1.25, Node.js 20, Darakub REST API (`https://api.darakub.com`), Darakub Container Registry (`registry.darakub.com`)

---

## Current State

| Item | Status | Issue |
|------|--------|-------|
| `.github/workflows/deploy.yml` | Exists | Go version mismatch (1.23 vs 1.25), no env config, no rollback |
| `backend/Dockerfile` | Good | Multi-stage, non-root user |
| `frontend/Dockerfile` | Good | Multi-stage, standalone mode |
| `backend/.gitlab-ci.yml` | Exists | GitLab CI (not needed for GitHub Actions) |
| `DEPLOY.md` | Updated | Has Darakub docs |

---

## File Structure

| File | Responsibility |
|------|---------------|
| `.github/workflows/ci.yml` | PR checks only (lint + test + typecheck) — fast feedback |
| `.github/workflows/deploy.yml` | Full deploy pipeline on push to main |
| `.github/actions/darakub-deploy/action.yml` | Reusable composite action for Darakub deploy |
| `.github/actions/darakub-health/action.yml` | Reusable composite action for health check |
| `docker-compose.darakub.yml` | Darakub service definitions (env, volumes, resources) |

---

## Task 1: Fix Go Version Mismatch & Separate CI from Deploy

**Covers:** Pipeline reliability, fast PR feedback

**Files:**
- Create: `.github/workflows/ci.yml`
- Modify: `.github/workflows/deploy.yml`

- [ ] **Step 1: Create CI-only workflow for PRs**

Create `.github/workflows/ci.yml`:

```yaml
name: CI — PR Checks

on:
  pull_request:
    branches: [main]

env:
  GO_VERSION: "1.25"
  NODE_VERSION: "20"

jobs:
  backend-lint:
    name: Backend Lint
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: backend
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache-dependency-path: backend/go.sum
      - run: go vet ./...

  backend-test:
    name: Backend Test
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: backend
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache-dependency-path: backend/go.sum
      - run: go test ./handlers/ -v -count=1 -timeout 120s

  frontend-check:
    name: Frontend Typecheck
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: frontend
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: npm
          cache-dependency-path: frontend/package-lock.json
      - run: npm ci
      - run: npx tsc --noEmit
```

- [ ] **Step 2: Fix deploy.yml Go version and remove redundant checks**

Modify `.github/workflows/deploy.yml` — change line 10 from `GO_VERSION: "1.23"` to `GO_VERSION: "1.25"`.

- [ ] **Step 3: Verify CI workflow syntax**

Run: `cat .github/workflows/ci.yml | head -5`
Expected: YAML header with name and on trigger.

- [ ] **Step 4: Commit**

```bash
git add .github/workflows/ci.yml .github/workflows/deploy.yml
git commit -m "ci: fix Go version to 1.25, separate CI from deploy workflow"
```

---

## Task 2: Create Reusable Darakub Deploy Action

**Covers:** DRY pipeline, consistent deploy logic

**Files:**
- Create: `.github/actions/darakub-deploy/action.yml`

- [ ] **Step 1: Create composite action for Darakub deployment**

Create `.github/actions/darakub-deploy/action.yml`:

```yaml
name: "Darakub Deploy"
description: "Deploy a service to Darakub via API"

inputs:
  service-name:
    description: "Darakub service name"
    required: true
  image:
    description: "Docker image to deploy"
    required: true
  api-key:
    description: "Darakub API key"
    required: true
  api-url:
    description: "Darakub API base URL"
    default: "https://api.darakub.com"
  env-vars:
    description: "JSON object of environment variables"
    default: "{}"
  cpu:
    description: "CPU limit"
    default: "1000m"
  memory:
    description: "Memory limit"
    default: "512Mi"
  health-path:
    description: "Health check path"
    default: "/api/v1/health"
  health-port:
    description: "Health check port"
    default: "3001"

runs:
  using: "composite"
  steps:
    - name: Deploy to Darakub
      shell: bash
      run: |
        echo "Deploying ${{ inputs.service-name }} with image ${{ inputs.image }}..."
        RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${{ inputs.api-url }}/api/v1/services/deploy" \
          -H "Authorization: Api-key ${{ inputs.api-key }}" \
          -H "Content-Type: application/json" \
          -d '{
            "serviceName": "${{ inputs.service-name }}",
            "image": "${{ inputs.image }}",
            "env": ${{ inputs.env-vars }},
            "resources": {
              "cpu": "${{ inputs.cpu }}",
              "memory": "${{ inputs.memory }}"
            },
            "healthCheck": {
              "path": "${{ inputs.health-path }}",
              "port": ${{ inputs.health-port }},
              "interval": 30
            }
          }')
        HTTP_CODE=$(echo "$RESPONSE" | tail -1)
        BODY=$(echo "$RESPONSE" | sed '$d')
        echo "HTTP Status: $HTTP_CODE"
        echo "Response: $BODY"
        if [ "$HTTP_CODE" -lt 200 ] || [ "$HTTP_CODE" -ge 300 ]; then
          echo "::error::Deploy failed with HTTP $HTTP_CODE"
          exit 1
        fi
        echo "Deploy initiated successfully"

    - name: Wait for deployment ready
      shell: bash
      run: |
        echo "Waiting for ${{ inputs.service-name }} to be ready..."
        for i in $(seq 1 30); do
          STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
            -H "Authorization: Api-key ${{ inputs.api-key }}" \
            "${{ inputs.api-url }}/api/v1/services/${{ inputs.service-name }}/status" || true)
          if [ "$STATUS" = "200" ]; then
            echo "Service is ready!"
            exit 0
          fi
          echo "Attempt $i/30 — status: $STATUS"
          sleep 10
        done
        echo "::error::Deployment timed out after 5 minutes"
        exit 1
```

- [ ] **Step 2: Verify action syntax**

Run: `cat .github/actions/darakub-deploy/action.yml | head -3`
Expected: YAML with name and description.

- [ ] **Step 3: Commit**

```bash
git add .github/actions/darakub-deploy/action.yml
git commit -m "ci: add reusable Darakub deploy composite action"
```

---

## Task 3: Create Reusable Health Check Action

**Covers:** DRY pipeline, consistent verification

**Files:**
- Create: `.github/actions/darakub-health/action.yml`

- [ ] **Step 1: Create composite action for health checking**

Create `.github/actions/darakub-health/action.yml`:

```yaml
name: "Darakub Health Check"
description: "Verify a deployed service is healthy"

inputs:
  url:
    description: "Full URL to check"
    required: true
  retries:
    description: "Number of retries"
    default: "5"
  delay:
    description: "Seconds between retries"
    default: "5"

runs:
  using: "composite"
  steps:
    - name: Health check
      shell: bash
      run: |
        echo "Checking health at ${{ inputs.url }}..."
        for i in $(seq 1 ${{ inputs.retries }}); do
          if curl -sf "${{ inputs.url }}" > /dev/null 2>&1; then
            echo "Health check passed on attempt $i"
            exit 0
          fi
          echo "Attempt $i/${{ inputs.retries }} failed, retrying in ${{ inputs.delay }}s..."
          sleep ${{ inputs.delay }}
        done
        echo "::error::Health check failed after ${{ inputs.retries }} attempts"
        exit 1
```

- [ ] **Step 2: Commit**

```bash
git add .github/actions/darakub-health/action.yml
git commit -m "ci: add reusable health check composite action"
```

---

## Task 4: Rewrite Deploy Workflow with Reusable Actions

**Counters:** Full pipeline reliability, error handling, rollback awareness

**Files:**
- Modify: `.github/workflows/deploy.yml`

- [ ] **Step 1: Rewrite deploy.yml with reusable actions**

Replace entire `.github/workflows/deploy.yml` with:

```yaml
name: Deploy to Darakub

on:
  push:
    branches: [main]

env:
  GO_VERSION: "1.25"
  NODE_VERSION: "20"
  DARAKUB_API: "https://api.darakub.com"

jobs:
  # ── Validate ────────────────────────────────────────
  backend-lint:
    name: Lint Backend
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: backend
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache-dependency-path: backend/go.sum
      - run: go vet ./...

  backend-test:
    name: Test Backend
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: backend
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache-dependency-path: backend/go.sum
      - run: go test ./handlers/ -v -count=1 -timeout 120s

  frontend-check:
    name: Typecheck Frontend
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: frontend
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: npm
          cache-dependency-path: frontend/package-lock.json
      - run: npm ci
      - run: npx tsc --noEmit

  # ── Build ───────────────────────────────────────────
  build-backend:
    name: Build Backend Image
    runs-on: ubuntu-latest
    needs: [backend-lint, backend-test]
    steps:
      - uses: actions/checkout@v4
      - uses: docker/setup-buildx-action@v3
      - uses: docker/build-push-action@v5
        with:
          context: ./backend
          push: false
          load: true
          tags: novafield-backend:${{ github.sha }}
          cache-from: type=gha,scope=backend
          cache-to: type=gha,mode=max,scope=backend
      - run: docker save novafield-backend:${{ github.sha }} | gzip > /tmp/backend.tar.gz
      - uses: actions/upload-artifact@v4
        with:
          name: image-backend
          path: /tmp/backend.tar.gz
          retention-days: 1

  build-frontend:
    name: Build Frontend Image
    runs-on: ubuntu-latest
    needs: [frontend-check]
    steps:
      - uses: actions/checkout@v4
      - uses: docker/setup-buildx-action@v3
      - uses: docker/build-push-action@v5
        with:
          context: ./frontend
          push: false
          load: true
          tags: novafield-frontend:${{ github.sha }}
          cache-from: type=gha,scope=frontend
          cache-to: type=gha,mode=max,scope=frontend
      - run: docker save novafield-frontend:${{ github.sha }} | gzip > /tmp/frontend.tar.gz
      - uses: actions/upload-artifact@v4
        with:
          name: image-frontend
          path: /tmp/frontend.tar.gz
          retention-days: 1

  # ── Deploy Backend ──────────────────────────────────
  deploy-backend:
    name: Deploy Backend
    runs-on: ubuntu-latest
    needs: [build-backend]
    environment:
      name: production
      url: ${{ vars.DARAKUB_BACKEND_URL }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/download-artifact@v4
        with:
          name: image-backend
          path: /tmp

      - name: Push to Darakub Registry
        run: |
          docker load < /tmp/backend.tar.gz
          docker tag novafield-backend:${{ github.sha }} registry.darakub.com/novafield/backend:${{ github.sha }}
          docker tag novafield-backend:${{ github.sha }} registry.darakub.com/novafield/backend:latest
          echo "${{ secrets.DARAKUB_REGISTRY_PASS }}" | docker login registry.darakub.com -u darakub --password-stdin
          docker push registry.darakub.com/novafield/backend:${{ github.sha }}
          docker push registry.darakub.com/novafield/backend:latest

      - name: Deploy via Darakub API
        uses: ./.github/actions/darakub-deploy
        with:
          service-name: novafield-backend
          image: registry.darakub.com/novafield/backend:${{ github.sha }}
          api-key: ${{ secrets.DARAKUB_API_KEY }}
          env-vars: |
            {
              "PORT": "3001",
              "FRONTEND_URL": "${{ vars.FRONTEND_URL }}",
              "CORS_ORIGINS": "${{ vars.CORS_ORIGINS }}",
              "SPOTIFY_CLIENT_ID": "${{ secrets.SPOTIFY_CLIENT_ID }}",
              "SPOTIFY_CLIENT_SECRET": "${{ secrets.SPOTIFY_CLIENT_SECRET }}",
              "SPOTIFY_REDIRECT_URI": "${{ vars.SPOTIFY_REDIRECT_URI }}"
            }
          cpu: "1000m"
          memory: "512Mi"
          health-path: "/api/v1/health"
          health-port: "3001"

      - name: Verify backend health
        uses: ./.github/actions/darakub-health
        with:
          url: ${{ vars.DARAKUB_BACKEND_URL }}/api/v1/health
          retries: 5
          delay: 5

  # ── Deploy Frontend ─────────────────────────────────
  deploy-frontend:
    name: Deploy Frontend
    runs-on: ubuntu-latest
    needs: [build-frontend, deploy-backend]
    environment:
      name: production
      url: ${{ vars.DARAKUB_FRONTEND_URL }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/download-artifact@v4
        with:
          name: image-frontend
          path: /tmp

      - name: Push to Darakub Registry
        run: |
          docker load < /tmp/frontend.tar.gz
          docker tag novafield-frontend:${{ github.sha }} registry.darakub.com/novafield/frontend:${{ github.sha }}
          docker tag novafield-frontend:${{ github.sha }} registry.darakub.com/novafield/frontend:latest
          echo "${{ secrets.DARAKUB_REGISTRY_PASS }}" | docker login registry.darakub.com -u darakub --password-stdin
          docker push registry.darakub.com/novafield/frontend:${{ github.sha }}
          docker push registry.darakub.com/novafield/frontend:latest

      - name: Deploy via Darakub API
        uses: ./.github/actions/darakub-deploy
        with:
          service-name: novafield-frontend
          image: registry.darakub.com/novafield/frontend:${{ github.sha }}
          api-key: ${{ secrets.DARAKUB_API_KEY }}
          env-vars: |
            {
              "NEXT_PUBLIC_API_URL": "${{ vars.NEXT_PUBLIC_API_URL }}",
              "NEXT_PUBLIC_WS_URL": "${{ vars.NEXT_PUBLIC_WS_URL }}"
            }
          cpu: "500m"
          memory: "512Mi"
          health-path: "/"
          health-port: "3000"

      - name: Verify frontend health
        uses: ./.github/actions/darakub-health
        with:
          url: ${{ vars.DARAKUB_FRONTEND_URL }}
          retries: 5
          delay: 5

  # ── Summary ─────────────────────────────────────────
  summary:
    name: Deploy Summary
    runs-on: ubuntu-latest
    needs: [deploy-backend, deploy-frontend]
    if: always()
    steps:
      - name: Write summary
        run: |
          echo "## 🚀 Deployment Summary" >> $GITHUB_STEP_SUMMARY
          echo "" >> $GITHUB_STEP_SUMMARY
          echo "| Service | Status | Image |" >> $GITHUB_STEP_SUMMARY
          echo "|---------|--------|-------|" >> $GITHUB_STEP_SUMMARY
          echo "| Backend | ${{ needs.deploy-backend.result }} | \`novafield/backend:${{ github.sha }}\` |" >> $GITHUB_STEP_SUMMARY
          echo "| Frontend | ${{ needs.deploy-frontend.result }} | \`novafield/frontend:${{ github.sha }}\` |" >> $GITHUB_STEP_SUMMARY
          echo "" >> $GITHUB_STEP_SUMMARY
          echo "**Commit:** \`${{ github.sha }}\`" >> $GITHUB_STEP_SUMMARY
          echo "**Branch:** \`${{ github.ref_name }}\`" >> $GITHUB_STEP_SUMMARY
          echo "**Timestamp:** $(date -u +'%Y-%m-%d %H:%M:%S UTC')" >> $GITHUB_STEP_SUMMARY

      - name: Fail if deploy failed
        if: needs.deploy-backend.result != 'success' || needs.deploy-frontend.result != 'success'
        run: |
          echo "::error::One or more deployments failed"
          exit 1
```

- [ ] **Step 2: Verify workflow syntax**

Run: `cat .github/workflows/deploy.yml | grep "uses:" | head -20`
Expected: List of action references including reusable actions.

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/deploy.yml
git commit -m "ci: rewrite deploy pipeline with reusable actions and error handling"
```

---

## Task 5: Create Darakub Docker Compose Override

**Covers:** Service configuration, resource limits, env management

**Files:**
- Create: `docker-compose.darakub.yml`

- [ ] **Step 1: Create Darakub-specific compose file**

Create `docker-compose.darakub.yml`:

```yaml
# Darakub service definitions
# Reference: https://darakub.com/docs/services
services:
  novafield-backend:
    image: registry.darakub.com/novafield/backend:latest
    ports:
      - "3001:3001"
    environment:
      - PORT=3001
      - FRONTEND_URL=${FRONTEND_URL}
      - CORS_ORIGINS=${CORS_ORIGINS}
      - SPOTIFY_CLIENT_ID=${SPOTIFY_CLIENT_ID}
      - SPOTIFY_CLIENT_SECRET=${SPOTIFY_CLIENT_SECRET}
      - SPOTIFY_REDIRECT_URI=${SPOTIFY_REDIRECT_URI}
    volumes:
      - novafield-data:/app/novafield.json
      - novafield-uploads:/app/uploads
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
    restart: unless-stopped

  novafield-frontend:
    image: registry.darakub.com/novafield/frontend:latest
    ports:
      - "3000:3000"
    environment:
      - NEXT_PUBLIC_API_URL=${NEXT_PUBLIC_API_URL}
      - NEXT_PUBLIC_WS_URL=${NEXT_PUBLIC_WS_URL}
    depends_on:
      novafield-backend:
        condition: service_healthy
    deploy:
      resources:
        limits:
          cpus: "0.5"
          memory: 512M
        reservations:
          cpus: "0.25"
          memory: 128M
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:3000"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 15s
    restart: unless-stopped

volumes:
  novafield-data:
  novafield-uploads:
```

- [ ] **Step 2: Commit**

```bash
git add docker-compose.darakub.yml
git commit -m "ci: add Darakub service compose definitions"
```

---

## Task 6: Update DEPLOY.md with Full Pipeline Docs

**Covers:** Documentation completeness

**Files:**
- Modify: `DEPLOY.md`

- [ ] **Step 1: Add pipeline architecture diagram to DEPLOY.md**

After the "Darakub Deployment" section header, add:

```markdown
### Pipeline Architecture

```
Push to main
    │
    ├──→ [backend-lint]     go vet
    ├──→ [backend-test]     go test ./handlers/
    ├──→ [frontend-check]   tsc --noEmit
    │         │
    │    (all pass)
    │         │
    ├──→ [build-backend]    Docker build + cache
    ├──→ [build-frontend]   Docker build + cache
    │         │
    │    (images ready)
    │         │
    ├──→ [deploy-backend]   Push → Darakub API → Health check
    │         │
    │    (backend healthy)
    │         │
    ├──→ [deploy-frontend]  Push → Darakub API → Health check
    │         │
    │    (all done)
    │         │
    └──→ [summary]          Deployment report
```
```

- [ ] **Step 2: Add required secrets table**

Add after the existing secrets table:

```markdown
### All Required GitHub Secrets

| Secret | Description | Where to find |
|--------|-------------|---------------|
| `DARAKUB_API_KEY` | Darakub API key | Darakub dashboard → Settings → API |
| `DARAKUB_REGISTRY_PASS` | Container registry password | Darakub dashboard → Settings → Registry |
| `SPOTIFY_CLIENT_ID` | Spotify OAuth client ID | Spotify Developer Dashboard |
| `SPOTIFY_CLIENT_SECRET` | Spotify OAuth client secret | Spotify Developer Dashboard |

### All Required GitHub Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DARAKUB_BACKEND_URL` | Backend public URL | `https://api.novafield.com` |
| `DARAKUB_FRONTEND_URL` | Frontend public URL | `https://novafield.com` |
| `NEXT_PUBLIC_API_URL` | API URL for frontend | `https://api.novafield.com/api/v1` |
| `NEXT_PUBLIC_WS_URL` | WebSocket URL for frontend | `wss://api.novafield.com` |
| `FRONTEND_URL` | Frontend URL for backend CORS | `https://novafield.com` |
| `CORS_ORIGINS` | Allowed CORS origins | `https://novafield.com,https://www.novafield.com` |
| `SPOTIFY_REDIRECT_URI` | Spotify callback URL | `https://api.novafield.com/api/v1/spotify/callback` |
```

- [ ] **Step 3: Commit**

```bash
git add DEPLOY.md
git commit -m "docs: add full pipeline architecture and secrets reference"
```

---

## Task 7: Add Branch Protection and Rollback Strategy

**Covers:** Production safety, rollback capability

**Files:**
- Modify: `.github/workflows/deploy.yml`

- [ ] **Step 1: Add rollback job to deploy.yml**

Add after the `summary` job:

```yaml
  # ── Rollback (manual trigger) ───────────────────────
  rollback:
    name: Rollback
    runs-on: ubuntu-latest
    if: github.event_name == 'workflow_dispatch'
    inputs:
      service:
        description: "Service to rollback (backend/frontend/all)"
        required: true
        type: choice
        options:
          - all
          - backend
          - frontend
      image-tag:
        description: "Image tag to rollback to (e.g., abc1234)"
        required: true
    steps:
      - uses: actions/checkout@v4

      - name: Rollback backend
        if: inputs.service == 'all' || inputs.service == 'backend'
        uses: ./.github/actions/darakub-deploy
        with:
          service-name: novafield-backend
          image: registry.darakub.com/novafield/backend:${{ inputs.image-tag }}
          api-key: ${{ secrets.DARAKUB_API_KEY }}

      - name: Rollback frontend
        if: inputs.service == 'all' || inputs.service == 'frontend'
        uses: ./.github/actions/darakub-deploy
        with:
          service-name: novafield-frontend
          image: registry.darakub.com/novafield/frontend:${{ inputs.image-tag }}
          api-key: ${{ secrets.DARAKUB_API_KEY }}
```

- [ ] **Step 2: Add workflow_dispatch trigger to deploy.yml**

Change the `on:` section to:

```yaml
on:
  push:
    branches: [main]
  workflow_dispatch:
    inputs:
      service:
        description: "Service to rollback"
        type: choice
        options:
          - all
          - backend
          - frontend
      image-tag:
        description: "Image tag to rollback to"
```

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/deploy.yml
git commit -m "ci: add manual rollback trigger with service selection"
```

---

## Task 8: Final Verification

**Covers:** All tasks verified

- [ ] **Step 1: Verify all files exist**

Run:
```bash
ls -la .github/workflows/ci.yml .github/workflows/deploy.yml .github/actions/darakub-deploy/action.yml .github/actions/darakub-health/action.yml docker-compose.darakub.yml
```

Expected: All 5 files listed.

- [ ] **Step 2: Verify Go builds**

Run: `cd backend && go vet ./...`
Expected: No output (clean).

- [ ] **Step 3: Verify frontend typechecks**

Run: `cd frontend && npx tsc --noEmit`
Expected: No output (clean).

- [ ] **Step 4: Verify workflow YAML is valid**

Run: `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/deploy.yml')); print('Valid YAML')"`
Expected: `Valid YAML`

- [ ] **Step 5: Final commit**

```bash
git add -A
git commit -m "ci: complete Darakub CI/CD pipeline — lint, test, build, deploy, rollback"
```

---

## Summary

| Task | Files | Description |
|------|-------|-------------|
| 1 | `ci.yml`, `deploy.yml` | Fix Go version, separate CI from deploy |
| 2 | `darakub-deploy/action.yml` | Reusable deploy action with error handling |
| 3 | `darakub-health/action.yml` | Reusable health check action |
| 4 | `deploy.yml` | Full rewrite with reusable actions |
| 5 | `docker-compose.darakub.yml` | Service definitions for Darakub |
| 6 | `DEPLOY.md` | Pipeline docs and secrets reference |
| 7 | `deploy.yml` | Rollback job with workflow_dispatch |
| 8 | — | Final verification |

### Pipeline Flow

```
PR → ci.yml (lint + test + typecheck)
Push to main → deploy.yml (lint → test → typecheck → build → deploy backend → deploy frontend → summary)
Manual → deploy.yml rollback (select service + image tag)
```

### Key Improvements Over Existing

1. **Go version fixed**: 1.23 → 1.25 (matches Dockerfile)
2. **CI/Deploy separation**: PRs get fast feedback, pushes get full deploy
3. **Reusable actions**: DRY deploy + health check logic
4. **Error handling**: HTTP status checks, timeout with retry
5. **Spotify env vars**: Secrets passed to backend
6. **Rollback support**: Manual trigger with service/tag selection
7. **Build caching**: Separate GHA cache scopes for backend/frontend
8. **Parallel builds**: Backend and frontend images build simultaneously
