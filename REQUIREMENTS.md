# REQUIREMENTS.md

## 1. Project Vision
**Gopher Colony** is a project designed to deepen the understanding of **WebSockets in Go** by simulating a real-world production environment.

The objective is to develop a Real-Time Strategy (RTS) or Simulation game server, inspired by mechanics found in *Dwarf Fortress*, *Farmville*, RimWorld, or *Factorio*. Although this is a side project, the fundamental premise is to be **"Production Ready"**: code, architecture, and practices must maintain a professional standard of quality, scalability, and maintainability, strictly avoiding "example code" shortcuts.

## 2. Technological Constraints
The technology stack is strictly delimited to focus on learning and maintaining control over dependencies.

*   **Language:** Go (current stable version).
*   **WebSockets:** [`github.com/coder/websocket`](https://github.com/coder/websocket) .
*   **Logging:** `log/slog` (Go standard library for structured logging).
*   **Other Dependencies:** Keep to a minimum. Always prefer the standard library (`stdlib`) unless complexity justifies an external dependency.
*   **Frontend:** HTML5, CSS, and Vanilla JavaScript (at the beginning, no heavy frameworks like React/Vue). Rendering will be based on HTML Grids/Canvas and the use of **Unicode/Emojis/FontAwesome** for graphics.

## 3. Architecture and Software Design
The system must follow robust design principles that facilitate testability and decoupling.

### 3.1. Hexagonal Architecture (Ports & Adapters)
*   The core game logic (Domain) must be independent of the network layer (WebSockets) and persistence.
*   Business logic must not import packages from the transport layer.
*   Interfaces must be used to define input and output ports.

### 3.2. Concurrency Model & Game Loop
*   **Agent Management:** The model of assigning "1 goroutine per entity/agent" is **discarded** due to inefficiency at scale.
*   **Centralized Ticker:** Game state updates will be driven by a **Centralized Simulation Loop (Ticker)** per room.
*   **Update Pattern:** Entities must behave as state machines and implement an `Update()` interface or pattern. The central loop will iterate through active entities and invoke this method.
*   **Thread Safety:** Since WebSockets and the Game Loop will access the same state, robust synchronization mechanisms (Mutexes, Channels, or Atomics) must be implemented to prevent *Race Conditions*.

### 3.3. Multi-Room (Multiplayer)
*   The server must support multiple concurrent game instances.
*   A player must be able to choose which "world" or "room" to join.
*   The state of one room must be completely isolated from others.

## 4. Game Mechanics (Functional Scope)

### 4.1. The World (Grid)
*   The map is a grid of cells (Tiles).
*   **Layers:** A cell can contain multiple stacked entities.
    *   *Example:* Terrain Layer (Soil) + Object Layer (Plant) + Effect Layer (Water).

### 4.2. Entities and Agents
*   Entities have a lifecycle and autonomy simulated by the Backend.
*   **State Change:** A tree grows over time; a crop matures; a machine processes resources.
*   State changes occur on the server and are notified to the client (Frontend) for visualization.

### 4.3. Player Interaction
*   Control via Mouse (Click on cells).
*   Available actions depend on the entity's state (Contextual).
    *   *Example:* If it is a Tree -> Chop. If it is empty soil -> Plant.

### 4.4. Production and Transport (Advanced Phases)
*   Resource gathering system (Wood, Stone, Food).
*   Construction of transport routes (Roads, Rails).
*   Market: Ability to buy/sell/trade resources between players connected to the same room.

## 5. Development Methodology
Development will be **incremental and evolutionary**.

1.  **Self-Contained Stages:** Each development phase must result in a functional and deployable artifact. Do not proceed to the next stage until the current one is stable.
2.  **Gradual Complexity:** Start with the bare minimum (MVP) and refactor to add complexity only when the requirement demands it.
3.  **Quality:**
    *   **Testing:** Unit tests for domain logic and integration tests for the WebSocket layer.
    *   **Error Handling:** Unchecked `panics` are not allowed. Proper handling of client disconnections and network errors.
    *   **Documentation:** Keep `README`, `PROTOCOL`, and `ARCHITECTURE` updated as the project evolves.
