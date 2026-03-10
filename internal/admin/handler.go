package admin

import (
	"fmt"

	"github.com/BrianB-22/BlackSector/internal/engine"
	"github.com/BrianB-22/BlackSector/internal/session"
	"github.com/rs/zerolog"
)

// Handler processes admin commands
type Handler struct {
	cli        *CLI
	tickEngine *engine.TickEngine
	sessionMgr *session.SessionManager
	logger     zerolog.Logger
}

// NewHandler creates a new admin command handler
func NewHandler(cli *CLI, tickEngine *engine.TickEngine, sessionMgr *session.SessionManager, logger zerolog.Logger) *Handler {
	return &Handler{
		cli:        cli,
		tickEngine: tickEngine,
		sessionMgr: sessionMgr,
		logger:     logger,
	}
}

// Start begins processing admin commands
// This should be called in a goroutine to avoid blocking
func (h *Handler) Start() {
	go h.processCommands()
	h.logger.Info().Msg("Admin command handler started")
}

// processCommands processes commands from the CLI channel
func (h *Handler) processCommands() {
	for {
		select {
		case cmd := <-h.cli.CommandChannel():
			h.handleCommand(cmd)
		case <-h.cli.ShutdownChannel():
			h.logger.Info().Msg("Admin CLI shutdown signal received")
			return
		}
	}
}

// handleCommand processes a single admin command
func (h *Handler) handleCommand(cmd Command) {
	h.logger.Debug().
		Str("command", cmd.Type).
		Msg("Processing admin command")
	
	switch cmd.Type {
	case CommandStatus:
		h.handleStatus()
	case CommandSessions:
		h.handleSessions()
	case CommandShutdown:
		h.handleShutdown()
	default:
		h.logger.Warn().
			Str("command", cmd.Type).
			Msg("Unknown command type in handler")
	}
}

// handleStatus displays current tick number and active session count
// Requirement 13.2
func (h *Handler) handleStatus() {
	tickNumber := h.tickEngine.TickNumber()
	activeCount := h.sessionMgr.ActiveSessionCount()
	
	output := fmt.Sprintf("Status:\n  Tick: %d\n  Active Sessions: %d\n", tickNumber, activeCount)
	h.cli.WriteOutput(output)
	
	h.logger.Info().
		Int64("tick", tickNumber).
		Int("active_sessions", activeCount).
		Msg("Status command executed")
}

// handleSessions displays a list of all active sessions with player names
// Requirement 13.3
func (h *Handler) handleSessions() {
	sessions := h.sessionMgr.ActiveSessions()
	
	if len(sessions) == 0 {
		h.cli.WriteOutput("No active sessions\n")
		h.logger.Info().Msg("Sessions command executed - no active sessions")
		return
	}
	
	output := fmt.Sprintf("Active Sessions (%d):\n", len(sessions))
	for _, sess := range sessions {
		// Get player name from database
		player, err := h.sessionMgr.GetPlayerByID(sess.PlayerID)
		if err != nil {
			h.logger.Error().
				Err(err).
				Str("player_id", sess.PlayerID).
				Msg("Failed to get player name for session")
			output += fmt.Sprintf("  Session: %s, Player ID: %s (name unavailable)\n", sess.SessionID, sess.PlayerID)
			continue
		}
		
		playerName := "unknown"
		if player != nil {
			playerName = player.PlayerName
		}
		
		output += fmt.Sprintf("  Session: %s, Player: %s (%s)\n", sess.SessionID, playerName, sess.PlayerID)
	}
	
	h.cli.WriteOutput(output)
	
	h.logger.Info().
		Int("session_count", len(sessions)).
		Msg("Sessions command executed")
}

// handleShutdown initiates graceful shutdown
// Requirement 13.4
func (h *Handler) handleShutdown() {
	h.cli.WriteOutput("Initiating graceful shutdown...\n")
	
	h.logger.Info().Msg("Shutdown command received from admin CLI")
	
	// Stop the tick engine
	h.tickEngine.Stop()
	
	h.logger.Info().Msg("Shutdown initiated")
}
