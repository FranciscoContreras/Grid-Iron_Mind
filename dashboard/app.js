// Configuration
const API_BASE_URL = ''; // Use relative URLs - works for both local and production
const CACHE_DURATION = 5 * 60 * 1000; // 5 minutes

// State
let state = {
    players: [],
    teams: [],
    currentPage: 1,
    playersPerPage: 20,
    totalPlayers: 0,
    filters: {
        search: '',
        position: '',
        status: '',
        team: ''
    }
};

// Cache
const cache = {
    get(key) {
        const item = localStorage.getItem(key);
        if (!item) return null;

        const { data, timestamp } = JSON.parse(item);
        if (Date.now() - timestamp > CACHE_DURATION) {
            localStorage.removeItem(key);
            return null;
        }
        return data;
    },
    set(key, data) {
        localStorage.setItem(key, JSON.stringify({
            data,
            timestamp: Date.now()
        }));
    }
};

// API Helper
async function apiCall(endpoint, params = {}) {
    const startTime = Date.now();

    try {
        let url = `${API_BASE_URL}${endpoint}`;

        // Add query parameters
        if (Object.keys(params).length > 0) {
            const queryString = new URLSearchParams(params).toString();
            url += `?${queryString}`;
        }

        const response = await fetch(url);
        const responseTime = Date.now() - startTime;

        // Update status indicator
        updateStatus(response.ok, responseTime);

        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        const data = await response.json();
        return { data, status: response.status, responseTime };
    } catch (error) {
        updateStatus(false, Date.now() - startTime);
        throw error;
    }
}

// UI Updates
function updateStatus(connected, responseTime) {
    const indicator = document.getElementById('statusIndicator');
    const statusText = document.getElementById('statusText');
    const responseTimeEl = document.getElementById('responseTime');

    if (connected) {
        indicator.classList.add('connected');
        statusText.textContent = 'Connected';
    } else {
        indicator.classList.remove('connected');
        statusText.textContent = 'Disconnected';
    }

    responseTimeEl.textContent = `${responseTime}ms`;
}

function showLoading(elementId) {
    const el = document.getElementById(elementId);
    if (el) el.style.display = 'block';
}

function hideLoading(elementId) {
    const el = document.getElementById(elementId);
    if (el) el.style.display = 'none';
}

function showError(elementId, message) {
    const el = document.getElementById(elementId);
    if (el) {
        el.textContent = message;
        el.style.display = 'block';
    }
}

function hideError(elementId) {
    const el = document.getElementById(elementId);
    if (el) el.style.display = 'none';
}

// Tab Navigation
function initTabs() {
    const tabs = document.querySelectorAll('.tab');
    const tabContents = document.querySelectorAll('.tab-content');

    tabs.forEach(tab => {
        tab.addEventListener('click', () => {
            const tabName = tab.dataset.tab;

            // Update active tab
            tabs.forEach(t => t.classList.remove('active'));
            tab.classList.add('active');

            // Update active content
            tabContents.forEach(content => {
                content.classList.remove('active');
                if (content.id === `${tabName}-tab`) {
                    content.classList.add('active');
                }
            });
        });
    });
}

// Players Tab
async function loadPlayers() {
    showLoading('playersLoading');
    hideError('playersError');

    try {
        const params = {
            limit: state.playersPerPage,
            offset: (state.currentPage - 1) * state.playersPerPage
        };

        if (state.filters.position) params.position = state.filters.position;
        if (state.filters.status) params.status = state.filters.status;
        if (state.filters.team) params.team = state.filters.team;

        const cacheKey = `players_${JSON.stringify(params)}`;
        let result = cache.get(cacheKey);

        if (!result) {
            result = await apiCall('/api/v1/players', params);
            cache.set(cacheKey, result);
        }

        state.players = result.data.data;
        state.totalPlayers = result.data.meta.total || 0;

        renderPlayers();
        updatePagination();
    } catch (error) {
        showError('playersError', `Failed to load players: ${error.message}`);
    } finally {
        hideLoading('playersLoading');
    }
}

