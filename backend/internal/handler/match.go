package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/saurabh22suman/oreo-tennis-scoring/backend/internal/model"
	"github.com/saurabh22suman/oreo-tennis-scoring/backend/internal/repository"
	"github.com/saurabh22suman/oreo-tennis-scoring/backend/internal/service"
)

// MatchHandler handles match endpoints.
type MatchHandler struct {
	svc       *service.MatchService
	matchRepo *repository.MatchRepository
}

// NewMatchHandler creates a new match handler.
func NewMatchHandler(svc *service.MatchService, matchRepo *repository.MatchRepository) *MatchHandler {
	return &MatchHandler{svc: svc, matchRepo: matchRepo}
}

// EventRequest represents a point event in a request.
type EventRequest struct {
	ID              uuid.UUID       `json:"id"`
	Timestamp       string          `json:"timestamp"`
	ServerPlayerID  uuid.UUID       `json:"server_player_id"`
	ServeType       model.ServeType `json:"serve_type"`
	PointWinnerTeam model.Team      `json:"point_winner_team"`
}

// EventsRequest represents a batch of events.
type EventsRequest struct {
	Events []EventRequest `json:"events"`
}

// validServeTypes is a set of valid serve types.
var validServeTypes = map[model.ServeType]bool{
	model.ServeTypeFirst:       true,
	model.ServeTypeSecond:      true,
	model.ServeTypeDoubleFault: true,
}

// validTeams is a set of valid teams.
var validTeams = map[model.Team]bool{
	model.TeamA: true,
	model.TeamB: true,
}

// validMatchTypes is a set of valid match types.
var validMatchTypes = map[model.MatchType]bool{
	model.MatchTypeSingles:           true,
	model.MatchTypeDoubles:           true,
	model.MatchTypeAustralianDoubles: true,
}

// Create starts a new match.
func (h *MatchHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req service.CreateMatchRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate inputs
	if req.VenueID == uuid.Nil {
		WriteError(w, http.StatusBadRequest, "venue_id is required")
		return
	}

	if !validMatchTypes[req.MatchType] {
		WriteError(w, http.StatusBadRequest, "match_type must be singles, doubles, or 1v2")
		return
	}

	if len(req.TeamA) == 0 {
		WriteError(w, http.StatusBadRequest, "team_a is required")
		return
	}

	if len(req.TeamB) == 0 {
		WriteError(w, http.StatusBadRequest, "team_b is required")
		return
	}

	match, err := h.svc.CreateMatch(r.Context(), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, match)
}

// AddEvents handles batch event submission (idempotent).
func (h *MatchHandler) AddEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Extract match ID from path: /api/matches/:id/events
	matchID := extractMatchIDFromPath(r.URL.Path)
	if matchID == uuid.Nil {
		WriteError(w, http.StatusBadRequest, "invalid match id")
		return
	}

	var req EventsRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Events) == 0 {
		WriteError(w, http.StatusBadRequest, "events array is required")
		return
	}

	if len(req.Events) > 1000 {
		WriteError(w, http.StatusBadRequest, "maximum 1000 events per request")
		return
	}

	// Convert and validate events
	events := make([]model.PointEvent, len(req.Events))
	for i, e := range req.Events {
		if e.ID == uuid.Nil {
			WriteError(w, http.StatusBadRequest, "event id is required")
			return
		}

		if e.ServerPlayerID == uuid.Nil {
			WriteError(w, http.StatusBadRequest, "server_player_id is required")
			return
		}

		if !validServeTypes[e.ServeType] {
			WriteError(w, http.StatusBadRequest, "serve_type must be first, second, or double_fault")
			return
		}

		if !validTeams[e.PointWinnerTeam] {
			WriteError(w, http.StatusBadRequest, "point_winner_team must be A or B")
			return
		}

		// Parse timestamp
		timestamp, err := parseTimestamp(e.Timestamp)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "invalid timestamp format")
			return
		}

		events[i] = model.PointEvent{
			ID:              e.ID,
			MatchID:         matchID,
			Timestamp:       timestamp,
			ServerPlayerID:  e.ServerPlayerID,
			ServeType:       e.ServeType,
			PointWinnerTeam: e.PointWinnerTeam,
		}
	}

	inserted, err := h.svc.AddEvents(r.Context(), matchID, events)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"inserted": inserted,
		"total":    len(events),
	})
}

// Complete marks a match as finished.
func (h *MatchHandler) Complete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Extract match ID from path: /api/matches/:id/complete
	matchID := extractMatchIDFromPath(r.URL.Path)
	if matchID == uuid.Nil {
		WriteError(w, http.StatusBadRequest, "invalid match id")
		return
	}

	if err := h.svc.CompleteMatch(r.Context(), matchID); err != nil {
		if err == repository.ErrNotFound {
			WriteError(w, http.StatusNotFound, "match not found or already completed")
			return
		}
		WriteError(w, http.StatusInternalServerError, "failed to complete match")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"message": "match completed",
	})
}

// Summary returns match statistics.
func (h *MatchHandler) Summary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Extract match ID from path: /api/matches/:id/summary
	matchID := extractMatchIDFromPath(r.URL.Path)
	if matchID == uuid.Nil {
		WriteError(w, http.StatusBadRequest, "invalid match id")
		return
	}

	summary, err := h.svc.GetMatchSummary(r.Context(), matchID)
	if err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}

	WriteJSON(w, http.StatusOK, summary)
}

// Delete removes a match (admin only).
func (h *MatchHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	matchID := extractIDFromPath(r.URL.Path)
	if matchID == uuid.Nil {
		WriteError(w, http.StatusBadRequest, "invalid match id")
		return
	}

	if err := h.svc.DeleteMatch(r.Context(), matchID); err != nil {
		if err == repository.ErrNotFound {
			WriteError(w, http.StatusNotFound, "match not found")
			return
		}
		WriteError(w, http.StatusInternalServerError, "failed to delete match")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"message": "match deleted",
	})
}

// List returns recent matches.
func (h *MatchHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	matches, err := h.matchRepo.List(r.Context(), 50)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list matches")
		return
	}

	WriteJSON(w, http.StatusOK, matches)
}

// extractMatchIDFromPath extracts UUID from paths like /api/matches/:id/events
func extractMatchIDFromPath(path string) uuid.UUID {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "matches" && i+1 < len(parts) {
			id, err := uuid.Parse(parts[i+1])
			if err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}

// parseTimestamp parses RFC3339 or ISO8601 timestamps.
func parseTimestamp(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
