package tui

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog"
	"github.com/BrianB-22/BlackSector/internal/db"
	"github.com/BrianB-22/BlackSector/internal/engine"
)

// ViewMode represents the current display mode of the TUI
type ViewMode string

const (
	ViewModeCommand ViewMode = "COMMAND" // Default command prompt view
	ViewModeMarket  ViewMode = "MARKET"  // Market listings (when docked)
	ViewModeCombat  ViewMode = "COMBAT"  // Combat interface
	ViewModeMission ViewMode = "MISSION" // Mission board
	ViewModeCargo   ViewMode = "CARGO"   // Cargo manifest
	ViewModeHelp    ViewMode = "HELP"    // Help screen
)

// GameView is the root bubbletea model for the BlackSector TUI.
// It receives state updates from the tick engine and renders the game state.
type GameView struct {
	// Session information
	sessionID string
	playerID  string

	// Current game state (from StateUpdate broadcasts)
	currentState *GameState

	// View state
	viewMode      ViewMode
	commandBuffer string
	errorMessage  string
	infoMessage   string

	// Channels for communication
	updateChan <-chan interface{} // Receives StateUpdate from tick engine
	commandOut chan<- interface{} // Sends commands to tick engine (via session)

	// Logger
	logger zerolog.Logger

	// Terminal dimensions
	width  int
	height int
}

// GameState represents the current game state for this player.
// This is populated from StateUpdate messages broadcast by the tick engine.
type GameState struct {
	// Tick information
	tickNumber int64
	timestamp  int64

	// Player state
	credits int64

	// Ship state
	ship  *db.Ship
	cargo []db.CargoSlot

	// Location state
	currentSystemID   int
	currentSystemName string
	dockedPortID      *int
	dockedPortName    string

	// Combat state (future)
	activeCombat interface{} // Will be *combat.CombatInstance in Phase 1

	// Mission state (future)
	activeMission interface{} // Will be *missions.MissionInstance in Phase 1
}

// NewGameView creates a new GameView instance.
// The updateChan receives StateUpdate messages from the tick engine.
// The commandOut channel is used to send commands back to the engine.
func NewGameView(
	sessionID string,
	playerID string,
	updateChan <-chan interface{},
	commandOut chan<- interface{},
	logger zerolog.Logger,
) *GameView {
	return &GameView{
		sessionID:     sessionID,
		playerID:      playerID,
		currentState:  &GameState{},
		viewMode:      ViewModeCommand,
		commandBuffer: "",
		updateChan:    updateChan,
		commandOut:    commandOut,
		logger:        logger,
		width:         80,  // Default width
		height:        24,  // Default height
	}
}

// Init initializes the bubbletea model.
// It returns a command to wait for state updates from the tick engine.
func (m *GameView) Init() tea.Cmd {
	return m.waitForStateUpdate()
}

// Update handles messages and updates the model state.
// This is called by bubbletea for every message received.
func (m *GameView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case StateUpdateMsg:
		// Update game state from tick engine broadcast
		m.updateGameState(msg)
		return m, m.waitForStateUpdate()

	case errMsg:
		m.errorMessage = msg.Error()
		return m, nil

	default:
		return m, nil
	}
}

// View renders the current state to a string for terminal display.
// The output includes status bar, main content area, and command prompt.
func (m *GameView) View() string {
	var b strings.Builder

	// Render status bar
	b.WriteString(m.renderStatusBar())
	b.WriteString("\n\n")

	// Render main content based on view mode
	switch m.viewMode {
	case ViewModeCommand:
		b.WriteString(m.renderCommandView())
	case ViewModeMarket:
		b.WriteString(m.renderMarketView())
	case ViewModeCombat:
		b.WriteString(m.renderCombatView())
	case ViewModeMission:
		b.WriteString(m.renderMissionView())
	case ViewModeCargo:
		b.WriteString(m.renderCargoView())
	case ViewModeHelp:
		b.WriteString(m.renderHelpView())
	default:
		b.WriteString("Unknown view mode\n")
	}

	b.WriteString("\n")

	// Render error/info messages
	if m.errorMessage != "" {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %s", m.errorMessage)))
		b.WriteString("\n")
	}
	if m.infoMessage != "" {
		b.WriteString(infoStyle.Render(m.infoMessage))
		b.WriteString("\n")
	}

	// Render command prompt
	b.WriteString(m.renderCommandPrompt())

	return b.String()
}