function renderPlayers() {
    const tbody = document.getElementById('playersTableBody');
    tbody.innerHTML = '';

    let filteredPlayers = state.players;

    // Client-side search filter
    if (state.filters.search) {
        const search = state.filters.search.toLowerCase();
        filteredPlayers = filteredPlayers.filter(p =>
            p.name.toLowerCase().includes(search)
        );
    }

    if (filteredPlayers.length === 0) {
        tbody.innerHTML = '<tr><td colspan="9" style="text-align: center; padding: 40px;">No players found</td></tr>';
        return;
    }

    filteredPlayers.forEach(player => {
        const row = document.createElement('tr');
        const headshotUrl = player.headshot_url || `https://a.espncdn.com/combiner/i?img=/i/headshots/nfl/players/full/${player.espn_athlete_id || 'default'}.png&w=350&h=254`;
        const yearsPro = player.years_pro ? `${player.years_pro} yrs` : 'Rookie';

        row.innerHTML = `
            <td><img src="${headshotUrl}" alt="${player.name}" class="player-headshot" onerror="this.src='https://a.espncdn.com/combiner/i?img=/i/teamlogos/nfl/500/scoreboard/nfl.png&h=50&w=50'"></td>
            <td><strong>${player.name}</strong></td>
            <td><span class="position-badge">${player.position}</span></td>
            <td>--</td>
            <td>#${player.jersey_number || '--'}</td>
            <td>${player.college || '--'}</td>
            <td>${yearsPro}</td>
            <td><span class="badge badge-${player.status}">${player.status}</span></td>
            <td><button class="btn" onclick="viewPlayerDetails('${player.id}')">View</button></td>
        `;
        tbody.appendChild(row);
    });
}

function updatePagination() {
    const totalPages = Math.ceil(state.totalPlayers / state.playersPerPage);
    document.getElementById('playersPageInfo').textContent = `Page ${state.currentPage} of ${totalPages}`;

    document.getElementById('prevPlayers').disabled = state.currentPage === 1;
    document.getElementById('nextPlayers').disabled = state.currentPage >= totalPages;
}

function initPlayerFilters() {
    document.getElementById('playerSearch').addEventListener('input', (e) => {
        state.filters.search = e.target.value;
        renderPlayers();
    });

    document.getElementById('positionFilter').addEventListener('change', (e) => {
        state.filters.position = e.target.value;
        state.currentPage = 1;
        loadPlayers();
    });

    document.getElementById('statusFilter').addEventListener('change', (e) => {
        state.filters.status = e.target.value;
        state.currentPage = 1;
        loadPlayers();
    });

    document.getElementById('teamFilter').addEventListener('change', (e) => {
        state.filters.team = e.target.value;
        state.currentPage = 1;
        loadPlayers();
    });

    document.getElementById('prevPlayers').addEventListener('click', () => {
        if (state.currentPage > 1) {
            state.currentPage--;
            loadPlayers();
        }
    });

    document.getElementById('nextPlayers').addEventListener('click', () => {
        state.currentPage++;
        loadPlayers();
    });

    document.getElementById('refreshPlayers').addEventListener('click', () => {
        localStorage.clear();
        loadPlayers();
    });
}

