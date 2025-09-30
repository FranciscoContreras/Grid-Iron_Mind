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
        // Use placeholder by default - images will attempt to load ESPN headshots via onerror chain
        const placeholderUrl = 'https://a.espncdn.com/combiner/i?img=/i/teamlogos/nfl/500/scoreboard/nfl.png&h=50&w=50';
        let headshotUrl = placeholderUrl;

        // Try headshot_url first, then ESPN with nfl_id
        if (player.headshot_url && player.headshot_url.trim() !== '') {
            headshotUrl = player.headshot_url;
        } else if (player.nfl_id) {
            headshotUrl = `https://a.espncdn.com/combiner/i?img=/i/headshots/nfl/players/full/${player.nfl_id}.png&w=350&h=254`;
        }

        const yearsPro = player.years_pro ? `${player.years_pro} yrs` : 'Rookie';

        row.innerHTML = `
            <td><img src="${headshotUrl}"
                     alt="${player.name}"
                     class="player-headshot"
                     style="max-width: 50px; max-height: 50px; object-fit: contain;"
                     onerror="this.onerror=null; this.src='${placeholderUrl}';"></td>
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

        // Load career stats with loading indicator
        let careerHTML = `
            <h3 style="margin-top: 20px; border-bottom: 2px solid var(--primary); padding-bottom: 10px;">
                📊 Career Statistics
            </h3>
            <div style="padding: 40px; text-align: center;">
                <div class="spinner"></div>
                <p style="margin-top: 20px; color: #666;">Loading career statistics...</p>
            </div>
        `;

        // Show modal with loading state
        const detailsContainer = document.getElementById('modalPlayerDetails');
        detailsContainer.innerHTML = `
            <div style="line-height: 2;">
                <h3 style="border-bottom: 2px solid var(--primary); padding-bottom: 10px; margin-bottom: 15px;">
                    ℹ️ Player Information
                </h3>
                <p><strong>Position:</strong> ${player.position}</p>
                <p><strong>Jersey:</strong> ${player.jersey_number || 'N/A'}</p>
                <p><strong>Status:</strong> <span class="badge badge-${player.status}">${player.status}</span></p>
                <p><strong>Height:</strong> ${player.height_inches ? Math.floor(player.height_inches / 12) + "'" + (player.height_inches % 12) + '"' : 'N/A'}</p>
                <p><strong>Weight:</strong> ${player.weight_pounds ? player.weight_pounds + ' lbs' : 'N/A'}</p>
                <p><strong>College:</strong> ${player.college || 'N/A'}</p>
                <p><strong>Draft:</strong> ${player.draft_year ? `${player.draft_year} - Round ${player.draft_round}, Pick ${player.draft_pick}` : 'N/A'}</p>
            </div>
            ${careerHTML}
        `;
        modal.classList.add('active');

        // Now fetch career stats
        try {
            const careerResult = await apiCall(`/api/v1/players/${playerId}/career`);
            const careerData = careerResult.data.data;

            if (careerData.career_stats && careerData.career_stats.length > 0) {
                const currentYear = new Date().getFullYear();
                const stats = careerData.career_stats;

                // Sort by season descending (most recent first)
                stats.sort((a, b) => b.season - a.season);

                // Calculate career totals and highlights
                const careerTotals = {
                    seasons: stats.length,
                    games: stats.reduce((sum, s) => sum + (s.games_played || 0), 0),
                    passingYards: stats.reduce((sum, s) => sum + (s.passing_yards || 0), 0),
                    passingTDs: stats.reduce((sum, s) => sum + (s.passing_tds || 0), 0),
                    passingInts: stats.reduce((sum, s) => sum + (s.passing_ints || 0), 0),
                    rushingYards: stats.reduce((sum, s) => sum + (s.rushing_yards || 0), 0),
                    rushingTDs: stats.reduce((sum, s) => sum + (s.rushing_tds || 0), 0),
                    receivingYards: stats.reduce((sum, s) => sum + (s.receiving_yards || 0), 0),
                    receivingTDs: stats.reduce((sum, s) => sum + (s.receiving_tds || 0), 0),
                    receptions: stats.reduce((sum, s) => sum + (s.receptions || 0), 0)
                };

                const yearsActive = `${stats[stats.length - 1].season}-${stats[0].season}`;

                // Determine primary position based on stats
                let primaryStats = [];
                if (careerTotals.passingYards > 0) {
                    primaryStats.push(`
                        <div style="text-align: center; padding: 15px; background: rgba(76, 175, 80, 0.1); border-radius: 8px;">
                            <div style="font-size: 28px; font-weight: bold; color: var(--primary);">${careerTotals.passingYards.toLocaleString()}</div>
                            <div style="font-size: 12px; color: #666; margin-top: 5px;">Career Passing Yards</div>
                            <div style="font-size: 14px; margin-top: 5px;">${careerTotals.passingTDs} TDs / ${careerTotals.passingInts} INTs</div>
                        </div>
                    `);
                }
                if (careerTotals.rushingYards > 0) {
                    primaryStats.push(`
                        <div style="text-align: center; padding: 15px; background: rgba(33, 150, 243, 0.1); border-radius: 8px;">
                            <div style="font-size: 28px; font-weight: bold; color: #2196F3;">${careerTotals.rushingYards.toLocaleString()}</div>
                            <div style="font-size: 12px; color: #666; margin-top: 5px;">Career Rushing Yards</div>
                            <div style="font-size: 14px; margin-top: 5px;">${careerTotals.rushingTDs} TDs</div>
                        </div>
                    `);
                }
                if (careerTotals.receivingYards > 0) {
                    primaryStats.push(`
                        <div style="text-align: center; padding: 15px; background: rgba(255, 152, 0, 0.1); border-radius: 8px;">
                            <div style="font-size: 28px; font-weight: bold; color: #FF9800;">${careerTotals.receivingYards.toLocaleString()}</div>
                            <div style="font-size: 12px; color: #666; margin-top: 5px;">Career Receiving Yards</div>
                            <div style="font-size: 14px; margin-top: 5px;">${careerTotals.receptions} Rec / ${careerTotals.receivingTDs} TDs</div>
                        </div>
                    `);
                }

                careerHTML = `
                    <h3 style="margin-top: 20px; border-bottom: 2px solid var(--primary); padding-bottom: 10px;">
                        📊 Career Statistics (${yearsActive})
                    </h3>

                    <!-- Career Summary -->
                    <div style="background: #f5f5f5; padding: 20px; border-radius: 8px; margin: 15px 0;">
                        <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(120px, 1fr)); gap: 15px; margin-bottom: 15px;">
                            <div style="text-align: center;">
                                <div style="font-size: 32px; font-weight: bold; color: var(--primary);">${careerTotals.seasons}</div>
                                <div style="font-size: 12px; color: #666;">Seasons</div>
                            </div>
                            <div style="text-align: center;">
                                <div style="font-size: 32px; font-weight: bold; color: var(--primary);">${careerTotals.games}</div>
                                <div style="font-size: 12px; color: #666;">Games Played</div>
                            </div>
                        </div>

                        ${primaryStats.length > 0 ? `
                            <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; margin-top: 15px;">
                                ${primaryStats.join('')}
                            </div>
                        ` : ''}
                    </div>

                    <!-- Season-by-Season Breakdown -->
                    <h4 style="margin: 20px 0 10px 0; color: #666;">Season-by-Season Breakdown</h4>
                    <div style="max-height: 400px; overflow-y: auto; border: 1px solid #ddd; border-radius: 4px;">
                        <table class="data-table" style="font-size: 14px;">
                            <thead style="position: sticky; top: 0; background: white; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
                                <tr>
                                    <th>Season</th>
                                    <th>Team</th>
                                    <th>GP</th>
                                    <th>Pass Yds</th>
                                    <th>Pass TD</th>
                                    <th>INT</th>
                                    <th>Rush Yds</th>
                                    <th>Rush TD</th>
                                    <th>Rec</th>
                                    <th>Rec Yds</th>
                                    <th>Rec TD</th>
                                </tr>
                            </thead>
                            <tbody>
                                ${stats.map(s => `
                                    <tr style="${s.season === currentYear ? 'background-color: rgba(76, 175, 80, 0.1); font-weight: bold;' : ''}">
                                        <td>${s.season}${s.season === currentYear ? ' ⭐' : ''}</td>
                                        <td>${s.team ? s.team.abbreviation : 'N/A'}</td>
                                        <td>${s.games_played || 0}</td>
                                        <td>${s.passing_yards > 0 ? s.passing_yards.toLocaleString() : '-'}</td>
                                        <td>${s.passing_tds > 0 ? s.passing_tds : '-'}</td>
                                        <td>${s.passing_ints > 0 ? s.passing_ints : '-'}</td>
                                        <td>${s.rushing_yards > 0 ? s.rushing_yards.toLocaleString() : '-'}</td>
                                        <td>${s.rushing_tds > 0 ? s.rushing_tds : '-'}</td>
                                        <td>${s.receptions > 0 ? s.receptions : '-'}</td>
                                        <td>${s.receiving_yards > 0 ? s.receiving_yards.toLocaleString() : '-'}</td>
                                        <td>${s.receiving_tds > 0 ? s.receiving_tds : '-'}</td>
                                    </tr>
                                `).join('')}
                            </tbody>
                        </table>
                    </div>
                `;
            } else {
                careerHTML = `
                    <h3 style="margin-top: 20px; border-bottom: 2px solid var(--primary); padding-bottom: 10px;">
                        📊 Career Statistics
                    </h3>
                    <p style="padding: 20px; text-align: center; color: #666;">
                        No career statistics available yet. Stats will be synced from ESPN.
                    </p>
                `;
            }
        } catch (error) {
            console.error('Failed to load career stats:', error);
            careerHTML = `
                <h3 style="margin-top: 20px; border-bottom: 2px solid var(--primary); padding-bottom: 10px;">
                    📊 Career Statistics
                </h3>
                <p style="padding: 20px; text-align: center; color: #666;">
                    Unable to load career statistics.
                </p>
            `;
        }

        // Fetch injury data
        let injuryHTML = '';
        try {
            const injuryResult = await apiCall(`/api/v1/players/${playerId}/injuries`);
            const injuries = injuryResult.data.injuries;

            if (injuries && injuries.length > 0) {
                const currentInjuries = injuries.filter(inj =>
                    inj.status !== 'Healthy' && inj.status !== 'Active'
                );

                if (currentInjuries.length > 0) {
                    injuryHTML = `
                        <div style="margin-top: 15px; padding: 12px; background: #fff3cd; border-left: 4px solid #ffc107; border-radius: 4px;">
                            <h4 style="margin: 0 0 10px 0; color: #856404;">⚠️ Current Injuries</h4>
                            ${currentInjuries.map(inj => `
                                <div style="margin-bottom: 8px;">
                                    <strong>${inj.status}</strong> - ${inj.injury_type || 'Unknown'}
                                    ${inj.body_location ? ` (${inj.body_location})` : ''}
                                    ${inj.return_date ? `<br><small>Expected return: ${new Date(inj.return_date).toLocaleDateString()}</small>` : ''}
                                </div>
                            `).join('')}
                        </div>
                    `;
                }
            }
        } catch (error) {
            console.error('Failed to load injuries:', error);
        }

        // Update the modal with loaded career stats and injuries
        detailsContainer.innerHTML = `
            <div style="line-height: 2;">
                <h3 style="border-bottom: 2px solid var(--primary); padding-bottom: 10px; margin-bottom: 15px;">
                    ℹ️ Player Information
                </h3>
                <p><strong>Position:</strong> ${player.position}</p>
                <p><strong>Jersey:</strong> ${player.jersey_number || 'N/A'}</p>
                <p><strong>Status:</strong> <span class="badge badge-${player.status}">${player.status}</span></p>
                <p><strong>Height:</strong> ${player.height_inches ? Math.floor(player.height_inches / 12) + "'" + (player.height_inches % 12) + '"' : 'N/A'}</p>
                <p><strong>Weight:</strong> ${player.weight_pounds ? player.weight_pounds + ' lbs' : 'N/A'}</p>
                <p><strong>College:</strong> ${player.college || 'N/A'}</p>
                <p><strong>Draft:</strong> ${player.draft_year ? `${player.draft_year} - Round ${player.draft_round}, Pick ${player.draft_pick}` : 'N/A'}</p>
                ${injuryHTML}
            </div>
            ${careerHTML}
        `;
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
        card.onclick = () => viewTeamDetail(team.id, team.name, team);

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

// Navigation State
let navigationState = {
    level: 'teams',
    currentTeam: null,
    currentGame: null,
    currentPlayer: null
};

// Level 1 → Level 2: Team Detail View
async function viewTeamDetail(teamId, teamName, teamData) {
    navigationState = {
        level: 'teamDetail',
        currentTeam: { id: teamId, name: teamName, data: teamData },
        currentGame: null,
        currentPlayer: null
    };

    // Hide teams grid
    document.getElementById('teamsGrid').style.display = 'none';

    // Show team detail view
    const teamDetailView = document.getElementById('teamDetailView');
    teamDetailView.style.display = 'block';

    // Update breadcrumb
    const breadcrumb = document.getElementById('teamBreadcrumb');
    breadcrumb.style.display = 'block';
    breadcrumb.innerHTML = `
        <span class="breadcrumb-item" onclick="navigateToTeams()">Teams</span>
        <span class="breadcrumb-separator"> > </span>
        <span class="breadcrumb-item active">${teamName}</span>
    `;

    // Show back button
    document.getElementById('backToTeams').style.display = 'inline-block';

    // Set team name with logo
    const teamLogo = teamData.abbreviation ?
        `https://a.espncdn.com/i/teamlogos/nfl/500/${teamData.abbreviation.toLowerCase()}.png` :
        'https://a.espncdn.com/combiner/i?img=/i/teamlogos/nfl/500/scoreboard/nfl.png&h=100&w=100';

    document.getElementById('teamDetailName').innerHTML = `
        <img src="${teamLogo}"
             alt="${teamName}"
             style="height: 60px; width: 60px; vertical-align: middle; margin-right: 15px; object-fit: contain;"
             onerror="this.src='https://a.espncdn.com/combiner/i?img=/i/teamlogos/nfl/500/scoreboard/nfl.png&h=100&w=100'">
        ${teamName}
    `;

    // Load games history and roster
    try {
        // Load games
        const gamesResult = await apiCall('/api/v1/games', { team: teamId, limit: 100 });
        const games = gamesResult.data.data || [];
        renderTeamGames(games);

        // Load roster
        const rosterResult = await apiCall(`/api/v1/teams/${teamId}/players`);
        const players = rosterResult.data.data || [];
        renderTeamRoster(players);
    } catch (error) {
        alert('Failed to load team details: ' + error.message);
    }
}

