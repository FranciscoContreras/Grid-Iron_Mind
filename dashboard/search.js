// Configuration
const API_BASE_URL = ''; // Relative URLs for both local and production

// State
let currentFilter = 'all';
let searchTimeout = null;
let allResults = { players: [], teams: [] };

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    const searchInput = document.getElementById('searchInput');
    const filterPills = document.querySelectorAll('.filter-pill');

    // Search input handler with debounce
    searchInput.addEventListener('input', (e) => {
        clearTimeout(searchTimeout);
        const query = e.target.value.trim();

        if (query.length === 0) {
            showEmptyState();
            return;
        }

        if (query.length < 2) {
            return; // Wait for at least 2 characters
        }

        searchTimeout = setTimeout(() => {
            performSearch(query);
        }, 300); // Debounce 300ms
    });

    // Filter pill handlers
    filterPills.forEach(pill => {
        pill.addEventListener('click', () => {
            filterPills.forEach(p => p.classList.remove('active'));
            pill.classList.add('active');
            currentFilter = pill.dataset.filter;
            renderResults();
        });
    });

    // Load teams on startup (for quick team search)
    loadTeams();
});

// API Calls
async function performSearch(query) {
    showLoading();

    try {
        // Search both players and teams
        const [playersResult, teamsResult] = await Promise.all([
            searchPlayers(query),
            Promise.resolve(filterTeams(query)) // Teams are already loaded
        ]);

        allResults.players = playersResult;
        allResults.teams = teamsResult;

        renderResults();
    } catch (error) {
        console.error('Search failed:', error);
        showError('Search failed. Please try again.');
    }
}

async function searchPlayers(query) {
    try {
        const response = await fetch(`${API_BASE_URL}/api/v1/players?limit=50`);
        const data = await response.json();
        const players = data.data || [];

        // Filter by name (case-insensitive)
        return players.filter(player =>
            player.name.toLowerCase().includes(query.toLowerCase())
        );
    } catch (error) {
        console.error('Failed to search players:', error);
        return [];
    }
}

async function loadTeams() {
    try {
        const response = await fetch(`${API_BASE_URL}/api/v1/teams`);
        const data = await response.json();
        allResults.teams = data.data || [];
    } catch (error) {
        console.error('Failed to load teams:', error);
    }
}

function filterTeams(query) {
    return allResults.teams.filter(team =>
        team.name.toLowerCase().includes(query.toLowerCase()) ||
        team.abbreviation.toLowerCase().includes(query.toLowerCase()) ||
        team.city.toLowerCase().includes(query.toLowerCase())
    );
}

// Rendering Functions
function renderResults() {
    const resultsGrid = document.getElementById('resultsGrid');
    const resultsCount = document.getElementById('resultsCount');
    const loadingState = document.getElementById('loadingState');
    const emptyState = document.getElementById('emptyState');

    loadingState.style.display = 'none';
    emptyState.style.display = 'none';

    let displayPlayers = [];
    let displayTeams = [];

    if (currentFilter === 'all' || currentFilter === 'players') {
        displayPlayers = allResults.players;
    }
    if (currentFilter === 'all' || currentFilter === 'teams') {
        displayTeams = allResults.teams;
    }

    const totalResults = displayPlayers.length + displayTeams.length;

    if (totalResults === 0) {
        resultsGrid.style.display = 'none';
        emptyState.style.display = 'block';
        document.querySelector('.empty-title').textContent = 'No results found';
        document.querySelector('.empty-text').textContent = 'Try searching with a different name';
        resultsCount.textContent = '0 results';
        return;
    }

    resultsCount.textContent = `${totalResults} result${totalResults !== 1 ? 's' : ''}`;
    resultsGrid.style.display = 'grid';
    resultsGrid.innerHTML = '';

    // Render teams first, then players
    displayTeams.forEach(team => {
        resultsGrid.appendChild(createTeamCard(team));
    });

    displayPlayers.forEach(player => {
        resultsGrid.appendChild(createPlayerCard(player));
    });
}