async function viewPlayerDetails(playerId) {
    try {
        const result = await apiCall(`/api/v1/players/${playerId}`);
        const player = result.data.data;

        const modal = document.getElementById('playerModal');
        document.getElementById('modalPlayerName').textContent = player.name;

        const details = `
            <div style="line-height: 2;">
                <p><strong>Position:</strong> ${player.position}</p>
                <p><strong>Jersey:</strong> ${player.jersey_number || 'N/A'}</p>
                <p><strong>Status:</strong> <span class="badge badge-${player.status}">${player.status}</span></p>
                <p><strong>Height:</strong> ${player.height_inches ? Math.floor(player.height_inches / 12) + "'" + (player.height_inches % 12) + '"' : 'N/A'}</p>
                <p><strong>Weight:</strong> ${player.weight_pounds ? player.weight_pounds + ' lbs' : 'N/A'}</p>
                <p><strong>College:</strong> ${player.college || 'N/A'}</p>
                <p><strong>Draft:</strong> ${player.draft_year ? `${player.draft_year} - Round ${player.draft_round}, Pick ${player.draft_pick}` : 'N/A'}</p>
            </div>
        `;

        document.getElementById('modalPlayerDetails').innerHTML = details;
        modal.classList.add('active');
    } catch (error) {
        alert('Failed to load player details: ' + error.message);
    }
}

// Teams Tab
async function loadTeams() {
    showLoading('teamsLoading');
    hideError('teamsError');

    try {
        const cacheKey = 'teams_all';
        let result = cache.get(cacheKey);

        if (!result) {
            result = await apiCall('/api/v1/teams');
            cache.set(cacheKey, result);
        }

        state.teams = result.data.data;
        renderTeams();
        populateTeamFilter();
    } catch (error) {
        showError('teamsError', `Failed to load teams: ${error.message}`);
    } finally {
        hideLoading('teamsLoading');
    }
}

function renderTeams() {
    const grid = document.getElementById('teamsGrid');
    grid.innerHTML = '';

    state.teams.forEach(team => {
        const card = document.createElement('div');
        card.className = 'team-card';
        card.onclick = () => viewTeamRoster(team.id, team.name);

        card.innerHTML = `
            <div class="team-abbr">${team.abbreviation}</div>
            <h3>${team.name}</h3>
            <div class="team-info">
                <p>${team.city}</p>
                <p>${team.conference} ${team.division}</p>
                <p>${team.stadium || ''}</p>
            </div>
        `;

        grid.appendChild(card);
    });
}

function populateTeamFilter() {
    const select = document.getElementById('teamFilter');
    select.innerHTML = '<option value="">All Teams</option>';

    state.teams.forEach(team => {
        const option = document.createElement('option');
        option.value = team.id;
        option.textContent = `${team.abbreviation} - ${team.name}`;
        select.appendChild(option);
    });
}

async function viewTeamRoster(teamId, teamName) {
    try {
        const result = await apiCall(`/api/v1/teams/${teamId}/players`);
        const players = result.data.data;

        const modal = document.getElementById('playerModal');
        document.getElementById('modalPlayerName').textContent = `${teamName} Roster`;

        let roster = '<div style="max-height: 400px; overflow-y: auto;">';
        roster += '<table class="data-table" style="margin: 0;">';
        roster += '<thead><tr><th>Name</th><th>Position</th><th>Jersey</th><th>Status</th></tr></thead>';
        roster += '<tbody>';

        players.forEach(player => {
            roster += `
                <tr>
                    <td>${player.name}</td>
                    <td>${player.position}</td>
                    <td>${player.jersey_number || '--'}</td>
                    <td><span class="badge badge-${player.status}">${player.status}</span></td>
                </tr>
            `;
        });

        roster += '</tbody></table></div>';

        document.getElementById('modalPlayerDetails').innerHTML = roster;
        modal.classList.add('active');
    } catch (error) {
        alert('Failed to load team roster: ' + error.message);
    }
}

// API Testing Tab
function initApiTesting() {
    const endpointSelect = document.getElementById('apiEndpoint');
    const idField = document.getElementById('apiIdField');
    const paramsField = document.getElementById('apiParamsField');

    endpointSelect.addEventListener('change', () => {
        const endpoint = endpointSelect.value;
        idField.style.display = endpoint.includes(':id') ? 'block' : 'none';
        paramsField.style.display = endpoint.includes(':id') && endpoint.includes('players') ? 'none' : 'block';
    });

    document.getElementById('executeApiCall').addEventListener('click', executeApiTest);
    document.getElementById('copyResponse').addEventListener('click', copyResponse);
}

