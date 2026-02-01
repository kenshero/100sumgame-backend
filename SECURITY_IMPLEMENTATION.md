# Security Implementation - Anti-Spam & Rate Limiting

## Overview
This document describes the security improvements implemented to prevent spam attacks and protect against UUID manipulation in the Sum-100 Game application.

## Security Features Implemented

### 1. Rate Limiting (Multiple Layers)

#### A. Global Rate Limiting (Middleware Level)
- **Location**: `backend/internal/middleware/rate_limit.go`
- **Implementation**: 100 requests per minute per IP address
- **Purpose**: Protects the server from DDoS-like attacks and excessive requests
- **Applied to**: All HTTP requests

#### B. Operation-Specific Rate Limiting (Service Level)
- **Location**: `backend/internal/service/rate_limiter.go` and `backend/internal/service/game_service.go`
- **Limits**:
  - CreateGame: 10 games per minute per guest
  - FillCells: 30 operations per minute per guest
  - VerifyGame: 10 verifications per minute per guest
- **Purpose**: Prevents spamming specific game operations before database writes
- **Applied to**: Individual GraphQL mutations

### 2. Cookie-Based Authentication with Session Management

#### A. Session Manager
- **Location**: `backend/internal/middleware/session.go`
- **Features**:
  - Server-side session generation and validation
  - HMAC-SHA256 signed cookies for tamper protection
  - 24-hour session expiration
  - Automatic cleanup of expired sessions
  - IP and UserAgent tracking (optional validation)

#### B. Authentication Middleware
- **Location**: `backend/internal/middleware/auth.go`
- **Features**:
  - Automatic session creation on first request
  - Session validation on every request
  - Automatic session refresh on valid requests
  - Session injection into request context
  - Automatic session recreation on invalid/expired sessions

#### C. Cookie Security Settings
```go
HttpOnly: true              // Prevents XSS attacks
Secure: true (production)   // Only sent over HTTPS
SameSite: StrictMode        // Prevents CSRF attacks
MaxAge: 86400              // 24 hours
Path: "/"                   // Available across all routes
```

### 3. Server-Side Session Validation

#### Changes in GraphQL Resolvers
- **Location**: `backend/internal/graphql/resolver/schema.resolver.go`
- **Modified Resolvers**:
  - `CreateGame`: Now uses session guest ID instead of client-provided guest ID
  - `CompleteGame`: Uses session guest ID for leaderboard submission
- **Benefit**: Clients can no longer manipulate guest IDs

### 4. Rate Limiting Before Database Operations

All critical game operations now check rate limits BEFORE database writes:
- `CreateGame`: Checks rate limit before inserting new game
- `FillCells`: Checks rate limit before updating game state
- `VerifyGame`: Checks rate limit before verification

This prevents database pollution from spam attacks.

## Architecture

```
HTTP Request
    ↓
[Rate Limit Middleware] - Global: 100 req/min/IP
    ↓
[Auth Middleware] - Session validation & creation
    ↓
[GraphQL Handler]
    ↓
[Resolver]
    ↓
[Service Layer] - Operation-specific rate limits
    ↓
[Database]
```

## Configuration

### Environment Variables

Add these to your `.env` file:

```bash
# Session secret key (MUST be changed in production)
SESSION_SECRET_KEY=your-super-secret-key-at-least-32-characters-long

# Environment mode (determines cookie security)
ENVIRONMENT=production  # Set to "production" for HTTPS
```

### Rate Limit Configuration

You can adjust rate limits in the following files:

#### Global Rate Limit (main.go)
```go
rateLimiter := middleware.NewRateLimiter(100, 1*time.Minute)
```

#### Operation Rate Limits (game_service.go)
```go
createGameLimiter:  NewOperationRateLimiter(10, 1*time.Minute),
fillCellsLimiter:   NewOperationRateLimiter(30, 1*time.Minute),
verifyGameLimiter:  NewOperationRateLimiter(10, 1*time.Minute),
```

