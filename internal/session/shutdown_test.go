package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/BrianB-22/BlackSector/internal/db"
	"github.com/BrianB-22/BlackSector/internal/protocol"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSendShutdownMessages tests that shutdown messages are sent to all active connections
func TestSendShutdownMessages(t *testing.T) {
	// Setup
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	database, err := db.InitDatabase(":memory:", logger)
	require.NoError(t, err)
	defer database.Close()

	// Create test players
	for i := 1; i <= 3; i++ {
		playerID := fmt.Sprintf("player-%d", i)
		_, err = database.Conn().Exec(`
			INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
			VALUES (?, ?, ?, ?, ?, ?)
		`, playerID, fmt.Sprintf("Player%d", i), "token-hash", 1000, 1234567890, false)
		require.NoError(t, err)
	}

	sessionMgr := NewSessionManager(database, logger)

	// Create mock connections
	conn1 := &bytes.Buffer{}
	conn2 := &bytes.Buffer{}
	conn3 := &bytes.Buffer{}

	// Create sessions
	session1, err := sessionMgr.CreateSession("player-1", "TEXT")
	require.NoError(t, err)
	session2, err := sessionMgr.CreateSession("player-2", "TEXT")
	require.NoError(t, err)
	session3, err := sessionMgr.CreateSession("player-3", "TEXT")
	require.NoError(t, err)

	// Store connections
	sessionMgr.mu.Lock()
	sessionMgr.connections[session1.SessionID] = conn1
	sessionMgr.connections[session2.SessionID] = conn2
	sessionMgr.connections[session3.SessionID] = conn3
	sessionMgr.mu.Unlock()

	// Send shutdown messages
	err = sessionMgr.SendShutdownMessages()
	require.NoError(t, err)

	// Verify all connections received the shutdown message
	for i, conn := range []*bytes.Buffer{conn1, conn2, conn3} {
		data := conn.Bytes()
		assert.Greater(t, len(data), 0, "Connection %d should have received data", i+1)

		// Parse the JSON message
		var msg protocol.ServerShutdown
		err := json.Unmarshal(bytes.TrimSpace(data), &msg)
		require.NoError(t, err, "Connection %d should have valid JSON", i+1)

		// Verify message structure
		assert.Equal(t, "server_shutdown", msg.Type)
		assert.Greater(t, msg.Timestamp, int64(0))
		assert.Equal(t, "Server is shutting down", msg.Payload.Message)
	}
}

// TestTerminateAllSessions tests that all sessions are terminated during shutdown
func TestTerminateAllSessions(t *testing.T) {
	// Setup
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	database, err := db.InitDatabase(":memory:", logger)
	require.NoError(t, err)
	defer database.Close()

	// Create test players
	for i := 1; i <= 3; i++ {
		playerID := fmt.Sprintf("player-%d", i)
		_, err = database.Conn().Exec(`
			INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
			VALUES (?, ?, ?, ?, ?, ?)
		`, playerID, fmt.Sprintf("Player%d", i), "token-hash", 1000, 1234567890, false)
		require.NoError(t, err)
	}

	sessionMgr := NewSessionManager(database, logger)

	// Create multiple sessions
	session1, err := sessionMgr.CreateSession("player-1", "TEXT")
	require.NoError(t, err)
	session2, err := sessionMgr.CreateSession("player-2", "TEXT")
	require.NoError(t, err)
	session3, err := sessionMgr.CreateSession("player-3", "TEXT")
	require.NoError(t, err)

	// Verify sessions are active
	assert.Equal(t, 3, sessionMgr.ActiveSessionCount())

	// Terminate all sessions
	err = sessionMgr.TerminateAllSessions()
	require.NoError(t, err)

	// Verify all sessions are terminated
	assert.Equal(t, 0, sessionMgr.ActiveSessionCount())

	// Verify sessions are removed from in-memory maps
	_, exists := sessionMgr.GetSession(session1.SessionID)
	assert.False(t, exists, "Session 1 should be removed from active sessions")
	_, exists = sessionMgr.GetSession(session2.SessionID)
	assert.False(t, exists, "Session 2 should be removed from active sessions")
	_, exists = sessionMgr.GetSession(session3.SessionID)
	assert.False(t, exists, "Session 3 should be removed from active sessions")

	// Verify no active sessions remain for these players
	activeSession1, err := database.GetActiveSessionByPlayerID("player-1")
	require.NoError(t, err)
	assert.Nil(t, activeSession1, "Player 1 should have no active session")

	activeSession2, err := database.GetActiveSessionByPlayerID("player-2")
	require.NoError(t, err)
	assert.Nil(t, activeSession2, "Player 2 should have no active session")

	activeSession3, err := database.GetActiveSessionByPlayerID("player-3")
	require.NoError(t, err)
	assert.Nil(t, activeSession3, "Player 3 should have no active session")
}