async function executeApiTest() {
    const endpointTemplate = document.getElementById('apiEndpoint').value;
    const id = document.getElementById('apiId').value;
    const paramsText = document.getElementById('apiParams').value;

    let endpoint = endpointTemplate.replace(':id', id);
    let params = {};

    if (paramsText.trim()) {
        try {
            params = JSON.parse(paramsText);
        } catch (e) {
            alert('Invalid JSON in parameters');
            return;
        }
    }

    try {
        const result = await apiCall(endpoint, params);

        document.getElementById('apiStatusCode').textContent = `Status: ${result.status}`;
        document.getElementById('apiStatusCode').className = 'status-badge success';
        document.getElementById('apiResponseTime').textContent = `${result.responseTime}ms`;
        document.getElementById('apiResponseBody').textContent = JSON.stringify(result.data, null, 2);
    } catch (error) {
        document.getElementById('apiStatusCode').textContent = 'Status: Error';
        document.getElementById('apiStatusCode').className = 'status-badge error';
        document.getElementById('apiResponseTime').textContent = '--';
        document.getElementById('apiResponseBody').textContent = JSON.stringify({
            error: error.message
        }, null, 2);
    }
}

function copyResponse() {
    const responseBody = document.getElementById('apiResponseBody').textContent;
    navigator.clipboard.writeText(responseBody).then(() => {
        alert('Response copied to clipboard!');
    });
}

// Modal
function initModal() {
    const modal = document.getElementById('playerModal');
    const closeBtn = document.querySelector('.modal-close');

    closeBtn.addEventListener('click', () => {
        modal.classList.remove('active');
    });

    window.addEventListener('click', (e) => {
        if (e.target === modal) {
            modal.classList.remove('active');
        }
    });
}

// Dark Mode
function initDarkMode() {
    const toggle = document.getElementById('darkModeToggle');
    const isDark = localStorage.getItem('darkMode') === 'true';

    if (isDark) {
        document.body.classList.add('dark-mode');
        toggle.textContent = '☀️';
    }

    toggle.addEventListener('click', () => {
        document.body.classList.toggle('dark-mode');
        const isDarkNow = document.body.classList.contains('dark-mode');
        localStorage.setItem('darkMode', isDarkNow);
        toggle.textContent = isDarkNow ? '☀️' : '🌙';
    });
}

// Initialize
// Games Functions
async function loadGames() {
    showLoading('gamesLoading');
    hideError('gamesError');

    const season = document.getElementById('seasonFilter').value;
    const week = document.getElementById('weekFilter').value;
    const status = document.getElementById('gameStatusFilter').value;

    const params = {};
    if (season) params.season = season;
    if (week) params.week = week;
    if (status) params.status = status;

    try {
        const result = await apiCall('/api/v1/games', params);
        const games = result.data.data || [];
        renderGames(games);
    } catch (error) {
        showError('gamesError', `Failed to load games: ${error.message}`);
    } finally {
        hideLoading('gamesLoading');
    }
}

function renderGames(games) {
    const tbody = document.getElementById('gamesTableBody');

    if (games.length === 0) {
        tbody.innerHTML = '<tr><td colspan="7" style="text-align: center; padding: 20px;">No games found</td></tr>';
        return;
    }

    tbody.innerHTML = games.map(game => `
        <tr>
            <td>${game.game_date ? new Date(game.game_date).toLocaleDateString() : 'TBD'}</td>
            <td>${game.away_team || 'N/A'}</td>
            <td style="text-align: center; font-weight: bold;">
                ${game.away_score !== null ? game.away_score : '-'} - ${game.home_score !== null ? game.home_score : '-'}
            </td>
            <td>${game.home_team || 'N/A'}</td>
            <td><span class="badge badge-${game.status}">${game.status || 'scheduled'}</span></td>
            <td>${game.weather_temp ? `${game.weather_temp}°F ${game.weather_condition || ''}` : 'N/A'}</td>
            <td>${game.venue_name || 'N/A'}${game.venue_city ? `, ${game.venue_city}` : ''}</td>
        </tr>
    `).join('');
}