// handleKeyPress processes keyboard input
func (m *GameView) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		// Exit the program
		return m, tea.Quit

	case tea.KeyEnter:
		// Process command
		return m, m.processCommand()

	case tea.KeyBackspace:
		// Remove last character from command buffer
		if len(m.commandBuffer) > 0 {
			m.commandBuffer = m.commandBuffer[:len(m.commandBuffer)-1]
		}
		return m, nil

	default:
		// Add character to command buffer
		if msg.Type == tea.KeyRunes {
			m.commandBuffer += string(msg.Runes)
		}
		return m, nil
	}
}

// processCommand parses and executes the current command buffer
func (m *GameView) processCommand() tea.Cmd {
	cmdInput := strings.TrimSpace(m.commandBuffer)
	m.commandBuffer = "" // Clear buffer

	if cmdInput == "" {
		return nil
	}

	// Clear previous messages
	m.errorMessage = ""
	m.infoMessage = ""

	// Parse command using the parser
	parsedCmd, err := ParseCommand(cmdInput)
	if err != nil {
		m.errorMessage = err.Error()
		return nil
	}

	// Handle local commands (view switching)
	if parsedCmd.IsLocal {
		switch parsedCmd.Type {
		case "help":
			m.viewMode = ViewModeHelp
		case "market":
			m.viewMode = ViewModeMarket
		case "cargo":
			m.viewMode = ViewModeCargo
		case "missions":
			m.viewMode = ViewModeMission
		case "system":
			// System info is displayed in command view
			m.viewMode = ViewModeCommand
			m.infoMessage = "System information displayed"
		}
		return nil
	}

	// Handle engine commands (navigation, trading)
	// These need to be sent to the tick engine via commandOut channel
	if err := m.sendCommandToEngine(parsedCmd); err != nil {
		m.errorMessage = err.Error()
		return nil
	}

	m.infoMessage = fmt.Sprintf("Command '%s' sent to server", parsedCmd.Type)
	return nil
}

// sendCommandToEngine sends a parsed command to the tick engine
func (m *GameView) sendCommandToEngine(parsedCmd *ParsedCommand) error {
	// For buy/sell commands, we need to fill in the port_id from current state
	if parsedCmd.Type == "buy" || parsedCmd.Type == "sell" {
		if m.currentState.dockedPortID == nil {
			return fmt.Errorf("must be docked at a port to trade")
		}

		// Update the payload with the current docked port ID
		var payload map[string]interface{}
		if err := json.Unmarshal(parsedCmd.Payload, &payload); err != nil {
			return fmt.Errorf("parse payload: %w", err)
		}
		payload["port_id"] = *m.currentState.dockedPortID

		updatedPayload, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshal updated payload: %w", err)
		}
		parsedCmd.Payload = updatedPayload
	}

	// Create engine command
	engineCmd := engine.Command{
		SessionID:   m.sessionID,
		PlayerID:    m.playerID,
		CommandType: parsedCmd.Type,
		Payload:     parsedCmd.Payload,
		EnqueuedAt:  time.Now().Unix(),
	}

	// Send to command channel (non-blocking)
	select {
	case m.commandOut <- engineCmd:
		m.logger.Debug().
			Str("session_id", m.sessionID).
			Str("player_id", m.playerID).
			Str("command_type", parsedCmd.Type).
			Msg("Command sent to engine")
		return nil
	default:
		return fmt.Errorf("command queue full, try again")
	}
}

// waitForStateUpdate returns a command that waits for the next state update
func (m *GameView) waitForStateUpdate() tea.Cmd {
	return func() tea.Msg {
		update := <-m.updateChan
		return StateUpdateMsg{update: update}
	}
}

