use anyhow::{Result, Context};
use postgres::Client;
use postgres_native_tls::MakeTlsConnector;
use native_tls::TlsConnector;
use log::info;

pub struct Database {
    client: Client,
}

impl Database {
    pub fn connect(database_url: &str) -> Result<Self> {
        info!("Connecting to database...");

        // Create TLS connector for Heroku PostgreSQL
        let connector = TlsConnector::builder()
            .danger_accept_invalid_certs(true) // Heroku uses self-signed certs
            .build()
            .context("Failed to create TLS connector")?;
        let connector = MakeTlsConnector::new(connector);

        let client = Client::connect(database_url, connector)
            .context("Failed to connect to database")?;

        info!("✅ Database connected");
        Ok(Database { client })
    }

    pub fn get_client(&mut self) -> &mut Client {
        &mut self.client
    }

    /// Check if database connection is healthy
    pub fn health_check(&mut self) -> Result<()> {
        self.client
            .query("SELECT 1", &[])
            .context("Health check failed")?;
        Ok(())
    }

    /// Get team ID by abbreviation
    pub fn get_team_id_by_abbr(&mut self, abbr: &str) -> Result<Option<uuid::Uuid>> {
        let row = self.client
            .query_opt(
                "SELECT id FROM teams WHERE abbreviation = $1",
                &[&abbr],
            )?;

        Ok(row.map(|r| r.get(0)))
    }

    /// Get player ID by NFL ID (gsis_id)
    pub fn get_player_id_by_nfl_id(&mut self, nfl_id: &str) -> Result<Option<uuid::Uuid>> {
        let row = self.client
            .query_opt(
                "SELECT id FROM players WHERE nfl_id = $1",
                &[&nfl_id],
            )?;

        Ok(row.map(|r| r.get(0)))
    }

    /// Get import progress status for a season and data type
    pub fn get_import_progress(&mut self, season: i32, data_type: &str) -> Result<Option<String>> {
        let row = self.client
            .query_opt(
                "SELECT status FROM import_progress WHERE season = $1 AND data_type = $2",
                &[&season, &data_type],
            )?;

        Ok(row.map(|r| r.get(0)))
    }

    /// Mark import progress
    pub fn mark_progress(
        &mut self,
        season: i32,
        data_type: &str,
        status: &str,
        records_imported: i32,
    ) -> Result<()> {
        self.client.execute(
            "INSERT INTO import_progress (season, data_type, status, records_imported, started_at, completed_at)
             VALUES ($1, $2, $3, $4, NOW(), CASE WHEN $3 = 'completed' THEN NOW() ELSE NULL END)
             ON CONFLICT (season, data_type)
             DO UPDATE SET
                 status = EXCLUDED.status,
                 records_imported = EXCLUDED.records_imported,
                 completed_at = EXCLUDED.completed_at",
            &[&season, &data_type, &status, &records_imported],
        )?;

        Ok(())
    }

    /// Get count of games for a season
    pub fn count_games(&mut self, season: i32) -> Result<i64> {
        let row = self.client
            .query_one(
                "SELECT COUNT(*) FROM games WHERE season = $1",
                &[&season],
            )?;

        Ok(row.get(0))
    }

    /// Get count of players for a season
    pub fn count_players(&mut self) -> Result<i64> {
        let row = self.client
            .query_one("SELECT COUNT(*) FROM players", &[])?;

        Ok(row.get(0))
    }

    /// Get count of game stats for a season
    pub fn count_game_stats(&mut self, season: i32) -> Result<i64> {
        let row = self.client
            .query_one(
                "SELECT COUNT(*) FROM game_stats WHERE season = $1",
                &[&season],
            )?;

        Ok(row.get(0))
    }
}
