package ingestion

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/google/uuid"
)

// CalculateStandings calculates team standings for a given season and week
//
// This function:
//   - Queries all games up to the specified week
//   - Calculates wins, losses, ties, points for/against
//   - Computes split records (home/away, division, conference)
//   - Determines current streak
//   - Ranks teams by division and conference
//   - Assigns playoff seeds
//   - Inserts/updates team_standings table
//
// Example usage:
//
//	if err := service.CalculateStandings(ctx, 2025, 4); err != nil {
//	    log.Fatal(err)
//	}
func (s *Service) CalculateStandings(ctx context.Context, season, week int) error {
	log.Printf("Calculating standings for season %d week %d...", season, week)

	// Get all teams
	teamsQuery := `
		SELECT id, abbreviation, conference, division
		FROM teams
		ORDER BY conference, division, abbreviation
	`

	teamsRows, err := s.dbPool.Query(ctx, teamsQuery)
	if err != nil {
		return fmt.Errorf("failed to query teams: %w", err)
	}
	defer teamsRows.Close()

	type TeamInfo struct {
		ID           uuid.UUID
		Abbreviation string
		Conference   string
		Division     string
	}

	var teams []TeamInfo
	teamMap := make(map[uuid.UUID]TeamInfo)

	for teamsRows.Next() {
		var team TeamInfo
		if err := teamsRows.Scan(&team.ID, &team.Abbreviation, &team.Conference, &team.Division); err != nil {
			log.Printf("Failed to scan team: %v", err)
			continue
		}
		teams = append(teams, team)
		teamMap[team.ID] = team
	}

	log.Printf("Calculating standings for %d teams", len(teams))

	// Calculate standings for each team
	type TeamStanding struct {
		TeamID            uuid.UUID
		Wins              int
		Losses            int
		Ties              int
		PointsFor         int
		PointsAgainst     int
		HomeWins          int
		HomeLosses        int
		AwayWins          int
		AwayLosses        int
		DivisionWins      int
		DivisionLosses    int
		ConferenceWins    int
		ConferenceLosses  int
		CurrentStreak     string
		RecentResults     []string // For streak calculation
	}

	standings := make(map[uuid.UUID]*TeamStanding)

	for _, team := range teams {
		standings[team.ID] = &TeamStanding{
			TeamID:        team.ID,
			RecentResults: []string{},
		}
	}

	// Query all completed games up to this week
	gamesQuery := `
		SELECT
			id, home_team_id, away_team_id,
			home_score, away_score,
			status
		FROM games
		WHERE season = $1
		AND week <= $2
		AND status = 'completed'
		ORDER BY week, game_date
	`

	gamesRows, err := s.dbPool.Query(ctx, gamesQuery, season, week)
	if err != nil {
		return fmt.Errorf("failed to query games: %w", err)
	}
	defer gamesRows.Close()

	gameCount := 0
	for gamesRows.Next() {
		var gameID uuid.UUID
		var homeTeamID, awayTeamID uuid.UUID
		var homeScore, awayScore int
		var status string

		if err := gamesRows.Scan(&gameID, &homeTeamID, &awayTeamID, &homeScore, &awayScore, &status); err != nil {
			log.Printf("Failed to scan game: %v", err)
			continue
		}

		gameCount++

		homeTeam := teamMap[homeTeamID]
		awayTeam := teamMap[awayTeamID]
		homeStanding := standings[homeTeamID]
		awayStanding := standings[awayTeamID]

		// Determine winner
		if homeScore > awayScore {
			// Home team wins
			homeStanding.Wins++
			homeStanding.HomeWins++
			homeStanding.RecentResults = append(homeStanding.RecentResults, "W")

			awayStanding.Losses++
			awayStanding.AwayLosses++
			awayStanding.RecentResults = append(awayStanding.RecentResults, "L")

			// Division game?
			if homeTeam.Division == awayTeam.Division {
				homeStanding.DivisionWins++
				awayStanding.DivisionLosses++
			}

			// Conference game?
			if homeTeam.Conference == awayTeam.Conference {
				homeStanding.ConferenceWins++
				awayStanding.ConferenceLosses++
			}

		} else if awayScore > homeScore {
			// Away team wins
			awayStanding.Wins++
			awayStanding.AwayWins++
			awayStanding.RecentResults = append(awayStanding.RecentResults, "W")

			homeStanding.Losses++
			homeStanding.HomeLosses++
			homeStanding.RecentResults = append(homeStanding.RecentResults, "L")

			// Division game?
			if homeTeam.Division == awayTeam.Division {
				awayStanding.DivisionWins++
				homeStanding.DivisionLosses++
			}

			// Conference game?
			if homeTeam.Conference == awayTeam.Conference {
				awayStanding.ConferenceWins++
				homeStanding.ConferenceLosses++
			}

		} else {
			// Tie
			homeStanding.Ties++
			awayStanding.Ties++
			homeStanding.RecentResults = append(homeStanding.RecentResults, "T")
			awayStanding.RecentResults = append(awayStanding.RecentResults, "T")
		}

		// Points
		homeStanding.PointsFor += homeScore
		homeStanding.PointsAgainst += awayScore
		awayStanding.PointsFor += awayScore
		awayStanding.PointsAgainst += homeScore
	}

	log.Printf("Processed %d games", gameCount)

	// Calculate streaks, rankings, and insert
	for _, team := range teams {
		standing := standings[team.ID]

		// Calculate win percentage
		totalGames := standing.Wins + standing.Losses + standing.Ties
		var winPct float64
		if totalGames > 0 {
			winPct = (float64(standing.Wins) + (float64(standing.Ties) * 0.5)) / float64(totalGames)
		}

		// Calculate point differential
		pointDiff := standing.PointsFor - standing.PointsAgainst

		// Calculate current streak
		streak := calculateStreak(standing.RecentResults)

		// Insert/Update standings
		query := `
			INSERT INTO team_standings (
				team_id, season, week,
				wins, losses, ties, win_pct,
				points_for, points_against, point_differential,
				home_wins, home_losses,
				away_wins, away_losses,
				division_wins, division_losses,
				conference_wins, conference_losses,
				current_streak,
				updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, NOW()
			)
			ON CONFLICT (team_id, season, week)
			DO UPDATE SET
				wins = EXCLUDED.wins,
				losses = EXCLUDED.losses,
				ties = EXCLUDED.ties,
				win_pct = EXCLUDED.win_pct,
				points_for = EXCLUDED.points_for,
				points_against = EXCLUDED.points_against,
				point_differential = EXCLUDED.point_differential,
				home_wins = EXCLUDED.home_wins,
				home_losses = EXCLUDED.home_losses,
				away_wins = EXCLUDED.away_wins,
				away_losses = EXCLUDED.away_losses,
				division_wins = EXCLUDED.division_wins,
				division_losses = EXCLUDED.division_losses,
				conference_wins = EXCLUDED.conference_wins,
				conference_losses = EXCLUDED.conference_losses,
				current_streak = EXCLUDED.current_streak,
				updated_at = NOW()
		`

		_, err := s.dbPool.Exec(ctx, query,
			team.ID, season, week,
			standing.Wins, standing.Losses, standing.Ties, winPct,
			standing.PointsFor, standing.PointsAgainst, pointDiff,
			standing.HomeWins, standing.HomeLosses,
			standing.AwayWins, standing.AwayLosses,
			standing.DivisionWins, standing.DivisionLosses,
			standing.ConferenceWins, standing.ConferenceLosses,
			streak,
		)

		if err != nil {
			log.Printf("Failed to insert standings for team %s: %v", team.Abbreviation, err)
			continue
		}
	}

	// Calculate rankings
	if err := s.calculateRankings(ctx, season, week); err != nil {
		log.Printf("Failed to calculate rankings: %v", err)
	}

	log.Printf("Standings calculation completed for week %d", week)
	return nil
}

