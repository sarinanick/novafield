#!/bin/bash
# ============================================
# NovaField AI — Darkube One-Click Deploy
# ============================================
# This script prepares your environment and
# gives you step-by-step instructions.
# ============================================

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${CYAN}╔══════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║  NovaField AI — Darkube Deploy Script     ║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════╝${NC}"
echo ""

# ------------------------------------------
# Step 0: Check prerequisites
# ------------------------------------------
echo -e "${YELLOW}[0/6] Checking prerequisites...${NC}"

if ! command -v gh &> /dev/null; then
    echo -e "${RED}✗ GitHub CLI (gh) not found. Install: https://cli.github.com/${NC}"
    exit 1
fi
echo -e "${GREEN}✓ GitHub CLI found${NC}"

if ! gh auth status &> /dev/null; then
    echo -e "${RED}✗ Not logged in to GitHub. Run: gh auth login${NC}"
    exit 1
fi
echo -e "${GREEN}✓ GitHub authenticated${NC}"

if [ ! -f "docker-compose.yml" ]; then
    echo -e "${RED}✗ Not in novafield root directory${NC}"
    exit 1
fi
echo -e "${GREEN}✓ In novafield directory${NC}"

echo ""

# ------------------------------------------
# Step 1: Ensure GitHub repos exist
# ------------------------------------------
echo -e "${YELLOW}[1/6] Checking GitHub repos...${NC}"

MAIN_REPO="sarinanick/novafield"
BACKEND_REPO="sarinanick/novafield-backend"

if gh repo view "$MAIN_REPO" &> /dev/null; then
    echo -e "${GREEN}✓ Main repo exists: $MAIN_REPO${NC}"
else
    echo -e "${YELLOW}Creating main repo...${NC}"
    gh repo create "$MAIN_REPO" --private --source=. --push
    echo -e "${GREEN}✓ Main repo created${NC}"
fi

if gh repo view "$BACKEND_REPO" &> /dev/null; then
    echo -e "${GREEN}✓ Backend repo exists: $BACKEND_REPO${NC}"
else
    echo -e "${YELLOW}Creating backend repo...${NC}"
    gh repo create "$BACKEND_REPO" --private --source=./backend --push
    echo -e "${GREEN}✓ Backend repo created${NC}"
fi

echo ""

# ------------------------------------------
# Step 2: Push latest code
# ------------------------------------------
echo -e "${YELLOW}[2/6] Pushing latest code...${NC}"

cd backend
if [ -n "$(git status --porcelain)" ]; then
    git add -A
    git commit -m "deploy: prepare for Darkube" --allow-empty
fi
git push github master 2>/dev/null || git push origin master
cd ..

if [ -n "$(git status --porcelain)" ]; then
    git add -A
    git commit -m "deploy: prepare for Darkube" --allow-empty
fi
git push github master 2>/dev/null || git push origin master

echo -e "${GREEN}✓ Code pushed${NC}"
echo ""

# ------------------------------------------
# Step 3: Generate env files
# ------------------------------------------
echo -e "${YELLOW}[3/6] Generating environment files...${NC}"

BACKEND_DOMAIN="YOUR_BACKEND_DOMAIN"
FRONTEND_DOMAIN="YOUR_FRONTEND_DOMAIN"

if [ -n "$1" ]; then
    BACKEND_DOMAIN="$1"
fi
if [ -n "$2" ]; then
    FRONTEND_DOMAIN="$2"
fi

cat > deploy/.env.backend << EOF
# Backend Environment Variables
# Set these in Darkube dashboard

PORT=3001
CORS_ORIGINS=https://${BACKEND_DOMAIN}
EOF

cat > deploy/.env.frontend << EOF
# Frontend Environment Variables
# Set these in Darkube dashboard

NEXT_PUBLIC_API_URL=https://${BACKEND_DOMAIN}/api/v1
NEXT_PUBLIC_WS_URL=wss://${BACKEND_DOMAIN}
EOF

echo -e "${GREEN}✓ Env files generated in deploy/${NC}"
echo ""

# ------------------------------------------
# Step 4: Create deployment checklist
# ------------------------------------------
echo -e "${YELLOW}[4/6] Creating deployment checklist...${NC}"

cat > deploy/CHECKLIST.md << 'EOF'
# Darkube Deployment Checklist

## Backend Service

