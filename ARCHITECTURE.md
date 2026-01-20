# Architecture

Gopher Colony follows the **Hexagonal Architecture** (also known as Ports & Adapters) to ensure separation of concerns and testability.

## Directory Structure

```
├── cmd/
│   └── server/             # Application entrypoint (main.go)
├── internal/
│   ├── core/               # The "Hexagon" - Pure Business Logic
│   │   ├── domain/         # Core entities (Room, Player) - independent of everything
│   │   ├── ports/          # Interfaces (Input/Output ports)
│   │   └── services/       # Application use-cases (ConnectionManager)
│   └── adapters/           # Interface with the outside world
│       ├── primary/        # Driving Adapters (Input)
│       │   └── http/       # HTTP/WS Server handlers
│       └── secondary/      # Driven Adapters (Output)
│           └── websockets/ # WebSocket implementation (wrapper around coder/websocket)
├── web/                    # Frontend
│   └── static/             # Static files (HTML, CSS, JS)
└── tests/
    └── integration/        # End-to-end integration tests
```

## Key Decisions

### 1. Hexagonal Architecture
We rigidly separate the "Core" from "Adapters".
- **Core** can only import other Core packages.
- **Adapters** import Core (to drive it) or implement Core interfaces (to be driven).
- **Domain** imports nothing.

### 2. Services vs Managers
- **Services** (like `ConnectionManager`) orchestrate the flow. They implement `ports.ConnectionService`.

### 3. Concurrency
- `sync.Mutex` / `sync.RWMutex` will be used for protecting state in Rooms.
- A central "Ticker" loop (to be implemented in Phase 4) will drive the simulation.
- 1 Goroutine per Agent is strictly **AVOIDED**.

### 4. Dependency Injection
Dependencies are injected manually in `main.go`. This keeps the graph visible and simple without needing magic DI frameworks.
