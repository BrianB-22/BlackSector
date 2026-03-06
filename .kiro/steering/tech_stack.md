# Tech Stack Steering

This project is **BlackSector** — a multiplayer text-based space trading game served over SSH.

## Language
Go 1.22+. Single binary deployment targeting Linux/amd64.

Module: `github.com/yourusername/blacksector` (update to match repo)

## Key Libraries
- SSH server: `github.com/gliderlabs/ssh` + `github.com/charmbracelet/wish`
- TUI: `github.com/charmbracelet/bubbletea` + `github.com/charmbracelet/lipgloss` + `github.com/charmbracelet/bubbles`
- Image rendering: `github.com/mattn/go-sixel` (sixel protocol) + `github.com/qeesung/image2ascii` (ASCII fallback)
- Database: `modernc.org/sqlite` (pure Go, no CGo)
- Logging: `github.com/rs/zerolog`
- Testing: `github.com/stretchr/testify` + `github.com/DATA-DOG/go-sqlmock`
- UUIDs: `github.com/google/uuid`

## Project Layout
```
cmd/server/          — main game server
cmd/universe-cli/    — world gen + admin tool
internal/config/     — server.json loading
internal/db/         — SQLite layer
internal/engine/     — tick loop
internal/combat/     — combat resolution
internal/economy/    — trading, banking
internal/navigation/ — jump system
internal/missions/   — mission tracking
internal/comms/      — IRN, messaging
internal/session/    — SSH session management
internal/tui/        — bubbletea views
internal/world/      — world config loading
config/              — server.json, world/, missions/
docs/                — all specification documents
migrations/          — SQL migration files
```

## Architecture Notes
- One goroutine per SSH session (bubbletea program)
- Single-writer tick engine goroutine owns all game state mutations
- Sessions send commands via channel to tick engine; tick engine broadcasts state updates back
- SQLite in WAL mode (see docs/10_data_models/database_schema.md §2)
- All DB writes happen in tick engine goroutine; sessions are read-only except registration

## Schemas as Living Documents
The database schema, world config schema, and server.json schema in the docs are **starting specifications, not immutable contracts**. During implementation, Kiro may modify schemas if there is a good technical reason (e.g., adding an index, adjusting a column type, splitting a table for performance). Document any schema changes in a comment or update the relevant spec file. The goal is a working game, not perfect adherence to the initial schema design.
