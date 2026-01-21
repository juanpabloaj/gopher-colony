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
document.getElementById('room-input').value = room;

function changeRoom() {
    const newRoom = document.getElementById('room-input').value;
    if (newRoom && newRoom !== room) {
        const url = new URL(window.location);
        url.searchParams.set('room', newRoom);
        window.location.href = url.toString();
    }
}

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

    const TERRAIN_MAP = {
        0: 'grass',
        1: 'water',
        2: 'stone',
        3: 'sapling',
        4: 'tree'
    };

    payload.tiles.forEach(tile => {
        // Default to Grass (0) if terrain is undefined (omitempty)
        // Use 'type' property (mapped from Go struct tag)
        const terrainType = tile.type !== undefined ? tile.type : 0;
        const terrainName = TERRAIN_MAP[terrainType] || 'grass';

        const div = document.createElement('div');
        div.id = `tile-${tile.x}-${tile.y}`; // Add ID for easy update
        div.className = `tile tile-${terrainName}`;
        div.title = `(${tile.x}, ${tile.y})`;
        div.onclick = () => sendClick(tile.x, tile.y);

        // Optional: Emoji icons
        if (terrainName === 'water') div.textContent = '🌊';
        else if (terrainName === 'stone') div.textContent = '🪨';
        else if (terrainName === 'sapling') div.textContent = '🌱';
        else if (terrainName === 'tree') div.textContent = '🌳';
        else div.textContent = ''; // Grass (empty or dot)

        gridDiv.appendChild(div);
    });
}

function updateTile(x, y, terrain) {
    const TERRAIN_MAP = {
        0: 'grass',
        1: 'water',
        2: 'stone',
        3: 'sapling',
        4: 'tree'
    };
    const terrainName = TERRAIN_MAP[terrain] || 'grass';

    const div = document.getElementById(`tile-${x}-${y}`);
    if (div) {
        div.className = `tile tile-${terrainName}`;
        if (terrainName === 'water') div.textContent = '🌊';
        else if (terrainName === 'stone') div.textContent = '🪨';
        else if (terrainName === 'sapling') div.textContent = '🌱';
        else if (terrainName === 'tree') div.textContent = '🌳';
        else div.textContent = ''; // Grass

        // Flash effect
        div.style.filter = "brightness(2)";
        setTimeout(() => div.style.filter = "", 200);
    }
}

function sendClick(x, y) {
    if (!wsGlobal || wsGlobal.readyState !== WebSocket.OPEN) return;

    const cmd = {
        type: "cmd",
        payload: {
            action: "click",
            x: x,
            y: y
        }
    };
    wsGlobal.send(JSON.stringify(cmd));
}

function connect() {
    log(`Connecting to ${wsUrl}...`);
    const ws = new WebSocket(wsUrl);
    wsGlobal = ws;

    ws.onopen = () => {
        statusDiv.textContent = 'Connected';
        statusDiv.classList.add('connected');
        log('WebSocket connection established.');
    };

    ws.onmessage = (event) => {
        const msg = JSON.parse(event.data);

        if (msg.type === 'init') {
            log('Received Map Initialization');
            renderGrid(msg.payload);
        } else if (msg.type === 'update') {
            // Handle delta update
            msg.payload.tiles.forEach(t => updateTile(t.x, t.y, t.type));
        } else if (msg.type === 'echo') {
            log(`Echo: ${JSON.stringify(msg.payload)}`);
        } else {
            log(`RX: ${JSON.stringify(msg)}`);
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
