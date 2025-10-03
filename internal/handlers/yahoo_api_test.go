package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/francisco/gridironmind/internal/config"
	"github.com/francisco/gridironmind/pkg/response"
	"golang.org/x/oauth2"
)

// YahooTestHandler handles Yahoo API test requests
type YahooTestHandler struct {
	clientID     string
	clientSecret string
	refreshToken string
	httpClient   *http.Client
}

// NewYahooTestHandler creates a new Yahoo test handler
func NewYahooTestHandler(cfg *config.Config) *YahooTestHandler {
	if cfg.YahooClientID == "" || cfg.YahooClientSecret == "" || cfg.YahooRefreshToken == "" {
		return &YahooTestHandler{}
	}

	// Create OAuth config
	oauth2Config := &oauth2.Config{
		ClientID:     cfg.YahooClientID,
		ClientSecret: cfg.YahooClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://api.login.yahoo.com/oauth2/request_auth",
			TokenURL: "https://api.login.yahoo.com/oauth2/get_token",
		},
		Scopes: []string{},
	}

	// Create token from refresh token
	token := &oauth2.Token{
		RefreshToken: cfg.YahooRefreshToken,
		TokenType:    "Bearer",
	}

	httpClient := oauth2Config.Client(context.Background(), token)

	return &YahooTestHandler{
		clientID:     cfg.YahooClientID,
		clientSecret: cfg.YahooClientSecret,
		refreshToken: cfg.YahooRefreshToken,
		httpClient:   httpClient,
	}
}

// HandleUserInfo retrieves authenticated user information
func (h *YahooTestHandler) HandleUserInfo(w http.ResponseWriter, r *http.Request) {
	if h.httpClient == nil {
		response.Error(w, http.StatusServiceUnavailable, "YAHOO_NOT_CONFIGURED", "Yahoo API is not configured")
		return
	}

	data, err := h.makeRawRequest(r.Context(), "/users;use_login=1")
	if err != nil {
		response.InternalError(w, fmt.Sprintf("Failed to fetch user info: %v", err))
		return
	}

	response.Success(w, data)
}

// HandleUserLeagues retrieves user's fantasy leagues
func (h *YahooTestHandler) HandleUserLeagues(w http.ResponseWriter, r *http.Request) {
	if h.httpClient == nil {
		response.Error(w, http.StatusServiceUnavailable, "YAHOO_NOT_CONFIGURED", "Yahoo API is not configured")
		return
	}

	data, err := h.makeRawRequest(r.Context(), "/users;use_login=1/games/leagues")
	if err != nil {
		response.InternalError(w, fmt.Sprintf("Failed to fetch leagues: %v", err))
		return
	}

	response.Success(w, data)
}

// HandleLeagueInfo retrieves specific league information
func (h *YahooTestHandler) HandleLeagueInfo(w http.ResponseWriter, r *http.Request) {
	if h.httpClient == nil {
		response.Error(w, http.StatusServiceUnavailable, "YAHOO_NOT_CONFIGURED", "Yahoo API is not configured")
		return
	}

	// Extract league key from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/yahoo/league/")
	leagueKey := strings.TrimSpace(path)

	if leagueKey == "" {
		response.BadRequest(w, "League key is required")
		return
	}

	endpoint := fmt.Sprintf("/league/%s", leagueKey)
	data, err := h.makeRawRequest(r.Context(), endpoint)
	if err != nil {
		response.InternalError(w, fmt.Sprintf("Failed to fetch league info: %v", err))
		return
	}

	response.Success(w, data)
}

// HandleTeamRoster retrieves team roster information
func (h *YahooTestHandler) HandleTeamRoster(w http.ResponseWriter, r *http.Request) {
	if h.httpClient == nil {
		response.Error(w, http.StatusServiceUnavailable, "YAHOO_NOT_CONFIGURED", "Yahoo API is not configured")
		return
	}

	// Extract team key from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/yahoo/team/")
	teamKey := strings.TrimSpace(path)

	if teamKey == "" {
		response.BadRequest(w, "Team key is required")
		return
	}

	endpoint := fmt.Sprintf("/team/%s/roster", teamKey)
	data, err := h.makeRawRequest(r.Context(), endpoint)
	if err != nil {
		response.InternalError(w, fmt.Sprintf("Failed to fetch team roster: %v", err))
		return
	}

	response.Success(w, data)
}

// HandleRawAPI makes a raw Yahoo API call with custom endpoint
func (h *YahooTestHandler) HandleRawAPI(w http.ResponseWriter, r *http.Request) {
	if h.httpClient == nil {
		response.Error(w, http.StatusServiceUnavailable, "YAHOO_NOT_CONFIGURED", "Yahoo API is not configured")
		return
	}

	endpoint := r.URL.Query().Get("url")
	if endpoint == "" {
		response.BadRequest(w, "URL parameter is required")
		return
	}

	// Ensure endpoint starts with /
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}

	data, err := h.makeRawRequest(r.Context(), endpoint)
	if err != nil {
		response.InternalError(w, fmt.Sprintf("Failed to make Yahoo API request: %v", err))
		return
	}

	response.Success(w, data)
}

// makeRawRequest is a helper to make raw Yahoo API requests and return JSON
func (h *YahooTestHandler) makeRawRequest(ctx context.Context, endpoint string) (interface{}, error) {
	baseURL := "https://fantasysports.yahooapis.com/fantasy/v2"
	reqURL := fmt.Sprintf("%s%s", baseURL, endpoint)

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "GridIronMind/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Yahoo API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var data interface{}
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return data, nil
}
