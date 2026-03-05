# Tech Stack

## Version: 1.0
## Status: Decided
## Owner: Core Architecture
## Last Updated: 2026-03-05

---

# 1. Language

**Go 1.22+**

Rationale:
- Single binary deployment — no runtime dependencies on the server
- Excellent concurrency primitives for the tick engine + concurrent SSH sessions
- Strong standard library (crypto/ssh, net, encoding/json)
- Fast compile times; easy cross-compilation for Linux server targets
- Goroutines map naturally to per-session player state

Minimum version: **Go 1.22** (for range-over-integer, improved slog)

---

# 2. Module Name

```
module github.com/yourusername/blacksector
```

Update to match actual GitHub repository before first commit.

---

# 3. Core Libraries

## 3.1 SSH Server

**`github.com/gliderlabs/ssh`** (primary)

Wraps `golang.org/x/crypto/ssh` with a simpler handler API. Used for the player-facing SSH server on port 2222.

**`github.com/charmbracelet/wish`** (middleware layer)

Provides bubbletea integration for SSH sessions. Used to bridge SSH connections into the bubbletea TUI framework.

```
go get github.com/gliderlabs/ssh
go get github.com/charmbracelet/wish
```

## 3.2 TUI Framework

**`github.com/charmbracelet/bubbletea`**

The core TUI application model. Uses The Elm Architecture (Model/Update/View). Each SSH session runs an independent bubbletea program.

**`github.com/charmbracelet/lipgloss`**

Declarative terminal styling: colors, borders, padding, alignment. Used for all UI layout and ANSI color rendering.

**`github.com/charmbracelet/bubbles`**

Pre-built components:
- `table` — market listings, cargo manifests, player lists
- `progress` — progress bars for shield recharge, fuel, mission progress
- `textinput` — command input line
- `spinner` — tick processing indicator
- `viewport` — scrollable text regions (combat log, message inbox)

```
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles
```

## 3.3 Image Rendering

Two-tier approach based on terminal capability detection:

**Tier 1 — Sixel/Kitty Protocol (capable terminals: WezTerm, iTerm2, foot, kitty)**

`github.com/mattn/go-sixel` — sixel image encoding for pixel-level image display.

Terminal capability is detected from the `TERM` and `COLORTERM` environment variables passed in the SSH session. If sixel or kitty graphics are supported, render ship images and system maps as pixel art.

**Tier 2 — Unicode block art (all terminals)**

Use Unicode block characters (`█▓▒░`) and lipgloss colors for "images" that work on any terminal. This is the default fallback.

ASCII art fallback via `github.com/qeesung/image2ascii` for terminals without sixel support.

```
go get github.com/mattn/go-sixel
go get github.com/qeesung/image2ascii
```

Image assets stored in `assets/images/`. At startup, pre-render all images to both sixel and ASCII formats and cache in memory.

## 3.4 Database

**`modernc.org/sqlite`**

Pure Go SQLite — no CGo, no external .so dependency. Simpler build, simpler deployment. Single-writer, multi-reader via WAL mode.

Do NOT use `mattn/go-sqlite3` — it requires CGo which complicates cross-compilation.

```
go get modernc.org/sqlite
```

All required PRAGMAs are applied on every connection open (see `docs/10_data_models/database_schema.md` Section 2).

## 3.5 Logging

**`github.com/rs/zerolog`**

Zero-allocation structured logging. Two log channels:

1. **`server.log`** — INFO level by default. Server events, player connections, errors.
2. **`debug.log`** — DEBUG level, enabled via `server.json`. Verbose: every tick event, every command, every DB query, every state transition. Optimized for AI-assisted troubleshooting.

```
go get github.com/rs/zerolog
```

Log configuration (see `docs/11_server/server_config_schema.md`):
```json
"logging": {
  "level": "info",
  "log_file": "server.log",
  "debug_log_enabled": false,
  "debug_log_path": "debug.log",
  "debug_log_include_sql": false,
  "debug_log_include_tick_detail": true
}
```

When `debug_log_enabled = true`, the debug log captures:
- Every tick start/end with duration
- Every command received (player, content, tick)
- Every state machine transition
- Every combat event (rolls, damage, outcomes)
- Every DB write (table, operation, affected rows) if `debug_log_include_sql = true`
- Every session connect/disconnect

## 3.6 Testing

**`github.com/stretchr/testify`** — assertions (`assert`, `require`) and mock support

**`github.com/DATA-DOG/go-sqlmock`** — mock SQLite for DB layer unit tests without touching real DB

All tests use table-driven format. Test coverage target: 80%+ for game logic packages.

```
go get github.com/stretchr/testify
go get github.com/DATA-DOG/go-sqlmock
```

## 3.7 UUID

**`github.com/google/uuid`**

Used for player_id, ship_id, account_id, and other primary keys requiring globally unique identifiers.

```
go get github.com/google/uuid
```

## 3.8 Config

**`encoding/json`** (stdlib)

Server config loaded from `config/server.json` at startup. No live reloading — changes require server restart. Schema documented in `docs/11_server/server_config_schema.md`.

---

# 4. Project Structure

```
blacksector/
├── cmd/
│   ├── server/         — main SSH game server (main.go)
│   └── universe-cli/   — world generation and admin CLI tool
├── internal/
│   ├── config/         — server.json loading and validation
│   ├── db/             — SQLite connection, migrations, query helpers
│   ├── engine/         — tick engine, game loop
│   ├── combat/         — combat resolution, NPC AI
│   ├── economy/        — trading, pricing, banking
│   ├── navigation/     — jump system, pathfinding
│   ├── missions/       — mission loading, tracking, completion
│   ├── comms/          — IRN, proximity messaging, mailbox
│   ├── session/        — SSH session management, command dispatch
│   ├── tui/            — bubbletea views and components
│   └── world/          — world config loading, system/port models
├── config/
│   ├── server.json     — server configuration
│   ├── missions/       — mission JSON files
│   └── world/          — world config JSON files
├── assets/
│   └── images/         — ship images, system art
├── docs/               — all specification documents
├── migrations/         — SQLite migration SQL files
└── Makefile
```

---

# 5. Build and Tooling

**Makefile targets:**

```makefile
make build          — build server binary (cmd/server)
make build-cli      — build universe-cli binary
make test           — run all tests
make test-verbose   — run tests with -v flag
make lint           — run golangci-lint
make run            — build and run server with default config
make migrate        — apply DB migrations
make clean          — remove build artifacts
```

**Linter:** `golangci-lint` with standard ruleset + `errcheck`, `govet`, `staticcheck`.

**Go version enforcement:** `go.mod` pins minimum Go version. CI fails if build requires a newer version.

---

# 6. Deployment Target

Single Linux binary. Deployed to a VPS or bare metal server.

Target OS: Linux/amd64 (primary), Linux/arm64 (Raspberry Pi compatible)

Cross-compile from macOS or Windows:
```bash
GOOS=linux GOARCH=amd64 go build -o blacksector-server ./cmd/server
```

No Docker required (single binary + SQLite file + config directory).

---

# End of Document
