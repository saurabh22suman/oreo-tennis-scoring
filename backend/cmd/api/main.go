package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/saurabh22suman/oreo-tennis-scoring/backend/internal/auth"
	"github.com/saurabh22suman/oreo-tennis-scoring/backend/internal/config"
	"github.com/saurabh22suman/oreo-tennis-scoring/backend/internal/database"
	"github.com/saurabh22suman/oreo-tennis-scoring/backend/internal/handler"
	"github.com/saurabh22suman/oreo-tennis-scoring/backend/internal/middleware"
	"github.com/saurabh22suman/oreo-tennis-scoring/backend/internal/repository"
	"github.com/saurabh22suman/oreo-tennis-scoring/backend/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	ctx := context.Background()
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Run migrations
	if err := database.RunMigrations(ctx, pool); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrations completed")

	// Initialize repositories
	playerRepo := repository.NewPlayerRepository(pool)
	venueRepo := repository.NewVenueRepository(pool)
	matchRepo := repository.NewMatchRepository(pool)
	tendenciesRepo := repository.NewTendenciesRepository(pool)

	// Initialize services
	matchSvc := service.NewMatchService(matchRepo, playerRepo, venueRepo)
	tendenciesSvc := service.NewTendenciesService(tendenciesRepo, venueRepo)

	// Initialize auth
	jwtService := auth.NewJWTService(cfg.JWTSecret)

	// Determine if running in production (HTTPS)
	isSecure := strings.HasPrefix(cfg.FrontendURL, "https://")

	// Initialize handlers
	authHandler := handler.NewAuthHandler(jwtService, cfg.AdminUsername, cfg.AdminPasswordHash, isSecure)
	playerHandler := handler.NewPlayerHandler(playerRepo)
	venueHandler := handler.NewVenueHandler(venueRepo)
	matchHandler := handler.NewMatchHandler(matchSvc, matchRepo)
	tendenciesHandler := handler.NewTendenciesHandler(tendenciesSvc)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService)
	corsMiddleware := middleware.NewCORS(cfg.GetAllowedOrigins())
	loginRateLimiter := middleware.NewRateLimiter(0.5, 5) // 1 request per 2 seconds, burst of 5

	// Setup router
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		handler.WriteJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
	})

	// Auth routes (rate limited)
	mux.Handle("/api/admin/login", loginRateLimiter.Limit(http.HandlerFunc(authHandler.Login)))
	mux.HandleFunc("/api/admin/logout", authHandler.Logout)

	// Protected admin routes
	mux.Handle("/api/admin/check", authMiddleware.RequireAuth(http.HandlerFunc(authHandler.CheckAuth)))
	mux.Handle("/api/admin/players", authMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			playerHandler.List(w, r)
		case http.MethodPost:
			playerHandler.Create(w, r)
		default:
			handler.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})))
	mux.Handle("/api/admin/players/", authMiddleware.RequireAuth(http.HandlerFunc(playerHandler.Update)))
	mux.Handle("/api/admin/venues", authMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			venueHandler.List(w, r)
		case http.MethodPost:
			venueHandler.Create(w, r)
		default:
			handler.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})))
	mux.Handle("/api/admin/venues/", authMiddleware.RequireAuth(http.HandlerFunc(venueHandler.Update)))
	mux.Handle("/api/admin/matches/", authMiddleware.RequireAuth(http.HandlerFunc(matchHandler.Delete)))
	mux.Handle("/api/admin/matches", authMiddleware.RequireAuth(http.HandlerFunc(matchHandler.List)))

	// Public routes
	mux.HandleFunc("/api/players", playerHandler.List)
	mux.HandleFunc("/api/venues", venueHandler.List)
	mux.HandleFunc("/api/matches", matchHandler.Create)

	// Venue-specific routes need path parsing
	mux.HandleFunc("/api/venues/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case strings.HasSuffix(path, "/tendencies"):
			tendenciesHandler.GetVenueTendencies(w, r)
		default:
			handler.WriteError(w, http.StatusNotFound, "not found")
		}
	})

	// Match-specific routes need path parsing
	mux.HandleFunc("/api/matches/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case strings.HasSuffix(path, "/events"):
			matchHandler.AddEvents(w, r)
		case strings.HasSuffix(path, "/complete"):
			matchHandler.Complete(w, r)
		case strings.HasSuffix(path, "/summary"):
			matchHandler.Summary(w, r)
		default:
			handler.WriteError(w, http.StatusNotFound, "not found")
		}
	})

	// Apply CORS middleware
	corsHandler := corsMiddleware.Handler(mux)

	// Apply body size limit middleware (1MB max)
	limitedHandler := middleware.LimitBody(middleware.MaxBodySize)(corsHandler)

	// Create server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      limitedHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %d", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
