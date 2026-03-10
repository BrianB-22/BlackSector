package session

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/BrianB-22/BlackSector/internal/db"
	"github.com/BrianB-22/BlackSector/internal/protocol"
)

// SessionManager manages active player sessions
type SessionManager struct {
	db             *db.Database
	logger         zerolog.Logger
	activeSessions map[string]*db.Session // map[sessionID]*Session
	connections    map[string]io.Writer   // map[sessionID]connection for sending messages
	updateChannels map[string]chan interface{} // map[sessionID]channel for state updates
	mu             sync.RWMutex
}

// NewSessionManager creates a new SessionManager
func NewSessionManager(database *db.Database, logger zerolog.Logger) *SessionManager {
	return &SessionManager{
		db:             database,
		logger:         logger,
		activeSessions: make(map[string]*db.Session),
		connections:    make(map[string]io.Writer),
		updateChannels: make(map[string]chan interface{}),
	}
}

// CreateSession creates a new session for a player
// Generates a unique session ID, sets state to CONNECTED, and persists to database
func (sm *SessionManager) CreateSession(playerID string, interfaceMode string) (*db.Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Generate unique session ID
	sessionID := uuid.New().String()
	now := time.Now().Unix()

	session := &db.Session{
		SessionID:      sessionID,
		PlayerID:       playerID,
		InterfaceMode:  interfaceMode,
		State:          db.SessionConnected,
		ConnectedAt:    now,
		DisconnectedAt: nil,
		LingerExpiryAt: nil,
		LastActivityAt: now,
	}

	// Persist to database
	if err := sm.db.InsertSession(session); err != nil {
		return nil, fmt.Errorf("failed to insert session: %w", err)
	}

	// Add to active sessions map
	sm.activeSessions[sessionID] = session

	sm.logger.Info().
		Str("session_id", sessionID).
		Str("player_id", playerID).
		Str("interface_mode", interfaceMode).
		Msg("Session created")

	return session, nil
}

// GetActiveSession returns the active session for a player, if one exists
// Only returns sessions with state CONNECTED
func (sm *SessionManager) GetActiveSession(playerID string) (*db.Session, error) {
	// First check in-memory map for quick lookup
	sm.mu.RLock()
	for _, session := range sm.activeSessions {
		if session.PlayerID == playerID && session.State == db.SessionConnected {
			sm.mu.RUnlock()
			return session, nil
		}
	}
	sm.mu.RUnlock()

	// Fall back to database query
	return sm.db.GetActiveSessionByPlayerID(playerID)
}

// UpdateSessionState updates the state of a session
// Sets appropriate timestamps based on the new state
func (sm *SessionManager) UpdateSessionState(sessionID string, state db.SessionState) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now().Unix()

	// Update database first
	if err := sm.db.UpdateSessionState(sessionID, state); err != nil {
		return fmt.Errorf("failed to update session state: %w", err)
	}

	// Update in-memory session if it exists
	if session, exists := sm.activeSessions[sessionID]; exists {
		session.State = state

		switch state {
		case db.SessionDisconnectedLingering:
			session.DisconnectedAt = &now
			// Calculate linger expiry (e.g., 5 minutes from now)
			lingerExpiry := now + 300
			session.LingerExpiryAt = &lingerExpiry
		case db.SessionTerminated:
			// Remove from active sessions map, connections map, and update channels
			delete(sm.activeSessions, sessionID)
			delete(sm.connections, sessionID)
			
			// Close and remove update channel
			if ch, exists := sm.updateChannels[sessionID]; exists {
				close(ch)
				delete(sm.updateChannels, sessionID)
			}
		}

		session.LastActivityAt = now
	}

	sm.logger.Info().
		Str("session_id", sessionID).
		Str("state", string(state)).
		Msg("Session state updated")

	return nil
}

// TerminateSession sets a session state to TERMINATED
func (sm *SessionManager) TerminateSession(sessionID string) error {
	return sm.UpdateSessionState(sessionID, db.SessionTerminated)
}

// GetSession returns a session by ID from the in-memory map
func (sm *SessionManager) GetSession(sessionID string) (*db.Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.activeSessions[sessionID]
	return session, exists
}