function createPlayerCard(player) {
    const card = document.createElement('div');
    card.className = 'result-card';
    card.onclick = () => showPlayerDetail(player);

    const position = player.position || 'N/A';
    const jerseyNumber = player.jersey_number || '--';
    const college = player.college || 'N/A';

    card.innerHTML = `
        <div class="card-header">
            <div class="card-logo" style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); display: flex; align-items: center; justify-content: center; color: white; font-weight: 700; font-size: 20px;">
                ${position}
            </div>
            <div class="card-title">
                <div class="card-name">${player.name}</div>
                <div class="card-subtitle">Player • ${position}</div>
            </div>
        </div>
        <div class="card-stats">
            <div class="stat-item">
                <div class="stat-label">Jersey</div>
                <div class="stat-value">${jerseyNumber}</div>
            </div>
            <div class="stat-item">
                <div class="stat-label">College</div>
                <div class="stat-value">${college.length > 10 ? college.substring(0, 10) + '...' : college}</div>
            </div>
            <div class="stat-item">
                <div class="stat-label">Status</div>
                <div class="stat-value">${player.status || 'Active'}</div>
            </div>
        </div>
    `;

    return card;
}

function createTeamCard(team) {
    const card = document.createElement('div');
    card.className = 'result-card';
    card.onclick = () => showTeamDetail(team);

    card.innerHTML = `
        <div class="card-header">
            <div class="card-logo" style="background: ${team.color || '#667eea'}; display: flex; align-items: center; justify-content: center; color: white; font-weight: 700; font-size: 18px;">
                ${team.abbreviation}
            </div>
            <div class="card-title">
                <div class="card-name">${team.name}</div>
                <div class="card-subtitle">Team • ${team.conference || 'NFL'}</div>
            </div>
        </div>
        <div class="card-stats">
            <div class="stat-item">
                <div class="stat-label">Division</div>
                <div class="stat-value">${team.division || 'N/A'}</div>
            </div>
            <div class="stat-item">
                <div class="stat-label">Stadium</div>
                <div class="stat-value">${team.stadium_name ? team.stadium_name.substring(0, 12) + (team.stadium_name.length > 12 ? '...' : '') : 'N/A'}</div>
            </div>
            <div class="stat-item">
                <div class="stat-label">Surface</div>
                <div class="stat-value">${team.stadium_surface || 'N/A'}</div>
            </div>
        </div>
    `;

    return card;
}

// Detail Views
async function showPlayerDetail(player) {
    hideSearchResults();
    const detailView = document.getElementById('playerDetailView');
    const header = document.getElementById('playerDetailHeader');
    const statsTable = document.getElementById('playerStatsTable');
    const gameHistory = document.getElementById('playerGameHistory');

    detailView.classList.add('active');

    // Header
    header.innerHTML = `
        <div style="display: flex; align-items: center; gap: 24px;">
            <div style="width: 80px; height: 80px; border-radius: 50%; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); display: flex; align-items: center; justify-content: center; color: white; font-weight: 700; font-size: 28px; border: 3px solid rgba(255, 255, 255, 0.5);">
                ${player.position || 'P'}
            </div>
            <div style="flex: 1;">
                <h2 style="font-size: 32px; font-weight: 700; margin-bottom: 8px;">${player.name}</h2>
                <div style="font-size: 16px; color: var(--soft-text-light);">
                    ${player.position || 'N/A'} • #${player.jersey_number || '--'} • ${player.college || 'N/A'}
                </div>
            </div>
        </div>
    `;

    // Load career stats
    try {
        const response = await fetch(`${API_BASE_URL}/api/v1/players/${player.id}/career`);
        const data = await response.json();
        const career = data.data;

        if (career && career.seasons && career.seasons.length > 0) {
            renderPlayerStats(statsTable, career.seasons);
            renderGameHistory(gameHistory, career.seasons);
        } else {
            statsTable.innerHTML = '<tbody><tr><td style="text-align: center; padding: 32px;">No stats available</td></tr></tbody>';
            gameHistory.innerHTML = '<p style="text-align: center; padding: 32px; color: var(--soft-text-light);">No game history available</p>';
        }
    } catch (error) {
        console.error('Failed to load player stats:', error);
        statsTable.innerHTML = '<tbody><tr><td style="text-align: center; padding: 32px;">Failed to load stats</td></tr></tbody>';
    }
}

function renderPlayerStats(table, seasons) {
    // Show most recent season
    const latestSeason = seasons[0];

    table.innerHTML = `
        <thead>
            <tr>
                <th>Season</th>
                <th>GP</th>
                <th>Pass Yds</th>
                <th>Pass TD</th>
                <th>Rush Yds</th>
                <th>Rush TD</th>
                <th>Rec</th>
                <th>Rec Yds</th>
                <th>Rec TD</th>
            </tr>
        </thead>
        <tbody>
            <tr>
                <td>${latestSeason.season}</td>
                <td>${latestSeason.games_played || 0}</td>
                <td>${latestSeason.passing_yards || 0}</td>
                <td>${latestSeason.passing_touchdowns || 0}</td>
                <td>${latestSeason.rushing_yards || 0}</td>
                <td>${latestSeason.rushing_touchdowns || 0}</td>
                <td>${latestSeason.receptions || 0}</td>
                <td>${latestSeason.receiving_yards || 0}</td>
                <td>${latestSeason.receiving_touchdowns || 0}</td>
            </tr>
        </tbody>
    `;
}

