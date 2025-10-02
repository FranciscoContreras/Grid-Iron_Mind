use anyhow::{Result, Context};
use std::env;

#[derive(Debug, Clone)]
pub struct Config {
    pub database_url: String,
    pub mode: String,
    pub year: i32,
    pub start_year: i32,
    pub end_year: i32,
    pub dry_run: bool,
    pub batch_size: usize,
    pub max_retries: u32,
}

impl Config {
    pub fn from_env() -> Result<Self> {
        // Parse command line arguments
        let args: Vec<String> = env::args().collect();

        let mode = Self::get_arg(&args, "--mode").unwrap_or_else(|| "full".to_string());
        let year = Self::get_arg(&args, "--year")
            .and_then(|s| s.parse().ok())
            .unwrap_or(2024);
        let start_year = Self::get_arg(&args, "--start-year")
            .and_then(|s| s.parse().ok())
            .unwrap_or(2010);
        let end_year = Self::get_arg(&args, "--end-year")
            .and_then(|s| s.parse().ok())
            .unwrap_or(2025);
        let dry_run = args.contains(&"--dry-run".to_string());

        let database_url = env::var("DATABASE_URL")
            .context("DATABASE_URL must be set in environment")?;

        Ok(Config {
            database_url,
            mode,
            year,
            start_year,
            end_year,
            dry_run,
            batch_size: 500,
            max_retries: 3,
        })
    }

    fn get_arg(args: &[String], key: &str) -> Option<String> {
        args.iter()
            .position(|arg| arg == key)
            .and_then(|i| args.get(i + 1))
            .map(|s| s.clone())
    }
}