// Career Stats Functions
async function loadCareerStats() {
    const playerId = document.getElementById('careerPlayerId').value.trim();

    if (!playerId) {
        showError('careerError', 'Please enter a player ID');
        return;
    }

    showLoading('careerLoading');
    hideError('careerError');
    document.getElementById('careerStatsContainer').style.display = 'none';

    try {
        const result = await apiCall(`/api/v1/players/${playerId}/career`);
        const data = result.data.data;
        renderCareerStats(data);
    } catch (error) {
        showError('careerError', `Failed to load career stats: ${error.message}`);
    } finally {
        hideLoading('careerLoading');
    }
}

function renderCareerStats(data) {
    document.getElementById('careerStatsContainer').style.display = 'block';
    document.getElementById('careerPlayerName').textContent = `Player ID: ${data.player_id}`;
    document.getElementById('careerSummary').textContent =
        `Total Seasons: ${data.total_seasons || 0} | Teams: ${data.team_history?.length || 0}`;

    // Render team history
    const teamHistoryBody = document.getElementById('teamHistoryBody');
    if (data.team_history && data.team_history.length > 0) {
        teamHistoryBody.innerHTML = data.team_history.map(history => `
            <tr>
                <td>${history.team_name || 'N/A'}</td>
                <td>${history.position}</td>
                <td>${history.season_start} - ${history.season_end || 'Present'}</td>
                <td>${history.is_current ? '✓' : ''}</td>
            </tr>
        `).join('');
    } else {
        teamHistoryBody.innerHTML = '<tr><td colspan="4" style="text-align: center;">No team history available</td></tr>';
    }

    // Render career stats
    const careerStatsBody = document.getElementById('careerStatsBody');
    if (data.career_stats && data.career_stats.length > 0) {
        careerStatsBody.innerHTML = data.career_stats.map(stat => `
            <tr>
                <td><strong>${stat.season}</strong></td>
                <td>${stat.team_name || 'N/A'}</td>
                <td>${stat.games_played || 0}</td>
                <td>${stat.passing_yards || 0}</td>
                <td>${stat.passing_tds || 0}</td>
                <td>${stat.rushing_yards || 0}</td>
                <td>${stat.rushing_tds || 0}</td>
                <td>${stat.receptions || 0}</td>
                <td>${stat.receiving_yards || 0}</td>
                <td>${stat.receiving_tds || 0}</td>
            </tr>
        `).join('');
    } else {
        careerStatsBody.innerHTML = '<tr><td colspan="10" style="text-align: center;">No career stats available</td></tr>';
    }
}

// Weather Analysis Functions
async function loadWeatherAnalysis() {
    showLoading('weatherLoading');
    hideError('weatherError');

    const season = document.getElementById('weatherSeason').value;

    try {
        const result = await apiCall('/api/v1/games', { season, limit: 1000 });
        const games = result.data.data || [];

        analyzeWeather(games);
        renderWeatherGames(games);
    } catch (error) {
        showError('weatherError', `Failed to load weather data: ${error.message}`);
    } finally {
        hideLoading('weatherLoading');
    }
}