function renderTeamGames(games) {
    const container = document.getElementById('teamGamesList');

    if (games.length === 0) {
        container.innerHTML = '<p style="padding: 20px; text-align: center;">No games found</p>';
        return;
    }

    container.innerHTML = `
        <table class="data-table">
            <thead>
                <tr>
                    <th>Date</th>
                    <th>Matchup</th>
                    <th>Score</th>
                    <th>Status</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>
                ${games.map(game => `
                    <tr>
                        <td>${game.game_date ? new Date(game.game_date).toLocaleDateString() : 'TBD'}</td>
                        <td>${game.away_team_name || game.away_team_abbr || 'N/A'} @ ${game.home_team_name || game.home_team_abbr || 'N/A'}</td>
                        <td style="font-weight: bold;">
                            ${game.away_score !== null ? `${game.away_score} - ${game.home_score}` : 'N/A'}
                        </td>
                        <td><span class="badge badge-${game.status}">${game.status || 'scheduled'}</span></td>
                        <td><button class="btn" onclick="viewGameDetail('${game.id}', '${game.away_team_name || game.away_team_abbr || 'Away'} @ ${game.home_team_name || game.home_team_abbr || 'Home'}', ${JSON.stringify(game).replace(/'/g, "&apos;")})">View Game</button></td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
    `;
}

function renderTeamRoster(players) {
    const container = document.getElementById('teamRosterList');

    if (players.length === 0) {
        container.innerHTML = '<p style="padding: 20px; text-align: center;">No roster data available</p>';
        return;
    }

    container.innerHTML = `
        <table class="data-table">
            <thead>
                <tr>
                    <th>Photo</th>
                    <th>Name</th>
                    <th>Position</th>
                    <th>Jersey</th>
                    <th>Status</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>
                ${players.map(player => {
                    // Use placeholder by default - images will attempt to load ESPN headshots via onerror chain
                    const placeholderUrl = 'https://a.espncdn.com/combiner/i?img=/i/teamlogos/nfl/500/scoreboard/nfl.png&h=50&w=50';
                    let headshotUrl = placeholderUrl;

                    // Try headshot_url first, then ESPN with nfl_id
                    if (player.headshot_url && player.headshot_url.trim() !== '') {
                        headshotUrl = player.headshot_url;
                    } else if (player.nfl_id) {
                        headshotUrl = `https://a.espncdn.com/combiner/i?img=/i/headshots/nfl/players/full/${player.nfl_id}.png&w=350&h=254`;
                    }

                    return `
                        <tr>
                            <td><img src="${headshotUrl}"
                                     alt="${player.name}"
                                     class="player-headshot"
                                     style="max-width: 50px; max-height: 50px; object-fit: contain;"
                                     onerror="this.onerror=null; this.src='${placeholderUrl}';"></td>
                            <td><strong>${player.name}</strong></td>
                            <td><span class="position-badge">${player.position}</span></td>
                            <td>#${player.jersey_number || '--'}</td>
                            <td><span class="badge badge-${player.status}">${player.status}</span></td>
                            <td><button class="btn" onclick="viewPlayerHistorical('${player.id}', '${player.name}')">View History</button></td>
                        </tr>
                    `;
                }).join('')}
            </tbody>
        </table>
    `;
}

// Level 2 → Level 3: Game Detail View
async function viewGameDetail(gameId, gameTitle, gameData) {
    navigationState.level = 'gameDetail';
    navigationState.currentGame = { id: gameId, title: gameTitle, data: gameData };

    // Hide team detail view
    document.getElementById('teamDetailView').style.display = 'none';

    // Show game detail view
    const gameDetailView = document.getElementById('gameDetailView');
    gameDetailView.style.display = 'block';

    // Update breadcrumb
    const breadcrumb = document.getElementById('teamBreadcrumb');
    breadcrumb.innerHTML = `
        <span class="breadcrumb-item" onclick="navigateToTeams()">Teams</span>
        <span class="breadcrumb-separator"> > </span>
        <span class="breadcrumb-item" onclick="navigateToTeamDetail()">${navigationState.currentTeam.name}</span>
        <span class="breadcrumb-separator"> > </span>
        <span class="breadcrumb-item active">${gameTitle}</span>
    `;

    // Set game title
    document.getElementById('gameDetailTitle').textContent = gameTitle;

    // Render weather/game conditions
    renderGameConditions(gameData);

    // Load player stats for this game
    try {
        const statsResult = await apiCall(`/api/v1/stats/game/${gameId}`);
        const stats = statsResult.data.data || [];
        renderGamePlayerStats(stats);
    } catch (error) {
        console.error('Failed to load player stats:', error);
        document.getElementById('gamePlayersList').innerHTML =
            '<p style="padding: 20px; text-align: center;">No player stats available for this game</p>';
    }
}

function renderGameConditions(game) {
    const container = document.getElementById('gameWeatherDetails');

    container.innerHTML = `
        <div class="weather-grid">
            <div class="weather-card">
                <h4>🌡️ Temperature</h4>
                <p style="font-size: 24px; font-weight: bold;">${game.weather_temp ? `${game.weather_temp}°F` : 'N/A'}</p>
                <p>${game.weather_condition || ''}</p>
            </div>
            <div class="weather-card">
                <h4>💨 Wind</h4>
                <p style="font-size: 24px; font-weight: bold;">${game.weather_wind_speed ? `${game.weather_wind_speed} mph` : 'N/A'}</p>
            </div>
            <div class="weather-card">
                <h4>🏟️ Venue</h4>
                <p style="font-size: 18px; font-weight: bold;">${game.venue_name || 'N/A'}</p>
                <p>${game.venue_city ? game.venue_city : ''}</p>
                <p>${game.venue_type ? `Type: ${game.venue_type}` : ''}</p>
            </div>
            <div class="weather-card">
                <h4>📅 Game Info</h4>
                <p><strong>Date:</strong> ${game.game_date ? new Date(game.game_date).toLocaleDateString() : 'TBD'}</p>
                <p><strong>Week:</strong> ${game.week || 'N/A'}</p>
                <p><strong>Season:</strong> ${game.season || 'N/A'}</p>
            </div>
        </div>
    `;
}

function renderGamePlayerStats(stats) {
    const container = document.getElementById('gamePlayersList');

    if (stats.length === 0) {
        container.innerHTML = '<p style="padding: 20px; text-align: center;">No player stats available for this game</p>';
        return;
    }

    // Separate by position type for better organization
    const passers = stats.filter(s => s.passing_attempts > 0);
    const rushers = stats.filter(s => s.rushing_attempts > 0);
    const receivers = stats.filter(s => s.receiving_receptions > 0);

    let html = '';

    // Passing stats
    if (passers.length > 0) {
        html += `
            <h3 style="margin-top: 20px; border-bottom: 2px solid var(--primary); padding-bottom: 10px;">
                🏈 Passing Stats
            </h3>
            <table class="data-table" style="font-size: 14px;">
                <thead>
                    <tr>
                        <th>Player</th>
                        <th>C/ATT</th>
                        <th>Yards</th>
                        <th>TD</th>
                        <th>INT</th>
                        <th>Rating</th>
                    </tr>
                </thead>
                <tbody>
                    ${passers.map(p => {
                        const rating = p.passing_attempts > 0 ?
                            ((p.passing_completions / p.passing_attempts) * 100).toFixed(1) : '0.0';
                        return `
                            <tr>
                                <td><strong>${p.player?.name || 'Unknown'}</strong></td>
                                <td>${p.passing_completions}/${p.passing_attempts}</td>
                                <td>${p.passing_yards}</td>
                                <td>${p.passing_touchdowns}</td>
                                <td>${p.passing_interceptions}</td>
                                <td>${rating}%</td>
                            </tr>
                        `;
                    }).join('')}
                </tbody>
            </table>
        `;
    }

    // Rushing stats
    if (rushers.length > 0) {
        html += `
            <h3 style="margin-top: 30px; border-bottom: 2px solid var(--primary); padding-bottom: 10px;">
                🏃 Rushing Stats
            </h3>
            <table class="data-table" style="font-size: 14px;">
                <thead>
                    <tr>
                        <th>Player</th>
                        <th>Attempts</th>
                        <th>Yards</th>
                        <th>Avg</th>
                        <th>TD</th>
                    </tr>
                </thead>
                <tbody>
                    ${rushers.map(p => {
                        const avg = p.rushing_attempts > 0 ?
                            (p.rushing_yards / p.rushing_attempts).toFixed(1) : '0.0';
                        return `
                            <tr>
                                <td><strong>${p.player?.name || 'Unknown'}</strong></td>
                                <td>${p.rushing_attempts}</td>
                                <td>${p.rushing_yards}</td>
                                <td>${avg}</td>
                                <td>${p.rushing_touchdowns}</td>
                            </tr>
                        `;
                    }).join('')}
                </tbody>
            </table>
        `;
    }

    // Receiving stats
    if (receivers.length > 0) {
        html += `
            <h3 style="margin-top: 30px; border-bottom: 2px solid var(--primary); padding-bottom: 10px;">
                🎯 Receiving Stats
            </h3>
            <table class="data-table" style="font-size: 14px;">
                <thead>
                    <tr>
                        <th>Player</th>
                        <th>Receptions</th>
                        <th>Targets</th>
                        <th>Yards</th>
                        <th>Avg</th>
                        <th>TD</th>
                    </tr>
                </thead>
                <tbody>
                    ${receivers.map(p => {
                        const avg = p.receiving_receptions > 0 ?
                            (p.receiving_yards / p.receiving_receptions).toFixed(1) : '0.0';
                        return `
                            <tr>
                                <td><strong>${p.player?.name || 'Unknown'}</strong></td>
                                <td>${p.receiving_receptions}</td>
                                <td>${p.receiving_targets}</td>
                                <td>${p.receiving_yards}</td>
                                <td>${avg}</td>
                                <td>${p.receiving_touchdowns}</td>
                            </tr>
                        `;
                    }).join('')}
                </tbody>
            </table>
        `;
    }

    if (html === '') {
        container.innerHTML = '<p style="padding: 20px; text-align: center;">No player performance data available for this game</p>';
    } else {
        container.innerHTML = html;
    }
}

// Level 3/2 → Level 4: Player Historical View
async function viewPlayerHistorical(playerId, playerName) {
    navigationState.level = 'playerHistorical';
    navigationState.currentPlayer = { id: playerId, name: playerName };

    // Hide previous views
    document.getElementById('teamDetailView').style.display = 'none';
    document.getElementById('gameDetailView').style.display = 'none';

    // Show player historical view
    const playerHistoricalView = document.getElementById('playerHistoricalView');
    playerHistoricalView.style.display = 'block';

    // Update breadcrumb
    const breadcrumb = document.getElementById('teamBreadcrumb');
    let breadcrumbHTML = `<span class="breadcrumb-item" onclick="navigateToTeams()">Teams</span>`;

    if (navigationState.currentTeam) {
        breadcrumbHTML += `
            <span class="breadcrumb-separator"> > </span>
            <span class="breadcrumb-item" onclick="navigateToTeamDetail()">${navigationState.currentTeam.name}</span>
        `;
    }

    if (navigationState.currentGame) {
        breadcrumbHTML += `
            <span class="breadcrumb-separator"> > </span>
            <span class="breadcrumb-item" onclick="navigateToGameDetail()">${navigationState.currentGame.title}</span>
        `;
    }

    breadcrumbHTML += `
        <span class="breadcrumb-separator"> > </span>
        <span class="breadcrumb-item active">${playerName}</span>
    `;

    breadcrumb.innerHTML = breadcrumbHTML;

    // Set player name
    document.getElementById('playerHistoricalName').textContent = `${playerName} - Career History`;

    // Load player career data
    try {
        const careerResult = await apiCall(`/api/v1/players/${playerId}/career`);
        const careerData = careerResult.data.data;
        renderPlayerHistoricalData(careerData);
    } catch (error) {
        document.getElementById('playerHistoricalData').innerHTML =
            `<p style="padding: 20px; text-align: center; color: var(--danger);">Failed to load career data: ${error.message}</p>`;
    }
}

function renderPlayerHistoricalData(data) {
    const container = document.getElementById('playerHistoricalData');

    let html = '<div style="margin-bottom: 30px;">';

    // Summary
    html += `
        <div class="stats-summary">
            <h3>Career Summary</h3>
            <p><strong>Total Seasons:</strong> ${data.total_seasons || 0}</p>
            <p><strong>Teams Played For:</strong> ${data.team_history?.length || 0}</p>
        </div>
    `;

    // Team History
    if (data.team_history && data.team_history.length > 0) {
        html += `
            <div style="margin-bottom: 30px;">
                <h3>Team History</h3>
                <table class="data-table">
                    <thead>
                        <tr>
                            <th>Team</th>
                            <th>Position</th>
                            <th>Years</th>
                            <th>Current</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${data.team_history.map(history => `
                            <tr>
                                <td><strong>${history.team_name || 'N/A'}</strong></td>
                                <td>${history.position}</td>
                                <td>${history.season_start} - ${history.season_end || 'Present'}</td>
                                <td>${history.is_current ? '✓ Current' : ''}</td>
                            </tr>
                        `).join('')}
                    </tbody>
                </table>
            </div>
        `;
    }

    // Season-by-Season Stats
    if (data.career_stats && data.career_stats.length > 0) {
        html += `
            <div>
                <h3>Season-by-Season Statistics</h3>
                <table class="data-table">
                    <thead>
                        <tr>
                            <th>Season</th>
                            <th>Team</th>
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
                        ${data.career_stats.map(stat => `
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
                        `).join('')}
                    </tbody>
                </table>
            </div>
        `;
    }

    html += '</div>';

    if (!data.team_history && !data.career_stats) {
        html = '<p style="padding: 20px; text-align: center;">No career data available for this player</p>';
    }

    container.innerHTML = html;
}

// Navigation Functions
function navigateToTeams() {
    navigationState = {
        level: 'teams',
        currentTeam: null,
        currentGame: null,
        currentPlayer: null
    };

    // Show teams grid
    document.getElementById('teamsGrid').style.display = 'grid';

    // Hide all other views
    document.getElementById('teamDetailView').style.display = 'none';
    document.getElementById('gameDetailView').style.display = 'none';
    document.getElementById('playerHistoricalView').style.display = 'none';

    // Hide breadcrumb and back button
    document.getElementById('teamBreadcrumb').style.display = 'none';
    document.getElementById('backToTeams').style.display = 'none';
}

function navigateToTeamDetail() {
    if (!navigationState.currentTeam) return;

    // Hide other views
    document.getElementById('gameDetailView').style.display = 'none';
    document.getElementById('playerHistoricalView').style.display = 'none';

    // Show team detail view
    document.getElementById('teamDetailView').style.display = 'block';

    // Update breadcrumb
    const breadcrumb = document.getElementById('teamBreadcrumb');
    breadcrumb.innerHTML = `
        <span class="breadcrumb-item" onclick="navigateToTeams()">Teams</span>
        <span class="breadcrumb-separator"> > </span>
        <span class="breadcrumb-item active">${navigationState.currentTeam.name}</span>
    `;

    navigationState.level = 'teamDetail';
    navigationState.currentGame = null;
    navigationState.currentPlayer = null;
}

function navigateToGameDetail() {
    if (!navigationState.currentGame) return;

    // Hide other views
    document.getElementById('teamDetailView').style.display = 'none';
    document.getElementById('playerHistoricalView').style.display = 'none';

    // Show game detail view
    document.getElementById('gameDetailView').style.display = 'block';

    // Update breadcrumb
    const breadcrumb = document.getElementById('teamBreadcrumb');
    breadcrumb.innerHTML = `
        <span class="breadcrumb-item" onclick="navigateToTeams()">Teams</span>
        <span class="breadcrumb-separator"> > </span>
        <span class="breadcrumb-item" onclick="navigateToTeamDetail()">${navigationState.currentTeam.name}</span>
        <span class="breadcrumb-separator"> > </span>
        <span class="breadcrumb-item active">${navigationState.currentGame.title}</span>
    `;

    navigationState.level = 'gameDetail';
    navigationState.currentPlayer = null;
}

function navigateBack() {
    switch (navigationState.level) {
        case 'playerHistorical':
            if (navigationState.currentGame) {
                navigateToGameDetail();
            } else if (navigationState.currentTeam) {
                navigateToTeamDetail();
            } else {
                navigateToTeams();
            }
            break;
        case 'gameDetail':
            navigateToTeamDetail();
            break;
        case 'teamDetail':
            navigateToTeams();
            break;
        default:
            navigateToTeams();
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
            <td>${game.away_team_name || game.away_team_abbr || 'N/A'}</td>
            <td style="text-align: center; font-weight: bold;">
                ${game.away_score !== null ? game.away_score : '-'} - ${game.home_score !== null ? game.home_score : '-'}
            </td>
            <td>${game.home_team_name || game.home_team_abbr || 'N/A'}</td>
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
            <td>${game.away_team_name || game.away_team_abbr || 'N/A'} @ ${game.home_team_name || game.home_team_abbr || 'N/A'}</td>
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
let syncLog = [];

function addToSyncLog(message, type = 'info') {
    const timestamp = new Date().toLocaleTimeString();
    const logEntry = { timestamp, message, type };
    syncLog.unshift(logEntry);

    // Keep only last 50 entries
    if (syncLog.length > 50) {
        syncLog = syncLog.slice(0, 50);
    }

    // Save to localStorage
    localStorage.setItem('syncLog', JSON.stringify(syncLog));

    updateSyncDisplay();
}

function updateSyncDisplay() {
    const statusDiv = document.getElementById('syncStatus');

    if (syncLog.length === 0) {
        statusDiv.innerHTML = '<p style="color: var(--text-secondary);">Ready to sync data. Click any sync button above to start.</p>';
        return;
    }

    let html = '';
    syncLog.forEach(entry => {
        let color = 'var(--text-primary)';
        let icon = 'ℹ️';

        if (entry.type === 'success') {
            color = 'var(--success)';
            icon = '✓';
        } else if (entry.type === 'error') {
            color = 'var(--danger)';
            icon = '✗';
        } else if (entry.type === 'loading') {
            color = 'var(--primary)';
            icon = '⟳';
        }

        html += `<div style="padding: 8px; border-bottom: 1px solid var(--border); color: ${color};">
            <span style="color: var(--text-secondary); font-size: 11px;">${entry.timestamp}</span>
            <span style="margin-left: 10px;">${icon} ${entry.message}</span>
        </div>`;
    });

    statusDiv.innerHTML = html;
}

function clearSyncLog() {
    syncLog = [];
    localStorage.removeItem('syncLog');
    updateSyncDisplay();
}

// Load sync log from localStorage on page load
function loadSyncLog() {
    const stored = localStorage.getItem('syncLog');
    if (stored) {
        syncLog = JSON.parse(stored);
        updateSyncDisplay();
    }
}

async function syncData(endpoint, body = {}) {
    addToSyncLog(`Starting ${endpoint} sync...`, 'loading');

    try {
        const response = await fetch(`${API_BASE_URL}/api/v1/admin/sync/${endpoint}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body)
        });

        const result = await response.json();

        if (response.ok) {
            addToSyncLog(`${result.data.message}`, 'success');

            // Refresh database status after sync
            setTimeout(refreshDatabaseStatus, 2000);
        } else {
            addToSyncLog(`Sync failed: ${result.error?.message || 'Unknown error'}`, 'error');
        }
    } catch (error) {
        addToSyncLog(`Sync failed: ${error.message}`, 'error');
    }
}