// updateGameState updates the local game state from a StateUpdate message
func (m *GameView) updateGameState(msg StateUpdateMsg) {
	// Parse the update as engine.StateUpdate
	update, ok := msg.update.(engine.StateUpdate)
	if !ok {
		// Try pointer type
		updatePtr, ok := msg.update.(*engine.StateUpdate)
		if !ok {
			m.logger.Warn().
				Str("session_id", m.sessionID).
				Type("type", msg.update).
				Msg("Received non-StateUpdate message")
			return
		}
		update = *updatePtr
	}

	// Update tick information
	m.currentState.tickNumber = update.TickNumber
	m.currentState.timestamp = update.Timestamp

	// Update player state if present
	if update.PlayerState != nil {
		m.currentState.credits = update.PlayerState.Credits
		m.currentState.ship = update.PlayerState.Ship
		m.currentState.cargo = update.PlayerState.Cargo

		// Update location information from ship
		if update.PlayerState.Ship != nil {
			m.currentState.currentSystemID = update.PlayerState.Ship.CurrentSystemID
			m.currentState.dockedPortID = update.PlayerState.Ship.DockedAtPortID
			
			// TODO: Get system name from world data
			// For now, just use the system ID as a string
			m.currentState.currentSystemName = fmt.Sprintf("System %d", update.PlayerState.Ship.CurrentSystemID)
			
			// Auto-switch to combat view if ship status is IN_COMBAT
			if update.PlayerState.Ship.Status == "IN_COMBAT" && m.viewMode != ViewModeCombat {
				m.viewMode = ViewModeCombat
				m.infoMessage = "Combat initiated!"
			}
			
			// Auto-switch back to command view if combat ended
			if update.PlayerState.Ship.Status != "IN_COMBAT" && m.viewMode == ViewModeCombat {
				m.viewMode = ViewModeCommand
			}
		}
	}

	// Process events
	for _, event := range update.Events {
		m.logger.Debug().
			Str("session_id", m.sessionID).
			Str("event_type", event.EventType).
			Str("message", event.Message).
			Msg("Game event received")
		
		// Display event as info message
		m.infoMessage = event.Message
	}

	m.logger.Debug().
		Str("session_id", m.sessionID).
		Int64("tick", update.TickNumber).
		Msg("State update processed")
}

// StateUpdateMsg wraps a state update from the tick engine
type StateUpdateMsg struct {
	update interface{}
}

// errMsg wraps an error message
type errMsg struct {
	err error
}

func (e errMsg) Error() string {
	return e.err.Error()
}

// Render methods (placeholders for now - will be implemented in subsequent tasks)

func (m *GameView) renderStatusBar() string {
	if m.currentState.ship == nil {
		return statusBarStyle.Render("BlackSector | Loading...")
	}

	// Format: System | Credits: 10000 | Hull: 100/100 | Shield: 50/50 | Energy: 100/100
	status := fmt.Sprintf(
		"System: %s | Credits: %d | Hull: %d/%d | Shield: %d/%d | Energy: %d/%d",
		m.currentState.currentSystemName,
		m.currentState.credits,
		m.currentState.ship.HullPoints,
		m.currentState.ship.MaxHullPoints,
		m.currentState.ship.ShieldPoints,
		m.currentState.ship.MaxShieldPoints,
		m.currentState.ship.EnergyPoints,
		m.currentState.ship.MaxEnergyPoints,
	)

	return statusBarStyle.Render(status)
}

func (m *GameView) renderCommandView() string {
	var b strings.Builder
	
	b.WriteString(titleStyle.Render("BlackSector"))
	b.WriteString("\n\n")
	
	// Show current location
	if m.currentState.ship != nil {
		b.WriteString(subtitleStyle.Render("Current Location"))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("System: %s (ID: %d)\n", m.currentState.currentSystemName, m.currentState.currentSystemID))
		
		if m.currentState.dockedPortID != nil {
			b.WriteString(fmt.Sprintf("Status: Docked at Port %d\n", *m.currentState.dockedPortID))
		} else {
			b.WriteString("Status: In space\n")
		}
		b.WriteString("\n")
	}
	
	// Show available commands
	b.WriteString(subtitleStyle.Render("Available Commands"))
	b.WriteString("\n")
	b.WriteString("  jump <system_id>  - Jump to another system\n")
	b.WriteString("  dock <port_id>    - Dock at a port\n")
	b.WriteString("  undock            - Undock from current port\n")
	b.WriteString("  buy <commodity> <qty> - Buy commodities\n")
	b.WriteString("  sell <commodity> <qty> - Sell commodities\n")
	b.WriteString("  market            - View market prices\n")
	b.WriteString("  cargo             - View cargo hold\n")
	b.WriteString("  missions          - View mission board\n")
	b.WriteString("  help              - Show help\n")
	
	return b.String()
}

