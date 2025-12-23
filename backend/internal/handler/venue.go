package handler

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/saurabh22suman/oreo-tennis-scoring/backend/internal/model"
	"github.com/saurabh22suman/oreo-tennis-scoring/backend/internal/repository"
)

// VenueHandler handles venue endpoints.
type VenueHandler struct {
	repo *repository.VenueRepository
}

// NewVenueHandler creates a new venue handler.
func NewVenueHandler(repo *repository.VenueRepository) *VenueHandler {
	return &VenueHandler{repo: repo}
}

// CreateVenueRequest represents a create venue request.
type CreateVenueRequest struct {
	Name    string        `json:"name"`
	Surface model.Surface `json:"surface"`
}

// UpdateVenueRequest represents an update venue request.
type UpdateVenueRequest struct {
	Name    *string        `json:"name,omitempty"`
	Surface *model.Surface `json:"surface,omitempty"`
	Active  *bool          `json:"active,omitempty"`
}

// validSurfaces is a set of valid surface types.
var validSurfaces = map[model.Surface]bool{
	model.SurfaceHard:  true,
	model.SurfaceClay:  true,
	model.SurfaceGrass: true,
}

// List returns all venues (admin: all, public: active only).
func (h *VenueHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	activeOnly := !strings.HasPrefix(r.URL.Path, "/api/admin")

	venues, err := h.repo.List(r.Context(), activeOnly)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list venues")
		return
	}

	WriteJSON(w, http.StatusOK, venues)
}

// Create adds a new venue.
func (h *VenueHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req CreateVenueRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		WriteError(w, http.StatusBadRequest, "name is required")
		return
	}

	if len(req.Name) > 100 {
		WriteError(w, http.StatusBadRequest, "name must be 100 characters or less")
		return
	}

	if !validSurfaces[req.Surface] {
		WriteError(w, http.StatusBadRequest, "surface must be hard, clay, or grass")
		return
	}

	venue := &model.Venue{
		Name:    req.Name,
		Surface: req.Surface,
		Active:  true,
	}

	if err := h.repo.Create(r.Context(), venue); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to create venue")
		return
	}

	WriteJSON(w, http.StatusCreated, venue)
}

// Update modifies an existing venue.
func (h *VenueHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id := extractIDFromPath(r.URL.Path)
	if id == uuid.Nil {
		WriteError(w, http.StatusBadRequest, "invalid venue id")
		return
	}

	var req UpdateVenueRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	venue, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrNotFound {
			WriteError(w, http.StatusNotFound, "venue not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "failed to get venue")
		return
	}

	if req.Name != nil {
		if *req.Name == "" {
			WriteError(w, http.StatusBadRequest, "name cannot be empty")
			return
		}
		if len(*req.Name) > 100 {
			WriteError(w, http.StatusBadRequest, "name must be 100 characters or less")
			return
		}
		venue.Name = *req.Name
	}
	if req.Surface != nil {
		if !validSurfaces[*req.Surface] {
			WriteError(w, http.StatusBadRequest, "surface must be hard, clay, or grass")
			return
		}
		venue.Surface = *req.Surface
	}
	if req.Active != nil {
		venue.Active = *req.Active
	}

	if err := h.repo.Update(r.Context(), venue); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to update venue")
		return
	}

	WriteJSON(w, http.StatusOK, venue)
}
