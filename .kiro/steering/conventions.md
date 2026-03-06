# Coding Conventions

## Error Handling
- Always wrap errors: `fmt.Errorf("context: %w", err)`
- Errors from DB and I/O must never be silently dropped
- Fatal errors (startup config missing, DB unreachable) log and `os.Exit(1)`
- Game logic errors (invalid command, insufficient credits) return typed errors to the session layer — never crash the server

## Testing
- All game logic packages must have tests
- Use table-driven tests for formulas and state machines:
  ```go
  tests := []struct{ name string; input X; want Y }{ ... }
  for _, tt := range tests { t.Run(tt.name, func(t *testing.T) { ... }) }
  ```
- Use `testify/assert` for non-fatal assertions, `testify/require` when a failure should stop the test
- Mock the DB with `go-sqlmock` in unit tests; use a real in-memory SQLite for integration tests
- Test coverage target: 80%+ for `internal/combat`, `internal/economy`, `internal/engine`

## Logging
- Use `zerolog` structured fields — never `fmt.Println` in production code
- Server log (INFO): connection events, tick warnings, errors
- Debug log (DEBUG, only when enabled): every command, every combat roll, every state change
- Log format: `log.Debug().Str("player", id).Int("tick", tick).Str("cmd", cmd).Msg("command received")`

## Database
- All queries use prepared statements or parameterized queries — no string concatenation in SQL
- Apply WAL PRAGMAs on every new connection (see database_schema.md §2)
- Wrap related writes in a single transaction
- Column naming: snake_case matching the schema in docs/10_data_models/database_schema.md

## Concurrency
- The tick engine owns game state. Sessions must not write state directly
- Commands are sent from session goroutines to tick engine via buffered channel
- State updates are broadcast from tick engine to sessions via per-session channel
- Use `sync.RWMutex` only for read caches (e.g., world config, ship class definitions loaded at startup)

## Package Naming
- `internal/` — all game logic (not exported)
- Short, lowercase package names: `combat`, `engine`, `economy`, `navigation`
- No stutter: `combat.Resolve()` not `combat.ResolveCombat()`

## Config and Constants
- All tunable values come from `config/server.json` (loaded into a `Config` struct at startup)
- No magic numbers in game logic — reference the Config struct
- Ship class stats loaded from world config at startup, not hardcoded

## TUI
- Each view is a separate bubbletea `Model` with its own `Init()`, `Update()`, `View()` methods
- Use lipgloss styles defined in a central `tui/styles.go` file
- Terminal capability detection (sixel support) happens once at session init and is stored in session state
- Image rendering: attempt sixel first; fall back to ASCII art if unsupported

## Specs are Guidance, Not Constraints
The docs/ specs define intended behavior. If implementation reveals a spec is impractical or ambiguous, implement the sensible solution and note the deviation. The goal is a working game.
