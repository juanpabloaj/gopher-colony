# Communication Protocol

This document defines the WebSocket communication protocol between the Client and the Server.

## Phase 1 (Connectivity)

**Format**: Text / UTF-8 Strings.

### Client -> Server

| Type | Content | Description |
|------|---------|-------------|
| Text | Free text | Currently, the server echoes any text sent. |

### Server -> Client

| Type | Content | Description |
|------|---------|-------------|
| Text | `echo: <original_msg>` | Verification response. |

### Connection Lifecycle

1. **Connect**: Client connects to `/ws`.
2. **Handshake**: Standard HTTP Upgrade.
3. **Session**: Connection is kept open.
4. **Disconnect**:
   - Server handles clean shutdowns (Close frame).
   - Client automatically attempts reconnects after 3s (implemented in `app.js`).

---

*Future phases will introduce a structured JSON protocol.*