// ActiveSessions returns a slice of all active sessions (state = CONNECTED)
func (sm *SessionManager) ActiveSessions() []*db.Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]*db.Session, 0, len(sm.activeSessions))
	for _, session := range sm.activeSessions {
		if session.State == db.SessionConnected {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// ActiveSessionCount returns the count of active sessions
func (sm *SessionManager) ActiveSessionCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	count := 0
	for _, session := range sm.activeSessions {
		if session.State == db.SessionConnected {
			count++
		}
	}

	return count
}

// HandshakeConfig contains configuration needed for the handshake protocol
type HandshakeConfig struct {
	ServerName     string
	MOTD           string
	TickIntervalMs int
}

// GameSession represents an active game session after successful handshake
type GameSession struct {
	SessionID string
	PlayerID  string
}

// HandleHandshake implements the server-side handshake protocol
// It sends HandshakeInit, waits for HandshakeResponse, validates credentials,
// and either creates a session or rejects the connection
func (sm *SessionManager) HandleHandshake(conn io.ReadWriter, config HandshakeConfig) (*GameSession, error) {
	// Step 1: Send HandshakeInit message
	initMsg := protocol.HandshakeInit{
		Type:            "handshake_init",
		Timestamp:       time.Now().Unix(),
		ProtocolVersion: "1.0",
		InterfaceMode:   "TEXT",
		ServerName:      config.ServerName,
		MOTD:            config.MOTD,
		Payload:         make(map[string]interface{}),
	}

	initJSON, err := json.Marshal(initMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal handshake_init: %w", err)
	}

	// Write the message with newline delimiter
	if _, err := conn.Write(append(initJSON, '\n')); err != nil {
		return nil, fmt.Errorf("failed to send handshake_init: %w", err)
	}

	sm.logger.Debug().
		Str("protocol_version", initMsg.ProtocolVersion).
		Str("interface_mode", initMsg.InterfaceMode).
		Msg("Sent handshake_init")

	// Step 2: Wait for HandshakeResponse with 30-second timeout
	responseChan := make(chan protocol.HandshakeResponse, 1)
	errorChan := make(chan error, 1)

	go func() {
		// Read response from connection
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if err != nil {
			errorChan <- fmt.Errorf("failed to read handshake_response: %w", err)
			return
		}

		var response protocol.HandshakeResponse
		if err := json.Unmarshal(buf[:n], &response); err != nil {
			errorChan <- fmt.Errorf("failed to unmarshal handshake_response: %w", err)
			return
		}

		responseChan <- response
	}()

	var response protocol.HandshakeResponse
	select {
	case response = <-responseChan:
		// Response received successfully
	case err := <-errorChan:
		return nil, sm.sendReject(conn, "", protocol.RejectReasonHandshakeTimeout, err)
	case <-time.After(30 * time.Second):
		return nil, sm.sendReject(conn, "", protocol.RejectReasonHandshakeTimeout, 
			fmt.Errorf("handshake timeout: no response within 30 seconds"))
	}

	sm.logger.Debug().
		Str("correlation_id", response.CorrelationID).
		Str("protocol_version", response.ProtocolVersion).
		Msg("Received handshake_response")

	// Step 3: Validate protocol version
	if response.ProtocolVersion != "1.0" {
		return nil, sm.sendReject(conn, response.CorrelationID, protocol.RejectReasonVersionMismatch,
			fmt.Errorf("protocol version mismatch: expected 1.0, got %s", response.ProtocolVersion))
	}

	// Step 4: Authenticate player token
	player, err := sm.db.GetPlayerByToken(response.Payload.PlayerToken)
	if err != nil {
		return nil, sm.sendReject(conn, response.CorrelationID, protocol.RejectReasonInvalidToken,
			fmt.Errorf("failed to authenticate token: %w", err))
	}
	if player == nil {
		return nil, sm.sendReject(conn, response.CorrelationID, protocol.RejectReasonInvalidToken,
			fmt.Errorf("invalid player token"))
	}

	sm.logger.Debug().
		Str("player_id", player.PlayerID).
		Str("player_name", player.PlayerName).
		Msg("Player authenticated")

	// Step 5: Check for existing active session
	existingSession, err := sm.GetActiveSession(player.PlayerID)
	if err != nil {
		return nil, sm.sendReject(conn, response.CorrelationID, protocol.RejectReasonInvalidToken,
			fmt.Errorf("failed to check for existing session: %w", err))
	}
	if existingSession != nil {
		return nil, sm.sendReject(conn, response.CorrelationID, protocol.RejectReasonSessionAlreadyActive,
			fmt.Errorf("player %s already has an active session", player.PlayerID))
	}

	// Step 6: Create session
	session, err := sm.CreateSession(player.PlayerID, "TEXT")
	if err != nil {
		return nil, sm.sendReject(conn, response.CorrelationID, protocol.RejectReasonInvalidToken,
			fmt.Errorf("failed to create session: %w", err))
	}

	// Store connection for later use (e.g., sending shutdown messages)
	sm.mu.Lock()
	sm.connections[session.SessionID] = conn
	sm.mu.Unlock()

	// Step 7: Send HandshakeAck
	ackMsg := protocol.HandshakeAck{
		Type:          "handshake_ack",
		Timestamp:     time.Now().Unix(),
		CorrelationID: response.CorrelationID,
		Payload: protocol.HandshakeAckPayload{
			SessionID:      session.SessionID,
			PlayerID:       player.PlayerID,
			TickIntervalMs: config.TickIntervalMs,
			InterfaceMode:  "TEXT",
		},
	}

	ackJSON, err := json.Marshal(ackMsg)
	if err != nil {
		// Session was created but we failed to send ack - clean up
		_ = sm.TerminateSession(session.SessionID)
		return nil, fmt.Errorf("failed to marshal handshake_ack: %w", err)
	}

	// Write the ack message with newline delimiter
	if _, err := conn.Write(append(ackJSON, '\n')); err != nil {
		// Session was created but we failed to send ack - clean up
		_ = sm.TerminateSession(session.SessionID)
		return nil, fmt.Errorf("failed to send handshake_ack: %w", err)
	}

	sm.logger.Info().
		Str("player_id", player.PlayerID).
		Str("player_name", player.PlayerName).
		Str("session_id", session.SessionID).
		Msg("Player connected")

	return &GameSession{
		SessionID: session.SessionID,
		PlayerID:  player.PlayerID,
	}, nil
}

// sendReject sends a HandshakeReject message and returns the error
func (sm *SessionManager) sendReject(conn io.Writer, correlationID string, reason string, err error) error {
	rejectMsg := protocol.HandshakeReject{
		Type:          "handshake_reject",
		Timestamp:     time.Now().Unix(),
		CorrelationID: correlationID,
		Payload: protocol.HandshakeRejectPayload{
			Reason: reason,
		},
	}

	rejectJSON, marshalErr := json.Marshal(rejectMsg)
	if marshalErr != nil {
		sm.logger.Error().
			Err(marshalErr).
			Msg("Failed to marshal handshake_reject")
		return fmt.Errorf("%w (also failed to marshal reject: %v)", err, marshalErr)
	}

	// Write the reject message with newline delimiter
	if _, writeErr := conn.Write(append(rejectJSON, '\n')); writeErr != nil {
		sm.logger.Error().
			Err(writeErr).
			Msg("Failed to send handshake_reject")
		return fmt.Errorf("%w (also failed to send reject: %v)", err, writeErr)
	}

	sm.logger.Info().
		Str("reason", reason).
		Str("correlation_id", correlationID).
		Err(err).
		Msg("Handshake rejected")

	return err
}

// SendShutdownMessages sends server_shutdown messages to all active sessions
// This is called during graceful shutdown to notify connected clients
func (sm *SessionManager) SendShutdownMessages() error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	shutdownMsg := protocol.ServerShutdown{
		Type:      "server_shutdown",
		Timestamp: time.Now().Unix(),
		Payload: protocol.ServerShutdownPayload{
			Message: "Server is shutting down",
		},
	}

	shutdownJSON, err := json.Marshal(shutdownMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal server_shutdown: %w", err)
	}

	// Send to all active connections
	sentCount := 0
	for sessionID, conn := range sm.connections {
		if _, err := conn.Write(append(shutdownJSON, '\n')); err != nil {
			sm.logger.Warn().
				Err(err).
				Str("session_id", sessionID).
				Msg("Failed to send shutdown message to session")
		} else {
			sentCount++
		}
	}

	sm.logger.Info().
		Int("sent_count", sentCount).
		Int("total_connections", len(sm.connections)).
		Msg("Sent shutdown messages to active sessions")

	return nil
}

