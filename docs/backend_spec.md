# OTS Backend Specification

## 1. Architecture
- Language: Go
- Framework: net/http or Fiber
- Database: PostgreSQL
- Authentication: Admin-only, bcrypt-hashed credentials via env vars
- Hosting: VPS (Dokploy-managed)
- Reverse Proxy: Dokploy built-in proxy (HTTPS)

## 2. Authentication Model

### Admin Credentials
Stored in environment variables:
- ADMIN_USERNAME
- ADMIN_PASSWORD_HASH (bcrypt)
- JWT_SECRET

### Login Flow
1. Admin submits username & password
2. Backend compares password with bcrypt hash
3. On success, issue JWT
4. JWT stored in HttpOnly, Secure cookie

No admin table exists in the database.

## 3. Data Models

### Players
- id (UUID, PK)
- name (string)
- active (boolean)
- created_at (timestamp)

### Venues
- id (UUID, PK)
- name (string)
- surface (hard | clay | grass)
- active (boolean)

### Matches
- id (UUID, PK)
- venue_id (UUID, FK)
- match_type (singles | doubles)
- started_at (timestamp)
- ended_at (timestamp)

### Match_Players
- match_id (UUID, FK)
- player_id (UUID, FK)
- team (A | B)

### Point_Events
- id (UUID, PK)
- match_id (UUID, FK)
- timestamp
- server_player_id (UUID)
- serve_type (first | second | double_fault)
- point_winner_team (A | B)

## 4. API Endpoints

### Auth
- POST /api/admin/login
- POST /api/admin/logout

### Admin (JWT required)
- GET /api/admin/players
- POST /api/admin/players
- PATCH /api/admin/players/:id
- GET /api/admin/venues
- POST /api/admin/venues
- PATCH /api/admin/venues/:id
- DELETE /api/admin/matches/:id

### Matches
- POST /api/matches
- POST /api/matches/:id/events (batch supported)
- POST /api/matches/:id/complete
- GET /api/matches/:id/summary

## 5. Offline Sync & Idempotency
- Events are accepted in batches
- Each event has a UUID
- Duplicate event IDs are ignored
- Conflict resolution by timestamp (last-write-wins)

## 6. Security
- HTTPS only
- JWT stored in HttpOnly cookies
- Admin login rate-limited
- CORS restricted to frontend domain
- Input validation on all endpoints

## 7. Deployment
- Dockerized services
- Env vars injected by Dokploy
- PostgreSQL volume-backed storage
- Stateless backend
