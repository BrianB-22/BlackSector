package session

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/BrianB-22/BlackSector/internal/db"
)

func setupTestDB(t *testing.T) *db.Database {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	database, err := db.InitDatabase(":memory:", logger)
	require.NoError(t, err)
	return database
}

func insertTestPlayer(t *testing.T, database *db.Database, playerID, playerName, tokenHash string) {
	query := `INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
	          VALUES (?, ?, ?, ?, ?, ?)`
	_, err := database.Conn().Exec(query, playerID, playerName, tokenHash, 10000, time.Now().Unix(), false)
	require.NoError(t, err)
}

func TestNewSessionManager(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	sm := NewSessionManager(database, logger)

	assert.NotNil(t, sm)
	assert.NotNil(t, sm.db)
	assert.NotNil(t, sm.activeSessions)
	assert.Equal(t, 0, len(sm.activeSessions))
}

func TestCreateSession(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	sm := NewSessionManager(database, logger)

	// Insert test player
	playerID := "player-123"
	insertTestPlayer(t, database, playerID, "TestPlayer", "hash123")

	// Create session
	session, err := sm.CreateSession(playerID, "TEXT")
	require.NoError(t, err)
	require.NotNil(t, session)

	// Verify session properties
	assert.NotEmpty(t, session.SessionID)
	assert.Equal(t, playerID, session.PlayerID)
	assert.Equal(t, "TEXT", session.InterfaceMode)
	assert.Equal(t, db.SessionConnected, session.State)
	assert.Greater(t, session.ConnectedAt, int64(0))
	assert.Nil(t, session.DisconnectedAt)
	assert.Nil(t, session.LingerExpiryAt)
	assert.Greater(t, session.LastActivityAt, int64(0))

	// Verify session is in active sessions map
	assert.Equal(t, 1, len(sm.activeSessions))
	assert.Equal(t, session, sm.activeSessions[session.SessionID])

	// Verify session was persisted to database
	dbSession, err := database.GetActiveSessionByPlayerID(playerID)
	require.NoError(t, err)
	require.NotNil(t, dbSession)
	assert.Equal(t, session.SessionID, dbSession.SessionID)
}

func TestCreateSession_UniqueSessionIDs(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	sm := NewSessionManager(database, logger)

	// Insert test players
	insertTestPlayer(t, database, "player-1", "Player1", "hash1")
	insertTestPlayer(t, database, "player-2", "Player2", "hash2")
	insertTestPlayer(t, database, "player-3", "Player3", "hash3")

	// Create multiple sessions
	session1, err := sm.CreateSession("player-1", "TEXT")
	require.NoError(t, err)

	session2, err := sm.CreateSession("player-2", "TEXT")
	require.NoError(t, err)

	session3, err := sm.CreateSession("player-3", "TEXT")
	require.NoError(t, err)

	// Verify all session IDs are unique
	assert.NotEqual(t, session1.SessionID, session2.SessionID)
	assert.NotEqual(t, session1.SessionID, session3.SessionID)
	assert.NotEqual(t, session2.SessionID, session3.SessionID)
}

func TestGetActiveSession(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	sm := NewSessionManager(database, logger)

	playerID := "player-123"
	insertTestPlayer(t, database, playerID, "TestPlayer", "hash123")

	// No active session initially
	session, err := sm.GetActiveSession(playerID)
	require.NoError(t, err)
	assert.Nil(t, session)

	// Create session
	createdSession, err := sm.CreateSession(playerID, "TEXT")
	require.NoError(t, err)

	// Should find active session
	session, err = sm.GetActiveSession(playerID)
	require.NoError(t, err)
	require.NotNil(t, session)
	assert.Equal(t, createdSession.SessionID, session.SessionID)
	assert.Equal(t, db.SessionConnected, session.State)
}

func TestGetActiveSession_OnlyReturnsConnected(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	sm := NewSessionManager(database, logger)

	playerID := "player-123"
	insertTestPlayer(t, database, playerID, "TestPlayer", "hash123")

	// Create session
	createdSession, err := sm.CreateSession(playerID, "TEXT")
	require.NoError(t, err)

	// Terminate session
	err = sm.TerminateSession(createdSession.SessionID)
	require.NoError(t, err)

	// Should not find active session
	session, err := sm.GetActiveSession(playerID)
	require.NoError(t, err)
	assert.Nil(t, session)
}

