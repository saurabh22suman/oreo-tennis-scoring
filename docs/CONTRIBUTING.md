# Contributing to Oreo Tennis Scoring (OTS)

Thank you for your interest in contributing! This document provides guidelines for contributing to OTS.

## Development Setup

```bash
# Clone repository
git clone https://github.com/saurabh22suman/oreo-tennis-scoring.git
cd oreo-tennis-scoring

# Run setup script
./setup-dev.sh

# Start development environment
docker-compose -f docker-compose.dev.yml up
```

## Project Structure

```
oreo-tennis-scoring/
├── backend/              # Go REST API
│   ├── cmd/api/         # Main entry point  
│   ├── internal/        # Internal packages
│   │   ├── auth/       # JWT & bcrypt
│   │   ├── config/     # Environment config
│   │   ├── database/   # PostgreSQL
│   │   ├── handler/    # HTTP handlers
│   │   ├── middleware/ # Auth, CORS, rate limiting
│   │   ├── model/      # Data models
│   │   ├── repository/ # Database queries
│   │   └── service/    # Business logic
│   └── migrations/     # SQL migrations
├── frontend/            # Svelte PWA
│   ├── src/
│   │   ├── routes/     # Page components
│   │   ├── services/   # API & IndexedDB
│   │   └── stores/     # Svelte stores
│   └── public/         # Static assets
├── docs/               # Documentation
└── scripts/            # Utility scripts
```

## Coding Standards

### Backend (Go)

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Keep handlers thin, logic in services
- All database operations in repositories
- Validate all input in handlers
- Return proper HTTP status codes

### Frontend (Svelte)

- Follow design tokens in `app.css`
- Mobile-first approach
- Offline-first for match screens
- No blocking API calls during live match
- Use Svelte stores for state management
- Keep components focused and reusable

## Pull Request Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Test thoroughly (manual testing required)
5. Commit with clear messages (`git commit -m 'Add amazing feature'`)
6. Push to your fork (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### PR Checklist

- [ ] Code follows project style
- [ ] Changes are tested
- [ ] No secrets or credentials in code
- [ ] README updated if needed
- [ ] Deployment guide updated if needed

## Testing

Currently, the project uses manual testing. When adding features:

1. Test all user flows end-to-end
2. Test offline functionality
3. Test on mobile devices (iOS + Android)
4. Test admin operations
5. Verify PWA installability

## Security

- Never commit `.env` files
- Never log sensitive data (passwords, JWT tokens)
- All admin routes must be protected
- Input validation on all endpoints
- Use prepared statements for SQL (already done via pgx)

## Architecture Principles

### Backend

- **Stateless**: No server-side sessions
- **Idempotent**: Events can be submitted multiple times safely
- **Validated**: All input validated before processing
- **Secured**: Bcrypt for passwords, JWT for auth, HttpOnly cookies

### Frontend

- **Offline-first**: Match scoring never fails due to network
- **IndexedDB**: Local storage for current match + event queue
- **Auto-sync**: Background sync when online
- **PWA**: Installable, responsive, fast

## Questions?

Open an issue for:
- Bug reports
- Feature requests
- Architecture questions
- Documentation improvements

Thank you for contributing! 🎾
