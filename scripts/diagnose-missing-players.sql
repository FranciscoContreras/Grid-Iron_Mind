-- Grid Iron Mind - Player Diagnostic Script
-- Checks for missing top fantasy players and provides database statistics

\echo '=== PLAYER DATABASE DIAGNOSTIC ==='
\echo ''

-- Total player count
\echo '=== Total Player Count ==='
SELECT
    COUNT(*) as total_players,
    COUNT(CASE WHEN status = 'active' THEN 1 END) as active_players,
    COUNT(CASE WHEN status = 'inactive' THEN 1 END) as inactive_players
FROM players;

\echo ''
\echo '=== Players by Position ==='
SELECT
    position,
    COUNT(*) as count
FROM players
WHERE status = 'active'
GROUP BY position
ORDER BY count DESC;

\echo ''
\echo '=== Recent Players (Last 30 days) ==='
SELECT COUNT(*) as recently_updated
FROM players
WHERE updated_at > NOW() - INTERVAL '30 days';

\echo ''
\echo '=== Top Fantasy Players Check (2024-2025 Season) ==='
\echo 'Searching for top fantasy performers...'
\echo ''

-- Check for specific top players
WITH top_players AS (
    SELECT unnest(ARRAY[
        'Saquon Barkley',
        'Lamar Jackson',
        'Josh Allen',
        'Jalen Hurts',
        'Derrick Henry',
        'Joe Burrow',
        'Ja''Marr Chase',
        'Amon-Ra St. Brown',
        'Justin Jefferson',
        'CeeDee Lamb',
        'Tyreek Hill',
        'Travis Kelce',
        'Sam LaPorta',
        'Christian McCaffrey',
        'Bijan Robinson',
        'Breece Hall',
        'Jahmyr Gibbs',
        'De''Von Achane',
        'Patrick Mahomes',
        'Kyler Murray',
        'A.J. Brown',
        'Nico Collins',
        'Puka Nacua',
        'Cooper Kupp',
        'Mike Evans',
        'Garrett Wilson',
        'Drake London',
        'Deebo Samuel',
        'George Kittle',
        'Trey McBride'
    ]) AS expected_name
)
SELECT
    tp.expected_name,
    CASE
        WHEN p.name IS NOT NULL THEN '✓ FOUND: ' || p.name
        ELSE '✗ MISSING'
    END as status,
    p.position,
    t.abbreviation as team
FROM top_players tp
LEFT JOIN players p ON (
    LOWER(p.name) = LOWER(tp.expected_name) OR
    p.name ILIKE '%' || split_part(tp.expected_name, ' ', 2) || '%'
)
LEFT JOIN teams t ON p.team_id = t.id
ORDER BY
    CASE WHEN p.name IS NULL THEN 0 ELSE 1 END DESC,
    tp.expected_name;

\echo ''
\echo '=== Missing Players Summary ==='
WITH top_players AS (
    SELECT unnest(ARRAY[
        'Saquon Barkley', 'Lamar Jackson', 'Josh Allen', 'Jalen Hurts',
        'Derrick Henry', 'Joe Burrow', 'Ja''Marr Chase', 'Amon-Ra St. Brown',
        'Justin Jefferson', 'CeeDee Lamb', 'Tyreek Hill', 'Travis Kelce',
        'Sam LaPorta', 'Christian McCaffrey', 'Bijan Robinson', 'Breece Hall',
        'Jahmyr Gibbs', 'De''Von Achane', 'Patrick Mahomes', 'Kyler Murray',
        'A.J. Brown', 'Nico Collins', 'Puka Nacua', 'Cooper Kupp',
        'Mike Evans', 'Garrett Wilson', 'Drake London', 'Deebo Samuel',
        'George Kittle', 'Trey McBride'
    ]) AS expected_name
)
SELECT
    COUNT(*) as total_checked,
    COUNT(p.name) as found_count,
    COUNT(*) - COUNT(p.name) as missing_count,
    ROUND(100.0 * COUNT(p.name) / COUNT(*), 1) || '%' as found_percentage
FROM top_players tp
LEFT JOIN players p ON (
    LOWER(p.name) = LOWER(tp.expected_name) OR
    p.name ILIKE '%' || split_part(tp.expected_name, ' ', 2) || '%'
);

\echo ''
\echo '=== Players with Similar Names (Potential Mismatches) ==='
SELECT name, position, status
FROM players
WHERE
    name ILIKE '%barkley%' OR
    name ILIKE '%jackson%' OR
    name ILIKE '%allen%' OR
    name ILIKE '%hurts%'
ORDER BY name;

\echo ''
\echo '=== Team Roster Completeness ==='
SELECT
    t.name as team,
    t.abbreviation,
    COUNT(p.id) as player_count
FROM teams t
LEFT JOIN players p ON p.team_id = t.id
GROUP BY t.id, t.name, t.abbreviation
ORDER BY player_count ASC;

\echo ''
\echo '=== Diagnostic Complete ==='
\echo 'If players are missing, run: make sync-update'
\echo 'Or trigger roster sync: POST /api/v1/admin/sync/rosters'
