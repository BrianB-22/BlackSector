# Headless Server Specification

## Version: 0.1

## Status: Draft

## Owner: Core Architecture

## Last Updated: 2026-03-05

---

# 1. Purpose

Defines the architecture of the BlackSector simulation server.

The server is a headless process — it has no graphical interface, no interactive terminal display, and no dependency on a display environment. All player interaction occurs through SSH (TEXT mode) or TLS/TCP (GUI mode, future).

The server is a single Go binary that runs as a background service.

---

# 2. Scope

IN SCOPE:

* server process architecture
* startup and shutdown sequence
* tick loop design
* subsystem initialization order
* signal handling
* admin interface
* port binding
* configuration loading

OUT OF SCOPE:

* protocol message format (see `protocol_spec_v1.md`)
* session lifecycle (see `session_management.md`)
* deployment environment (see `deployment_model.md`)
* telemetry and logging detail (see `telemetry_logging.md`)

---

# 3. Design Principles

* Single binary — no runtime dependencies beyond the OS and SQLite
* Single-threaded tick engine — simulation is deterministic and not parallelized
* Headless — no stdout display, no TTY requirement
* Config-driven — all tunable parameters in configuration files
* Graceful shutdown — in-progress tick completes before exit
* Server authoritative — clients cannot influence simulation directly

---

# 4. Server Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    BlackSector Server                    │
│                                                         │
│  ┌──────────────┐   ┌──────────────┐                    │
│  │  SSH Listener│   │  TCP Listener│                    │
│  │  port 2222   │   │  port 2223   │                    │
│  └──────┬───────┘   └──────┬───────┘                    │
│         │                  │                            │
│         └────────┬─────────┘                            │
│                  │                                      │
│         ┌────────▼────────┐                             │
│         │ Session Manager │                             │
│         └────────┬────────┘                             │
│                  │                                      │
│         ┌────────▼────────┐                             │
│         │  Command Queue  │                             │
│         └────────┬────────┘                             │
│                  │                                      │
│  ┌───────────────▼───────────────────────────────────┐  │
│  │                   Tick Engine                     │  │
│  │                                                   │  │
│  │  Navigation │ Combat │ Mining │ Economy │ Missions │  │
│  └───────────────────────────────────────────────────┘  │
│                  │                                      │
│         ┌────────▼────────┐                             │
│         │   Persistence   │                             │
│         │ SQLite + Snapshots                            │
│         └─────────────────┘                             │
│                                                         │
│  ┌──────────────────────────────────────────────────┐   │
│  │              Admin Interface (stdin)              │   │
│  └──────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

---

# 5. Startup Sequence

The server follows a strict startup order:

```
1. Load server configuration (config/server.json)
2. Initialize logger
3. Open SQLite database
4. Load static world data (regions, systems, ports, jump_connections)
5. Load configuration data (commodities, ship_classes, upgrades, economic_events)
6. Load mission definitions from config/missions/
7. Load snapshot (if present) or initialize empty world state
8. Initialize tick engine
9. Initialize subsystems (navigation, combat, mining, economy, missions, exploration)
10. Initialize session manager
11. Start admin interface (stdin reader goroutine)
12. Bind SSH listener on port 2222
13. Bind TLS/TCP listener on port 2223 (if TLS certificate configured)
14. Begin tick loop
15. Log: server_start event
```

If any step in 1–11 fails, the server exits with a descriptive error message and non-zero exit code.

If TLS certificate is not configured, port 2223 is not bound and GUI mode is unavailable.

---

# 6. Tick Loop

The tick loop is the heartbeat of the simulation.

```
while running:
    tick_start = now()

    1. Process command queue (dequeue all commands for this tick)
    2. Execute navigation updates
    3. Execute combat resolution
    4. Execute mining resolution
    5. Execute economic tick (port inventory, price adjustment)
    6. Execute economic event evaluation
    7. Emit game events to event bus
    8. Execute mission evaluation (consumes events from bus)
    9. Distribute mission rewards
    10. Execute AI trader decisions
    11. Flush state to SQLite (dirty records only)
    12. Emit protocol messages to connected sessions
    13. Write event log entries (async)
    14. Evaluate snapshot trigger (write if interval reached)

    tick_duration = now() - tick_start
    if tick_duration > slow_tick_threshold:
        log tick_slow event
    sleep max(0, tick_interval_ms - tick_duration)
    tick_number++
```

The tick loop runs on the main goroutine. No simulation logic executes outside the tick loop.

Default tick interval: 2000ms (configurable).
Slow tick threshold: 500ms (configurable, logs warning).

---

# 7. Subsystem Initialization

Each subsystem is initialized before the tick loop begins. Subsystems are stateless modules that operate on shared world state.

| Subsystem  | Responsibilities                                       |
| ---------- | ------------------------------------------------------ |
| Navigation | Ship movement, jump point traversal, waypoint routing  |
| Combat     | Combat round resolution, damage calculation            |
| Mining     | Yield extraction, depletion, hazard evaluation         |
| Economy    | Port pricing, supply/demand, commodity flow            |
| EconEvents | Economic event spawning, modifier application          |
| Missions   | Objective tracking, reward distribution                |
| Exploration| Scan resolution, anomaly discovery                     |
| AITraders  | NPC trade decisions, route selection, movement         |