func TestUpdateSessionState(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	sm := NewSessionManager(database, logger)

	playerID := "player-123"
	insertTestPlayer(t, database, playerID, "TestPlayer", "hash123")

	// Create session
	session, err := sm.CreateSession(playerID, "TEXT")
	require.NoError(t, err)

	// Update to DISCONNECTED_LINGERING
	err = sm.UpdateSessionState(session.SessionID, db.SessionDisconnectedLingering)
	require.NoError(t, err)

	// Verify state updated in memory
	updatedSession, exists := sm.GetSession(session.SessionID)
	require.True(t, exists)
	assert.Equal(t, db.SessionDisconnectedLingering, updatedSession.State)
	assert.NotNil(t, updatedSession.DisconnectedAt)
	assert.NotNil(t, updatedSession.LingerExpiryAt)

	// Verify linger expiry is set to ~5 minutes from now
	expectedExpiry := time.Now().Unix() + 300
	assert.InDelta(t, expectedExpiry, *updatedSession.LingerExpiryAt, 5)
}

func TestTerminateSession(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	sm := NewSessionManager(database, logger)

	playerID := "player-123"
	insertTestPlayer(t, database, playerID, "TestPlayer", "hash123")

	// Create session
	session, err := sm.CreateSession(playerID, "TEXT")
	require.NoError(t, err)

	// Verify session is in active sessions
	assert.Equal(t, 1, len(sm.activeSessions))

	// Terminate session
	err = sm.TerminateSession(session.SessionID)
	require.NoError(t, err)

	// Verify session removed from active sessions map
	assert.Equal(t, 0, len(sm.activeSessions))

	// Verify session no longer exists in map
	_, exists := sm.GetSession(session.SessionID)
	assert.False(t, exists)
}

func TestActiveSessions(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	sm := NewSessionManager(database, logger)

	// Insert test players
	insertTestPlayer(t, database, "player-1", "Player1", "hash1")
	insertTestPlayer(t, database, "player-2", "Player2", "hash2")
	insertTestPlayer(t, database, "player-3", "Player3", "hash3")

	// Create sessions
	session1, err := sm.CreateSession("player-1", "TEXT")
	require.NoError(t, err)

	session2, err := sm.CreateSession("player-2", "TEXT")
	require.NoError(t, err)

	session3, err := sm.CreateSession("player-3", "TEXT")
	require.NoError(t, err)

	// All sessions should be active
	activeSessions := sm.ActiveSessions()
	assert.Equal(t, 3, len(activeSessions))

	// Disconnect one session
	err = sm.UpdateSessionState(session2.SessionID, db.SessionDisconnectedLingering)
	require.NoError(t, err)

	// Should only return 2 active sessions
	activeSessions = sm.ActiveSessions()
	assert.Equal(t, 2, len(activeSessions))

	// Verify correct sessions are returned
	sessionIDs := make(map[string]bool)
	for _, s := range activeSessions {
		sessionIDs[s.SessionID] = true
	}
	assert.True(t, sessionIDs[session1.SessionID])
	assert.False(t, sessionIDs[session2.SessionID])
	assert.True(t, sessionIDs[session3.SessionID])
}

func TestActiveSessionCount(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	sm := NewSessionManager(database, logger)

	// Initially no active sessions
	assert.Equal(t, 0, sm.ActiveSessionCount())

	// Insert test players and create sessions
	insertTestPlayer(t, database, "player-1", "Player1", "hash1")
	insertTestPlayer(t, database, "player-2", "Player2", "hash2")

	session1, err := sm.CreateSession("player-1", "TEXT")
	require.NoError(t, err)
	assert.Equal(t, 1, sm.ActiveSessionCount())

	_, err = sm.CreateSession("player-2", "TEXT")
	require.NoError(t, err)
	assert.Equal(t, 2, sm.ActiveSessionCount())

	// Terminate one session
	err = sm.TerminateSession(session1.SessionID)
	require.NoError(t, err)
	assert.Equal(t, 1, sm.ActiveSessionCount())
}