// TestShutdownSequence tests the complete shutdown sequence
func TestShutdownSequence(t *testing.T) {
	// Setup
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	database, err := db.InitDatabase(":memory:", logger)
	require.NoError(t, err)
	defer database.Close()

	// Create test players
	for i := 1; i <= 2; i++ {
		playerID := fmt.Sprintf("player-%d", i)
		_, err = database.Conn().Exec(`
			INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
			VALUES (?, ?, ?, ?, ?, ?)
		`, playerID, fmt.Sprintf("Player%d", i), "token-hash", 1000, 1234567890, false)
		require.NoError(t, err)
	}

	sessionMgr := NewSessionManager(database, logger)

	// Create sessions with connections
	conn1 := &bytes.Buffer{}
	conn2 := &bytes.Buffer{}

	session1, err := sessionMgr.CreateSession("player-1", "TEXT")
	require.NoError(t, err)
	session2, err := sessionMgr.CreateSession("player-2", "TEXT")
	require.NoError(t, err)

	sessionMgr.mu.Lock()
	sessionMgr.connections[session1.SessionID] = conn1
	sessionMgr.connections[session2.SessionID] = conn2
	sessionMgr.mu.Unlock()

	// Step 1: Send shutdown messages
	err = sessionMgr.SendShutdownMessages()
	require.NoError(t, err)

	// Verify messages were sent
	assert.Greater(t, conn1.Len(), 0)
	assert.Greater(t, conn2.Len(), 0)

	// Step 2: Terminate all sessions
	err = sessionMgr.TerminateAllSessions()
	require.NoError(t, err)

	// Verify all sessions are terminated
	assert.Equal(t, 0, sessionMgr.ActiveSessionCount())

	// Verify connections are cleaned up
	sessionMgr.mu.RLock()
	assert.Equal(t, 0, len(sessionMgr.connections))
	sessionMgr.mu.RUnlock()
}

// TestConnectionStorageOnHandshake tests that connections are stored during handshake
func TestConnectionStorageOnHandshake(t *testing.T) {
	// Setup
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
	database, err := db.InitDatabase(":memory:", logger)
	require.NoError(t, err)
	defer database.Close()

	// Insert test player directly into database
	_, err = database.Conn().Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES (?, ?, ?, ?, ?, ?)
	`, "player-123", "TestPlayer", "test-token-hash", 1000, 1234567890, false)
	require.NoError(t, err)

	sessionMgr := NewSessionManager(database, logger)

	// Create mock connection using the existing mockConn from manager_test.go
	conn := &testConn{
		readData:  bytes.NewBuffer(nil),
		writeData: bytes.NewBuffer(nil),
	}

	// Prepare handshake response
	response := protocol.HandshakeResponse{
		Type:            "handshake_response",
		Timestamp:       1234567890,
		ProtocolVersion: "1.0",
		CorrelationID:   "test-correlation",
		Payload: protocol.HandshakeResponsePayload{
			PlayerToken: "test-token-hash",
		},
	}
	responseJSON, err := json.Marshal(response)
	require.NoError(t, err)
	conn.readData.Write(responseJSON)

	// Execute handshake
	config := HandshakeConfig{
		ServerName:     "Test Server",
		MOTD:           "Test MOTD",
		TickIntervalMs: 2000,
	}

	session, err := sessionMgr.HandleHandshake(conn, config)
	require.NoError(t, err)
	require.NotNil(t, session)

	// Verify connection is stored
	sessionMgr.mu.RLock()
	storedConn, exists := sessionMgr.connections[session.SessionID]
	sessionMgr.mu.RUnlock()

	assert.True(t, exists, "Connection should be stored")
	assert.Equal(t, conn, storedConn, "Stored connection should match")
}

// testConn implements io.ReadWriter for testing
type testConn struct {
	readData  *bytes.Buffer
	writeData *bytes.Buffer
}

func (m *testConn) Read(p []byte) (n int, err error) {
	return m.readData.Read(p)
}

func (m *testConn) Write(p []byte) (n int, err error) {
	return m.writeData.Write(p)
}
