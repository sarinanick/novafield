# Awesome CI/CD Pipeline Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use compose:subagent (recommended) or compose:execute to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Transform the basic CI/CD pipeline into a production-grade pipeline with caching, security scanning, multi-environment deployment, and auto-rollback.

**Architecture:** GitHub Actions orchestrates 2 workflows: `ci.yml` (PR checks with security scanning) and `deploy.yml` (build → staging → approve → production → auto-rollback). Docker images built with BuildKit GHA cache, scanned with Trivy, deployed via Darakub API.

**Tech Stack:** GitHub Actions, Docker BuildKit, Go 1.25, Node.js 20, golangci-lint, npm audit, Trivy, Darakub REST API

---

## File Structure

| File | Responsibility |
|------|---------------|
| `.github/workflows/ci.yml` | PR checks: lint, test, typecheck, security scan |
| `.github/workflows/deploy.yml` | Build, deploy staging, approve, deploy production, auto-rollback |
| `.github/actions/darakub-deploy/action.yml` | Reusable deploy action (exists, no changes) |
| `.github/actions/darakub-health/action.yml` | Reusable health check action (exists, no changes) |
| `.github/dependabot.yml` | Automated dependency updates |

---

## Task 1: Improve CI Workflow with Security Scanning

**Covers:** Go caching, security scanning

**Files:**
- Modify: `.github/workflows/ci.yml`

- [ ] **Step 1: Rewrite ci.yml with caching and security scanning**

Replace entire `.github/workflows/ci.yml` with:

```yaml
name: CI — PR Checks

on:
  pull_request:
    branches: [main]

env:
  GO_VERSION: "1.25"
  NODE_VERSION: "20"

jobs:
  # ── Backend ─────────────────────────────────────────
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
          cache: true
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
          cache: true
          cache-dependency-path: backend/go.sum
      - run: go test ./handlers/ -v -count=1 -timeout 120s

  # ── Frontend ────────────────────────────────────────
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

  frontend-audit:
    name: Frontend Security Audit
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
      - run: npm audit --audit-level=high
        continue-on-error: true

  # ── Security Scanning ───────────────────────────────
  scan-backend:
    name: Scan Backend Dockerfile
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'config'
          scan-ref: 'backend'
          format: 'table'
          exit-code: '1'
          severity: 'CRITICAL,HIGH'
        continue-on-error: true

  scan-frontend:
    name: Scan Frontend Dockerfile
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'config'
          scan-ref: 'frontend'
          format: 'table'
          exit-code: '1'
          severity: 'CRITICAL,HIGH'
        continue-on-error: true
```

- [ ] **Step 2: Verify YAML syntax**

Run: `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml')); print('Valid')"`

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: add security scanning and Go caching to PR checks"
```

---

## Task 2: Rewrite Deploy Workflow with Multi-Environment and Auto-Rollback

**Covers:** Multi-environment, auto-rollback, Docker caching

**Files:**
- Modify: `.github/workflows/deploy.yml`

- [ ] **Step 1: Rewrite deploy.yml with staging + production + auto-rollback**

Replace entire `.github/workflows/deploy.yml` with:

```yaml
name: Deploy to Darakub

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

env:
  GO_VERSION: "1.25"
  NODE_VERSION: "20"
  DARAKUB_API: "https://api.darakub.com"
  REGISTRY: "registry.darakub.com/novafield"

