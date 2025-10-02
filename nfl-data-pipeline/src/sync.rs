use anyhow::Result;
use log::{info, warn, error};
use indicatif::{ProgressBar, ProgressStyle};
use csv::ReaderBuilder;
use chrono::Datelike;

use crate::config::Config;
use crate::database::Database;
use crate::downloader::Downloader;
use crate::parser::{RosterPlayer, PlayerStat, Game};
use crate::transformer;

pub struct DataPipeline {
    pub config: Config,
    downloader: Downloader,
    database: Database,
}

impl DataPipeline {
    pub fn new(config: Config) -> Result<Self> {
        let downloader = Downloader::new(config.max_retries);
        let database = Database::connect(&config.database_url)?;

        Ok(DataPipeline {
            config,
            downloader,
            database,
        })
    }

    /// Run full import for all years
    pub fn run_full_import(&mut self) -> Result<()> {
        let pb = ProgressBar::new((self.config.end_year - self.config.start_year + 1) as u64);
        pb.set_style(
            ProgressStyle::default_bar()
                .template("[{elapsed_precise}] {bar:40.cyan/blue} {pos}/{len} {msg}")
                .unwrap(),
        );

        for year in self.config.start_year..=self.config.end_year {
            pb.set_message(format!("Importing {}", year));

            match self.import_year(year) {
                Ok(_) => {
                    pb.inc(1);
                }
                Err(e) => {
                    error!("Failed to import year {}: {}", year, e);
                    pb.inc(1);
                }
            }
        }

        pb.finish_with_message("Import complete!");
        Ok(())
    }

    /// Import data for a single year
    pub fn import_year(&mut self, year: i32) -> Result<()> {
        info!("📅 Importing data for year {}...", year);

        // 1. Import rosters (players)
        match self.import_rosters(year) {
            Ok(count) => info!("  ✅ Rosters: {} players", count),
            Err(e) => warn!("  ⚠️  Rosters failed: {}", e),
        }

        // 2. Import schedule (games) - SKIPPED: Use ESPN API via Go importer instead
        // NFLverse schedule format is different, easier to use ESPN for schedules
        info!("  ⏭️  Schedule: Skipping (use Go importer with ESPN API)");

        // 3. Import player stats
        match self.import_player_stats(year) {
            Ok(count) => info!("  ✅ Player Stats: {} records", count),
            Err(e) => warn!("  ⚠️  Player Stats failed: {}", e),
        }

        // 4. Import Next Gen Stats (2016+)
        if year >= 2016 {
            match self.import_ngs_passing(year) {
                Ok(count) => info!("  ✅ NGS Passing: {} records", count),
                Err(e) => warn!("  ⚠️  NGS Passing failed: {}", e),
            }
        }

        info!("✅ Year {} import complete", year);
        Ok(())
    }

    /// Import rosters for a year
    fn import_rosters(&mut self, year: i32) -> Result<usize> {
        info!("  [1/4] Importing rosters for {}...", year);

        let csv_data = self.downloader.download_roster(year)?;
        let mut reader = ReaderBuilder::new()
            .from_reader(csv_data.as_bytes());

        let mut imported = 0;
        let mut batch = Vec::new();

        for result in reader.deserialize::<RosterPlayer>() {
            match result {
                Ok(player) => {
                    batch.push(player);

                    if batch.len() >= self.config.batch_size {
                        self.upsert_players_batch(&batch)?;
                        imported += batch.len();
                        batch.clear();
                    }
                }
                Err(e) => warn!("Failed to parse roster row: {}", e),
            }
        }

        // Insert remaining
        if !batch.is_empty() {
            self.upsert_players_batch(&batch)?;
            imported += batch.len();
        }

        if !self.config.dry_run {
            self.database.mark_progress(year, "rosters", "completed", imported as i32)?;
        }

        Ok(imported)
    }

