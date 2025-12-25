package handler

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/saurabh22suman/oreo-tennis-scoring/backend/internal/service"
)

// TendenciesHandler handles venue tendencies endpoints.
type TendenciesHandler struct {
	svc *service.TendenciesService
}

// NewTendenciesHandler creates a new tendencies handler.
func NewTendenciesHandler(svc *service.TendenciesService) *TendenciesHandler {
	return &TendenciesHandler{svc: svc}
}

// GetVenueTendencies handles GET /api/venues/:id/tendencies
// Returns team and player tendencies for a specific venue.
// Per OTS_Venue_Team_Player_Tendencies_Spec.md:
// - Teams: Doubles only, minimum 3 matches at venue
// - Players: Minimum 5 matches at venue
// - No rankings, neutral ordering
func (h *TendenciesHandler) GetVenueTendencies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Extract venue ID from path: /api/venues/:id/tendencies
	venueID := extractVenueIDFromPath(r.URL.Path)
	if venueID == uuid.Nil {
		WriteError(w, http.StatusBadRequest, "invalid venue id")
		return
	}

	tendencies, err := h.svc.GetVenueTendencies(r.Context(), venueID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteError(w, http.StatusNotFound, "venue not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "failed to get tendencies")
		return
	}

	WriteJSON(w, http.StatusOK, tendencies)
}

// extractVenueIDFromPath extracts the venue UUID from a path like /api/venues/:id/tendencies
func extractVenueIDFromPath(path string) uuid.UUID {
	// Path format: /api/venues/{uuid}/tendencies
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 {
		return uuid.Nil
	}

	// Find "venues" and get the next part
	for i, part := range parts {
		if part == "venues" && i+1 < len(parts) {
			id, err := uuid.Parse(parts[i+1])
			if err != nil {
				return uuid.Nil
			}
			return id
		}
	}
	return uuid.Nil
}
