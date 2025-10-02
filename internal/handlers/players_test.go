package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/francisco/gridironmind/internal/db"
	"github.com/francisco/gridironmind/internal/models"
	"github.com/google/uuid"
)

// Mock PlayerQueries for testing
type mockPlayerQueries struct {
	listPlayersFunc    func(ctx context.Context, filters db.PlayerFilters) ([]*models.Player, int, error)
	getPlayerByIDFunc  func(ctx context.Context, id uuid.UUID) (*models.Player, error)
}

func (m *mockPlayerQueries) ListPlayers(ctx context.Context, filters db.PlayerFilters) ([]*models.Player, int, error) {
	if m.listPlayersFunc != nil {
		return m.listPlayersFunc(ctx, filters)
	}
	return []*models.Player{}, 0, nil
}

func (m *mockPlayerQueries) GetPlayerByID(ctx context.Context, id uuid.UUID) (*models.Player, error) {
	if m.getPlayerByIDFunc != nil {
		return m.getPlayerByIDFunc(ctx, id)
	}
	return nil, nil
}

func TestHandlePlayers_MethodNotAllowed(t *testing.T) {
	handler := NewPlayersHandler()
	handler.autoFetchEnabled = false // Disable auto-fetch for testing

	req := httptest.NewRequest(http.MethodPost, "/api/v1/players", nil)
	w := httptest.NewRecorder()

	handler.HandlePlayers(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("HandlePlayers() status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["error"] == nil {
		t.Error("HandlePlayers() should return error for non-GET method")
	}
}

func TestListPlayers_Success(t *testing.T) {
	handler := NewPlayersHandler()
	handler.autoFetchEnabled = false

	playerID := uuid.New()
	teamID := uuid.New()

	// Mock successful response
	handler.queries = &mockPlayerQueries{
		listPlayersFunc: func(ctx context.Context, filters db.PlayerFilters) ([]*models.Player, int, error) {
			players := []*models.Player{
				{
					ID:       playerID,
					Name:     "Patrick Mahomes",
					Position: "QB",
					TeamID:   &teamID,
				},
			}
			return players, 1, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/players", nil)
	w := httptest.NewRecorder()

	handler.HandlePlayers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("listPlayers() status = %d, want %d", w.Code, http.StatusOK)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["data"] == nil {
		t.Error("listPlayers() should return data field")
	}

	if response["meta"] == nil {
		t.Error("listPlayers() should return meta field with pagination")
	}

	meta := response["meta"].(map[string]interface{})
	if meta["total"] != float64(1) {
		t.Errorf("listPlayers() total = %v, want 1", meta["total"])
	}
}

func TestListPlayers_WithFilters(t *testing.T) {
	handler := NewPlayersHandler()
	handler.autoFetchEnabled = false

	tests := []struct {
		name          string
		queryParams   string
		expectedLimit int
		expectedPos   string
	}{
		{
			name:          "Position filter QB",
			queryParams:   "?position=QB",
			expectedLimit: 50,
			expectedPos:   "QB",
		},
		{
			name:          "Custom limit",
			queryParams:   "?limit=25",
			expectedLimit: 25,
			expectedPos:   "",
		},
		{
			name:          "Position and limit",
			queryParams:   "?position=WR&limit=10",
			expectedLimit: 10,
			expectedPos:   "WR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler.queries = &mockPlayerQueries{
				listPlayersFunc: func(ctx context.Context, filters db.PlayerFilters) ([]*models.Player, int, error) {
					if filters.Limit != tt.expectedLimit {
						t.Errorf("Expected limit %d, got %d", tt.expectedLimit, filters.Limit)
					}
					if tt.expectedPos != "" && filters.Position != tt.expectedPos {
						t.Errorf("Expected position %s, got %s", tt.expectedPos, filters.Position)
					}
					return []*models.Player{}, 0, nil
				},
			}

			req := httptest.NewRequest(http.MethodGet, "/api/v1/players"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.HandlePlayers(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("listPlayers() status = %d, want %d", w.Code, http.StatusOK)
			}
		})
	}
}

func TestListPlayers_InvalidPosition(t *testing.T) {
	handler := NewPlayersHandler()
	handler.autoFetchEnabled = false

	req := httptest.NewRequest(http.MethodGet, "/api/v1/players?position=INVALID", nil)
	w := httptest.NewRecorder()

	handler.HandlePlayers(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("listPlayers() status = %d, want %d for invalid position", w.Code, http.StatusBadRequest)
	}
}

func TestListPlayers_InvalidStatus(t *testing.T) {
	handler := NewPlayersHandler()
	handler.autoFetchEnabled = false

	req := httptest.NewRequest(http.MethodGet, "/api/v1/players?status=invalid", nil)
	w := httptest.NewRecorder()

	handler.HandlePlayers(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("listPlayers() status = %d, want %d for invalid status", w.Code, http.StatusBadRequest)
	}
}

func TestListPlayers_InvalidTeamID(t *testing.T) {
	handler := NewPlayersHandler()
	handler.autoFetchEnabled = false

	req := httptest.NewRequest(http.MethodGet, "/api/v1/players?team=not-a-uuid", nil)
	w := httptest.NewRecorder()

	handler.HandlePlayers(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("listPlayers() status = %d, want %d for invalid team ID", w.Code, http.StatusBadRequest)
	}
}

func TestGetPlayer_Success(t *testing.T) {
	handler := NewPlayersHandler()
	handler.autoFetchEnabled = false

	playerID := uuid.New()
	teamID := uuid.New()

	handler.queries = &mockPlayerQueries{
		getPlayerByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Player, error) {
			if id != playerID {
				return nil, nil
			}
			return &models.Player{
				ID:       playerID,
				Name:     "Patrick Mahomes",
				Position: "QB",
				TeamID:   &teamID,
			}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/players/"+playerID.String(), nil)
	w := httptest.NewRecorder()

	handler.HandlePlayers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("getPlayer() status = %d, want %d", w.Code, http.StatusOK)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["data"] == nil {
		t.Error("getPlayer() should return data field")
	}

	data := response["data"].(map[string]interface{})
	if data["name"] != "Patrick Mahomes" {
		t.Errorf("getPlayer() name = %v, want Patrick Mahomes", data["name"])
	}
}

func TestGetPlayer_InvalidID(t *testing.T) {
	handler := NewPlayersHandler()
	handler.autoFetchEnabled = false

	req := httptest.NewRequest(http.MethodGet, "/api/v1/players/invalid-uuid", nil)
	w := httptest.NewRecorder()

	handler.HandlePlayers(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("getPlayer() status = %d, want %d for invalid UUID", w.Code, http.StatusBadRequest)
	}
}

func TestGetPlayer_NotFound(t *testing.T) {
	handler := NewPlayersHandler()
	handler.autoFetchEnabled = false

	handler.queries = &mockPlayerQueries{
		getPlayerByIDFunc: func(ctx context.Context, id uuid.UUID) (*models.Player, error) {
			return nil, nil
		},
	}

	playerID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/players/"+playerID.String(), nil)
	w := httptest.NewRecorder()

	handler.HandlePlayers(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("getPlayer() status = %d, want %d for not found", w.Code, http.StatusNotFound)
	}
}

func TestListPlayers_Pagination(t *testing.T) {
	handler := NewPlayersHandler()
	handler.autoFetchEnabled = false

	tests := []struct {
		name           string
		limit          string
		offset         string
		expectedLimit  int
		expectedOffset int
	}{
		{"Default pagination", "", "", 50, 0},
		{"Custom limit", "25", "", 25, 0},
		{"Custom offset", "", "50", 50, 50},
		{"Both limit and offset", "10", "20", 10, 20},
		{"Invalid limit defaults to 50", "-10", "", 50, 0},
		{"Max limit enforced", "200", "", 100, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler.queries = &mockPlayerQueries{
				listPlayersFunc: func(ctx context.Context, filters db.PlayerFilters) ([]*models.Player, int, error) {
					if filters.Limit != tt.expectedLimit {
						t.Errorf("Expected limit %d, got %d", tt.expectedLimit, filters.Limit)
					}
					if filters.Offset != tt.expectedOffset {
						t.Errorf("Expected offset %d, got %d", tt.expectedOffset, filters.Offset)
					}
					return []*models.Player{}, 0, nil
				},
			}

			url := "/api/v1/players?"
			if tt.limit != "" {
				url += "limit=" + tt.limit + "&"
			}
			if tt.offset != "" {
				url += "offset=" + tt.offset
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			handler.HandlePlayers(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("listPlayers() status = %d, want %d", w.Code, http.StatusOK)
			}
		})
	}
}