func (m *GameView) renderMarketView() string {
	var b strings.Builder
	
	b.WriteString(titleStyle.Render("Market Prices"))
	b.WriteString("\n\n")
	
	// Check if docked
	if m.currentState.dockedPortID == nil {
		b.WriteString(warningStyle.Render("You must be docked at a port to view market prices"))
		b.WriteString("\n")
		return b.String()
	}
	
	b.WriteString(fmt.Sprintf("Port ID: %d\n\n", *m.currentState.dockedPortID))
	
	// TODO: Fetch actual market prices from the economy system
	// For now, show placeholder
	b.WriteString(mutedStyle.Render("Market prices will be displayed here"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("(Economy integration pending)"))
	b.WriteString("\n")
	
	return b.String()
}

func (m *GameView) renderCombatView() string {
	var b strings.Builder
	
	b.WriteString(titleStyle.Render("⚔ COMBAT ⚔"))
	b.WriteString("\n\n")
	
	// Check if in combat
	if m.currentState.activeCombat == nil {
		b.WriteString(warningStyle.Render("No active combat"))
		b.WriteString("\n")
		return b.String()
	}
	
	// TODO: Parse activeCombat as *combat.CombatInstance when integrated
	// For now, display placeholder
	b.WriteString(subtitleStyle.Render("Enemy Pirate"))
	b.WriteString("\n")
	b.WriteString(combatEnemyStyle.Render("Pirate Ship"))
	b.WriteString("\n")
	b.WriteString("  Hull: ??? / ???\n")
	b.WriteString("  Shield: ??? / ???\n")
	b.WriteString("\n")
	
	b.WriteString(subtitleStyle.Render("Your Ship"))
	b.WriteString("\n")
	if m.currentState.ship != nil {
		hullStyle := getHealthStyle(m.currentState.ship.HullPoints, m.currentState.ship.MaxHullPoints)
		shieldStyle := getHealthStyle(m.currentState.ship.ShieldPoints, m.currentState.ship.MaxShieldPoints)
		
		b.WriteString(combatPlayerStyle.Render(m.currentState.ship.ShipClass))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  Hull: %s\n", hullStyle.Render(fmt.Sprintf("%d / %d", m.currentState.ship.HullPoints, m.currentState.ship.MaxHullPoints))))
		b.WriteString(fmt.Sprintf("  Shield: %s\n", shieldStyle.Render(fmt.Sprintf("%d / %d", m.currentState.ship.ShieldPoints, m.currentState.ship.MaxShieldPoints))))
		b.WriteString(fmt.Sprintf("  Energy: %d / %d\n", m.currentState.ship.EnergyPoints, m.currentState.ship.MaxEnergyPoints))
	}
	b.WriteString("\n")
	
	b.WriteString(subtitleStyle.Render("Combat Log"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("Turn 1: Combat initiated"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("(Recent combat events will appear here)"))
	b.WriteString("\n\n")
	
	b.WriteString(subtitleStyle.Render("Available Actions"))
	b.WriteString("\n")
	b.WriteString("  attack     - Attack the enemy\n")
	b.WriteString("  flee       - Attempt to escape\n")
	b.WriteString("  surrender  - Surrender (lose 40% credits)\n")
	b.WriteString("\n")
	
	return b.String()
}

func (m *GameView) renderMissionView() string {
	var b strings.Builder
	
	b.WriteString(titleStyle.Render("Mission Board"))
	b.WriteString("\n\n")
	
	// Check if docked
	if m.currentState.dockedPortID == nil {
		b.WriteString(warningStyle.Render("You must be docked at a port to view missions"))
		b.WriteString("\n")
		return b.String()
	}
	
	// Check if player has active mission
	if m.currentState.activeMission != nil {
		b.WriteString(subtitleStyle.Render("Active Mission"))
		b.WriteString("\n")
		b.WriteString(highlightStyle.Render("Mission in progress"))
		b.WriteString("\n")
		b.WriteString(mutedStyle.Render("Complete your current mission before accepting another"))
		b.WriteString("\n\n")
		
		// TODO: Display active mission details when integrated
		b.WriteString("Mission: ???\n")
		b.WriteString("Objective: ???\n")
		b.WriteString("Progress: ???\n")
		b.WriteString("\n")
		
		b.WriteString(subtitleStyle.Render("Commands"))
		b.WriteString("\n")
		b.WriteString("  mission_abandon  - Abandon current mission\n")
		b.WriteString("\n")
		
		return b.String()
	}
	
	// Display available missions
	b.WriteString(subtitleStyle.Render("Available Missions"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Port ID: %d\n\n", *m.currentState.dockedPortID))
	
	// TODO: Fetch actual missions from the mission system
	// For now, show placeholder
	b.WriteString(mutedStyle.Render("No missions available at this port"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("(Mission integration pending)"))
	b.WriteString("\n\n")
	
	b.WriteString(subtitleStyle.Render("Commands"))
	b.WriteString("\n")
	b.WriteString("  mission_accept <mission_id>  - Accept a mission\n")
	b.WriteString("  mission_list                 - Refresh mission list\n")
	b.WriteString("\n")
	
	return b.String()
}

func (m *GameView) renderCargoView() string {
	var b strings.Builder
	
	b.WriteString(titleStyle.Render("Cargo Hold"))
	b.WriteString("\n\n")
	
	if m.currentState.ship == nil {
		b.WriteString("No ship data available\n")
		return b.String()
	}
	
	// Show cargo capacity
	usedCapacity := 0
	for _, slot := range m.currentState.cargo {
		usedCapacity += slot.Quantity
	}
	
	b.WriteString(fmt.Sprintf("Capacity: %d / %d\n\n", usedCapacity, m.currentState.ship.CargoCapacity))
	
	// Show cargo contents
	if len(m.currentState.cargo) == 0 {
		b.WriteString(mutedStyle.Render("Cargo hold is empty"))
		b.WriteString("\n")
	} else {
		b.WriteString(tableHeaderStyle.Render("Commodity"))
		b.WriteString("  ")
		b.WriteString(tableHeaderStyle.Render("Quantity"))
		b.WriteString("\n")
		
		for i, slot := range m.currentState.cargo {
			style := tableRowStyle
			if i%2 == 1 {
				style = tableRowAltStyle
			}
			
			b.WriteString(style.Render(fmt.Sprintf("%-20s  %d", slot.CommodityID, slot.Quantity)))
			b.WriteString("\n")
		}
	}
	
	return b.String()
}

func (m *GameView) renderHelpView() string {
	var b strings.Builder
	
	b.WriteString(titleStyle.Render("BlackSector - Command Reference"))
	b.WriteString("\n\n")
	
	b.WriteString(subtitleStyle.Render("Navigation"))
	b.WriteString("\n")
	b.WriteString("  jump <system_id>  - Jump to another system via jump connection\n")
	b.WriteString("  dock <port_id>    - Dock at a port in your current system\n")
	b.WriteString("  undock            - Undock from the current port\n")
	b.WriteString("\n")
	
	b.WriteString(subtitleStyle.Render("Trading"))
	b.WriteString("\n")
	b.WriteString("  buy <commodity> <quantity>   - Buy commodities at current port\n")
	b.WriteString("  sell <commodity> <quantity>  - Sell commodities at current port\n")
	b.WriteString("  market                       - View market prices at current port\n")
	b.WriteString("\n")
	
	b.WriteString(subtitleStyle.Render("Combat"))
	b.WriteString("\n")
	b.WriteString("  attack     - Attack the enemy in combat\n")
	b.WriteString("  flee       - Attempt to escape from combat\n")
	b.WriteString("  surrender  - Surrender (lose 40% of credits)\n")
	b.WriteString("\n")
	
	b.WriteString(subtitleStyle.Render("Missions"))
	b.WriteString("\n")
	b.WriteString("  missions                     - View mission board\n")
	b.WriteString("  mission_accept <mission_id>  - Accept a mission\n")
	b.WriteString("  mission_abandon              - Abandon current mission\n")
	b.WriteString("\n")
	
	b.WriteString(subtitleStyle.Render("Information"))
	b.WriteString("\n")
	b.WriteString("  cargo   - View your cargo hold\n")
	b.WriteString("  system  - View current system information\n")
	b.WriteString("  help    - Show this help screen\n")
	b.WriteString("\n")
	
	b.WriteString(subtitleStyle.Render("Commodities"))
	b.WriteString("\n")
	b.WriteString("  food_supplies, fuel_cells, raw_ore, refined_ore,\n")
	b.WriteString("  machinery, electronics, luxury_goods\n")
	b.WriteString("\n")
	
	b.WriteString(mutedStyle.Render("Press Ctrl+C or Esc to exit"))
	b.WriteString("\n")
	
	return b.String()
}

func (m *GameView) renderCommandPrompt() string {
	prompt := fmt.Sprintf("> %s", m.commandBuffer)
	return promptStyle.Render(prompt)
}
