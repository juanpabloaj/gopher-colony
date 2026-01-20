const statusDiv = document.getElementById('status');
const logsDiv = document.getElementById('logs');

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

function connect() {
    log(`Connecting to ${wsUrl}...`);
    const ws = new WebSocket(wsUrl);

    ws.onopen = () => {
        statusDiv.textContent = 'Connected';
        statusDiv.classList.add('connected');
        log('WebSocket connection established.');

        // Send a test hello
        ws.send('Hello from Frontend Gopher!');
    };

    ws.onmessage = (event) => {
        log(`RX: ${event.data}`);
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
