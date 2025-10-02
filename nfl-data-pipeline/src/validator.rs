use anyhow::{Result, anyhow};
use log::{warn, info};

use crate::parser::{RosterPlayer, PlayerStat, Game};

pub struct DataValidator;

impl DataValidator {
    pub fn new() -> Self {
        DataValidator
    }

    /// Validate roster player data
    pub fn validate_player(&self, player: &RosterPlayer) -> Result<()> {
        // Required fields
        if player.gsis_id.is_empty() {
            return Err(anyhow!("Player missing gsis_id"));
        }
        if player.full_name.is_empty() {
            return Err(anyhow!("Player missing full_name"));
        }
        if player.team.is_empty() {
            return Err(anyhow!("Player missing team"));
        }
        if player.position.is_empty() {
            return Err(anyhow!("Player missing position"));
        }

        // Validate season range
        if player.season < 1999 || player.season > 2030 {
            return Err(anyhow!("Invalid season: {}", player.season));
        }

        // Validate position codes
        let valid_positions = vec![
            "QB", "RB", "WR", "TE", "FB", "HB",
            "OL", "OT", "OG", "C", "G", "T",
            "DL", "DE", "DT", "NT",
            "LB", "ILB", "OLB", "MLB",
            "DB", "CB", "S", "FS", "SS",
            "K", "P", "LS",
        ];
        if !valid_positions.contains(&player.position.as_str()) {
            warn!("Unknown position code: {}", player.position);
        }

        // Validate height (if present)
        if let Some(height) = &player.height {
            if !height.contains('-') {
                warn!("Invalid height format: {}", height);
            }
        }

        // Validate weight (if present)
        if let Some(weight) = player.weight {
            if weight < 150 || weight > 400 {
                warn!("Unusual weight: {} for {}", weight, player.full_name);
            }
        }

        Ok(())
    }

    /// Validate game data
    pub fn validate_game(&self, game: &Game) -> Result<()> {
        // Required fields
        if game.game_id.is_empty() {
            return Err(anyhow!("Game missing game_id"));
        }
        if game.home_team.is_empty() {
            return Err(anyhow!("Game missing home_team"));
        }
        if game.away_team.is_empty() {
            return Err(anyhow!("Game missing away_team"));
        }
        if game.gameday.is_empty() {
            return Err(anyhow!("Game missing gameday"));
        }

        // Validate season
        if game.season < 1999 || game.season > 2030 {
            return Err(anyhow!("Invalid season: {}", game.season));
        }

        // Validate week
        if game.week < 1 || game.week > 22 {
            return Err(anyhow!("Invalid week: {}", game.week));
        }

        // Validate game type
        let valid_types = vec!["REG", "PRE", "POST", "WC", "DIV", "CON", "SB"];
        if !valid_types.contains(&game.game_type.as_str()) {
            warn!("Unknown game type: {}", game.game_type);
        }

        // Validate scores (if present)
        if let Some(home_score) = game.home_score {
            if home_score < 0 || home_score > 100 {
                warn!("Unusual home score: {} in {}", home_score, game.game_id);
            }
        }
        if let Some(away_score) = game.away_score {
            if away_score < 0 || away_score > 100 {
                warn!("Unusual away score: {} in {}", away_score, game.game_id);
            }
        }

        Ok(())
    }

    /// Validate player stat data
    pub fn validate_stat(&self, stat: &PlayerStat) -> Result<()> {
        // Required fields
        if stat.player_id.is_empty() {
            return Err(anyhow!("Stat missing player_id"));
        }

        // Validate season
        if stat.season < 1999 || stat.season > 2030 {
            return Err(anyhow!("Invalid season: {}", stat.season));
        }

        // Validate week
        if stat.week < 1 || stat.week > 22 {
            return Err(anyhow!("Invalid week: {}", stat.week));
        }

        // Validate season type
        let valid_types = vec!["REG", "PRE", "POST"];
        if !valid_types.contains(&stat.season_type.as_str()) {
            warn!("Unknown season type: {}", stat.season_type);
        }

        // Validate reasonable stat ranges (warnings only)
        if let Some(yards) = stat.passing_yards {
            if yards < 0.0 || yards > 600.0 {
                warn!("Unusual passing yards: {} for {}", yards, stat.player_id);
            }
        }
        if let Some(yards) = stat.rushing_yards {
            if yards < -20.0 || yards > 300.0 {
                warn!("Unusual rushing yards: {} for {}", yards, stat.player_id);
            }
        }
        if let Some(yards) = stat.receiving_yards {
            if yards < 0.0 || yards > 300.0 {
                warn!("Unusual receiving yards: {} for {}", yards, stat.player_id);
            }
        }

        Ok(())
    }

