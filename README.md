# 🎾 Oreo Tennis Scoring (OTS)

A lightweight, offline-first Progressive Web App for tracking tennis matches, tournaments, and serve statistics.

![OTS Logo](frontend/public/oreo-logo.png)

## ✨ Features

### Match Scoring
- 🎾 **Real-time Scoring** - Track points, games, sets with tennis notation
- 📊 **Serve Statistics** - Automatic calculation of first/second serve percentages
- 🎯 **Two Scoring Modes**:
  - **Standard Mode**: Traditional tennis (best of 3 sets)
  - **Short Format**: Quick recreational play (best of 3 games)
- 👥 **Singles & Doubles** - Support for both match types

### Tournament Management
- 🏆 **Tournament Engine** - Full tournament orchestration
- 🎲 **Random Team Generation** - Deterministic doubles team pairing
- 🔄 **Round-Robin Stage** - Every team plays every other team
- 🥇 **Knockout Stage** - Automatic semifinals and finals
- 📈 **Live Standings** - Real-time tournament rankings

### Progressive Web App
- 📱 **Mobile-First** - Install on any device, works offline
- 🌙 **Dark Mode** - Eye-friendly interface optimized for outdoor use
- ⚡ **Blazing Fast** - Sub-1s load time, minimal bundle size
- 💾 **Offline Support** - IndexedDB for local storage
- 🔄 **Auto-sync** - Syncs when connection restores

### Admin Features
- 🔒 **Secure Authentication** - JWT with bcrypt password hashing
- 👤 **Player Management** - Add, edit, activate/deactivate players
- 🏟️ **Venue Management** - Track multiple playing locations
- 📋 **Match History** - View and delete past matches

## 🏗️ Architecture

### Tech Stack

- **Frontend**: Svelte 5 + Vite (PWA)
- **Backend**: Go 1.21+ (net/http, no frameworks)
- **Database**: PostgreSQL 15+
- **Deployment**: Docker + Dokploy
- **Storage**: IndexedDB (offline), PostgreSQL (persistent)

### Core Engines

**Scoring Engine** (`backend/internal/scoring/`)
- Pure, stateless tennis scoring logic
- Supports Standard and Short-Format modes
- Tennis notation display (0, 15, 30, 40, Deuce, Ad)
- Fully tested (13 unit tests, 100% pass rate)

**Tournament Engine** (`backend/internal/tournament/`)
- Deterministic team generation (seeded randomization)
- Round-robin match generation (T × (T-1) / 2 formula)
- Standings calculation with ranking
- Knockout bracket logic (3, 4, 5+ teams)
- Fully tested (17 unit tests, 100% pass rate)

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- (Optional) Go 1.21+ and Node.js 18+ for local development

### Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
# Database
POSTGRES_DB=ots
POSTGRES_USER=ots
POSTGRES_PASSWORD=your_secure_password
DATABASE_URL=postgres://ots:password@postgres:5432/ots?sslmode=disable

# Admin Credentials
ADMIN_USERNAME=admin
ADMIN_PASSWORD_HASH=<bcrypt-hash>  # See below for generation
JWT_SECRET=<random-32-char-string>

# URLs
FRONTEND_URL=http://localhost
CORS_ORIGIN=http://localhost
VITE_API_URL=http://localhost:8080/api
```

### Generate Credentials

```bash
# Generate bcrypt hash for admin password
htpasswd -bnBC 10 "" your_password | tr -d ':\n'

# Generate JWT secret
openssl rand -hex 32
```

### Run with Docker (Recommended)

```bash
# Development mode
docker compose -f docker-compose.dev.yml up --build

# Production mode
docker compose -f docker-compose.prod.yml up --build -d
```

Access the app:
- **Frontend**: http://localhost
- **Backend API**: http://localhost:8080
- **Health Check**: http://localhost:8080/health

### Local Development (Without Docker)

#### Backend

```bash
cd backend
go mod download
go run cmd/api/main.go
```

#### Frontend

```bash
cd frontend
npm install
npm run dev
```

## 📚 Documentation

- **[Project Summary](docs/PROJECT_SUMMARY.md)** - Complete feature overview
- **[Implementation Complete](docs/IMPLEMENTATION_COMPLETE.md)** - Core logic details
- **[Scoring Spec](docs/OTS_Tennis_Scoring_Spec.md)** - Tennis scoring rules
- **[Tournament Spec](docs/OTS_Tournament_Spec.md)** - Tournament rules
- **[Quick Reference](docs/QUICK_REFERENCE.md)** - Code usage examples
- **[Deployment Guide](docs/DOKPLOY_DEPLOYMENT.md)** - Dokploy deployment
- **[UI Design Spec](docs/ui_design_spec.md)** - Frontend design guidelines
- **[Backend Spec](docs/backend_spec.md)** - API specifications
- **[Frontend Spec](docs/frontend_spec.md)** - Frontend architecture

## 🔌 API Endpoints

### Public Endpoints
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/api/players` | List active players |
| GET | `/api/venues` | List active venues |
| POST | `/api/matches` | Create new match |
| POST | `/api/matches/:id/events` | Submit point events (batch) |
| POST | `/api/matches/:id/complete` | Complete match |
| GET | `/api/matches/:id/summary` | Get match summary |

