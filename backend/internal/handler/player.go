package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/saurabh22suman/oreo-tennis-scoring/backend/internal/model"
	"github.com/saurabh22suman/oreo-tennis-scoring/backend/internal/repository"
)

// PlayerHandler handles player endpoints.
type PlayerHandler struct {
	repo *repository.PlayerRepository
}

// NewPlayerHandler creates a new player handler.
func NewPlayerHandler(repo *repository.PlayerRepository) *PlayerHandler {
	return &PlayerHandler{repo: repo}
}

// CreatePlayerRequest represents a create player request.
type CreatePlayerRequest struct {
	Name string `json:"name"`
}

// UpdatePlayerRequest represents an update player request.
type UpdatePlayerRequest struct {
	Name   *string `json:"name,omitempty"`
	Active *bool   `json:"active,omitempty"`
}

// List returns all players (admin: all, public: active only).
func (h *PlayerHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Check if this is an admin request (based on path)
	activeOnly := !strings.HasPrefix(r.URL.Path, "/api/admin")

	players, err := h.repo.List(r.Context(), activeOnly)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list players")
		return
	}

	WriteJSON(w, http.StatusOK, players)
}

// Create adds a new player.
func (h *PlayerHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req CreatePlayerRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		WriteError(w, http.StatusBadRequest, "name is required")
		return
	}

	// Validate and sanitize name
	name, valid := ValidateNameWithLength(req.Name, 100)
	if !valid {
		WriteError(w, http.StatusBadRequest, "name contains invalid characters or exceeds 100 characters")
		return
	}

	player := &model.Player{
		Name:   name,
		Active: true,
	}

	if err := h.repo.Create(r.Context(), player); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to create player")
		return
	}

	WriteJSON(w, http.StatusCreated, player)
}

// Update modifies an existing player.
func (h *PlayerHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Extract player ID from path
	id := extractIDFromPath(r.URL.Path)
	if id == uuid.Nil {
		WriteError(w, http.StatusBadRequest, "invalid player id")
		return
	}

	var req UpdatePlayerRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Get existing player
	player, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrNotFound {
			WriteError(w, http.StatusNotFound, "player not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "failed to get player")
		return
	}

	// Apply updates
	if req.Name != nil {
		name, valid := ValidateNameWithLength(*req.Name, 100)
		if !valid {
			WriteError(w, http.StatusBadRequest, "name contains invalid characters or exceeds 100 characters")
			return
		}
		player.Name = name
	}
	if req.Active != nil {
		player.Active = *req.Active
	}

	if err := h.repo.Update(r.Context(), player); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to update player")
		return
	}

	WriteJSON(w, http.StatusOK, player)
}

// GetByID retrieves a player by ID.
func (h *PlayerHandler) GetByID(ctx context.Context, id uuid.UUID) (*model.Player, error) {
	return h.repo.GetByID(ctx, id)
}

// extractIDFromPath extracts UUID from URL path like /api/admin/players/:id
func extractIDFromPath(path string) uuid.UUID {
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return uuid.Nil
	}
	idStr := parts[len(parts)-1]
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil
	}
	return id
}