---

# 8. Listener Architecture

## SSH Listener (port 2222)

Accepts SSH connections. Authentication is handled by the SSH transport layer. After authentication succeeds, the application handshake begins.

Each SSH session runs in a goroutine. Commands from the session are enqueued into the command queue. Protocol responses are sent back to the session goroutine for transmission.

## TLS/TCP Listener (port 2223)

Accepts TLS connections. TLS handshake must complete before any protocol traffic is exchanged. Plaintext connections are rejected immediately.

Requires a valid TLS certificate and key, configured in `server.json`.

If certificate is absent or invalid, port 2223 is not opened.

---

# 9. Command Queue

Player commands submitted via protocol messages are enqueued into the per-session command queue.

Commands are dequeued at the start of each tick (step 1 of tick loop).

Maximum commands per tick per session: 3 (configurable). Excess commands are rejected with `error_rate_limited`.

See `command_queue.md` for full detail.

---

# 10. Admin Interface

The server reads admin commands from stdin.

Admin interface is a simple line-oriented command processor. It runs in a separate goroutine and enqueues admin commands for execution at the start of the next tick.

The admin interface is the only direct human interaction with the server process. It does not require authentication — physical access to the server process is assumed.

See `cli_management_tool.md` for available commands.

---

# 11. Graceful Shutdown

Shutdown is triggered by:

* `SIGTERM` or `SIGINT` signal
* `shutdown` admin command

Shutdown sequence:

```
1. Stop accepting new connections
2. Send server_shutdown protocol message to all connected sessions
3. Wait for current tick to complete (if in progress)
4. Write final snapshot
5. Write server_stop event log entry
6. Close all sessions
7. Close SQLite database
8. Exit with code 0
```

In-progress tick is never interrupted. The server may take up to one full tick interval to shut down cleanly.

---

# 12. Signal Handling

| Signal  | Behavior              |
| ------- | --------------------- |
| SIGTERM | Graceful shutdown     |
| SIGINT  | Graceful shutdown     |
| SIGHUP  | Reload configuration  |
| SIGKILL | Immediate exit        |

SIGHUP triggers a hot-reload of mission definitions and economic event definitions. World structure (systems, ports) is not reloaded via SIGHUP.

---

# 13. Configuration

Server configuration is loaded from `config/server.json` at startup.

Key configuration fields:

```json
{
  "server_name": "Black Sector",
  "ssh_port": 2222,
  "tcp_port": 2223,
  "tls_cert_file": "certs/server.crt",
  "tls_key_file": "certs/server.key",
  "tick_interval_ms": 2000,
  "slow_tick_threshold_ms": 500,
  "snapshot_interval_ticks": 100,
  "snapshot_directory": "snapshots/",
  "snapshot_retention_count": 10,
  "database_path": "data/blacksector.db",
  "log_directory": "logs/",
  "max_concurrent_players": 100,
  "handshake_timeout_seconds": 30,
  "linger_timeout_seconds": 300
}
```

---

# 14. Performance Targets

| Metric                       | Target      |
| ---------------------------- | ----------- |
| Tick duration (normal)       | < 100ms     |
| Tick duration (slow warning) | < 500ms     |
| Concurrent players supported | 50–100      |
| Ports supported              | up to 1,000 |
| Systems supported            | 500–1,000   |
| Command enqueue latency      | < 1ms       |

---

# 15. PRNG Seeding and Determinism

The simulation uses two distinct PRNG contexts:

## 15.1 Universe Generation PRNG

Used during world generation only (procedural placement of systems, ports, resources, security ratings).

```
seed = universe_seed (integer from server.json "universe_seed" field)
rng  = rand.New(rand.NewSource(seed))
```

All procedural generation runs in deterministic order from this seed. Given the same `universe_seed`, the generated universe is always identical.

## 15.2 Tick PRNG

Used during simulation for per-tick randomness (pirate spawn checks, hazard rolls, mining yields, combat variance).

```
tick_seed = universe_seed XOR (uint64(tick_number) << 32)
rng       = rand.New(rand.NewSource(int64(tick_seed)))
```

A new PRNG is constructed fresh each tick from this seed. This ensures:
- Deterministic results for any given tick number
- Results differ between ticks without global mutable state
- Replay is possible: re-run from tick N with same seed to reproduce tick N outcomes

**Important:** Do not use a single shared global `rand` — use `rand.New(rand.NewSource(...))` explicitly. The global `rand` is not seeded deterministically and is not safe for concurrent use.

## 15.3 server.json universe_seed field

```json
"universe_seed": 42
```

This value seeds both universe generation and all tick randomness. Changing it produces a different universe and different simulation outcomes.

---

# 16. Non-Goals (v1)

* Multi-process or distributed architecture
* In-process hot reload of world structure
* Clustering or high availability
* Web-based admin dashboard
* Metrics endpoint (Prometheus, etc.)

---

# End of Document