// TerminateAllSessions updates all session states to TERMINATED
// This is called during graceful shutdown after connections are closed
func (sm *SessionManager) TerminateAllSessions() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Get all session IDs to terminate
	sessionIDs := make([]string, 0, len(sm.activeSessions))
	for sessionID := range sm.activeSessions {
		sessionIDs = append(sessionIDs, sessionID)
	}

	// Terminate each session
	terminatedCount := 0
	for _, sessionID := range sessionIDs {
		if err := sm.db.UpdateSessionState(sessionID, db.SessionTerminated); err != nil {
			sm.logger.Error().
				Err(err).
				Str("session_id", sessionID).
				Msg("Failed to terminate session")
		} else {
			terminatedCount++
			// Remove from maps
			delete(sm.activeSessions, sessionID)
			delete(sm.connections, sessionID)
		}
	}

	sm.logger.Info().
		Int("terminated_count", terminatedCount).
		Msg("Terminated all sessions")

	return nil
}
// GetPlayerByID retrieves a player by their player ID
func (sm *SessionManager) GetPlayerByID(playerID string) (*db.Player, error) {
	return sm.db.GetPlayerByID(playerID)
}


// RegisterSessionForUpdates creates a buffered channel for state updates
// Returns the channel that the session should listen on for updates
func (sm *SessionManager) RegisterSessionForUpdates(sessionID string) (chan interface{}, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if session exists
	if _, exists := sm.activeSessions[sessionID]; !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	// Create buffered channel (buffer size 10 to handle bursts)
	updateChan := make(chan interface{}, 10)
	sm.updateChannels[sessionID] = updateChan

	sm.logger.Debug().
		Str("session_id", sessionID).
		Msg("Session registered for state updates")

	return updateChan, nil
}

