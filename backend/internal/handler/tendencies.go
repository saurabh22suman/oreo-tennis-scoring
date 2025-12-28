package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

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

	// Parse date filter parameters
	dateFilter := parseDateFilter(r)

	tendencies, err := h.svc.GetVenueTendencies(r.Context(), venueID, dateFilter)
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

// parseDateFilter extracts date filtering parameters from the request.
// Supports:
// - period=day (today only)
// - period=week (last 7 days)
// - period=month&month=1&year=2025 (specific month)
// - period=all (no date filter, default)
func parseDateFilter(r *http.Request) service.DateFilter {
	query := r.URL.Query()
	period := query.Get("period")

	now := time.Now()

	switch period {
	case "day":
		// Today: start of day to end of day
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endOfDay := startOfDay.Add(24 * time.Hour)
		return service.DateFilter{
			Enabled:   true,
			StartDate: startOfDay,
			EndDate:   endOfDay,
		}

	case "week":
		// Last 7 days
		endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
		startOfWeek := endOfDay.AddDate(0, 0, -6)
		startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, now.Location())
		return service.DateFilter{
			Enabled:   true,
			StartDate: startOfWeek,
			EndDate:   endOfDay.Add(time.Second),
		}

	case "month":
		// Specific month (month=1-12, year=YYYY)
		monthStr := query.Get("month")
		yearStr := query.Get("year")

		month, err := strconv.Atoi(monthStr)
		if err != nil || month < 1 || month > 12 {
			month = int(now.Month())
		}

		year, err := strconv.Atoi(yearStr)
		if err != nil || year < 2020 || year > 2100 {
			year = now.Year()
		}

		startOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, now.Location())
		endOfMonth := startOfMonth.AddDate(0, 1, 0)

		return service.DateFilter{
			Enabled:   true,
			StartDate: startOfMonth,
			EndDate:   endOfMonth,
		}

	default:
		// "all" or unspecified - no date filter
		return service.DateFilter{Enabled: false}
	}
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
