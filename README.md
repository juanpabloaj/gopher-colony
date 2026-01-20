# Gopher Colony

**Gopher Colony** is a real-time strategy/simulation game server built with Go, focusing on clean architecture, concurrency patterns, and production-grade engineering practices.

## Phase 1: Foundation and Connectivity

Currently in **Phase 1**, the project establishes the core infrastructure:
- **Language**: Go 1.25+
- **Architecture**: Hexagonal (Ports & Adapters)
- **Communication**: WebSockets via `github.com/coder/websocket`
- **Frontend**: The main application interface (HTML5/Vanilla JS).

## Getting Started

### Prerequisites
- Go 1.25 or higher
- Make (optional, for future scripts)

### Running the Server
```bash
go run cmd/server/main.go
```
The server will start on port `8080`.

### Client
Open your browser and navigate to:
http://localhost:8080

You should see the Phase 1 dashboard. It automatically attempts to connect to the WebSocket endpoint at `/ws`.

## Tests
Run integration tests:
```bash
go test -v ./tests/integration/...
```

## Project Structure
See [ARCHITECTURE.md](ARCHITECTURE.md) for details on the codebase structure and design decisions.

## Protocol
See [PROTOCOL.md](PROTOCOL.md) for details on the WebSocket communication format.
