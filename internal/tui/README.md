# TUI Package

The `tui` package provides the TEXT mode terminal user interface for BlackSector using the Bubble Tea framework.

## Architecture

### GameView Model

`GameView` is the root bubbletea model that implements the `tea.Model` interface:
- `Init()` - Initializes the model and returns a command to wait for state updates
- `Update(tea.Msg)` - Handles messages (keyboard input, state updates, window resize)
- `View()` - Renders the current state to a string for terminal display

### Communication Flow

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│   Session   │────────▶│   GameView   │────────▶│ Tick Engine │
│  Goroutine  │         │  (bubbletea) │         │             │
└─────────────┘         └──────────────┘         └─────────────┘
       ▲                        │                        │
       │                        │                        │
       │                        ▼                        │
       │                 Update Channel                  │
       │                (StateUpdate)                    │
       │                                                 │
       └─────────────────────────────────────────────────┘
                    Command Queue (Commands)
```

1. User types commands in the TUI
2. GameView parses commands and sends them to the tick engine via command queue
3. Tick engine processes commands and broadcasts StateUpdate messages
4. GameView receives StateUpdate via update channel and re-renders

### View Modes

The TUI supports multiple view modes:

- **COMMAND** - Default view with status bar and command prompt
- **MARKET** - Market listings when docked at a port
- **COMBAT** - Combat interface during pirate encounters
- **MISSION** - Mission board and active mission status
- **CARGO** - Cargo manifest display
- **HELP** - Help screen with available commands

### State Management

`GameState` holds the local copy of game state received from StateUpdate broadcasts:
- Player credits
- Ship status (hull, shields, energy)
- Cargo contents
- Current system and docked port
- Active combat instance (future)
- Active mission instance (future)

## Styling

All visual styles are defined in `styles.go` using lipgloss:

- **Status bar** - Cyan background with ship stats
- **Command prompt** - Cyan bold text
- **Error messages** - Red bold text
- **Info messages** - Sky blue text
- **Success messages** - Green bold text
- **Health indicators** - Color-coded based on percentage (green > 70%, orange > 30%, red < 30%)

## Usage Example

```go
// Create channels
updateChan := make(chan interface{}, 10)
commandOut := make(chan interface{}, 10)

// Create GameView
view := tui.NewGameView(sessionID, playerID, updateChan, commandOut, logger)

// Run bubbletea program
program := tea.NewProgram(view)
if err := program.Start(); err != nil {
    log.Fatal(err)
}
```

## Future Extensions

Phase 2 will add:
- Sixel image rendering for GUI mode
- ASCII art fallback for terminals without sixel support
- More complex view modes (scanning, mining, fleet management)
- Real-time event notifications
- Animated transitions between views

## Testing

Run tests with:
```bash
go test ./internal/tui -v
```

Current test coverage focuses on:
- GameView initialization
- View mode constants
- Basic model structure

Integration tests with full state updates will be added in subsequent tasks.