func TestGetSession(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	sm := NewSessionManager(database, logger)

	playerID := "player-123"
	insertTestPlayer(t, database, playerID, "TestPlayer", "hash123")

	// Create session
	createdSession, err := sm.CreateSession(playerID, "TEXT")
	require.NoError(t, err)

	// Get session by ID
	session, exists := sm.GetSession(createdSession.SessionID)
	assert.True(t, exists)
	assert.Equal(t, createdSession.SessionID, session.SessionID)

	// Try to get non-existent session
	_, exists = sm.GetSession("non-existent-id")
	assert.False(t, exists)
}

// mockConn is a mock connection for testing handshake protocol
type mockConn struct {
	readData  []byte
	writeData []byte
	readErr   error
	writeErr  error
}

func (m *mockConn) Read(p []byte) (n int, err error) {
	if m.readErr != nil {
		return 0, m.readErr
	}
	n = copy(p, m.readData)
	return n, nil
}

func (m *mockConn) Write(p []byte) (n int, err error) {
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	m.writeData = append(m.writeData, p...)
	return len(p), nil
}

func TestHandleHandshake_Success(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	sm := NewSessionManager(database, logger)

	// Insert test player
	playerID := "player-123"
	tokenHash := "valid-token-hash"
	insertTestPlayer(t, database, playerID, "TestPlayer", tokenHash)

	// Create mock connection with valid handshake response
	response := `{"type":"handshake_response","timestamp":1234567890,"protocol_version":"1.0","correlation_id":"test-correlation","payload":{"player_token":"valid-token-hash"}}`
	conn := &mockConn{
		readData: []byte(response),
	}

	config := HandshakeConfig{
		ServerName:     "Test Server",
		MOTD:           "Welcome to the test",
		TickIntervalMs: 2000,
	}

	// Execute handshake
	session, err := sm.HandleHandshake(conn, config)
	
	// Should succeed and return a session
	assert.NoError(t, err)
	require.NotNil(t, session)
	
	// Verify session properties
	assert.NotEmpty(t, session.SessionID)
	assert.Equal(t, playerID, session.PlayerID)
	assert.Equal(t, "TEXT", session.InterfaceMode)
	assert.Equal(t, db.SessionConnected, session.State)

	// Verify HandshakeInit was sent
	writeDataStr := string(conn.writeData)
	assert.Contains(t, writeDataStr, "handshake_init")
	assert.Contains(t, writeDataStr, "1.0")
	assert.Contains(t, writeDataStr, "TEXT")
	assert.Contains(t, writeDataStr, "Test Server")
	assert.Contains(t, writeDataStr, "Welcome to the test")
	
	// Verify HandshakeAck was sent
	assert.Contains(t, writeDataStr, "handshake_ack")
	assert.Contains(t, writeDataStr, session.SessionID)
	assert.Contains(t, writeDataStr, playerID)
	assert.Contains(t, writeDataStr, "2000") // tick_interval_ms
	
	// Verify session was created in database
	dbSession, err := database.GetActiveSessionByPlayerID(playerID)
	require.NoError(t, err)
	require.NotNil(t, dbSession)
	assert.Equal(t, session.SessionID, dbSession.SessionID)
}

func TestHandleHandshake_VersionMismatch(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	sm := NewSessionManager(database, logger)

	// Create mock connection with invalid protocol version
	response := `{"type":"handshake_response","timestamp":1234567890,"protocol_version":"2.0","correlation_id":"test-correlation","payload":{"player_token":"valid-token-hash"}}`
	conn := &mockConn{
		readData: []byte(response),
	}

	config := HandshakeConfig{
		ServerName:     "Test Server",
		MOTD:           "Welcome",
		TickIntervalMs: 2000,
	}

	// Execute handshake
	session, err := sm.HandleHandshake(conn, config)
	
	// Should return error
	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "protocol version mismatch")

	// Verify HandshakeReject was sent
	assert.Contains(t, string(conn.writeData), "handshake_reject")
	assert.Contains(t, string(conn.writeData), "version_mismatch")
}

