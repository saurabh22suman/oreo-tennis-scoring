package handler

import (
	"net/http"
	"time"

	"github.com/saurabh22suman/oreo-tennis-scoring/backend/internal/auth"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	jwtService   *auth.JWTService
	adminUser    string
	adminPwdHash string
	isSecure     bool
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(jwtService *auth.JWTService, adminUser, adminPwdHash string, isSecure bool) *AuthHandler {
	return &AuthHandler{
		jwtService:   jwtService,
		adminUser:    adminUser,
		adminPwdHash: adminPwdHash,
		isSecure:     isSecure,
	}
}

// LoginRequest represents a login request.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login handles admin authentication.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req LoginRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		WriteError(w, http.StatusBadRequest, "username and password required")
		return
	}

	// Check username
	if req.Username != h.adminUser {
		WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Verify password against bcrypt hash
	if err := auth.VerifyPassword(req.Password, h.adminPwdHash); err != nil {
		WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Generate JWT token
	token, expiresAt, err := h.jwtService.GenerateToken(req.Username)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	// Set HttpOnly, Secure cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   h.isSecure,
		SameSite: http.SameSiteStrictMode,
	})

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "login successful",
		"expires_at": expiresAt.Format(time.RFC3339),
	})
}

// Logout clears the auth cookie.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.isSecure,
		SameSite: http.SameSiteStrictMode,
	})

	WriteJSON(w, http.StatusOK, map[string]string{
		"message": "logout successful",
	})
}

// CheckAuth returns the current auth status.
func (h *AuthHandler) CheckAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]bool{
		"authenticated": true,
	})
}
