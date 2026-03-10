package session

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/BrianB-22/BlackSector/internal/protocol"
)

// HandleSSHHandshake implements a simplified handshake for SSH TEXT mode
// The SSH username is used to identify the player (no token required)
func (sm *SessionManager) HandleSSHHandshake(conn io.ReadWriter, sshUsername string, config HandshakeConfig) (*GameSession, error) {
	// Look up player by SSH username
	player, err := sm.db.GetPlayerBySSHUsername(sshUsername)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup player by SSH username: %w", err)
	}
	if player == nil {
		return nil, fmt.Errorf("player not found for SSH username: %s", sshUsername)
	}

	sm.logger.Debug().
		Str("player_id", player.PlayerID).
		Str("player_name", player.PlayerName).
		Str("ssh_username", sshUsername).
		Msg("Player authenticated via SSH username")

	// Check for existing active session
	existingSession, err := sm.GetActiveSession(player.PlayerID)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing session: %w", err)
	}
	if existingSession != nil {
		return nil, fmt.Errorf("player %s already has an active session", player.PlayerID)
	}

	// Create session
	session, err := sm.CreateSession(player.PlayerID, "TEXT")
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Store connection for later use
	sm.mu.Lock()
	sm.connections[session.SessionID] = conn
	sm.mu.Unlock()

	// Send HandshakeInit
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
		_ = sm.TerminateSession(session.SessionID)
		return nil, fmt.Errorf("failed to marshal handshake_init: %w", err)
	}

	if _, err := conn.Write(append(initJSON, '\n')); err != nil {
		_ = sm.TerminateSession(session.SessionID)
		return nil, fmt.Errorf("failed to send handshake_init: %w", err)
	}

	sm.logger.Debug().
		Str("session_id", session.SessionID).
		Str("player_id", player.PlayerID).
		Msg("Sent handshake_init")

	// Send HandshakeAck (no response needed for SSH mode)
	ackMsg := protocol.HandshakeAck{
		Type:          "handshake_ack",
		Timestamp:     time.Now().Unix(),
		CorrelationID: "", // No correlation ID for SSH mode
		Payload: protocol.HandshakeAckPayload{
			SessionID:      session.SessionID,
			PlayerID:       player.PlayerID,
			TickIntervalMs: config.TickIntervalMs,
			InterfaceMode:  "TEXT",
		},
	}

	ackJSON, err := json.Marshal(ackMsg)
	if err != nil {
		_ = sm.TerminateSession(session.SessionID)
		return nil, fmt.Errorf("failed to marshal handshake_ack: %w", err)
	}

	if _, err := conn.Write(append(ackJSON, '\n')); err != nil {
		_ = sm.TerminateSession(session.SessionID)
		return nil, fmt.Errorf("failed to send handshake_ack: %w", err)
	}

	sm.logger.Info().
		Str("player_id", player.PlayerID).
		Str("player_name", player.PlayerName).
		Str("session_id", session.SessionID).
		Msg("SSH handshake completed")

	return &GameSession{
		SessionID: session.SessionID,
		PlayerID:  player.PlayerID,
	}, nil
}