### Admin Endpoints (JWT Required)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/admin/login` | Admin login |
| POST | `/api/admin/logout` | Admin logout |
| GET | `/api/admin/players` | List all players |
| POST | `/api/admin/players` | Create player |
| PATCH | `/api/admin/players/:id` | Update player |
| GET | `/api/admin/venues` | List all venues |
| POST | `/api/admin/venues` | Create venue |
| PATCH | `/api/admin/venues/:id` | Update venue |
| GET | `/api/admin/matches` | List all matches |
| DELETE | `/api/admin/matches/:id` | Delete match |

## 📁 Project Structure

```
oreo-tennis-scoring/
├── backend/
│   ├── cmd/api/              # Application entrypoint
│   ├── internal/
│   │   ├── auth/             # JWT & bcrypt authentication
│   │   ├── config/           # Environment configuration
│   │   ├── database/         # PostgreSQL connection
│   │   ├── handler/          # HTTP request handlers
│   │   ├── middleware/       # Auth, CORS, logging
│   │   ├── model/            # Data models
│   │   ├── repository/       # Database queries
│   │   ├── service/          # Business logic
│   │   ├── scoring/          # ⭐ Tennis scoring engine
│   │   └── tournament/       # ⭐ Tournament engine
│   ├── migrations/           # SQL migrations
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── routes/           # Page components
│   │   ├── services/         # API & IndexedDB services
│   │   ├── stores/           # Svelte stores (state)
│   │   └── App.svelte        # Main app component
│   ├── public/               # Static assets, PWA manifest
│   └── Dockerfile
├── docs/                     # 📚 All documentation
├── docker-compose.dev.yml    # Development environment
├── docker-compose.prod.yml   # Production environment
├── .env.example              # Environment template
└── README.md                 # You are here!
```

## 🎮 Usage

### Starting a Match

1. **Home Screen**: Click "Start New Match"
2. **Match Setup**: Select venue and match type (Singles/Doubles, Standard/Short)
3. **Player Selection**: Choose players for Team A and Team B
4. **Live Match**: Tap the court side that won the point
5. **Match Summary**: View final statistics and serve percentages

### Starting a Tournament

1. **Home Screen**: Click "Start New Tournament"
2. **Tournament Setup**: Select venue and players (minimum 4, even number)
3. **Team Creation**: Generate random teams or create manually
4. **Round Robin**: Play all matches
5. **Knockout**: Semifinals and Final (top teams advance)
6. **Winner**: Tournament champion declared!

### Admin Panel

1. Click "Admin" on home screen
2. Login with admin credentials
3. Manage players, venues, and view match history

## 🧪 Testing

### Backend Tests

```bash
cd backend

# Test scoring engine (13 tests)
go test ./internal/scoring/... -v

# Test tournament engine (17 tests)
go test ./internal/tournament/... -v

# All tests
go test ./... -v
```

**Test Coverage**: 30 unit tests, 100% pass rate

## 🚢 Deployment

### Dokploy (Recommended)

See **[DOKPLOY_DEPLOYMENT.md](docs/DOKPLOY_DEPLOYMENT.md)** for complete guide.

Quick steps:
1. Create project in Dokploy
2. Set environment variables
3. Upload `docker-compose.prod.yml`
4. Configure domain & SSL
5. Deploy!

### Manual Docker

```bash
# Build and run
docker compose -f docker-compose.prod.yml up --build -d

# View logs
docker compose -f docker-compose.prod.yml logs -f

# Stop
docker compose -f docker-compose.prod.yml down
```

## 🔒 Security

- ✅ JWT authentication with HttpOnly cookies
- ✅ Bcrypt password hashing (cost 12)
- ✅ CORS protection
- ✅ Rate limiting (100 req/min per IP)
- ✅ SQL injection prevention (parameterized queries)
- ✅ XSS protection (Content-Security-Policy headers)
- ✅ No sensitive data in frontend

## 🎨 PWA Installation

### Mobile
1. Open app in browser
2. Tap browser menu (⋮)
3. Select "Add to Home Screen"
4. App installs with custom icon

### Desktop
1. Open app in browser
2. Click install icon in address bar
3. App installs as standalone application

## 📊 Key Metrics

- **Backend**: ~3,500 lines of production Go code
- **Frontend**: ~2,000 lines of Svelte code
- **Core Engines**: 1,374 lines (30 tests)
- **Bundle Size**: ~74 KB (gzipped: 25 KB)
- **Load Time**: < 1 second
- **Lighthouse Score**: 95+ (Performance, Accessibility, Best Practices, SEO)

## 🤝 Contributing

Contributions welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

See **[CONTRIBUTING.md](docs/CONTRIBUTING.md)** for guidelines.

## 📄 License

MIT License - See LICENSE file for details

## 🙏 Acknowledgments

- Tennis scoring rules: International Tennis Federation (ITF)
- PWA best practices: Google Web Fundamentals
- Icon design: Oreo the Tennis Bunny 🐰

---

**Built with ❤️ for tennis enthusiasts**

*Specifications treated as LAW. Correctness over cleverness.*