func TestHandleHandshake_InvalidToken(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	sm := NewSessionManager(database, logger)

	// Insert test player with different token
	insertTestPlayer(t, database, "player-123", "TestPlayer", "correct-token-hash")

	// Create mock connection with invalid token
	response := `{"type":"handshake_response","timestamp":1234567890,"protocol_version":"1.0","correlation_id":"test-correlation","payload":{"player_token":"wrong-token-hash"}}`
	conn := &mockConn{
		readData: []byte(response),
	}

	config := HandshakeConfig{
		ServerName:     "Test Server",
		MOTD:           "Welcome",
		TickIntervalMs: 2000,
	}

	// Execute handshake
	session, err := sm.HandleHandshake(conn, config)
	
	// Should return error
	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "invalid player token")

	// Verify HandshakeReject was sent
	assert.Contains(t, string(conn.writeData), "handshake_reject")
	assert.Contains(t, string(conn.writeData), "invalid_token")
}

func TestHandleHandshake_SessionAlreadyActive(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	sm := NewSessionManager(database, logger)

	// Insert test player
	playerID := "player-123"
	tokenHash := "valid-token-hash"
	insertTestPlayer(t, database, playerID, "TestPlayer", tokenHash)

	// Create an existing active session
	_, err := sm.CreateSession(playerID, "TEXT")
	require.NoError(t, err)

	// Create mock connection with valid handshake response
	response := `{"type":"handshake_response","timestamp":1234567890,"protocol_version":"1.0","correlation_id":"test-correlation","payload":{"player_token":"valid-token-hash"}}`
	conn := &mockConn{
		readData: []byte(response),
	}

	config := HandshakeConfig{
		ServerName:     "Test Server",
		MOTD:           "Welcome",
		TickIntervalMs: 2000,
	}

	// Execute handshake
	session, err := sm.HandleHandshake(conn, config)
	
	// Should return error
	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "already has an active session")

	// Verify HandshakeReject was sent
	assert.Contains(t, string(conn.writeData), "handshake_reject")
	assert.Contains(t, string(conn.writeData), "session_already_active")
}

func TestHandleHandshake_HandshakeInit(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	sm := NewSessionManager(database, logger)

	// Insert test player
	insertTestPlayer(t, database, "player-123", "TestPlayer", "valid-token-hash")

	// Create mock connection
	response := `{"type":"handshake_response","timestamp":1234567890,"protocol_version":"1.0","correlation_id":"test-correlation","payload":{"player_token":"valid-token-hash"}}`
	conn := &mockConn{
		readData: []byte(response),
	}

	config := HandshakeConfig{
		ServerName:     "Black Sector",
		MOTD:           "Watch your back out there.",
		TickIntervalMs: 2000,
	}

	// Execute handshake
	_, err := sm.HandleHandshake(conn, config)
	require.NoError(t, err)

	// The writeData contains both handshake_init and handshake_ack separated by newlines
	// Split by newline to get individual messages
	writeDataStr := string(conn.writeData)
	messages := []string{}
	start := 0
	for i, c := range writeDataStr {
		if c == '\n' {
			if i > start {
				messages = append(messages, writeDataStr[start:i])
			}
			start = i + 1
		}
	}

	// First message should be handshake_init
	require.GreaterOrEqual(t, len(messages), 1, "Expected at least one message")
	
	var initMsg map[string]interface{}
	err = json.Unmarshal([]byte(messages[0]), &initMsg)
	require.NoError(t, err)

	// Verify all required fields
	assert.Equal(t, "handshake_init", initMsg["type"])
	assert.Equal(t, "1.0", initMsg["protocol_version"])
	assert.Equal(t, "TEXT", initMsg["interface_mode"])
	assert.Equal(t, "Black Sector", initMsg["server_name"])
	assert.Equal(t, "Watch your back out there.", initMsg["motd"])
	assert.NotNil(t, initMsg["timestamp"])
	assert.NotNil(t, initMsg["payload"])
}
