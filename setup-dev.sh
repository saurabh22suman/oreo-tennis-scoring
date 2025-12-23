#!/bin/bash

# Development setup script for OTS

set -e

echo "🎾 Oreo Tennis Scoring - Development Setup"
echo "==========================================="
echo ""

# Check if .env exists
if [ ! -f .env ]; then
    echo "📝 Creating .env file from template..."
    cp .env.example .env
    
    echo ""
    echo "⚠️  IMPORTANT: Edit .env file with your settings:"
    echo "   1. Set POSTGRES_PASSWORD"
    echo "   2. Generate admin password hash: go run scripts/hash_password.go 'your-password'"
    echo "   3. Generate JWT secret: openssl rand -base64 32"
    echo ""
    read -p "Press Enter after updating .env file..."
fi

echo ""
echo "🔧 Installing backend dependencies..."
cd backend
go mod download
cd ..

echo ""
echo "🔧 Installing frontend dependencies..."
cd frontend
npm install
cd ..

echo ""
echo "✅ Setup complete!"
echo ""
echo "To start development:"
echo "  Option 1 - Docker Compose (recommended):"
echo "    docker-compose -f docker-compose.dev.yml up --build"
echo ""
echo "  Option 2 - Manual:"
echo "    Terminal 1: cd backend && go run cmd/api/main.go"
echo "    Terminal 2: cd frontend && npm run dev"
echo "    Terminal 3: docker run -p 5432:5432 -e POSTGRES_PASSWORD=ots_dev_password postgres:15-alpine"
echo ""
echo "Access the app:"
echo "  Frontend: http://localhost:5173"
echo "  Backend:  http://localhost:8080"
echo "  Admin login: username='admin', password='admin123' (dev only)"
echo ""
