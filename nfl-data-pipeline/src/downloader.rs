use anyhow::{Result, anyhow};
use log::warn;
use std::time::Duration;
use reqwest::blocking::Client;

pub struct Downloader {
    client: Client,
    max_retries: u32,
}

impl Downloader {
    pub fn new(max_retries: u32) -> Self {
        let client = Client::builder()
            .timeout(Duration::from_secs(60))
            .build()
            .expect("Failed to create HTTP client");

        Downloader {
            client,
            max_retries,
        }
    }

    /// Download player stats CSV for a given year
    pub fn download_player_stats(&self, year: i32) -> Result<String> {
        let url = format!(
            "https://github.com/nflverse/nflverse-data/releases/download/player_stats/player_stats_{}.csv",
            year
        );
        self.download_with_retry(&url)
    }

    /// Download roster CSV for a given year
    pub fn download_roster(&self, year: i32) -> Result<String> {
        let url = format!(
            "https://github.com/nflverse/nflverse-data/releases/download/rosters/roster_{}.csv",
            year
        );
        self.download_with_retry(&url)
    }

    /// Download schedule CSV for a given year
    pub fn download_schedule(&self, year: i32) -> Result<String> {
        let url = format!(
            "https://github.com/nflverse/nflverse-data/releases/download/schedules/sched_{}.csv",
            year
        );
        self.download_with_retry(&url)
    }

    /// Download Next Gen Stats (passing) for a given year
    pub fn download_ngs_passing(&self, year: i32) -> Result<String> {
        if year < 2016 {
            return Err(anyhow!("NGS data only available from 2016 onwards"));
        }
        let url = format!(
            "https://github.com/nflverse/nflverse-data/releases/download/nextgen_stats/ngs_{}_passing.csv",
            year
        );
        self.download_with_retry(&url)
    }

    /// Download Next Gen Stats (rushing) for a given year
    pub fn download_ngs_rushing(&self, year: i32) -> Result<String> {
        if year < 2016 {
            return Err(anyhow!("NGS data only available from 2016 onwards"));
        }
        let url = format!(
            "https://github.com/nflverse/nflverse-data/releases/download/nextgen_stats/ngs_{}_rushing.csv",
            year
        );
        self.download_with_retry(&url)
    }

    /// Download Next Gen Stats (receiving) for a given year
    pub fn download_ngs_receiving(&self, year: i32) -> Result<String> {
        if year < 2016 {
            return Err(anyhow!("NGS data only available from 2016 onwards"));
        }
        let url = format!(
            "https://github.com/nflverse/nflverse-data/releases/download/nextgen_stats/ngs_{}_receiving.csv",
            year
        );
        self.download_with_retry(&url)
    }

    /// Download with automatic retries
    fn download_with_retry(&self, url: &str) -> Result<String> {
        let mut last_error = None;

        for attempt in 1..=self.max_retries {
            match self.client.get(url).send() {
                Ok(response) => {
                    if response.status().is_success() {
                        return response
                            .text()
                            .map_err(|e| anyhow!("Failed to read response: {}", e));
                    } else if response.status() == 404 {
                        return Err(anyhow!("Data not found (404): {}", url));
                    } else {
                        warn!(
                            "HTTP {} for {}, attempt {}/{}",
                            response.status(),
                            url,
                            attempt,
                            self.max_retries
                        );
                        last_error = Some(anyhow!("HTTP {}", response.status()));
                    }
                }
                Err(e) => {
                    warn!(
                        "Request failed for {}: {}, attempt {}/{}",
                        url, e, attempt, self.max_retries
                    );
                    last_error = Some(anyhow!("Request error: {}", e));
                }
            }

            // Exponential backoff
            if attempt < self.max_retries {
                std::thread::sleep(Duration::from_secs(2u64.pow(attempt)));
            }
        }

        Err(last_error.unwrap_or_else(|| anyhow!("Download failed after {} retries", self.max_retries)))
    }
}