function analyzeWeather(games) {
    let temps = [];
    let winds = [];
    let venueTypes = { indoor: 0, outdoor: 0, retractable: 0 };

    games.forEach(game => {
        if (game.weather_temp) temps.push(game.weather_temp);
        if (game.weather_wind_speed) winds.push(game.weather_wind_speed);
        if (game.venue_type) {
            venueTypes[game.venue_type.toLowerCase()] = (venueTypes[game.venue_type.toLowerCase()] || 0) + 1;
        }
    });

    // Temperature stats
    const avgTemp = temps.length > 0 ? (temps.reduce((a, b) => a + b, 0) / temps.length).toFixed(1) : '--';
    const maxTemp = temps.length > 0 ? Math.max(...temps) : '--';
    const minTemp = temps.length > 0 ? Math.min(...temps) : '--';

    document.getElementById('tempStats').innerHTML = `
        <p>Average: ${avgTemp}°F</p>
        <p>High: ${maxTemp}°F</p>
        <p>Low: ${minTemp}°F</p>
    `;

    // Wind stats
    const avgWind = winds.length > 0 ? (winds.reduce((a, b) => a + b, 0) / winds.length).toFixed(1) : '--';
    const highWind = winds.filter(w => w > 15).length;

    document.getElementById('windStats').innerHTML = `
        <p>Average Wind: ${avgWind} mph</p>
        <p>High Wind Games (>15mph): ${highWind}</p>
    `;

    // Venue stats
    document.getElementById('venueStats').innerHTML = `
        <p>Indoor: ${venueTypes.indoor || 0}</p>
        <p>Outdoor: ${venueTypes.outdoor || 0}</p>
        <p>Retractable: ${venueTypes.retractable || 0}</p>
    `;
}

function renderWeatherGames(games) {
    const tbody = document.getElementById('weatherGamesBody');

    const filteredGames = games.filter(g => g.weather_temp || g.weather_condition);

    if (filteredGames.length === 0) {
        tbody.innerHTML = '<tr><td colspan="7" style="text-align: center; padding: 20px;">No weather data available</td></tr>';
        return;
    }

    tbody.innerHTML = filteredGames.map(game => `
        <tr>
            <td>${game.game_date ? new Date(game.game_date).toLocaleDateString() : 'TBD'}</td>
            <td>${game.away_team || 'N/A'} @ ${game.home_team || 'N/A'}</td>
            <td>${game.weather_temp ? `${game.weather_temp}°F` : 'N/A'}</td>
            <td>${game.weather_condition || 'N/A'}</td>
            <td>${game.weather_wind_speed ? `${game.weather_wind_speed} mph` : 'N/A'}</td>
            <td>${game.venue_name || 'N/A'}</td>
            <td>${game.away_score !== null ? `${game.away_score}-${game.home_score}` : 'N/A'}</td>
        </tr>
    `).join('');
}

// Fantasy Stats Functions
async function loadFantasyStats() {
    showLoading('fantasyLoading');
    hideError('fantasyError');

    const season = document.getElementById('fantasySeason').value;
    const week = document.getElementById('fantasyWeek').value;
    const position = document.getElementById('fantasyPosition').value;

    try {
        // This would call a fantasy stats endpoint when available
        // For now, we'll use player stats as a placeholder
        const result = await apiCall('/api/v1/players', {
            limit: 50,
            position: position || undefined
        });

        renderFantasyLeaders(result.data.data || []);
    } catch (error) {
        showError('fantasyError', `Failed to load fantasy stats: ${error.message}`);
    } finally {
        hideLoading('fantasyLoading');
    }
}

function renderFantasyLeaders(players) {
    const tbody = document.getElementById('fantasyLeadersBody');
    const scoringType = document.getElementById('fantasyScoring').value;

    if (players.length === 0) {
        tbody.innerHTML = '<tr><td colspan="12" style="text-align: center; padding: 20px;">No data available</td></tr>';
        return;
    }

    // Calculate fantasy points (simplified)
    const playersWithPoints = players.map(p => ({
        ...p,
        fantasyPoints: calculateFantasyPoints(p, scoringType)
    })).sort((a, b) => b.fantasyPoints - a.fantasyPoints);

    tbody.innerHTML = playersWithPoints.slice(0, 25).map((player, index) => `
        <tr>
            <td><strong>${index + 1}</strong></td>
            <td>${player.name}</td>
            <td><span class="position-badge">${player.position}</span></td>
            <td>--</td>
            <td><strong>${player.fantasyPoints.toFixed(1)}</strong></td>
            <td>--</td>
            <td>--</td>
            <td>--</td>
            <td>--</td>
            <td>--</td>
            <td>--</td>
            <td>--</td>
        </tr>
    `).join('');
}

