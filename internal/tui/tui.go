// Package tui provides the TEXT mode terminal user interface for BlackSector.
//
// The TUI is built using the Bubble Tea framework (github.com/charmbracelet/bubbletea)
// and renders ANSI-colored terminal output using lipgloss for styling.
//
// # Architecture
//
// The TUI follows the Bubble Tea pattern with a root GameView model that implements
// the tea.Model interface (Init, Update, View methods). The GameView receives state
// updates from the tick engine via a channel and sends commands back to the engine
// via the session manager.
//
// Communication flow:
//   - Session goroutine runs the bubbletea program
//   - GameView receives StateUpdate messages from tick engine via update channel
//   - User input is parsed into Command structs and sent to tick engine command queue
//   - GameView renders the current state to terminal output
//
// # View Modes
//
// The TUI supports multiple view modes:
//   - COMMAND: Default mode with status bar and command prompt
//   - MARKET: Market listings when docked at a port
//   - COMBAT: Combat interface during pirate encounters
//   - MISSION: Mission board and active mission status
//   - CARGO: Cargo manifest display
//
// # State Management
//
// GameView maintains a local copy of game state received from StateUpdate broadcasts.
// This includes:
//   - Player credits
//   - Ship status (hull, shields, energy, position)
//   - Cargo contents
//   - Current system and docked port
//   - Active combat instance
//   - Active mission instance
//
// The state is updated each tick and the view is re-rendered to reflect changes.
//
// # Styling
//
// All visual styling is defined in styles.go using lipgloss. This includes:
//   - Status bar colors and formatting
//   - Command prompt styling
//   - Error message colors
//   - Market listing tables
//   - Combat interface elements
//
// # Future Extensions
//
// Phase 2 will add:
//   - Sixel image rendering for GUI mode
//   - ASCII art fallback for terminals without sixel support
//   - More complex view modes (scanning, mining, fleet management)
//   - Real-time event notifications
package tui