    /// Validate batch of players
    pub fn validate_player_batch(&self, players: &[RosterPlayer]) -> (usize, usize) {
        let mut valid = 0;
        let mut invalid = 0;

        for player in players {
            match self.validate_player(player) {
                Ok(_) => valid += 1,
                Err(e) => {
                    warn!("Invalid player {}: {}", player.gsis_id, e);
                    invalid += 1;
                }
            }
        }

        info!("Player validation: {} valid, {} invalid", valid, invalid);
        (valid, invalid)
    }

    /// Validate batch of games
    pub fn validate_game_batch(&self, games: &[Game]) -> (usize, usize) {
        let mut valid = 0;
        let mut invalid = 0;

        for game in games {
            match self.validate_game(game) {
                Ok(_) => valid += 1,
                Err(e) => {
                    warn!("Invalid game {}: {}", game.game_id, e);
                    invalid += 1;
                }
            }
        }

        info!("Game validation: {} valid, {} invalid", valid, invalid);
        (valid, invalid)
    }

    /// Validate batch of stats
    pub fn validate_stat_batch(&self, stats: &[PlayerStat]) -> (usize, usize) {
        let mut valid = 0;
        let mut invalid = 0;

        for stat in stats {
            match self.validate_stat(stat) {
                Ok(_) => valid += 1,
                Err(e) => {
                    warn!("Invalid stat for {}: {}", stat.player_id, e);
                    invalid += 1;
                }
            }
        }

        info!("Stat validation: {} valid, {} invalid", valid, invalid);
        (valid, invalid)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_validate_player() {
        let validator = DataValidator::new();

        let valid_player = RosterPlayer {
            season: 2024,
            team: "KC".to_string(),
            position: "QB".to_string(),
            depth_chart_position: Some("QB".to_string()),
            jersey_number: Some(15),
            status: Some("ACT".to_string()),
            full_name: "Patrick Mahomes".to_string(),
            first_name: Some("Patrick".to_string()),
            last_name: Some("Mahomes".to_string()),
            birth_date: Some("1995-09-17".to_string()),
            height: Some("6-3".to_string()),
            weight: Some(230),
            college: Some("Texas Tech".to_string()),
            gsis_id: "00-0033873".to_string(),
            espn_id: None,
            sportradar_id: None,
            yahoo_id: None,
            rotowire_id: None,
            pff_id: None,
            pfr_id: None,
            fantasy_data_id: None,
            sleeper_id: None,
            years_exp: Some(7),
            headshot_url: None,
            entry_year: Some(2017),
            rookie_year: Some(2017),
            draft_club: Some("KC".to_string()),
            draft_number: Some(10),
        };

        assert!(validator.validate_player(&valid_player).is_ok());
    }

    #[test]
    fn test_validate_game() {
        let validator = DataValidator::new();

        let valid_game = Game {
            game_id: "2024_01_KC_BAL".to_string(),
            season: 2024,
            game_type: "REG".to_string(),
            week: 1,
            gameday: "2024-09-05".to_string(),
            weekday: Some("Thursday".to_string()),
            gametime: Some("20:20".to_string()),
            away_team: "KC".to_string(),
            away_score: Some(27),
            home_team: "BAL".to_string(),
            home_score: Some(20),
            location: Some("home".to_string()),
            result: Some(7),
            total: Some(47),
            overtime: Some(0),
            old_game_id: None,
            gsis: None,
            nfl_detail_id: None,
            pfr: None,
            pff: None,
            espn: None,
            ftn: None,
            away_rest: None,
            home_rest: None,
            away_moneyline: None,
            home_moneyline: None,
            spread_line: None,
            away_spread_odds: None,
            home_spread_odds: None,
            total_line: None,
            under_odds: None,
            over_odds: None,
            div_game: None,
            roof: None,
            surface: None,
            temp: None,
            wind: None,
            away_qb_id: None,
            home_qb_id: None,
            away_qb_name: None,
            home_qb_name: None,
            away_coach: None,
            home_coach: None,
            referee: None,
            stadium_id: None,
            stadium: None,
        };

        assert!(validator.validate_game(&valid_game).is_ok());
    }
}