// calculateStreak determines current win/loss streak from recent results
func calculateStreak(results []string) string {
	if len(results) == 0 {
		return ""
	}

	// Start from most recent game
	lastResult := results[len(results)-1]
	count := 1

	// Count consecutive same results
	for i := len(results) - 2; i >= 0; i-- {
		if results[i] == lastResult {
			count++
		} else {
			break
		}
	}

	return fmt.Sprintf("%s%d", lastResult, count)
}

// calculateRankings calculates division and conference rankings
func (s *Service) calculateRankings(ctx context.Context, season, week int) error {
	// Get all standings for this week
	query := `
		SELECT
			ts.team_id,
			ts.wins,
			ts.losses,
			ts.ties,
			ts.win_pct,
			ts.point_differential,
			ts.division_wins,
			ts.conference_wins,
			t.conference,
			t.division
		FROM team_standings ts
		JOIN teams t ON ts.team_id = t.id
		WHERE ts.season = $1 AND ts.week = $2
		ORDER BY t.conference, t.division, ts.win_pct DESC, ts.point_differential DESC
	`

	rows, err := s.dbPool.Query(ctx, query, season, week)
	if err != nil {
		return fmt.Errorf("failed to query standings: %w", err)
	}
	defer rows.Close()

	type Standing struct {
		TeamID           uuid.UUID
		Wins             int
		Losses           int
		Ties             int
		WinPct           float64
		PointDiff        int
		DivisionWins     int
		ConferenceWins   int
		Conference       string
		Division         string
	}

	var allStandings []Standing

	for rows.Next() {
		var s Standing
		if err := rows.Scan(&s.TeamID, &s.Wins, &s.Losses, &s.Ties, &s.WinPct, &s.PointDiff,
			&s.DivisionWins, &s.ConferenceWins, &s.Conference, &s.Division); err != nil {
			log.Printf("Failed to scan standing: %v", err)
			continue
		}
		allStandings = append(allStandings, s)
	}

	// Group by division
	divisionGroups := make(map[string][]Standing)
	conferenceGroups := make(map[string][]Standing)

	for _, s := range allStandings {
		divKey := s.Conference + "_" + s.Division
		divisionGroups[divKey] = append(divisionGroups[divKey], s)
		conferenceGroups[s.Conference] = append(conferenceGroups[s.Conference], s)
	}

	// Sort and rank divisions
	for divKey, standings := range divisionGroups {
		sort.Slice(standings, func(i, j int) bool {
			if standings[i].WinPct != standings[j].WinPct {
				return standings[i].WinPct > standings[j].WinPct
			}
			if standings[i].DivisionWins != standings[j].DivisionWins {
				return standings[i].DivisionWins > standings[j].DivisionWins
			}
			return standings[i].PointDiff > standings[j].PointDiff
		})

		// Assign division ranks
		for rank, s := range standings {
			_, err := s.dbPool.Exec(ctx,
				`UPDATE team_standings SET division_rank = $1 WHERE team_id = $2 AND season = $3 AND week = $4`,
				rank+1, s.TeamID, season, week)
			if err != nil {
				log.Printf("Failed to update division rank: %v", err)
			}
		}

		log.Printf("Ranked %s: %d teams", divKey, len(standings))
	}

	// Sort and rank conferences (for playoff seeds)
	for conf, standings := range conferenceGroups {
		sort.Slice(standings, func(i, j int) bool {
			// Division winners first
			iDivWinner := standings[i].Division != "" && divisionGroups[standings[i].Conference+"_"+standings[i].Division][0].TeamID == standings[i].TeamID
			jDivWinner := standings[j].Division != "" && divisionGroups[standings[j].Conference+"_"+standings[j].Division][0].TeamID == standings[j].TeamID

			if iDivWinner && !jDivWinner {
				return true
			}
			if !iDivWinner && jDivWinner {
				return false
			}

			// Then by win percentage
			if standings[i].WinPct != standings[j].WinPct {
				return standings[i].WinPct > standings[j].WinPct
			}

			// Then by conference wins
			if standings[i].ConferenceWins != standings[j].ConferenceWins {
				return standings[i].ConferenceWins > standings[j].ConferenceWins
			}

			// Then by point differential
			return standings[i].PointDiff > standings[j].PointDiff
		})

		// Assign conference ranks and playoff seeds (top 7)
		for rank, s := range standings {
			var playoffSeed *int
			if rank < 7 {
				seed := rank + 1
				playoffSeed = &seed
			}

			_, err := s.dbPool.Exec(ctx,
				`UPDATE team_standings SET conference_rank = $1, playoff_seed = $2 WHERE team_id = $3 AND season = $4 AND week = $5`,
				rank+1, playoffSeed, s.TeamID, season, week)
			if err != nil {
				log.Printf("Failed to update conference rank: %v", err)
			}
		}

		log.Printf("Ranked %s conference: %d teams", conf, len(standings))
	}

	return nil
}

// CalculateStandingsSeason calculates standings for all weeks in a season
func (s *Service) CalculateStandingsSeason(ctx context.Context, season int) error {
	log.Printf("Calculating standings for entire season %d...", season)

	// Determine max week with completed games
	var maxWeek int
	err := s.dbPool.QueryRow(ctx, `
		SELECT COALESCE(MAX(week), 0)
		FROM games
		WHERE season = $1 AND status = 'completed'
	`, season).Scan(&maxWeek)

	if err != nil {
		return fmt.Errorf("failed to get max week: %w", err)
	}

	if maxWeek == 0 {
		return fmt.Errorf("no completed games found for season %d", season)
	}

	log.Printf("Found completed games through week %d", maxWeek)

	// Calculate for each week
	for week := 1; week <= maxWeek; week++ {
		log.Printf("Calculating standings for week %d/%d...", week, maxWeek)
		if err := s.CalculateStandings(ctx, season, week); err != nil {
			log.Printf("Failed to calculate standings for week %d: %v", week, err)
			continue
		}
	}

	log.Printf("Completed standings calculation for season %d", season)
	return nil
}
