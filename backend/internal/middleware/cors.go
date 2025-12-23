package middleware

import (
	"github.com/rs/cors"
)

// NewCORS creates a CORS middleware with the specified origins.
func NewCORS(allowedOrigins []string) *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true, // Required for cookies
		MaxAge:           300,  // 5 minutes
	})
}
