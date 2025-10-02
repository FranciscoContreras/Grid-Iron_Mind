use anyhow::Result;
use env_logger;
use log::{info, error};

mod config;
mod downloader;
mod parser;
mod transformer;
mod database;
mod sync;
mod validator;

use config::Config;
use sync::DataPipeline;

fn main() -> Result<()> {
    // Initialize logging
    env_logger::init();

    // Load configuration
    dotenv::dotenv().ok();
    let config = Config::from_env()?;

    info!("🏈 NFL Data Pipeline Starting");
    info!("Mode: {}", config.mode);
    info!("Year range: {}-{}", config.start_year, config.end_year);

    // Create pipeline
    let mut pipeline = DataPipeline::new(config)?;

    // Execute based on mode
    match pipeline.config.mode.as_str() {
        "full" => {
            info!("📥 Full import: {} seasons", pipeline.config.end_year - pipeline.config.start_year + 1);
            pipeline.run_full_import()?;
        },
        "year" => {
            info!("📅 Single year import: {}", pipeline.config.year);
            pipeline.import_year(pipeline.config.year)?;
        },
        "update" => {
            info!("🔄 Incremental update");
            pipeline.run_update()?;
        },
        "validate" => {
            info!("✅ Validating existing data");
            pipeline.validate_data()?;
        },
        _ => {
            error!("Invalid mode: {}", pipeline.config.mode);
            std::process::exit(1);
        }
    }

    info!("✅ Pipeline completed successfully!");
    Ok(())
}
