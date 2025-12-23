#!/bin/bash

# Quick verification script for OTS installation

echo "🎾 OTS Installation Verification"
echo "================================="
echo ""

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

check_file() {
    if [ -f "$1" ]; then
        echo -e "${GREEN}✓${NC} $1"
        return 0
    else
        echo -e "${RED}✗${NC} $1 - MISSING"
        return 1
    fi
}

check_dir() {
    if [ -d "$1" ]; then
        echo -e "${GREEN}✓${NC} $1/"
        return 0
    else
        echo -e "${RED}✗${NC} $1/ - MISSING"
        return 1
    fi
}

# Check backend files
echo "Backend Files:"
check_file "backend/go.mod"
check_file "backend/cmd/api/main.go"
check_dir "backend/internal/auth"
check_dir "backend/internal/handler"
check_dir "backend/internal/repository"
check_file "backend/Dockerfile"

echo ""

# Check frontend files
echo "Frontend Files:"
check_file "frontend/package.json"
check_file "frontend/src/App.svelte"
check_file "frontend/src/main.js"
check_dir "frontend/src/routes"
check_dir "frontend/src/services"
check_dir "frontend/src/stores"
check_file "frontend/public/manifest.json"
check_file "frontend/public/sw.js"
check_file "frontend/Dockerfile"

echo ""

# Check configuration files
echo "Configuration Files:"
check_file ".env.example"
check_file "docker-compose.dev.yml"
check_file "docker-compose.prod.yml"
check_file ".gitignore"

echo ""

# Check documentation
echo "Documentation:"
check_file "README.md"
check_file "DEPLOYMENT.md"
check_file "CONTRIBUTING.md"
check_file "PROJECT_SUMMARY.md"

echo ""

# Check for .env file
if [ -f ".env" ]; then
    echo -e "${GREEN}✓${NC} .env file exists"
    
    # Check critical env vars
    if grep -q "ADMIN_PASSWORD_HASH=\$2a" .env 2>/dev/null; then
        echo -e "${GREEN}  ✓${NC} Admin password hash set"
    else
        echo -e "${YELLOW}  !${NC} Admin password hash not set (run: go run scripts/hash_password.go 'password')"
    fi
    
    if grep -q "JWT_SECRET=.*[a-zA-Z0-9]{32}" .env 2>/dev/null; then
        echo -e "${GREEN}  ✓${NC} JWT secret appears set"
    else
        echo -e "${YELLOW}  !${NC} JWT secret might be too short (run: openssl rand -base64 32)"
    fi
else
    echo -e "${YELLOW}!${NC} .env file not found (copy from .env.example)"
fi

echo ""

# Check if Go dependencies are downloaded
if [ -d "backend/vendor" ] || [ -f "backend/go.sum" ]; then
    echo -e "${GREEN}✓${NC} Go dependencies downloaded"
else
    echo -e "${YELLOW}!${NC} Go dependencies not downloaded (run: cd backend && go mod download)"
fi

# Check if Node dependencies are installed
if [ -d "frontend/node_modules" ]; then
    echo -e "${GREEN}✓${NC} Node dependencies installed"
else
    echo -e "${YELLOW}!${NC} Node dependencies not installed (run: cd frontend && npm install)"
fi

echo ""
echo "================================="
echo ""

# Final recommendation
if [ -f ".env" ] && [ -f "backend/go.sum" ] && [ -d "frontend/node_modules" ]; then
    echo -e "${GREEN}✅ Ready to start development!${NC}"
    echo ""
    echo "Run: docker-compose -f docker-compose.dev.yml up --build"
else
    echo -e "${YELLOW}⚠️  Some setup steps remaining${NC}"
    echo ""
    echo "Run: ./setup-dev.sh"
fi

echo ""