    /// Import schedule for a year
    fn import_schedule(&mut self, year: i32) -> Result<usize> {
        info!("  [2/4] Importing schedule for {}...", year);

        let csv_data = self.downloader.download_schedule(year)?;
        let mut reader = ReaderBuilder::new()
            .from_reader(csv_data.as_bytes());

        let mut imported = 0;
        let mut batch = Vec::new();

        for result in reader.deserialize::<Game>() {
            match result {
                Ok(game) => {
                    // Only import regular season games
                    if game.game_type == "REG" {
                        batch.push(game);

                        if batch.len() >= self.config.batch_size {
                            self.upsert_games_batch(&batch)?;
                            imported += batch.len();
                            batch.clear();
                        }
                    }
                }
                Err(e) => warn!("Failed to parse schedule row: {}", e),
            }
        }

        // Insert remaining
        if !batch.is_empty() {
            self.upsert_games_batch(&batch)?;
            imported += batch.len();
        }

        if !self.config.dry_run {
            self.database.mark_progress(year, "schedule", "completed", imported as i32)?;
        }

        Ok(imported)
    }

    /// Import player stats for a year
    fn import_player_stats(&mut self, year: i32) -> Result<usize> {
        info!("  [3/4] Importing player stats for {}...", year);

        let csv_data = self.downloader.download_player_stats(year)?;
        let mut reader = ReaderBuilder::new()
            .from_reader(csv_data.as_bytes());

        let mut imported = 0;
        let mut batch = Vec::new();

        for result in reader.deserialize::<PlayerStat>() {
            match result {
                Ok(stat) => {
                    // Only import regular season stats
                    if stat.season_type == "REG" {
                        batch.push(stat);

                        if batch.len() >= self.config.batch_size {
                            self.upsert_stats_batch(&batch)?;
                            imported += batch.len();
                            batch.clear();
                        }
                    }
                }
                Err(e) => warn!("Failed to parse stat row: {}", e),
            }
        }

        // Insert remaining
        if !batch.is_empty() {
            self.upsert_stats_batch(&batch)?;
            imported += batch.len();
        }

        if !self.config.dry_run {
            self.database.mark_progress(year, "player_stats", "completed", imported as i32)?;
        }

        Ok(imported)
    }

    /// Import NGS passing stats
    fn import_ngs_passing(&mut self, year: i32) -> Result<usize> {
        info!("  [4/4] Importing NGS passing for {}...", year);

        let csv_data = self.downloader.download_ngs_passing(year)?;
        // Parsing logic here (similar to above)

        Ok(0) // Placeholder
    }

    /// Run incremental update
    pub fn run_update(&mut self) -> Result<()> {
        info!("🔄 Running incremental update...");

        // Get current year
        let current_year = chrono::Utc::now().year();

        // Update current season
        self.import_year(current_year)?;

        Ok(())
    }

    /// Validate existing data
    pub fn validate_data(&mut self) -> Result<()> {
        info!("✅ Validating data...");

        for year in self.config.start_year..=self.config.end_year {
            let games = self.database.count_games(year)?;
            let stats = self.database.count_game_stats(year)?;

            info!("  {} - Games: {}, Stats: {}", year, games, stats);
        }

        let total_players = self.database.count_players()?;
        info!("  Total players: {}", total_players);

        Ok(())
    }

    // Batch upsert methods (placeholder - implement actual SQL)
    fn upsert_players_batch(&mut self, players: &[RosterPlayer]) -> Result<()> {
        if self.config.dry_run {
            return Ok(());
        }

        for player in players {
            self.upsert_player(player)?;
        }

        Ok(())
    }

    fn upsert_player(&mut self, player: &RosterPlayer) -> Result<()> {
        let team_abbr = transformer::normalize_team_abbr(&player.team);
        let team_id = self.database.get_team_id_by_abbr(&team_abbr)?;

        let height_inches = player.height.as_ref().and_then(|h| transformer::height_to_inches(h));

        let client = self.database.get_client();

        // Convert to explicit types to match PostgreSQL expectations
        let nfl_id: &str = &player.gsis_id;
        let name: &str = &player.full_name;
        let position: &str = &player.position;
        let status: &str = player.status.as_deref().unwrap_or("active");
        let college: Option<&str> = player.college.as_deref();

        client.execute(
            "INSERT INTO players (id, nfl_id, name, position, team_id, jersey_number, height_inches, weight_pounds, college, status, created_at, updated_at)
             VALUES (uuid_generate_v4(), $1::text, $2::text, $3::text, $4, $5, $6, $7, $8::text, $9::text, NOW(), NOW())
             ON CONFLICT (nfl_id) DO UPDATE SET
                 name = EXCLUDED.name,
                 position = EXCLUDED.position,
                 team_id = EXCLUDED.team_id,
                 jersey_number = EXCLUDED.jersey_number,
                 height_inches = EXCLUDED.height_inches,
                 weight_pounds = EXCLUDED.weight_pounds,
                 college = EXCLUDED.college,
                 status = EXCLUDED.status,
                 updated_at = NOW()",
            &[
                &nfl_id,
                &name,
                &position,
                &team_id,
                &player.jersey_number,
                &height_inches,
                &player.weight,
                &college,
                &status,
            ],
        )?;

        Ok(())
    }