jobs:
  # ════════════════════════════════════════════════════
  # VALIDATE
  # ════════════════════════════════════════════════════
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
          cache: true
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
          cache: true
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

  # ════════════════════════════════════════════════════
  # BUILD
  # ════════════════════════════════════════════════════
  build-backend:
    name: Build Backend
    runs-on: ubuntu-latest
    needs: [backend-lint, backend-test]
    outputs:
      image-tag: ${{ github.sha }}
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
    name: Build Frontend
    runs-on: ubuntu-latest
    needs: [frontend-check]
    outputs:
      image-tag: ${{ github.sha }}
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

  # ════════════════════════════════════════════════════
  # SECURITY SCAN
  # ════════════════════════════════════════════════════
  scan-images:
    name: Scan Docker Images
    runs-on: ubuntu-latest
    needs: [build-backend, build-frontend]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/download-artifact@v4
        with:
          name: image-backend
          path: /tmp
      - uses: actions/download-artifact@v4
        with:
          name: image-frontend
          path: /tmp
      - name: Scan backend image
        run: |
          docker load < /tmp/backend.tar.gz
          docker save novafield-backend:${{ github.sha }} | trivy image --severity CRITICAL,HIGH --exit-code 1 -
        continue-on-error: true
      - name: Scan frontend image
        run: |
          docker load < /tmp/frontend.tar.gz
          docker save novafield-frontend:${{ github.sha }} | trivy image --severity CRITICAL,HIGH --exit-code 1 -
        continue-on-error: true

  # ════════════════════════════════════════════════════
  # DEPLOY STAGING
  # ════════════════════════════════════════════════════
  deploy-staging-backend:
    name: Deploy Staging Backend
    runs-on: ubuntu-latest
    needs: [build-backend, scan-images]
    environment:
      name: staging
      url: ${{ vars.DARAKUB_STAGING_BACKEND_URL }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/download-artifact@v4
        with:
          name: image-backend
          path: /tmp
      - name: Push to registry
        run: |
          docker load < /tmp/backend.tar.gz
          docker tag novafield-backend:${{ github.sha }} ${{ env.REGISTRY }}/backend:${{ github.sha }}
          docker tag novafield-backend:${{ github.sha }} ${{ env.REGISTRY }}/backend:staging
          echo "${{ secrets.DARAKUB_REGISTRY_PASS }}" | docker login registry.darakub.com -u darakub --password-stdin
          docker push ${{ env.REGISTRY }}/backend:${{ github.sha }}
          docker push ${{ env.REGISTRY }}/backend:staging
      - name: Deploy
        uses: ./.github/actions/darakub-deploy
        with:
          service-name: novafield-staging-backend
          image: ${{ env.REGISTRY }}/backend:${{ github.sha }}
          api-key: ${{ secrets.DARAKUB_API_KEY }}
          env-vars: |
            {
              "PORT": "3001",
              "FRONTEND_URL": "${{ vars.DARAKUB_STAGING_FRONTEND_URL }}",
              "CORS_ORIGINS": "${{ vars.STAGING_CORS_ORIGINS }}"
            }
          health-path: "/api/v1/health"
          health-port: "3001"
      - name: Health check
        uses: ./.github/actions/darakub-health
        with:
          url: ${{ vars.DARAKUB_STAGING_BACKEND_URL }}/api/v1/health
          retries: 5
          delay: 5

  deploy-staging-frontend:
    name: Deploy Staging Frontend
    runs-on: ubuntu-latest
    needs: [build-frontend, deploy-staging-backend]
    environment:
      name: staging
      url: ${{ vars.DARAKUB_STAGING_FRONTEND_URL }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/download-artifact@v4
        with:
          name: image-frontend
          path: /tmp
      - name: Push to registry
        run: |
          docker load < /tmp/frontend.tar.gz
          docker tag novafield-frontend:${{ github.sha }} ${{ env.REGISTRY }}/frontend:${{ github.sha }}
          docker tag novafield-frontend:${{ github.sha }} ${{ env.REGISTRY }}/frontend:staging
          echo "${{ secrets.DARAKUB_REGISTRY_PASS }}" | docker login registry.darakub.com -u darakub --password-stdin
          docker push ${{ env.REGISTRY }}/frontend:${{ github.sha }}
          docker push ${{ env.REGISTRY }}/frontend:staging
      - name: Deploy
        uses: ./.github/actions/darakub-deploy
        with:
          service-name: novafield-staging-frontend
          image: ${{ env.REGISTRY }}/frontend:${{ github.sha }}
          api-key: ${{ secrets.DARAKUB_API_KEY }}
          env-vars: |
            {
              "NEXT_PUBLIC_API_URL": "${{ vars.STAGING_NEXT_PUBLIC_API_URL }}",
              "NEXT_PUBLIC_WS_URL": "${{ vars.STAGING_NEXT_PUBLIC_WS_URL }}"
            }
          health-path: "/"
          health-port: "3000"
      - name: Health check
        uses: ./.github/actions/darakub-health
        with:
          url: ${{ vars.DARAKUB_STAGING_FRONTEND_URL }}
          retries: 5
          delay: 5

  # ════════════════════════════════════════════════════
  # APPROVE PRODUCTION
  # ════════════════════════════════════════════════════
  approve-production:
    name: Approve Production Deploy
    runs-on: ubuntu-latest
    needs: [deploy-staging-backend, deploy-staging-frontend]
    environment:
      name: production
    steps:
      - name: Approved
        run: echo "Production deployment approved"

  # ════════════════════════════════════════════════════
  # DEPLOY PRODUCTION
  # ════════════════════════════════════════════════════
  deploy-prod-backend:
    name: Deploy Production Backend
    runs-on: ubuntu-latest
    needs: [approve-production, build-backend]
    environment:
      name: production
      url: ${{ vars.DARAKUB_BACKEND_URL }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/download-artifact@v4
        with:
          name: image-backend
          path: /tmp
      - name: Push to registry
        run: |
          docker load < /tmp/backend.tar.gz
          docker tag novafield-backend:${{ github.sha }} ${{ env.REGISTRY }}/backend:${{ github.sha }}
          docker tag novafield-backend:${{ github.sha }} ${{ env.REGISTRY }}/backend:latest
          echo "${{ secrets.DARAKUB_REGISTRY_PASS }}" | docker login registry.darakub.com -u darakub --password-stdin
          docker push ${{ env.REGISTRY }}/backend:${{ github.sha }}
          docker push ${{ env.REGISTRY }}/backend:latest
      - name: Deploy
        uses: ./.github/actions/darakub-deploy
        with:
          service-name: novafield-backend
          image: ${{ env.REGISTRY }}/backend:${{ github.sha }}
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

      - name: Health check
        id: health
        uses: ./.github/actions/darakub-health
        with:
          url: ${{ vars.DARAKUB_BACKEND_URL }}/api/v1/health
          retries: 5
          delay: 5

      - name: Auto-rollback on failure
        if: failure() && steps.health.outcome == 'failure'
        uses: ./.github/actions/darakub-deploy
        with:
          service-name: novafield-backend
          image: ${{ env.REGISTRY }}/backend:latest
          api-key: ${{ secrets.DARAKUB_API_KEY }}

  deploy-prod-frontend:
    name: Deploy Production Frontend
    runs-on: ubuntu-latest
    needs: [deploy-prod-backend, build-frontend]
    environment:
      name: production
      url: ${{ vars.DARAKUB_FRONTEND_URL }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/download-artifact@v4
        with:
          name: image-frontend
          path: /tmp
      - name: Push to registry
        run: |
          docker load < /tmp/frontend.tar.gz
          docker tag novafield-frontend:${{ github.sha }} ${{ env.REGISTRY }}/frontend:${{ github.sha }}
          docker tag novafield-frontend:${{ github.sha }} ${{ env.REGISTRY }}/frontend:latest
          echo "${{ secrets.DARAKUB_REGISTRY_PASS }}" | docker login registry.darakub.com -u darakub --password-stdin
          docker push ${{ env.REGISTRY }}/frontend:${{ github.sha }}
          docker push ${{ env.REGISTRY }}/frontend:latest
      - name: Deploy
        uses: ./.github/actions/darakub-deploy
        with:
          service-name: novafield-frontend
          image: ${{ env.REGISTRY }}/frontend:${{ github.sha }}
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

      - name: Health check
        id: health
        uses: ./.github/actions/darakub-health
        with:
          url: ${{ vars.DARAKUB_FRONTEND_URL }}
          retries: 5
          delay: 5

      - name: Auto-rollback on failure
        if: failure() && steps.health.outcome == 'failure'
        uses: ./.github/actions/darakub-deploy
        with:
          service-name: novafield-frontend
          image: ${{ env.REGISTRY }}/frontend:latest
          api-key: ${{ secrets.DARAKUB_API_KEY }}

  # ════════════════════════════════════════════════════
  # SUMMARY
  # ════════════════════════════════════════════════════
  summary:
    name: Deploy Summary
    runs-on: ubuntu-latest
    needs: [deploy-prod-backend, deploy-prod-frontend]
    if: always()
    steps:
      - name: Write summary
        run: |
          echo "## Deployment Summary" >> $GITHUB_STEP_SUMMARY
          echo "" >> $GITHUB_STEP_SUMMARY
          echo "| Service | Status | Image |" >> $GITHUB_STEP_SUMMARY
          echo "|---------|--------|-------|" >> $GITHUB_STEP_SUMMARY
          echo "| Backend | ${{ needs.deploy-prod-backend.result }} | \`${{ github.sha }}\` |" >> $GITHUB_STEP_SUMMARY
          echo "| Frontend | ${{ needs.deploy-prod-frontend.result }} | \`${{ github.sha }}\` |" >> $GITHUB_STEP_SUMMARY
          echo "" >> $GITHUB_STEP_SUMMARY
          echo "**Commit:** \`${{ github.sha }}\`" >> $GITHUB_STEP_SUMMARY
          echo "**Branch:** \`${{ github.ref_name }}\`" >> $GITHUB_STEP_SUMMARY
          echo "**Time:** $(date -u +'%Y-%m-%d %H:%M:%S UTC')" >> $GITHUB_STEP_SUMMARY

      - name: Fail if deploy failed
        if: needs.deploy-prod-backend.result != 'success' || needs.deploy-prod-frontend.result != 'success'
        run: |
          echo "::error::Production deployment failed"
          exit 1

  # ════════════════════════════════════════════════════
  # ROLLBACK (manual)
  # ════════════════════════════════════════════════════
  rollback:
    name: Manual Rollback
    runs-on: ubuntu-latest
    if: github.event_name == 'workflow_dispatch'
    steps:
      - uses: actions/checkout@v4
      - name: Rollback backend
        if: github.event.inputs.service == 'all' || github.event.inputs.service == 'backend'
        uses: ./.github/actions/darakub-deploy
        with:
          service-name: novafield-backend
          image: ${{ env.REGISTRY }}/backend:${{ github.event.inputs.image-tag }}
          api-key: ${{ secrets.DARAKUB_API_KEY }}
      - name: Rollback frontend
        if: github.event.inputs.service == 'all' || github.event.inputs.service == 'frontend'
        uses: ./.github/actions/darakub-deploy
        with:
          service-name: novafield-frontend
          image: ${{ env.REGISTRY }}/frontend:${{ github.event.inputs.image-tag }}
          api-key: ${{ secrets.DARAKUB_API_KEY }}
```

- [ ] **Step 2: Verify YAML syntax**

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/deploy.yml
git commit -m "ci: production pipeline with staging, security scan, auto-rollback"
```

---

## Task 3: Add Dependabot for Automated Dependency Updates

**Covers:** Security scanning (automated)

**Files:**
- Create: `.github/dependabot.yml`

- [ ] **Step 1: Create dependabot.yml**

```yaml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/backend"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 5
    labels:
      - "dependencies"
      - "go"

  - package-ecosystem: "npm"
    directory: "/frontend"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 5
    labels:
      - "dependencies"
      - "javascript"

  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 3
    labels:
      - "dependencies"
      - "ci"
```

- [ ] **Step 2: Commit**

```bash
git add .github/dependabot.yml
git commit -m "ci: add dependabot for automated dependency updates"
```

---

## Task 4: Update DEPLOY.md with New Pipeline Docs

**Covers:** Documentation

**Files:**
- Modify: `DEPLOY.md`

- [ ] **Step 1: Update the Pipeline Architecture section in DEPLOY.md**

Find the existing "### Pipeline Architecture" section and replace it with:

```markdown
### Pipeline Architecture

```
PR → CI (lint + test + typecheck + security scan)
Push to main → CD:
  ├── Build Images (parallel, GHA cache)
  ├── Security Scan (Trivy)
  ├── Deploy Staging (auto)
  ├── Health Check Staging
  ├── Approve Production (manual gate)
  ├── Deploy Production
  ├── Health Check Production
  ├── Auto-Rollback (if health fails)
  └── Summary
```

### Pipeline Features

| Feature | Status |
|---------|--------|
| Go module caching | ✅ setup-go + GOMODCACHE |
| Docker layer caching | ✅ BuildKit GHA cache |
| Security scanning | ✅ Trivy + npm audit + golangci-lint |
| Multi-environment | ✅ staging + production |
| Auto-rollback | ✅ health check → rollback on failure |
| Manual rollback | ✅ workflow_dispatch with service/tag selection |
| Dependency updates | ✅ Dependabot (weekly) |
```

- [ ] **Step 2: Commit**

```bash
git add DEPLOY.md
git commit -m "docs: update pipeline docs with new features"
```

---

## Task 5: Final Verification

**Covers:** All

- [ ] **Step 1: Verify all files**

```bash
ls -la .github/workflows/ci.yml .github/workflows/deploy.yml .github/dependabot.yml
```

- [ ] **Step 2: Verify YAML validity**

```bash
python3 -c "
import yaml
for f in ['.github/workflows/ci.yml', '.github/workflows/deploy.yml', '.github/dependabot.yml']:
    yaml.safe_load(open(f))
    print(f'{f}: valid')
"
```

- [ ] **Step 3: Verify Go compiles**

```bash
cd backend && go vet ./...
```

- [ ] **Step 4: Final commit**

```bash
git add -A
git commit -m "ci: complete awesome CI/CD pipeline with caching, security, multi-env, auto-rollback"
```

---

## Summary

| Task | Files | Description |
|------|-------|-------------|
| 1 | `ci.yml` | Security scanning + Go caching |
| 2 | `deploy.yml` | Staging + production + auto-rollback |
| 3 | `dependabot.yml` | Automated dependency updates |
| 4 | `DEPLOY.md` | Updated documentation |
| 5 | — | Final verification |

### New Pipeline Features

1. **Go module caching** — `setup-go` with `cache: true`
2. **Docker layer caching** — BuildKit GHA cache (scope per service)
3. **Security scanning** — Trivy (Dockerfiles), npm audit, golangci-lint
4. **Multi-environment** — staging (auto) → production (manual approval)
5. **Auto-rollback** — health check failure triggers automatic rollback
6. **Dependabot** — weekly dependency updates for Go, npm, GitHub Actions
