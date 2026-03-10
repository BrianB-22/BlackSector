package db

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a temporary in-memory database for testing
func setupTestDB(t *testing.T) *Database {
	// Create a temporary directory for the test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Copy all migration files to a location the test can find
	migrationDir := filepath.Join(tmpDir, "migrations")
	require.NoError(t, os.MkdirAll(migrationDir, 0755))
	
	// List of all migration files to copy
	migrationFiles := []string{
		"001_initial_schema.sql",
		"002_add_auth_columns.sql",
		"003_add_combat_instances.sql",
		"004_add_mission_current_objective.sql",
		"005_add_performance_indexes.sql",
	}
	
	// Copy each migration file from the project root
	for _, migrationFile := range migrationFiles {
		migrationContent, err := os.ReadFile(filepath.Join("../../migrations", migrationFile))
		require.NoError(t, err, "Failed to read migration file %s", migrationFile)
		
		require.NoError(t, os.WriteFile(filepath.Join(migrationDir, migrationFile), migrationContent, 0644))
	}

	// Change to temp directory so the database can find the migrations
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() { os.Chdir(oldWd) })

	logger := zerolog.New(zerolog.NewTestWriter(t)).With().Timestamp().Logger()
	db, err := InitDatabase(dbPath, logger)
	require.NoError(t, err)
	require.NotNil(t, db)

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// insertTestPlayer inserts a test player into the database
func insertTestPlayer(t *testing.T, db *Database, player *Player) {
	query := `
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, last_login_at, is_banned)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	
	var lastLoginAt interface{}
	if player.LastLoginAt != nil {
		lastLoginAt = *player.LastLoginAt
	}
	
	isBanned := 0
	if player.IsBanned {
		isBanned = 1
	}

	_, err := db.conn.Exec(query, player.PlayerID, player.PlayerName, player.TokenHash,
		player.Credits, player.CreatedAt, lastLoginAt, isBanned)
	require.NoError(t, err)
}

func TestGetPlayerByToken(t *testing.T) {
	db := setupTestDB(t)

	t.Run("returns player when token exists", func(t *testing.T) {
		// Insert a test player
		now := time.Now().Unix()
		testPlayer := &Player{
			PlayerID:   "player-123",
			PlayerName: "TestPlayer",
			TokenHash:  "hash-abc123",
			Credits:    1000,
			CreatedAt:  now,
			IsBanned:   false,
		}
		insertTestPlayer(t, db, testPlayer)

		// Query by token
		player, err := db.GetPlayerByToken("hash-abc123")
		require.NoError(t, err)
		require.NotNil(t, player)

		assert.Equal(t, "player-123", player.PlayerID)
		assert.Equal(t, "TestPlayer", player.PlayerName)
		assert.Equal(t, "hash-abc123", player.TokenHash)
		assert.Equal(t, int64(1000), player.Credits)
		assert.Equal(t, now, player.CreatedAt)
		assert.Nil(t, player.LastLoginAt)
		assert.False(t, player.IsBanned)
	})

	t.Run("returns nil when token does not exist", func(t *testing.T) {
		player, err := db.GetPlayerByToken("nonexistent-token")
		require.NoError(t, err)
		assert.Nil(t, player)
	})

	t.Run("handles player with last login", func(t *testing.T) {
		now := time.Now().Unix()
		lastLogin := now - 3600
		testPlayer := &Player{
			PlayerID:    "player-456",
			PlayerName:  "PlayerWithLogin",
			TokenHash:   "hash-def456",
			Credits:     2000,
			CreatedAt:   now - 7200,
			LastLoginAt: &lastLogin,
			IsBanned:    false,
		}
		insertTestPlayer(t, db, testPlayer)

		player, err := db.GetPlayerByToken("hash-def456")
		require.NoError(t, err)
		require.NotNil(t, player)
		require.NotNil(t, player.LastLoginAt)
		assert.Equal(t, lastLogin, *player.LastLoginAt)
	})

	t.Run("handles banned player", func(t *testing.T) {
		now := time.Now().Unix()
		testPlayer := &Player{
			PlayerID:   "player-789",
			PlayerName: "BannedPlayer",
			TokenHash:  "hash-ghi789",
			Credits:    0,
			CreatedAt:  now,
			IsBanned:   true,
		}
		insertTestPlayer(t, db, testPlayer)

		player, err := db.GetPlayerByToken("hash-ghi789")
		require.NoError(t, err)
		require.NotNil(t, player)
		assert.True(t, player.IsBanned)
	})
}

func TestInsertSession(t *testing.T) {
	db := setupTestDB(t)

	// Insert a test player first
	now := time.Now().Unix()
	testPlayer := &Player{
		PlayerID:   "player-123",
		PlayerName: "TestPlayer",
		TokenHash:  "hash-abc123",
		Credits:    1000,
		CreatedAt:  now,
	}
	insertTestPlayer(t, db, testPlayer)

	t.Run("inserts session successfully", func(t *testing.T) {
		session := &Session{
			SessionID:      "session-001",
			PlayerID:       "player-123",
			InterfaceMode:  "TEXT",
			State:          SessionConnected,
			ConnectedAt:    now,
			LastActivityAt: now,
		}

		err := db.InsertSession(session)
		require.NoError(t, err)

		// Verify the session was inserted
		var count int
		err = db.conn.QueryRow("SELECT COUNT(*) FROM sessions WHERE session_id = ?", "session-001").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("inserts session with nullable fields", func(t *testing.T) {
		disconnectedAt := now + 100
		lingerExpiryAt := now + 200

		session := &Session{
			SessionID:      "session-002",
			PlayerID:       "player-123",
			InterfaceMode:  "GUI",
			State:          SessionDisconnectedLingering,
			ConnectedAt:    now,
			DisconnectedAt: &disconnectedAt,
			LingerExpiryAt: &lingerExpiryAt,
			LastActivityAt: now + 50,
		}

		err := db.InsertSession(session)
		require.NoError(t, err)

		// Verify the session was inserted with nullable fields
		var storedDisconnectedAt, storedLingerExpiryAt *int64
		err = db.conn.QueryRow(
			"SELECT disconnected_at, linger_expiry_at FROM sessions WHERE session_id = ?",
			"session-002",
		).Scan(&storedDisconnectedAt, &storedLingerExpiryAt)
		require.NoError(t, err)
		require.NotNil(t, storedDisconnectedAt)
		require.NotNil(t, storedLingerExpiryAt)
		assert.Equal(t, disconnectedAt, *storedDisconnectedAt)
		assert.Equal(t, lingerExpiryAt, *storedLingerExpiryAt)
	})
}

func TestUpdateSessionState(t *testing.T) {
	db := setupTestDB(t)

	// Insert a test player and session
	now := time.Now().Unix()
	testPlayer := &Player{
		PlayerID:   "player-123",
		PlayerName: "TestPlayer",
		TokenHash:  "hash-abc123",
		Credits:    1000,
		CreatedAt:  now,
	}
	insertTestPlayer(t, db, testPlayer)

	session := &Session{
		SessionID:      "session-001",
		PlayerID:       "player-123",
		InterfaceMode:  "TEXT",
		State:          SessionConnected,
		ConnectedAt:    now,
		LastActivityAt: now,
	}
	require.NoError(t, db.InsertSession(session))

	t.Run("updates session state successfully", func(t *testing.T) {
		err := db.UpdateSessionState("session-001", SessionDisconnectedLingering)
		require.NoError(t, err)

		// Verify the state was updated
		var state string
		err = db.conn.QueryRow("SELECT state FROM sessions WHERE session_id = ?", "session-001").Scan(&state)
		require.NoError(t, err)
		assert.Equal(t, string(SessionDisconnectedLingering), state)
	})

	t.Run("returns error for nonexistent session", func(t *testing.T) {
		err := db.UpdateSessionState("nonexistent-session", SessionTerminated)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "session not found")
	})
}

func TestGetActiveSessionByPlayerID(t *testing.T) {
	db := setupTestDB(t)

	// Insert a test player
	now := time.Now().Unix()
	testPlayer := &Player{
		PlayerID:   "player-123",
		PlayerName: "TestPlayer",
		TokenHash:  "hash-abc123",
		Credits:    1000,
		CreatedAt:  now,
	}
	insertTestPlayer(t, db, testPlayer)

	t.Run("returns active session when it exists", func(t *testing.T) {
		session := &Session{
			SessionID:      "session-001",
			PlayerID:       "player-123",
			InterfaceMode:  "TEXT",
			State:          SessionConnected,
			ConnectedAt:    now,
			LastActivityAt: now,
		}
		require.NoError(t, db.InsertSession(session))

		activeSession, err := db.GetActiveSessionByPlayerID("player-123")
		require.NoError(t, err)
		require.NotNil(t, activeSession)

		assert.Equal(t, "session-001", activeSession.SessionID)
		assert.Equal(t, "player-123", activeSession.PlayerID)
		assert.Equal(t, "TEXT", activeSession.InterfaceMode)
		assert.Equal(t, SessionConnected, activeSession.State)
		assert.Equal(t, now, activeSession.ConnectedAt)
		assert.Equal(t, now, activeSession.LastActivityAt)
	})

	t.Run("returns nil when no active session exists", func(t *testing.T) {
		activeSession, err := db.GetActiveSessionByPlayerID("player-nonexistent")
		require.NoError(t, err)
		assert.Nil(t, activeSession)
	})

	t.Run("ignores non-connected sessions", func(t *testing.T) {
		// Insert a player with a terminated session
		testPlayer2 := &Player{
			PlayerID:   "player-456",
			PlayerName: "TestPlayer2",
			TokenHash:  "hash-def456",
			Credits:    2000,
			CreatedAt:  now,
		}
		insertTestPlayer(t, db, testPlayer2)

		terminatedSession := &Session{
			SessionID:      "session-002",
			PlayerID:       "player-456",
			InterfaceMode:  "TEXT",
			State:          SessionTerminated,
			ConnectedAt:    now,
			LastActivityAt: now,
		}
		require.NoError(t, db.InsertSession(terminatedSession))

		activeSession, err := db.GetActiveSessionByPlayerID("player-456")
		require.NoError(t, err)
		assert.Nil(t, activeSession)
	})

	t.Run("returns most recent active session", func(t *testing.T) {
		// Insert a player with multiple connected sessions (edge case, but test it)
		testPlayer3 := &Player{
			PlayerID:   "player-789",
			PlayerName: "TestPlayer3",
			TokenHash:  "hash-ghi789",
			Credits:    3000,
			CreatedAt:  now,
		}
		insertTestPlayer(t, db, testPlayer3)

		// Insert older session
		olderSession := &Session{
			SessionID:      "session-003",
			PlayerID:       "player-789",
			InterfaceMode:  "TEXT",
			State:          SessionConnected,
			ConnectedAt:    now - 100,
			LastActivityAt: now - 100,
		}
		require.NoError(t, db.InsertSession(olderSession))

		// Insert newer session
		newerSession := &Session{
			SessionID:      "session-004",
			PlayerID:       "player-789",
			InterfaceMode:  "TEXT",
			State:          SessionConnected,
			ConnectedAt:    now,
			LastActivityAt: now,
		}
		require.NoError(t, db.InsertSession(newerSession))

		activeSession, err := db.GetActiveSessionByPlayerID("player-789")
		require.NoError(t, err)
		require.NotNil(t, activeSession)
		assert.Equal(t, "session-004", activeSession.SessionID)
	})
}
