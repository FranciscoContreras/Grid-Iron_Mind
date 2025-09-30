// Configuration
const API_BASE_URL = 'http://localhost:3000'; // Change to your Vercel URL when deployed
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
        tbody.innerHTML = '<tr><td colspan="6" style="text-align: center; padding: 40px;">No players found</td></tr>';
        return;
    }

    filteredPlayers.forEach(player => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td><strong>${player.name}</strong></td>
            <td>${player.position}</td>
            <td>--</td>
            <td>${player.jersey_number || '--'}</td>
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
document.addEventListener('DOMContentLoaded', () => {
    initTabs();
    initPlayerFilters();
    initApiTesting();
    initModal();
    initDarkMode();

    // Load initial data
    loadPlayers();
    loadTeams();

    // Refresh teams button
    document.getElementById('refreshTeams').addEventListener('click', () => {
        localStorage.clear();
        loadTeams();
    });
});

// Make viewPlayerDetails available globally
window.viewPlayerDetails = viewPlayerDetails;