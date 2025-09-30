#!/usr/bin/env python3
"""
NFLverse Data Loader (Python version)

Downloads player statistics from nflverse CSV files and loads into PostgreSQL.

Prerequisites:
    pip install pandas psycopg2-binary requests

Usage:
    python scripts/load_nflverse_data.py

Environment variables:
    DATABASE_URL - PostgreSQL connection string (Heroku format)
"""

import os
import sys
import pandas as pd
import psycopg2
from urllib.parse import urlparse

def parse_database_url(url):
    """Parse Heroku DATABASE_URL into connection parameters"""
    result = urlparse(url)
    return {
        'database': result.path[1:],
        'user': result.username,
        'password': result.password,
        'host': result.hostname,
        'port': result.port or 5432
    }

def get_db_connection():
    """Connect to PostgreSQL database"""
    database_url = os.getenv('DATABASE_URL')
    if not database_url:
        print("ERROR: DATABASE_URL environment variable not set")
        sys.exit(1)

    # Heroku uses postgres:// but psycopg2 expects postgresql://
    if database_url.startswith('postgres://'):
        database_url = database_url.replace('postgres://', 'postgresql://', 1)

    print("Connecting to database...")
    conn = psycopg2.connect(database_url, sslmode='require')
    print("Connected!\n")
    return conn

def fetch_nflverse_player_stats(season):
    """Download player stats CSV from nflverse GitHub"""
    url = f"https://github.com/nflverse/nflverse-data/releases/download/player_stats/player_stats_{season}.csv"
    print(f"Fetching player stats for {season} from nflverse...")

    try:
        df = pd.read_csv(url)
        print(f"  Downloaded {len(df)} records")
        return df
    except Exception as e:
        print(f"  ERROR: Failed to download: {e}")
        return None

def aggregate_season_stats(df):
    """Aggregate weekly stats into season totals per player"""
    print("Aggregating weekly stats into season totals...")

    # Group by player and aggregate stats
    agg_stats = df.groupby(['player_id', 'player_display_name', 'position', 'season', 'recent_team']).agg({
        'week': 'count',  # games played
        'passing_yards': 'sum',
        'passing_tds': 'sum',
        'interceptions': 'sum',
        'passing_attempts': 'sum',
        'completions': 'sum',
        'rushing_yards': 'sum',
        'rushing_tds': 'sum',
        'rushing_attempts': 'sum',
        'receiving_yards': 'sum',
        'receiving_tds': 'sum',
        'receptions': 'sum',
        'targets': 'sum'
    }).reset_index()

    # Rename columns
    agg_stats.rename(columns={
        'week': 'games',
        'player_display_name': 'player_name',
        'interceptions': 'passing_ints'
    }, inplace=True)

    print(f"  Aggregated to {len(agg_stats)} player-season records\n")
    return agg_stats

def get_team_mapping(conn):
    """Get mapping of team abbreviations to UUIDs"""
    cursor = conn.cursor()
    cursor.execute("SELECT id, abbreviation FROM teams")
    mapping = {row[1]: row[0] for row in cursor.fetchall()}
    cursor.close()
    return mapping

def get_player_mapping(conn):
    """Get mapping of player names to UUIDs"""
    cursor = conn.cursor()
    cursor.execute("SELECT id, LOWER(name) as name FROM players")
    mapping = {row[1]: row[0] for row in cursor.fetchall()}
    cursor.close()
    return mapping

def load_stats_to_database(conn, stats, team_mapping, player_mapping):
    """Load aggregated stats into player_career_stats table"""
    print("Loading stats into database...")

    cursor = conn.cursor()
    inserted = 0
    updated = 0
    failed = 0

    for _, row in stats.iterrows():
        # Find player by name (case-insensitive)
        player_name = row['player_name'].lower()
        player_id = player_mapping.get(player_name)

        if not player_id:
            # Try without suffixes like Jr., Sr., III
            simplified_name = player_name.split(',')[0].split(' jr')[0].split(' sr')[0].split(' iii')[0]
            player_id = player_mapping.get(simplified_name)

        if not player_id:
            print(f"  ⚠️  Player not found: {row['player_name']}")
            failed += 1
            continue

        # Get team ID
        team_abbr = row['recent_team']
        team_id = team_mapping.get(team_abbr)

        # Upsert career stats
        query = """
            INSERT INTO player_career_stats (
                player_id, season, team_id, games_played,
                passing_yards, passing_tds, passing_ints,
                passing_attempts, passing_completions,
                rushing_yards, rushing_tds, rushing_attempts,
                receiving_yards, receiving_tds, receptions, receiving_targets
            ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            ON CONFLICT (player_id, season)
            DO UPDATE SET
                team_id = EXCLUDED.team_id,
                games_played = EXCLUDED.games_played,
                passing_yards = EXCLUDED.passing_yards,
                passing_tds = EXCLUDED.passing_tds,
                passing_ints = EXCLUDED.passing_ints,
                passing_attempts = EXCLUDED.passing_attempts,
                passing_completions = EXCLUDED.passing_completions,
                rushing_yards = EXCLUDED.rushing_yards,
                rushing_tds = EXCLUDED.rushing_tds,
                rushing_attempts = EXCLUDED.rushing_attempts,
                receiving_yards = EXCLUDED.receiving_yards,
                receiving_tds = EXCLUDED.receiving_tds,
                receptions = EXCLUDED.receptions,
                receiving_targets = EXCLUDED.receiving_targets
        """

        try:
            cursor.execute(query, (
                player_id, int(row['season']), team_id, int(row['games']),
                int(row['passing_yards']), int(row['passing_tds']), int(row['passing_ints']),
                int(row['passing_attempts']), int(row['completions']),
                int(row['rushing_yards']), int(row['rushing_tds']), int(row['rushing_attempts']),
                int(row['receiving_yards']), int(row['receiving_tds']),
                int(row['receptions']), int(row['targets'])
            ))

            if cursor.rowcount > 0:
                inserted += 1
            else:
                updated += 1

        except Exception as e:
            print(f"  ❌ Error for {row['player_name']}: {e}")
            failed += 1

    conn.commit()
    cursor.close()

    print(f"\n✅ Completed: {inserted} inserted, {updated} updated, {failed} failed\n")

def main():
    """Main execution"""
    print("NFLverse Data Loader (Python)")
    print("=" * 50)
    print()

    # Connect to database
    conn = get_db_connection()

    # Get mappings
    print("Loading team and player mappings...")
    team_mapping = get_team_mapping(conn)
    player_mapping = get_player_mapping(conn)
    print(f"  Found {len(team_mapping)} teams, {len(player_mapping)} players\n")

    # Process each season
    seasons = [2023, 2024]
    for season in seasons:
        print(f"\n{'='*50}")
        print(f"Processing {season} Season")
        print('='*50)

        # Fetch data
        df = fetch_nflverse_player_stats(season)
        if df is None:
            continue

        # Aggregate stats
        stats = aggregate_season_stats(df)

        # Load to database
        load_stats_to_database(conn, stats, team_mapping, player_mapping)

    # Close connection
    conn.close()
    print("\n✅ All done!\n")

if __name__ == '__main__':
    main()