function renderGameHistory(container, seasons) {
    container.innerHTML = '';

    seasons.forEach(season => {
        const seasonCard = document.createElement('div');
        seasonCard.style.cssText = `
            background: linear-gradient(135deg, rgba(255, 255, 255, 0.25) 0%, rgba(255, 255, 255, 0.15) 100%);
            backdrop-filter: blur(40px) saturate(200%);
            -webkit-backdrop-filter: blur(40px) saturate(200%);
            border: 1px solid rgba(255, 255, 255, 0.3);
            border-radius: var(--radius-2xl);
            padding: 24px;
            margin-bottom: 16px;
        `;

        seasonCard.innerHTML = `
            <h3 style="font-size: 18px; font-weight: 600; margin-bottom: 12px;">
                ${season.season} Season
            </h3>
            <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(120px, 1fr)); gap: 16px;">
                <div>
                    <div style="font-size: 12px; color: var(--soft-text-lighter); margin-bottom: 4px;">Games Played</div>
                    <div style="font-size: 20px; font-weight: 600;">${season.games_played || 0}</div>
                </div>
                <div>
                    <div style="font-size: 12px; color: var(--soft-text-lighter); margin-bottom: 4px;">Pass Yards</div>
                    <div style="font-size: 20px; font-weight: 600;">${season.passing_yards || 0}</div>
                </div>
                <div>
                    <div style="font-size: 12px; color: var(--soft-text-lighter); margin-bottom: 4px;">Rush Yards</div>
                    <div style="font-size: 20px; font-weight: 600;">${season.rushing_yards || 0}</div>
                </div>
                <div>
                    <div style="font-size: 12px; color: var(--soft-text-lighter); margin-bottom: 4px;">Rec Yards</div>
                    <div style="font-size: 20px; font-weight: 600;">${season.receiving_yards || 0}</div>
                </div>
            </div>
        `;

        container.appendChild(seasonCard);
    });
}

async function showTeamDetail(team) {
    hideSearchResults();
    const detailView = document.getElementById('teamDetailView');
    const header = document.getElementById('teamDetailHeader');
    const scheduleTable = document.getElementById('teamScheduleTable');
    const pastSeasons = document.getElementById('teamPastSeasons');

    detailView.classList.add('active');

    // Header
    header.innerHTML = `
        <div style="display: flex; align-items: center; gap: 24px;">
            <div style="width: 80px; height: 80px; border-radius: 50%; background: ${team.color || '#667eea'}; display: flex; align-items: center; justify-content: center; color: white; font-weight: 700; font-size: 24px; border: 3px solid rgba(255, 255, 255, 0.5);">
                ${team.abbreviation}
            </div>
            <div style="flex: 1;">
                <h2 style="font-size: 32px; font-weight: 700; margin-bottom: 8px;">${team.name}</h2>
                <div style="font-size: 16px; color: var(--soft-text-light);">
                    ${team.conference || 'NFL'} • ${team.division || 'Division'} • ${team.stadium_name || 'Stadium'}
                </div>
            </div>
        </div>
    `;

    // Load current season games
    try {
        const response = await fetch(`${API_BASE_URL}/api/v1/games?season=2025`);
        const data = await response.json();
        const games = data.data || [];

        const teamGames = games.filter(game =>
            game.home_team_id === team.id || game.away_team_id === team.id
        );

        renderTeamSchedule(scheduleTable, teamGames, team);
        renderPastSeasons(pastSeasons, team);
    } catch (error) {
        console.error('Failed to load team games:', error);
        scheduleTable.innerHTML = '<tbody><tr><td style="text-align: center; padding: 32px;">Failed to load schedule</td></tr></tbody>';
    }
}