// UnregisterSessionForUpdates removes the update channel for a session
// Should be called when a session disconnects
func (sm *SessionManager) UnregisterSessionForUpdates(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if ch, exists := sm.updateChannels[sessionID]; exists {
		close(ch)
		delete(sm.updateChannels, sessionID)

		sm.logger.Debug().
			Str("session_id", sessionID).
			Msg("Session unregistered from state updates")
	}
}

// GetUpdateChannel returns the update channel for a session
func (sm *SessionManager) GetUpdateChannel(sessionID string) (chan interface{}, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	ch, exists := sm.updateChannels[sessionID]
	return ch, exists
}

// BroadcastToSession sends a state update to a specific session
// Uses non-blocking send with timeout to avoid blocking the tick loop
func (sm *SessionManager) BroadcastToSession(sessionID string, update interface{}) error {
	sm.mu.RLock()
	ch, exists := sm.updateChannels[sessionID]
	sm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no update channel for session %s", sessionID)
	}

	// Non-blocking send with timeout
	select {
	case ch <- update:
		return nil
	case <-time.After(50 * time.Millisecond):
		sm.logger.Warn().
			Str("session_id", sessionID).
			Msg("State update channel full or slow, dropping update")
		return fmt.Errorf("update channel timeout for session %s", sessionID)
	}
}

// BroadcastToAllSessions sends a state update to all active sessions
// Returns the number of successful broadcasts and any errors encountered
func (sm *SessionManager) BroadcastToAllSessions(updates map[string]interface{}) (int, []error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	successCount := 0
	var errors []error

	for sessionID, update := range updates {
		ch, exists := sm.updateChannels[sessionID]
		if !exists {
			errors = append(errors, fmt.Errorf("no update channel for session %s", sessionID))
			continue
		}

		// Non-blocking send with timeout
		select {
		case ch <- update:
			successCount++
		case <-time.After(50 * time.Millisecond):
			sm.logger.Warn().
				Str("session_id", sessionID).
				Msg("State update channel full or slow, dropping update")
			errors = append(errors, fmt.Errorf("update channel timeout for session %s", sessionID))
		}
	}

	return successCount, errors
}