- [ ] Create service: `novafield-backend`
- [ ] Source: Git → `https://github.com/sarinanick/novafield-backend`
- [ ] Build: Dockerfile
- [ ] Dockerfile Path: `Dockerfile`
- [ ] Port: `3001`
- [ ] Add environment variables (see deploy/.env.backend)
- [ ] Add persistent volume: `/app/novafield.json`
- [ ] Add persistent volume: `/app/uploads`
- [ ] Deploy

## Frontend Service

- [ ] Create service: `novafield-frontend`
- [ ] Source: Git → `https://github.com/sarinanick/novafield`
- [ ] Build: Dockerfile
- [ ] Dockerfile Path: `frontend/Dockerfile`
- [ ] Port: `3000`
- [ ] Add environment variables (see deploy/.env.frontend)
- [ ] Deploy

## Post-Deploy

- [ ] Test: GET /api/v1/health
- [ ] Test: Register user
- [ ] Test: WebSocket connection
- [ ] Configure custom domains
- [ ] Enable SSL/TLS
EOF

echo -e "${GREEN}✓ Checklist created: deploy/CHECKLIST.md${NC}"
echo ""

# ------------------------------------------
# Step 5: Verify Dockerfiles
# ------------------------------------------
echo -e "${YELLOW}[5/6] Verifying Dockerfiles...${NC}"

if [ -f "backend/Dockerfile" ]; then
    echo -e "${GREEN}✓ Backend Dockerfile exists${NC}"
else
    echo -e "${RED}✗ Backend Dockerfile missing${NC}"
fi

if [ -f "frontend/Dockerfile" ]; then
    echo -e "${GREEN}✓ Frontend Dockerfile exists${NC}"
else
    echo -e "${RED}✗ Frontend Dockerfile missing${NC}"
fi

echo ""

# ------------------------------------------
# Step 6: Print instructions
# ------------------------------------------
echo -e "${YELLOW}[6/6] Deploy Instructions${NC}"
echo ""
echo -e "${CYAN}═══════════════════════════════════════════${NC}"
echo -e "${CYAN}  WHAT TO DO ON DARKUBE:${NC}"
echo -e "${CYAN}═══════════════════════════════════════════${NC}"
echo ""
echo -e "${GREEN}1. Go to https://darkube.app/dashboard${NC}"
echo -e "${GREEN}2. Create Backend Service:${NC}"
echo -e "   - Name: novafield-backend"
echo -e "   - Source: Git → sarinanick/novafield-backend"
echo -e "   - Dockerfile Path: Dockerfile"
echo -e "   - Port: 3001"
echo ""
echo -e "${GREEN}3. Set Backend Environment Variables:${NC}"
echo -e "   - PORT=3001"
echo -e "   - CORS_ORIGINS=https://YOUR_BACKEND_DOMAIN"
echo ""
echo -e "${GREEN}4. Add Backend Persistent Storage:${NC}"
echo -e "   - /app/novafield.json"
echo -e "   - /app/uploads"
echo ""
echo -e "${GREEN}5. Create Frontend Service:${NC}"
echo -e "   - Name: novafield-frontend"
echo -e "   - Source: Git → sarinanick/novafield"
echo -e "   - Dockerfile Path: frontend/Dockerfile"
echo -e "   - Port: 3000"
echo ""
echo -e "${GREEN}6. Set Frontend Environment Variables:${NC}"
echo -e "   - NEXT_PUBLIC_API_URL=https://YOUR_BACKEND_DOMAIN/api/v1"
echo -e "   - NEXT_PUBLIC_WS_URL=wss://YOUR_BACKEND_DOMAIN"
echo ""
echo -e "${GREEN}7. Deploy both services${NC}"
echo ""
echo -e "${CYAN}═══════════════════════════════════════════${NC}"
echo -e "${CYAN}  INFRASTRUCTURE NEEDED:${NC}"
echo -e "${CYAN}═══════════════════════════════════════════${NC}"
echo ""
echo -e "${YELLOW}Required:${NC}"
echo -e "  ✓ Darkube account"
echo -e "  ✓ 2 services (backend + frontend)"
echo -e "  ✓ Persistent storage (2 volumes)"
echo ""
echo -e "${YELLOW}Optional (for better performance):${NC}"
echo -e "  ○ Redis (session cache) — not required yet"
echo -e "  ○ PostgreSQL (if you outgrow JSON DB) — not required yet"
echo -e "  ○ Custom domain + SSL"
echo ""
echo -e "${CYAN}═══════════════════════════════════════════${NC}"
echo ""
echo -e "${GREEN}Done! Files ready in deploy/${NC}"
echo -e "${GREEN}Run: cat deploy/CHECKLIST.md${NC}"