## Security Benefits

### 1. Prevents Spam Attacks
- Rate limiting prevents users from creating excessive requests
- Multiple layers of rate limiting provide defense in depth

### 2. Prevents UUID Manipulation
- Guest IDs are now generated server-side
- Signed cookies prevent tampering
- Clients can no longer change their guest ID

### 3. Database Protection
- Rate limits are checked before database writes
- Prevents database pollution from spam
- Reduces unnecessary database load

### 4. Protection Against Common Attacks
- **XSS**: HttpOnly cookies prevent JavaScript access
- **CSRF**: SameSite cookies prevent cross-site attacks
- **DDoS**: Global rate limiting prevents resource exhaustion
- **Brute Force**: Operation-specific limits prevent automated abuse

## Migration from localStorage to Cookies

### Frontend Changes Required

The frontend no longer needs to:
- Generate or store guest IDs in localStorage
- Send guest IDs in GraphQL mutations

The frontend should:
- Let the backend handle session creation automatically
- Include credentials in requests (most HTTP clients do this automatically with cookies)

### Example GraphQL Mutation Changes

#### Before (Client-Side Guest ID)
```graphql
mutation {
  createGame(guestID: "client-generated-uuid") {
    id
    guestId
  }
}
```

#### After (Server-Side Session)
```graphql
mutation {
  createGame(guestID: "ignored-by-server") {
    id
    guestId
  }
}
```

The `guestID` parameter is now ignored; the server uses the session's guest ID.

## Monitoring and Logging

### Recommended Monitoring

1. **Rate Limit Exceeded Events**: Monitor how often rate limits are hit
2. **Session Creation Rate**: Monitor for unusual session creation patterns
3. **Failed Authentication Attempts**: Monitor for invalid session attempts
4. **IP Blocking**: Consider implementing automatic IP blocking for repeated violations

### Adding Logging

You can add logging in the rate limiting middleware:

```go
if !limiter.Allow(ip) {
    log.Printf("Rate limit exceeded for IP: %s", ip)
    // ... return error
}
```

## Testing

### Testing Rate Limiting

```bash
# Test global rate limit (should fail after 100 requests)
for i in {1..150}; do
  curl -X POST http://localhost:8080/graphql \
    -H "Content-Type: application/json" \
    -d '{"query":"{ __typename }"}'
done
```

### Testing Session Management

```bash
# First request should create a session
curl -c cookies.txt http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"{ __typename }"}'

# Subsequent requests should use the session
curl -b cookies.txt http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"{ __typename }"}'
```

## Production Checklist

- [ ] Set a strong `SESSION_SECRET_KEY` in production environment
- [ ] Set `ENVIRONMENT=production` to enable secure cookies
- [ ] Use HTTPS (required for secure cookies)
- [ ] Configure appropriate rate limits based on your traffic
- [ ] Set up monitoring for rate limit violations
- [ ] Consider implementing IP blocking for repeated violations
- [ ] Review and adjust rate limits based on usage patterns
- [ ] Consider using Redis for distributed rate limiting (if using multiple servers)

## Future Enhancements

### 1. CAPTCHA
- Add CAPTCHA for sensitive operations (e.g., score submission)
- Recommended: Google reCAPTCHA v3 (invisible)

### 2. IP-Based Restrictions
- Limit number of sessions per IP per hour
- Implement temporary IP blocking for repeated violations

### 3. Distributed Rate Limiting
- Use Redis for rate limiting across multiple server instances
- Share session data across servers

### 4. Advanced Fingerprinting
- Browser fingerprinting to detect multiple accounts
- Device fingerprinting for additional security

### 5. Anomaly Detection
- Machine learning to detect suspicious patterns
- Automatic response to detected attacks

## Support

For questions or issues related to security implementation:
1. Check this documentation
2. Review the implementation files
3. Consult the code comments
4. Test in development environment before production deployment