// Database Status Functions
async function refreshDatabaseStatus() {
    try {
        // Get teams count
        const teamsResult = await apiCall('/api/v1/teams');
        const teamsCount = teamsResult.data.data?.length || 0;
        document.getElementById('dbTeamsCount').textContent = teamsCount;

        // Get players count
        const playersResult = await apiCall('/api/v1/players', { limit: 1 });
        const playersCount = playersResult.data.meta?.total || 0;
        document.getElementById('dbPlayersCount').textContent = playersCount;

        // Get games count
        const gamesResult = await apiCall('/api/v1/games', { limit: 1, season: 2024 });
        const gamesCount = gamesResult.data.meta?.total || 0;
        document.getElementById('dbGamesCount').textContent = gamesCount;

        // Get weather data count (games with weather)
        const weatherResult = await apiCall('/api/v1/games', { limit: 1000, season: 2024 });
        const weatherCount = (weatherResult.data.data || []).filter(g => g.weather_temp).length;
        document.getElementById('dbWeatherCount').textContent = weatherCount;

        addToSyncLog('Database status refreshed', 'success');
    } catch (error) {
        addToSyncLog(`Failed to refresh database status: ${error.message}`, 'error');
    }
}

// Quick sync functions
async function quickSyncAll() {
    addToSyncLog('🚀 Starting comprehensive sync (this may take several minutes)...', 'loading');

    // Step 1: Sync teams
    await syncData('teams');
    await new Promise(resolve => setTimeout(resolve, 3000));

    // Step 2: Sync rosters
    await syncData('rosters');
    await new Promise(resolve => setTimeout(resolve, 3000));

    // Step 3: Sync historical games 2020-2024
    await syncData('historical/seasons', { start_year: 2020, end_year: 2024 });
    await new Promise(resolve => setTimeout(resolve, 5000));

    // Step 4: Sync NFLverse stats for recent seasons
    for (let year = 2022; year <= 2024; year++) {
        await syncData('nflverse/stats', { season: year });
        await new Promise(resolve => setTimeout(resolve, 2000));
    }

    // Step 5: Enrich with weather data
    for (let year = 2022; year <= 2024; year++) {
        await syncData('weather', { season: year });
        await new Promise(resolve => setTimeout(resolve, 2000));
    }

    addToSyncLog('🎉 Comprehensive sync completed!', 'success');
    refreshDatabaseStatus();
}

