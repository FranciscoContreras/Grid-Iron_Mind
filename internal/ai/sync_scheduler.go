package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// SyncScheduler uses AI to intelligently schedule data syncs
type SyncScheduler struct {
	aiService *Service
}

// NewSyncScheduler creates a new AI-powered sync scheduler
func NewSyncScheduler(aiService *Service) *SyncScheduler {
	return &SyncScheduler{
		aiService: aiService,
	}
}

// SyncRecommendation represents an AI recommendation for data sync
type SyncRecommendation struct {
	SyncType     string                 `json:"sync_type"`      // games, stats, injuries, rosters
	Priority     string                 `json:"priority"`       // critical, high, medium, low
	Reason       string                 `json:"reason"`
	EstimatedTime string                `json:"estimated_time"` // e.g., "2 minutes"
	DataAge      string                 `json:"data_age"`       // how old current data is
	NextSyncIn   time.Duration          `json:"next_sync_in"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// SchedulePlan represents a complete sync schedule plan
type SchedulePlan struct {
	Timestamp       time.Time            `json:"timestamp"`
	Recommendations []SyncRecommendation `json:"recommendations"`
	GameDayMode     bool                 `json:"game_day_mode"`
	AIProvider      string               `json:"ai_provider"`
	Reasoning       string               `json:"reasoning"`
}

// GenerateSyncPlan creates an intelligent sync schedule based on current state
func (ss *SyncScheduler) GenerateSyncPlan(ctx context.Context) (*SchedulePlan, error) {
	log.Println("[SYNC SCHEDULER] Generating intelligent sync plan...")

	// Gather context about current state
	now := time.Now()
	dayOfWeek := now.Weekday()
	hourOfDay := now.Hour()

	// Determine if it's game day (Thursday, Sunday, Monday during NFL season)
	isGameDay := (dayOfWeek == time.Thursday || dayOfWeek == time.Sunday || dayOfWeek == time.Monday)
	isNFLSeason := now.Month() >= time.September && now.Month() <= time.February

	contextData := map[string]interface{}{
		"current_time":     now.Format(time.RFC3339),
		"day_of_week":      dayOfWeek.String(),
		"hour_of_day":      hourOfDay,
		"is_game_day":      isGameDay,
		"is_nfl_season":    isNFLSeason,
		"is_business_hours": hourOfDay >= 8 && hourOfDay <= 20,
	}

	// Ask AI to create sync plan
	contextJSON, _ := json.MarshalIndent(contextData, "", "  ")

	prompt := fmt.Sprintf(`You are an NFL data sync scheduler. Create an intelligent sync schedule based on current conditions.

Current Context:
%s

Available Sync Types:
1. "games" - Game schedules and scores (5-10 min)
2. "stats" - Player game statistics (10-20 min)
3. "injuries" - Injury reports (2-5 min)
4. "rosters" - Team rosters (10-15 min)

Scheduling Rules:
- Game days (Thu/Sun/Mon): Prioritize games and stats, sync every 15-30 min during games
- Off days: Sync injuries daily, rosters weekly, stats after games complete
- Business hours: Lower priority unless game day
- Night time: Only critical syncs

Respond with ONLY valid JSON:
{
  "game_day_mode": true/false,
  "reasoning": "why this schedule makes sense",
  "recommendations": [
    {
      "sync_type": "games|stats|injuries|rosters",
      "priority": "critical|high|medium|low",
      "reason": "why sync this now",
      "next_sync_in_minutes": 30
    }
  ]
}

Consider:
- Resource efficiency (don't over-sync)
- Data freshness needs (critical data more frequent)
- API rate limits (spread out syncs)
- User access patterns (sync before peak hours)

Respond with ONLY the JSON, no other text.`, string(contextJSON))

	response, provider, err := ss.aiService.AnswerQuery(ctx, prompt, "Sync schedule generation")
	if err != nil {
		return nil, fmt.Errorf("failed to generate sync plan: %w", err)
	}

	// Parse AI response
	var aiPlan struct {
		GameDayMode     bool   `json:"game_day_mode"`
		Reasoning       string `json:"reasoning"`
		Recommendations []struct {
			SyncType         string `json:"sync_type"`
			Priority         string `json:"priority"`
			Reason           string `json:"reason"`
			NextSyncInMinutes int    `json:"next_sync_in_minutes"`
		} `json:"recommendations"`
	}

	if err := json.Unmarshal([]byte(response), &aiPlan); err != nil {
		return nil, fmt.Errorf("failed to parse AI plan: %w", err)
	}

	// Convert to schedule plan
	plan := &SchedulePlan{
		Timestamp:       now,
		Recommendations: []SyncRecommendation{},
		GameDayMode:     aiPlan.GameDayMode,
		AIProvider:      string(provider),
		Reasoning:       aiPlan.Reasoning,
	}

	for _, rec := range aiPlan.Recommendations {
		plan.Recommendations = append(plan.Recommendations, SyncRecommendation{
			SyncType:     rec.SyncType,
			Priority:     rec.Priority,
			Reason:       rec.Reason,
			NextSyncIn:   time.Duration(rec.NextSyncInMinutes) * time.Minute,
			EstimatedTime: ss.estimateSyncTime(rec.SyncType),
		})
	}

	log.Printf("[SYNC SCHEDULER] Generated plan with %d recommendations (game day: %v)",
		len(plan.Recommendations), plan.GameDayMode)

	return plan, nil
}

// ShouldSyncNow determines if a specific sync should run now
func (ss *SyncScheduler) ShouldSyncNow(ctx context.Context, syncType string, lastSyncTime time.Time) (bool, string, error) {
	timeSinceSync := time.Since(lastSyncTime)

	prompt := fmt.Sprintf(`Should we sync %s data right now?

Last Sync: %s ago (at %s)
Current Time: %s
Day: %s

Decision Criteria:
- Games: Sync every 30min on game days, hourly off days
- Stats: Sync every hour on game days, daily off days
- Injuries: Sync daily during week, twice daily on game days
- Rosters: Sync weekly unless trade deadline

Respond with ONLY valid JSON:
{
  "should_sync": true/false,
  "reason": "brief explanation"
}

Respond with ONLY the JSON, no other text.`,
		syncType,
		formatDuration(timeSinceSync),
		lastSyncTime.Format(time.RFC3339),
		time.Now().Format(time.RFC3339),
		time.Now().Weekday().String())

	response, _, err := ss.aiService.AnswerQuery(ctx, prompt, "Sync decision")
	if err != nil {
		return false, "", err
	}

	var decision struct {
		ShouldSync bool   `json:"should_sync"`
		Reason     string `json:"reason"`
	}

	if err := json.Unmarshal([]byte(response), &decision); err != nil {
		return false, "", err
	}

	return decision.ShouldSync, decision.Reason, nil
}

// PredictDataUsage predicts when users will query specific data
func (ss *SyncScheduler) PredictDataUsage(ctx context.Context, dataType string) ([]time.Time, error) {
	prompt := fmt.Sprintf(`Predict when users are most likely to query %s data in the next 24 hours.

Consider:
- Fantasy football lineup setting deadlines
- Pre-game research (2-3 hours before kickoff)
- Post-game stat checking (1-2 hours after games)
- Weekly analysis (Tuesday-Wednesday)

Respond with ONLY valid JSON:
{
  "peak_times": [
    {"hour": 14, "day_offset": 0, "reason": "pre-game research"}
  ]
}

Use 24-hour format. day_offset: 0=today, 1=tomorrow.
Respond with ONLY the JSON, no other text.`, dataType)

	response, _, err := ss.aiService.AnswerQuery(ctx, prompt, "Usage prediction")
	if err != nil {
		return nil, err
	}

	var result struct {
		PeakTimes []struct {
			Hour      int    `json:"hour"`
			DayOffset int    `json:"day_offset"`
			Reason    string `json:"reason"`
		} `json:"peak_times"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	// Convert to actual times
	peakTimes := []time.Time{}
	now := time.Now()

	for _, pt := range result.PeakTimes {
		peakTime := time.Date(
			now.Year(), now.Month(), now.Day()+pt.DayOffset,
			pt.Hour, 0, 0, 0, now.Location(),
		)
		peakTimes = append(peakTimes, peakTime)
	}

	return peakTimes, nil
}

// estimateSyncTime provides estimated duration for sync operations
func (ss *SyncScheduler) estimateSyncTime(syncType string) string {
	estimates := map[string]string{
		"games":    "5-10 minutes",
		"stats":    "10-20 minutes",
		"injuries": "2-5 minutes",
		"rosters":  "10-15 minutes",
	}

	if est, ok := estimates[syncType]; ok {
		return est
	}
	return "5-10 minutes"
}

// formatDuration formats a duration in human-readable format
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%d hours", int(d.Hours()))
	}
	return fmt.Sprintf("%d days", int(d.Hours()/24))
}
