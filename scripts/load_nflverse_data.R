#!/usr/bin/env Rscript

# NFLverse Data Loader
# Fetches player statistics from nflverse and loads into PostgreSQL database
#
# Prerequisites:
#   install.packages(c("nflreadr", "DBI", "RPostgres", "dplyr", "tidyr"))
#
# Usage:
#   Rscript scripts/load_nflverse_data.R
#
# Environment variables:
#   DATABASE_URL - PostgreSQL connection string (Heroku format)

library(nflreadr)
library(DBI)
library(RPostgres)
library(dplyr)
library(tidyr)

# Parse Heroku DATABASE_URL
parse_database_url <- function(url) {
  # Format: postgres://user:password@host:port/database
  pattern <- "postgres://([^:]+):([^@]+)@([^:]+):([^/]+)/(.*)"
  matches <- regmatches(url, regexec(pattern, url))[[1]]

  if (length(matches) == 0) {
    stop("Invalid DATABASE_URL format")
  }

  list(
    user = matches[2],
    password = matches[3],
    host = matches[4],
    port = as.integer(matches[5]),
    dbname = matches[6]
  )
}

# Get database connection
get_db_connection <- function() {
  database_url <- Sys.getenv("DATABASE_URL")

  if (database_url == "") {
    stop("DATABASE_URL environment variable not set")
  }

  db_params <- parse_database_url(database_url)

  dbConnect(
    RPostgres::Postgres(),
    dbname = db_params$dbname,
    host = db_params$host,
    port = db_params$port,
    user = db_params$user,
    password = db_params$password,
    sslmode = "require"
  )
}

# Load player stats from nflverse
load_player_stats <- function(seasons = 2020:2024) {
  cat("Fetching player stats from nflverse...\n")

  # Load player stats - this uses nflverse's weekly stats
  stats <- load_player_stats(seasons) %>%
    group_by(player_id, player_name, position, season, recent_team) %>%
    summarise(
      games = n(),
      passing_yards = sum(passing_yards, na.rm = TRUE),
      passing_tds = sum(passing_tds, na.rm = TRUE),
      passing_ints = sum(interceptions, na.rm = TRUE),
      rushing_yards = sum(rushing_yards, na.rm = TRUE),
      rushing_tds = sum(rushing_tds, na.rm = TRUE),
      receiving_yards = sum(receiving_yards, na.rm = TRUE),
      receiving_tds = sum(receiving_tds, na.rm = TRUE),
      receptions = sum(receptions, na.rm = TRUE),
      targets = sum(targets, na.rm = TRUE),
      .groups = "drop"
    )

  cat(sprintf("Fetched %d player-season records\n", nrow(stats)))
  return(stats)
}

# Map nflverse team abbreviations to our database team IDs
get_team_mapping <- function(conn) {
  teams <- dbGetQuery(conn, "SELECT id, abbreviation FROM teams")
  setNames(teams$id, teams$abbreviation)
}

# Map nflverse player IDs to our database player UUIDs
get_player_mapping <- function(conn) {
  # Our database uses nfl_id which is the ESPN athlete ID
  # nflverse uses gsis_id as player_id
  # We need to match by name as a fallback
  players <- dbGetQuery(conn, "SELECT id, nfl_id, name FROM players")
  players
}

# Load career stats into database
load_career_stats_to_db <- function(conn, stats) {
  cat("Loading career stats into database...\n")

  team_mapping <- get_team_mapping(conn)
  players_db <- get_player_mapping(conn)

  inserted <- 0
  failed <- 0

  for (i in 1:nrow(stats)) {
    row <- stats[i, ]

    # Try to find player by name (fuzzy matching would be better)
    player_match <- players_db %>%
      filter(tolower(name) == tolower(row$player_name))

    if (nrow(player_match) == 0) {
      cat(sprintf("Player not found: %s\n", row$player_name))
      failed <- failed + 1
      next
    }

    player_id <- player_match$id[1]

    # Get team ID
    team_abbr <- row$recent_team
    team_id <- team_mapping[team_abbr]

    if (is.na(team_id) || is.null(team_id)) {
      team_id <- NULL
    }

    # Upsert career stats
    query <- "
      INSERT INTO player_career_stats (
        player_id, season, team_id, games_played,
        passing_yards, passing_tds, passing_ints,
        rushing_yards, rushing_tds,
        receiving_yards, receiving_tds, receptions, receiving_targets
      ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
      ON CONFLICT (player_id, season)
      DO UPDATE SET
        team_id = EXCLUDED.team_id,
        games_played = EXCLUDED.games_played,
        passing_yards = EXCLUDED.passing_yards,
        passing_tds = EXCLUDED.passing_tds,
        passing_ints = EXCLUDED.passing_ints,
        rushing_yards = EXCLUDED.rushing_yards,
        rushing_tds = EXCLUDED.rushing_tds,
        receiving_yards = EXCLUDED.receiving_yards,
        receiving_tds = EXCLUDED.receiving_tds,
        receptions = EXCLUDED.receptions,
        receiving_targets = EXCLUDED.receiving_targets
    "

    tryCatch({
      dbExecute(
        conn, query,
        params = list(
          player_id, row$season, team_id, row$games,
          row$passing_yards, row$passing_tds, row$passing_ints,
          row$rushing_yards, row$rushing_tds,
          row$receiving_yards, row$receiving_tds, row$receptions, row$targets
        )
      )
      inserted <- inserted + 1
    }, error = function(e) {
      cat(sprintf("Error inserting stats for %s: %s\n", row$player_name, e$message))
      failed <- failed + 1
    })
  }

  cat(sprintf("\nCompleted: %d inserted, %d failed\n", inserted, failed))
}

# Main execution
main <- function() {
  cat("NFLverse Data Loader\n")
  cat("===================\n\n")

  # Connect to database
  cat("Connecting to database...\n")
  conn <- get_db_connection()
  cat("Connected!\n\n")

  # Load stats from nflverse
  stats <- load_player_stats(seasons = 2023:2024)

  # Load into database
  load_career_stats_to_db(conn, stats)

  # Disconnect
  dbDisconnect(conn)
  cat("\nDone!\n")
}

# Run if called as script
if (!interactive()) {
  main()
}