async function syncCurrent2024() {
    addToSyncLog('📅 Starting 2024 season sync...', 'loading');

    await syncData('teams');
    await new Promise(resolve => setTimeout(resolve, 2000));

    await syncData('rosters');
    await new Promise(resolve => setTimeout(resolve, 2000));

    await syncData('games');
    await new Promise(resolve => setTimeout(resolve, 2000));

    await syncData('nflverse/stats', { season: 2024 });
    await new Promise(resolve => setTimeout(resolve, 2000));

    await syncData('weather', { season: 2024 });

    addToSyncLog('✅ 2024 season sync completed!', 'success');
    refreshDatabaseStatus();
}

document.addEventListener('DOMContentLoaded', () => {
    initTabs();
    initPlayerFilters();
    initApiTesting();
    initModal();
    initDarkMode();

    // Load sync log from localStorage
    loadSyncLog();

    // Load initial data
    loadPlayers();
    loadTeams();

    // Refresh database status on load
    refreshDatabaseStatus();

    // Refresh buttons
    document.getElementById('refreshTeams').addEventListener('click', () => {
        localStorage.clear();
        loadTeams();
    });

    // Back button
    document.getElementById('backToTeams').addEventListener('click', navigateBack);

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

    // Weather sync button
    document.getElementById('syncWeather').addEventListener('click', () => {
        const season = parseInt(document.getElementById('weatherSyncSeason').value);
        syncData('weather', { season });
    });

    // Database status
    document.getElementById('refreshDbStatus').addEventListener('click', refreshDatabaseStatus);
    document.getElementById('clearSyncLog').addEventListener('click', clearSyncLog);

    // Quick action buttons
    document.getElementById('quickSyncAll').addEventListener('click', quickSyncAll);
    document.getElementById('syncCurrent2024').addEventListener('click', syncCurrent2024);
});

// Make functions available globally
window.viewPlayerDetails = viewPlayerDetails;
window.viewTeamDetail = viewTeamDetail;
window.viewGameDetail = viewGameDetail;
window.viewPlayerHistorical = viewPlayerHistorical;
window.navigateToTeams = navigateToTeams;
window.navigateToTeamDetail = navigateToTeamDetail;
window.navigateToGameDetail = navigateToGameDetail;
window.navigateBack = navigateBack;