function renderTeamSchedule(table, games, team) {
    if (games.length === 0) {
        table.innerHTML = '<tbody><tr><td style="text-align: center; padding: 32px;">No games scheduled</td></tr></tbody>';
        return;
    }

    const rows = games.slice(0, 10).map(game => {
        const isHome = game.home_team_id === team.id;
        const opponent = isHome ? 'vs Away' : '@ Home';
        const score = game.home_score && game.away_score
            ? `${game.home_score} - ${game.away_score}`
            : 'Scheduled';

        return `
            <tr>
                <td>${new Date(game.game_date).toLocaleDateString()}</td>
                <td>${opponent}</td>
                <td>${score}</td>
                <td>${game.status || 'Scheduled'}</td>
            </tr>
        `;
    }).join('');

    table.innerHTML = `
        <thead>
            <tr>
                <th>Date</th>
                <th>Matchup</th>
                <th>Score</th>
                <th>Status</th>
            </tr>
        </thead>
        <tbody>${rows}</tbody>
    `;
}

function renderPastSeasons(container, team) {
    const seasons = [2024, 2023, 2022, 2021];

    container.innerHTML = '';

    seasons.forEach(season => {
        const seasonCard = document.createElement('div');
        seasonCard.style.cssText = `
            background: linear-gradient(135deg, rgba(255, 255, 255, 0.25) 0%, rgba(255, 255, 255, 0.15) 100%);
            backdrop-filter: blur(40px) saturate(200%);
            -webkit-backdrop-filter: blur(40px) saturate(200%);
            border: 1px solid rgba(255, 255, 255, 0.3);
            border-radius: var(--radius-2xl);
            padding: 24px;
            margin-bottom: 16px;
            cursor: pointer;
            transition: all 0.3s ease;
        `;

        seasonCard.innerHTML = `
            <h3 style="font-size: 18px; font-weight: 600; margin-bottom: 8px;">
                ${season} Season
            </h3>
            <p style="color: var(--soft-text-light); font-size: 14px;">
                Click to load season details
            </p>
        `;

        seasonCard.onclick = async () => {
            // Load season games
            try {
                const response = await fetch(`${API_BASE_URL}/api/v1/games?season=${season}`);
                const data = await response.json();
                const games = data.data || [];

                const teamGames = games.filter(game =>
                    game.home_team_id === team.id || game.away_team_id === team.id
                );

                const wins = teamGames.filter(game => {
                    const isHome = game.home_team_id === team.id;
                    return (isHome && game.home_score > game.away_score) ||
                           (!isHome && game.away_score > game.home_score);
                }).length;

                seasonCard.innerHTML = `
                    <h3 style="font-size: 18px; font-weight: 600; margin-bottom: 12px;">
                        ${season} Season
                    </h3>
                    <div style="display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px;">
                        <div>
                            <div style="font-size: 12px; color: var(--soft-text-lighter); margin-bottom: 4px;">Games Played</div>
                            <div style="font-size: 20px; font-weight: 600;">${teamGames.length}</div>
                        </div>
                        <div>
                            <div style="font-size: 12px; color: var(--soft-text-lighter); margin-bottom: 4px;">Wins</div>
                            <div style="font-size: 20px; font-weight: 600;">${wins}</div>
                        </div>
                        <div>
                            <div style="font-size: 12px; color: var(--soft-text-lighter); margin-bottom: 4px;">Losses</div>
                            <div style="font-size: 20px; font-weight: 600;">${teamGames.length - wins}</div>
                        </div>
                    </div>
                `;
            } catch (error) {
                console.error('Failed to load season:', error);
            }
        };

        container.appendChild(seasonCard);
    });
}

// UI State Management
function showLoading() {
    document.getElementById('loadingState').style.display = 'block';
    document.getElementById('emptyState').style.display = 'none';
    document.getElementById('resultsGrid').style.display = 'none';
}

function showEmptyState() {
    document.getElementById('loadingState').style.display = 'none';
    document.getElementById('emptyState').style.display = 'block';
    document.getElementById('resultsGrid').style.display = 'none';
    document.querySelector('.empty-title').textContent = 'Ready to explore';
    document.querySelector('.empty-text').textContent = 'Start typing to search for your favorite players and teams';
    document.getElementById('resultsCount').textContent = 'Search for players or teams to get started';
}

function showError(message) {
    document.getElementById('loadingState').style.display = 'none';
    document.getElementById('emptyState').style.display = 'block';
    document.getElementById('resultsGrid').style.display = 'none';
    document.querySelector('.empty-title').textContent = 'Error';
    document.querySelector('.empty-text').textContent = message;
}

function hideSearchResults() {
    document.getElementById('resultsContainer').style.display = 'none';
}

function goBack() {
    document.getElementById('playerDetailView').classList.remove('active');
    document.getElementById('teamDetailView').classList.remove('active');
    document.getElementById('resultsContainer').style.display = 'block';
}
