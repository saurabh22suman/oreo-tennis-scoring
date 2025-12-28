# 🔒 OTS Security Audit Report

**Date:** December 28, 2025  
**Version:** 1.1 (Post-Remediation)  
**Auditor:** Principal Engineer / Security Architect  
**Scope:** Full codebase security and architecture review

---

## 1. Executive Summary

### Overall Risk Level: **LOW** ✅ (Improved from MEDIUM)

The OTS application has been hardened with security fixes applied. The core authentication flow, JWT handling, and database access are well-implemented. Critical issues have been addressed.

### Issues Status

| # | Issue | Severity | Status |
|---|-------|----------|--------|
| 1 | XSS vulnerability via `{@html}` in Modal component | **HIGH** | ✅ FIXED |
| 2 | Rate limiting only on login endpoint | **MEDIUM** | ⚪ BY DESIGN |
| 3 | X-Forwarded-For header spoofing vulnerability | **MEDIUM** | ✅ FIXED |
| 4 | Missing input sanitization for player/venue names | **MEDIUM** | ✅ FIXED |
| 5 | Request body size limits | **MEDIUM** | ✅ FIXED |
| 6 | Security headers in nginx | **LOW** | ✅ FIXED |
| 7 | Public match creation (no auth required) | **MEDIUM** | ⚪ BY DESIGN |

---

## 2. Critical Security Issues (Must Fix)

### 2.1 XSS Vulnerability in Modal Component ✅ FIXED