function calculateFantasyPoints(player, scoringType) {
    // Placeholder calculation - would be enhanced with actual stats
    let points = 0;

    // This is a simplified example - real implementation would use actual game stats
    if (player.position === 'QB') points = Math.random() * 30;
    else if (player.position === 'RB') points = Math.random() * 25;
    else if (player.position === 'WR') points = Math.random() * 22;
    else if (player.position === 'TE') points = Math.random() * 15;

    if (scoringType === 'ppr') points += Math.random() * 5;
    else if (scoringType === 'half_ppr') points += Math.random() * 2.5;

    return points;
}

// Admin Sync Functions
async function syncData(endpoint, body = {}) {
    const statusDiv = document.getElementById('syncStatus');
    statusDiv.innerHTML = '<p class="loading">Syncing...</p>';

    try {
        const response = await fetch(`${API_BASE_URL}/api/v1/admin/sync/${endpoint}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body)
        });

        const result = await response.json();

        if (response.ok) {
            statusDiv.innerHTML = `<p class="success">✓ ${result.data.message}</p>`;
        } else {
            statusDiv.innerHTML = `<p class="error">✗ Sync failed: ${result.error?.message || 'Unknown error'}</p>`;
        }
    } catch (error) {
        statusDiv.innerHTML = `<p class="error">✗ Sync failed: ${error.message}</p>`;
    }
}

document.addEventListener('DOMContentLoaded', () => {
    initTabs();
    initPlayerFilters();
    initApiTesting();
    initModal();
    initDarkMode();

    // Load initial data
    loadPlayers();
    loadTeams();

    // Refresh buttons
    document.getElementById('refreshTeams').addEventListener('click', () => {
        localStorage.clear();
        loadTeams();
    });

    document.getElementById('refreshGames').addEventListener('click', loadGames);

    // Career stats
    document.getElementById('loadCareerStats').addEventListener('click', loadCareerStats);

    // Weather analysis
    document.getElementById('refreshWeather').addEventListener('click', loadWeatherAnalysis);

    // Fantasy stats
    document.getElementById('refreshFantasy').addEventListener('click', loadFantasyStats);

    // Admin sync buttons - ESPN
    document.getElementById('syncTeams').addEventListener('click', () => syncData('teams'));
    document.getElementById('syncRosters').addEventListener('click', () => syncData('rosters'));
    document.getElementById('syncGames').addEventListener('click', () => syncData('games'));
    document.getElementById('syncFull').addEventListener('click', () => syncData('full'));

    // Admin sync buttons - Historical
    document.getElementById('syncHistoricalSeason').addEventListener('click', () => {
        const year = parseInt(document.getElementById('historicalYear').value);
        syncData('historical/season', { year });
    });

    document.getElementById('syncMultipleSeasons').addEventListener('click', () => {
        const start_year = parseInt(document.getElementById('multiSeasonStart').value);
        const end_year = parseInt(document.getElementById('multiSeasonEnd').value);
        syncData('historical/seasons', { start_year, end_year });
    });

    // Admin sync buttons - NFLverse
    document.getElementById('syncNFLverseStats').addEventListener('click', () => {
        const season = parseInt(document.getElementById('nflverseStatsSeason').value);
        syncData('nflverse/stats', { season });
    });

    document.getElementById('syncNFLverseSchedule').addEventListener('click', () => {
        const season = parseInt(document.getElementById('nflverseScheduleSeason').value);
        syncData('nflverse/schedule', { season });
    });

    document.getElementById('syncNFLverseNextGen').addEventListener('click', () => {
        const season = parseInt(document.getElementById('nflverseNextGenSeason').value);
        const stat_type = document.getElementById('nflverseStatType').value;
        syncData('nflverse/nextgen', { season, stat_type });
    });
});

// Make viewPlayerDetails available globally
window.viewPlayerDetails = viewPlayerDetails;