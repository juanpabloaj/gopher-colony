const statusDiv = document.getElementById('status');
const logsDiv = document.getElementById('logs');
const gridDiv = document.getElementById('game-grid');

// Determine protocol (ws or wss)
const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
// Get room from URL param or default to lobby
const params = new URLSearchParams(window.location.search);
const room = params.get('room') || 'lobby';
const wsUrl = `${protocol}//${window.location.host}/ws?room=${room}`;

// Update UI
document.querySelector('h1').innerHTML += ` <span class="badge" style="background:#64748b">Room: ${room}</span>`;

function log(msg) {
    const entry = document.createElement('div');
    entry.className = 'log-entry';

    const time = new Date().toLocaleTimeString();
    entry.innerHTML = `<span class="time">[${time}]</span> ${msg}`;

    logsDiv.appendChild(entry);
    logsDiv.scrollTop = logsDiv.scrollHeight;
}

function renderGrid(payload) {
    gridDiv.innerHTML = '';

    // Set grid dimensions
    gridDiv.style.gridTemplateColumns = `repeat(${payload.width}, 20px)`;
    gridDiv.style.gridTemplateRows = `repeat(${payload.height}, 20px)`;

    payload.tiles.forEach(tile => {
        const div = document.createElement('div');
        div.className = `tile tile-${tile.Terrain}`;
        div.title = `(${tile.X}, ${tile.Y})`;

        // Optional: Emoji icons
        if (tile.Terrain === 'water') div.textContent = '🌊';
        else if (tile.Terrain === 'stone') div.textContent = '🪨';
        else div.textContent = '🌱'; // Grass

        gridDiv.appendChild(div);
    });
}

function connect() {
    log(`Connecting to ${wsUrl}...`);
    const ws = new WebSocket(wsUrl);

    ws.onopen = () => {
        statusDiv.textContent = 'Connected';
        statusDiv.classList.add('connected');
        log('WebSocket connection established.');
    };

    ws.onmessage = (event) => {
        try {
            const msg = JSON.parse(event.data);

            if (msg.type === 'init') {
                log('Received Map Initialization');
                renderGrid(msg.payload);
            } else if (msg.type === 'echo') {
                log(`Echo: ${JSON.stringify(msg.payload)}`);
            } else {
                log(`RX: ${JSON.stringify(msg)}`);
            }
        } catch (e) {
            // Fallback for Phase 1 text messages (if any)
            log(`RX [Text]: ${event.data}`);
        }
    };

    ws.onclose = () => {
        statusDiv.textContent = 'Disconnected';
        statusDiv.classList.remove('connected');
        log('WebSocket connection closed. Retrying in 3s...');

        setTimeout(connect, 3000);
    };

    ws.onerror = (error) => {
        console.error("WS Error", error);
        log('WebSocket error occurred.');
    };
}

// Start connection
connect();
