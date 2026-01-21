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

// Terrain Configuration
const TERRAIN_MAP = {
    0: 'grass',
    1: 'water',
    2: 'stone',
    3: 'sapling',
    4: 'tree'
};

function updateTileVisuals(div, terrainType) {
    const defaultType = 0; // Grass
    // Handle cases where terrainType might be undefined
    const type = terrainType !== undefined ? terrainType : defaultType;
    const terrainName = TERRAIN_MAP[type] || 'grass';

    // Update Class
    div.className = `tile tile-${terrainName}`;

    // Update Emoji Content
    if (terrainName === 'water') div.textContent = '🌊';
    else if (terrainName === 'stone') div.textContent = '🪨';
    else if (terrainName === 'sapling') div.textContent = '🌱';
    else if (terrainName === 'tree') div.textContent = '🌳';
    else div.textContent = ''; // Grass
}

function renderGrid(payload) {
    gridDiv.innerHTML = '';
    gopherMap.clear(); // Reset gophers

    // Set grid dimensions
    gridDiv.style.gridTemplateColumns = `repeat(${payload.width}, 20px)`;
    gridDiv.style.gridTemplateRows = `repeat(${payload.height}, 20px)`;

    payload.tiles.forEach(tile => {
        const div = document.createElement('div');
        div.id = `tile-${tile.x}-${tile.y}`; // Add ID for easy update
        div.title = `(${tile.x}, ${tile.y})`;
        div.onclick = () => sendClick(tile.x, tile.y);

        // Apply visual styling
        updateTileVisuals(div, tile.type);

        gridDiv.appendChild(div);
    });
}

function updateTile(x, y, terrain) {
    const div = document.getElementById(`tile-${x}-${y}`);
    if (div) {
        updateTileVisuals(div, terrain);

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
            // Render initial gophers
            if (msg.payload.gophers) {
                msg.payload.gophers.forEach(g => updateGopher(g));
            }
        } else if (msg.type === 'update') {
            // Handle delta update
            if (msg.payload.tiles) {
                msg.payload.tiles.forEach(t => updateTile(t.x, t.y, t.type));
            }
            if (msg.payload.gophers) {
                msg.payload.gophers.forEach(g => updateGopher(g));
            }
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

// Map to track gopher DOM elements by ID
const gopherMap = new Map();

function updateGopher(gopher) {
    let div = gopherMap.get(gopher.id);

    if (!div) {
        // Create new Gopher
        div = document.createElement('div');
        div.className = 'gopher';
        div.id = `gopher-${gopher.id}`;
        gridDiv.appendChild(div);
        gopherMap.set(gopher.id, div);
    }

    // Update visual content based on inventory
    let html = '<span class="gopher-body">🐹</span>';
    if (gopher.inventory && gopher.inventory.wood > 0) {
        html += '<span class="gopher-cargo">🪵</span>';
        div.classList.add('carrying');
    } else {
        div.classList.remove('carrying');
    }
    div.innerHTML = html;

    // Calculate position
    // Tiles are 20px, Gap is 1px, Padding is 10px
    const CELL_SIZE = 20;
    const GAP = 1;
    const PADDING = 10;

    const left = PADDING + gopher.x * (CELL_SIZE + GAP);
    const top = PADDING + gopher.y * (CELL_SIZE + GAP);

    div.style.left = `${left}px`;
    div.style.top = `${top}px`;
}

// Start connection
connect();