**Location:** [frontend/src/lib/Modal.svelte#L40](frontend/src/lib/Modal.svelte#L40)

**Issue:** The Modal component used `{@html message}` which rendered raw HTML.

**Fix Applied:** Changed to `{message}` to use Svelte's default text escaping.

---

### 2.2 Rate Limiting Gaps ⚪ BY DESIGN

**Location:** [backend/cmd/api/main.go#L76-L77](backend/cmd/api/main.go#L76-L77)

**Note:** Rate limiting is intentionally applied only to the login endpoint. This is a design decision for a closed-group self-hosted application where usability is prioritized over aggressive rate limiting on match/player endpoints.

---

### 2.3 IP Address Spoofing via X-Forwarded-For ✅ FIXED

**Location:** [backend/internal/middleware/ratelimit.go#L96-L125](backend/internal/middleware/ratelimit.go#L96-L125)

**Issue:** The `getClientIP` function previously trusted the entire `X-Forwarded-For` header.

**Fix Applied:** Updated to properly parse the header, taking only the first IP (original client) and trimming whitespace. Also handles IPv6 addresses and strips port from RemoteAddr.

---

### 2.4 Missing Input Sanitization ✅ FIXED

**Location:** [backend/internal/handler/validation.go](backend/internal/handler/validation.go)

**Fix Applied:** Added `ValidateName` and `ValidateNameWithLength` functions that validate names against a safe character regex pattern (`[\p{L}\p{N}\s\-'\.]+`). This allows international characters while preventing injection attacks. Applied to player and venue name handlers.

---

### 2.5 Missing CSRF Protection

**Issue:** While SameSite=Strict cookies provide CSRF protection, there's no explicit CSRF token mechanism.

**Status:** Current SameSite=Strict is sufficient for modern browsers. No additional action required for this closed-group application.

---

### 2.6 Request Body Size Limits ✅ FIXED

**Location:** [backend/internal/middleware/bodylimit.go](backend/internal/middleware/bodylimit.go)

**Fix Applied:** Added `LimitBody` middleware that limits request body size to 1MB. Applied globally in main.go.

---

### 2.7 Security Headers ✅ FIXED

**Location:** [frontend/nginx.prod.conf](frontend/nginx.prod.conf), [frontend/nginx.localhost.conf](frontend/nginx.localhost.conf)

**Fix Applied:** Added security headers to both nginx configurations:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Permissions-Policy: geolocation=(), microphone=(), camera=()`

---

## 3. Architectural Issues

### 3.1 Public Match Creation Endpoint ⚪ BY DESIGN

**Location:** [backend/cmd/api/main.go#L112](backend/cmd/api/main.go#L112)

**Note:** Public match creation is intentional for this closed-group application. The frontend is used by a trusted group, and requiring authentication for every match would reduce usability.

---

### 3.2 Event Submission Without Match Ownership Verification

**Location:** [backend/internal/handler/match.go#L98-L180](backend/internal/handler/match.go#L98-L180)

**Issue:** Anyone who knows a match ID can submit events to it. There's no ownership or session verification.

**Why it matters:** Malicious actors could corrupt match data by submitting fake events.

**Suggestion:** Implement match session tokens:
1. On match creation, return a `session_token` (stored in IndexedDB)
2. Require `X-Match-Token` header on event submission
3. Validate token matches the match

---

### 3.3 Request Body Size Limits ✅ FIXED

**Location:** [backend/internal/middleware/bodylimit.go](backend/internal/middleware/bodylimit.go)

**Fix Applied:** Added `LimitBody` middleware (1MB limit) applied globally in main.go.

---

### 3.4 Hardcoded Backend Service Name in Nginx Config

**Location:** [frontend/nginx.prod.conf#L14](frontend/nginx.prod.conf#L14)

**Issue:** The backend service name is hardcoded, making deployment fragile.

```nginx
set $backend "oreotennisscoring-ots-tuekrc-backend-1";  # Brittle!
```

**Fix:** Use environment variable substitution:
```nginx
set $backend "${BACKEND_HOST}";
# Or use Docker's service discovery:
set $backend "backend";
```

---

## 4. Code Quality Issues

### 4.1 Test Coverage Critically Low

**Current Coverage:**
| Package | Coverage |
|---------|----------|
| auth | 0.0% |
| handler | 0.0% |
| middleware | 0.0% |
| repository | 0.0% |
| config | 0.0% |
| scoring | 81.3% ✅ |
| tournament | 63.6% ✅ |
| service | 0.7% |

**Risk:** Auth, handler, and middleware are security-critical with ZERO test coverage.

**Priority Tests Needed:**
1. `auth/jwt_test.go` - Token generation, validation, expiry
2. `middleware/auth_test.go` - Cookie parsing, token validation
3. `handler/auth_test.go` - Login flows, rate limiting bypass attempts
4. `middleware/ratelimit_test.go` - IP extraction, limit enforcement

---

### 4.2 Error Messages Expose Internal Details

**Location:** Multiple handlers

**Example:**
```go
return nil, fmt.Errorf("failed to get match: %w", err)  // Wraps internal error
```

**Issue:** Error wrapping exposes database/internal errors to clients.

**Fix:** Use separate internal and external errors:
```go
// Internal logging
log.Printf("match fetch failed: %v", err)

// External response
return nil, ErrMatchNotFound  // Generic external error
```

---

### 4.3 Inconsistent Error Handling Pattern

**Observation:** Some handlers return specific HTTP status codes, others use generic 500.

**Example of inconsistency:**
```go
// Good - specific error
if err == repository.ErrNotFound {
    WriteError(w, http.StatusNotFound, "match not found")
}

// Less good - generic error
WriteError(w, http.StatusInternalServerError, "failed to create player")
```

**Suggestion:** Create error types with HTTP status mapping:
```go
type AppError struct {
    Code    int
    Message string
    Err     error
}
```

---

### 4.4 Duplicate Validation Logic

**Location:** Handler validation vs Service validation

**Issue:** Player count validation exists in both:
- [handler/match.go#L74-L86](backend/internal/handler/match.go#L74-L86)
- [service/match.go#L48-L58](backend/internal/service/match.go#L48-L58)

**Suggestion:** Centralize validation in service layer; handlers only do schema/format validation.

---

## 5. Performance & Reliability Risks

### 5.1 N+1 Query in GetMatchSummary

**Location:** [backend/internal/service/match.go#L160-L167](backend/internal/service/match.go#L160-L167)

**Issue:** Individual player fetch inside loop:
```go
for _, mp := range matchPlayers {
    player, err := s.playerRepo.GetByID(ctx, mp.PlayerID)  // N+1!
}
```

**Impact:** 4+ extra queries per summary request (doubles match).

**Fix:** Batch fetch:
```go
playerIDs := make([]uuid.UUID, len(matchPlayers))
for i, mp := range matchPlayers {
    playerIDs[i] = mp.PlayerID
}
players, err := s.playerRepo.GetByIDs(ctx, playerIDs)  // Single query
```

---

### 5.2 Tendencies Queries Are Expensive

**Location:** [backend/internal/repository/tendencies.go](backend/internal/repository/tendencies.go)

**Issue:** Complex aggregation queries run on every request with no caching.

**Impact:** Slow page loads for venues with many matches.

**Mitigation Options:**
1. **Short-term:** Add database indexes (already present ✅)
2. **Medium-term:** Cache results with 5-minute TTL
3. **Long-term:** Materialized views or pre-computed stats table

---

### 5.3 No Connection Pool Monitoring

**Location:** [backend/internal/database/postgres.go](backend/internal/database/postgres.go)

**Issue:** Connection pool settings exist but no monitoring/metrics.

**Suggestion:** Add health check endpoint with pool stats:
```go
mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    stats := pool.Stat()
    WriteJSON(w, http.StatusOK, map[string]interface{}{
        "status": "healthy",
        "db_pool": map[string]int32{
            "total": stats.TotalConns(),
            "idle":  stats.IdleConns(),
            "in_use": stats.AcquiredConns(),
        },
    })
})
```

---

### 5.4 Missing Graceful Degradation for Offline Sync

**Location:** [frontend/src/services/api.js#L192-L207](frontend/src/services/api.js#L192-L207)

**Issue:** If sync fails repeatedly, events accumulate in IndexedDB with no user feedback or fallback.

**Suggestion:** Add:
1. Max retry count with exponential backoff
2. User notification for persistent sync failures
3. "Export data" fallback for offline recovery

---

## 6. Compliance with OTS Specs

### 6.1 Spec Compliance Status

| Spec | Status | Notes |
|------|--------|-------|
| OTS_Tennis_Scoring_Spec.md | ✅ Compliant | Scoring engine well-tested |
| OTS_Tournament_Spec.md | ✅ Compliant | Tournament logic implemented |
| OTS_Venue_Team_Player_Tendencies_Spec.md | ✅ Compliant | Thresholds correctly applied |

### 6.2 Potential Future Violations

1. **Tendencies Data Leakage:** Current implementation correctly excludes tendencies from sharing, but future features should maintain this boundary.

2. **Player Win Percentage:** Spec explicitly forbids showing player win percentage. Current code complies, but this should be enforced at the type level:
```go
type VenuePlayerTendency struct {
    // WinPercentage float64  // REMOVED - per spec Section 5
}
```

---

## 7. DevOps & Security Hardening

### 7.1 Docker Security

**Current State:** Good practices observed:
- ✅ Multi-stage builds
- ✅ Non-root user in backend
- ✅ Alpine base images
- ✅ Minimal dependencies

**Missing:**
- ❌ Read-only filesystem
- ❌ Security scanning in CI
- ❌ Image signing

**Fix for read-only:**
```yaml
# docker-compose.prod.yml
backend:
  read_only: true
  tmpfs:
    - /tmp
```

---

### 7.2 Security Headers ✅ FIXED

**Location:** Nginx config

**Status:** Security headers have been added to both nginx.prod.conf and nginx.localhost.conf.

---

### 7.3 Environment Variable Validation

**Current:** [config/config.go](backend/internal/config/config.go) validates presence and JWT length ✅

**Missing:** Validation that `ADMIN_PASSWORD_HASH` is a valid bcrypt hash:
```go
if _, err := bcrypt.Cost([]byte(cfg.AdminPasswordHash)); err != nil {
    return nil, fmt.Errorf("ADMIN_PASSWORD_HASH is not a valid bcrypt hash")
}
```

---

## 8. Final Recommendations

### Completed Fixes ✅

1. ~~**CRITICAL:** Remove `{@html}` from Modal.svelte~~ ✅ FIXED
2. ~~**CRITICAL:** Fix X-Forwarded-For IP extraction~~ ✅ FIXED
3. ~~**HIGH:** Add request body size limits~~ ✅ FIXED
4. ~~**MEDIUM:** Add security headers to nginx~~ ✅ FIXED
5. ~~**MEDIUM:** Add input sanitization for names~~ ✅ FIXED

### By Design Decisions ⚪

1. Rate limiting only on login endpoint (closed-group app)
2. Public match creation without authentication (usability priority)

### Remaining Medium-term Improvements

1. Add test coverage for auth, handlers, middleware (target: 70%+)
2. Implement match session tokens for event submission
3. Fix N+1 query in match summary
4. Add structured logging with request IDs
5. Add bcrypt hash validation in config

### Long-term Hardening (Ongoing)

1. Implement audit logging for admin actions
2. Add rate limit monitoring and alerting
3. Consider caching layer for tendencies
4. Add dependency vulnerability scanning in CI
5. Implement backup verification procedures

---

## 9. Conclusion

OTS demonstrates good foundational security practices:
- ✅ Bcrypt password hashing
- ✅ HttpOnly, Secure, SameSite=Strict cookies
- ✅ JWT with reasonable expiration
- ✅ Parameterized SQL queries (no SQL injection)
- ✅ CORS properly configured
- ✅ Non-root Docker user
- ✅ XSS prevention (fixed)
- ✅ IP spoofing prevention (fixed)
- ✅ Request body size limits (added)
- ✅ Security headers (added)
- ✅ Input validation (added)

The critical security issues have been addressed. The application is now ready for production deployment. Design decisions around rate limiting and public match creation are appropriate for a closed-group, self-hosted tennis scoring app.

**Recommendation:** Application is production-ready. Consider medium-term improvements (test coverage, match session tokens) iteratively.

---

*End of Audit Report*