    fn upsert_games_batch(&mut self, games: &[Game]) -> Result<()> {
        if self.config.dry_run {
            return Ok(());
        }

        for game in games {
            self.upsert_game(game)?;
        }

        Ok(())
    }

    fn upsert_game(&mut self, game: &Game) -> Result<()> {
        let home_team_abbr = transformer::normalize_team_abbr(&game.home_team);
        let away_team_abbr = transformer::normalize_team_abbr(&game.away_team);

        let home_team_id = self.database.get_team_id_by_abbr(&home_team_abbr)?;
        let away_team_id = self.database.get_team_id_by_abbr(&away_team_abbr)?;

        if home_team_id.is_none() {
            warn!("Home team {} not found", home_team_abbr);
            return Ok(());
        }
        if away_team_id.is_none() {
            warn!("Away team {} not found", away_team_abbr);
            return Ok(());
        }

        let client = self.database.get_client();
        client.execute(
            "INSERT INTO games (id, nfl_game_id, season, week, game_date, home_team_id, away_team_id, home_score, away_score, status, created_at, updated_at)
             VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
             ON CONFLICT (nfl_game_id) DO UPDATE SET
                 home_score = EXCLUDED.home_score,
                 away_score = EXCLUDED.away_score,
                 status = EXCLUDED.status,
                 updated_at = NOW()",
            &[
                &game.game_id,
                &game.season,
                &game.week,
                &game.gameday,
                &home_team_id,
                &away_team_id,
                &game.home_score,
                &game.away_score,
                &"final",
            ],
        )?;

        Ok(())
    }

    fn upsert_stats_batch(&mut self, stats: &[PlayerStat]) -> Result<()> {
        if self.config.dry_run {
            return Ok(());
        }

        for stat in stats {
            if let Err(e) = self.upsert_stat(stat) {
                warn!("Failed to upsert stat for {}: {}", stat.player_display_name.as_ref().unwrap_or(&"unknown".to_string()), e);
            }
        }

        Ok(())
    }

    fn upsert_stat(&mut self, stat: &PlayerStat) -> Result<()> {
        // Get player ID
        let player_id = self.database.get_player_id_by_nfl_id(&stat.player_id)?;

        if player_id.is_none() {
            // Player not found, skip
            return Ok(());
        }

        let client = self.database.get_client();
        client.execute(
            "INSERT INTO game_stats (id, player_id, season, week, passing_yards, rushing_yards, receiving_yards, passing_tds, rushing_tds, receiving_tds, receptions, targets, attempts, completions, interceptions, created_at, updated_at)
             VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW(), NOW())
             ON CONFLICT (player_id, season, week) DO UPDATE SET
                 passing_yards = EXCLUDED.passing_yards,
                 rushing_yards = EXCLUDED.rushing_yards,
                 receiving_yards = EXCLUDED.receiving_yards,
                 passing_tds = EXCLUDED.passing_tds,
                 rushing_tds = EXCLUDED.rushing_tds,
                 receiving_tds = EXCLUDED.receiving_tds,
                 receptions = EXCLUDED.receptions,
                 targets = EXCLUDED.targets,
                 attempts = EXCLUDED.attempts,
                 completions = EXCLUDED.completions,
                 interceptions = EXCLUDED.interceptions,
                 updated_at = NOW()",
            &[
                &player_id,
                &stat.season,
                &stat.week,
                &stat.passing_yards.map(|v| v as i32),
                &stat.rushing_yards.map(|v| v as i32),
                &stat.receiving_yards.map(|v| v as i32),
                &stat.passing_tds,
                &stat.rushing_tds,
                &stat.receiving_tds,
                &stat.receptions.map(|v| v as i32),
                &stat.targets.map(|v| v as i32),
                &stat.attempts.map(|v| v as i32),
                &stat.completions.map(|v| v as i32),
                &stat.interceptions,
            ],
        )?;

        Ok(())
    }
}
