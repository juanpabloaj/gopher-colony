# Communication Protocol

This document defines the WebSocket communication protocol between the Client and the Server.

## Transport layer
- **URL**: `ws://HOST/ws?room=[room_id]`
- **Format**: JSON (UTF-8).

## Message Structure

All messages follow this standard envelope:
```json
{
  "type": "string",  // Message Type (cmd, init, update, echo)
  "payload": { ... } // Detailed data
}
```

### 1. Client -> Server (Commands)

**Type**: `cmd`

| Field | Type | Description |
|-------|------|-------------|
| `action` | string | The action to perform (e.g., "click") |
| `x` | int | Target X coordinate |
| `y` | int | Target Y coordinate |

**Example:**
```json
{
  "type": "cmd",
  "payload": {
    "action": "click",
    "x": 10,
    "y": 5
  }
}
```

### 2. Server -> Client

#### A. Initial Game State
**Type**: `init`
Sent immediately upon connection.

**Payload**:
- `id`: Room ID.
- `width`, `height`: World dimensions.
- `tiles`: Array of Tile objects.

**Tile Object (Optimized)**:
- `x`, `y`: Coordinates.
- `type`: Integer enum (0=Grass, 1=Water, 2=Stone).
- **Note**: `type` is `omitempty`. If missing/0, it is Grass.

**Example**:
```json
{
  "type": "init",
  "payload": {
    "width": 32,
    "height": 32,
    "tiles": [
      {"x":0, "y":0}, 
      {"x":0, "y":1, "type":2} 
    ]
  }
}
```

#### B. State Update (Broadcast)
**Type**: `update`
Sent to all clients in the room when state changes.

**Payload**:
- `tiles`: Array of changed tiles.

**Example**:
```json
{
  "type": "update",
  "payload": {
    "tiles": [
      {"x":10, "y":5, "type":2}
    ]
  }
}
```
