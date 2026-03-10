package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitDatabase_NewDatabase(t *testing.T) {
	// Create temporary directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Create a test logger
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	// Initialize database
	db, err := InitDatabase(dbPath, logger)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	// Verify database file was created
	_, err = os.Stat(dbPath)
	assert.NoError(t, err, "Database file should exist")

	// Verify WAL files are created (indicates WAL mode is active)
	// WAL file may not exist immediately after creation, which is fine
	walPath := dbPath + "-wal"
	_, walErr := os.Stat(walPath)
	// Either WAL file exists or doesn't exist yet - both are acceptable
	assert.True(t, walErr == nil || os.IsNotExist(walErr), "WAL file check should not fail unexpectedly")
}

func TestInitDatabase_PRAGMAs(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	db, err := InitDatabase(dbPath, logger)
	require.NoError(t, err)
	defer db.Close()

	// Verify journal_mode is WAL
	var journalMode string
	err = db.conn.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	require.NoError(t, err)
	assert.Equal(t, "wal", journalMode, "Journal mode should be WAL")

	// Verify synchronous is NORMAL (value 1)
	var synchronous int
	err = db.conn.QueryRow("PRAGMA synchronous").Scan(&synchronous)
	require.NoError(t, err)
	assert.Equal(t, 1, synchronous, "Synchronous should be NORMAL (1)")

	// Verify foreign_keys is ON
	var foreignKeys int
	err = db.conn.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys)
	require.NoError(t, err)
	assert.Equal(t, 1, foreignKeys, "Foreign keys should be ON (1)")

	// Verify busy_timeout is 5000
	var busyTimeout int
	err = db.conn.QueryRow("PRAGMA busy_timeout").Scan(&busyTimeout)
	require.NoError(t, err)
	assert.Equal(t, 5000, busyTimeout, "Busy timeout should be 5000ms")
}

func TestInitDatabase_SchemaInitialization(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	db, err := InitDatabase(dbPath, logger)
	require.NoError(t, err)
	defer db.Close()

	// Verify key tables exist
	tables := []string{
		"players",
		"sessions",
		"ships",
		"systems",
		"regions",
		"ports",
		"commodities",
	}

	for _, table := range tables {
		var name string
		query := "SELECT name FROM sqlite_master WHERE type='table' AND name=?"
		err := db.conn.QueryRow(query, table).Scan(&name)
		assert.NoError(t, err, "Table %s should exist", table)
		assert.Equal(t, table, name)
	}
}

func TestInitDatabase_ExistingDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	// Create database first time
	db1, err := InitDatabase(dbPath, logger)
	require.NoError(t, err)
	db1.Close()

	// Open existing database
	db2, err := InitDatabase(dbPath, logger)
	require.NoError(t, err)
	defer db2.Close()

	// Verify it's the same database by checking tables still exist
	var name string
	err = db2.conn.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='players'").Scan(&name)
	assert.NoError(t, err)
	assert.Equal(t, "players", name)
}

func TestInitDatabase_InvalidPath(t *testing.T) {
	// Use an invalid path that cannot be created
	dbPath := "/root/impossible/path/test.db"
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	db, err := InitDatabase(dbPath, logger)
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "failed to create database directory")
}

func TestInitDatabase_MissingMigrationFile(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	// Change to a directory where migrations don't exist
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	db, err := InitDatabase(dbPath, logger)
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "failed to read migration file")
}

func TestDatabase_Close(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	db, err := InitDatabase(dbPath, logger)
	require.NoError(t, err)

	// Close database
	err = db.Close()
	assert.NoError(t, err)

	// Verify connection is closed by attempting a query
	var result int
	err = db.conn.QueryRow("SELECT 1").Scan(&result)
	assert.Error(t, err)
	// The error message varies between "sql: database is closed" and "sql: connection is already closed"
	assert.Contains(t, err.Error(), "closed")
}

func TestDatabase_Conn(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	db, err := InitDatabase(dbPath, logger)
	require.NoError(t, err)
	defer db.Close()

	// Get connection and verify it works
	conn := db.Conn()
	assert.NotNil(t, conn)

	var result int
	err = conn.QueryRow("SELECT 1").Scan(&result)
	assert.NoError(t, err)
	assert.Equal(t, 1, result)
}

func TestDatabase_ForeignKeyConstraints(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	db, err := InitDatabase(dbPath, logger)
	require.NoError(t, err)
	defer db.Close()

	// Insert a player
	_, err = db.conn.Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES ('player1', 'TestPlayer', 'hash123', 1000, 1234567890, 0)
	`)
	require.NoError(t, err)

	// Try to insert a session with invalid player_id (should fail due to foreign key)
	_, err = db.conn.Exec(`
		INSERT INTO sessions (session_id, player_id, interface_mode, state, connected_at, last_activity_at)
		VALUES ('session1', 'nonexistent', 'TEXT', 'CONNECTED', 1234567890, 1234567890)
	`)
	assert.Error(t, err, "Foreign key constraint should prevent invalid player_id")
	assert.Contains(t, err.Error(), "FOREIGN KEY constraint failed")

	// Insert session with valid player_id (should succeed)
	_, err = db.conn.Exec(`
		INSERT INTO sessions (session_id, player_id, interface_mode, state, connected_at, last_activity_at)
		VALUES ('session1', 'player1', 'TEXT', 'CONNECTED', 1234567890, 1234567890)
	`)
	assert.NoError(t, err, "Valid foreign key should allow insert")
}

func TestDatabase_BeginTx(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	db, err := InitDatabase(dbPath, logger)
	require.NoError(t, err)
	defer db.Close()

	// Begin a transaction
	tx, err := db.BeginTx()
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	// Clean up
	tx.Rollback()
}

func TestDatabase_CommitTx_Success(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	db, err := InitDatabase(dbPath, logger)
	require.NoError(t, err)
	defer db.Close()

	// Begin transaction
	tx, err := db.BeginTx()
	require.NoError(t, err)

	// Insert a player within the transaction
	_, err = tx.Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES ('player1', 'TestPlayer', 'hash123', 1000, 1234567890, 0)
	`)
	require.NoError(t, err)

	// Commit the transaction
	err = db.CommitTx(tx)
	assert.NoError(t, err)

	// Verify the player was inserted
	var playerName string
	err = db.conn.QueryRow("SELECT player_name FROM players WHERE player_id = 'player1'").Scan(&playerName)
	assert.NoError(t, err)
	assert.Equal(t, "TestPlayer", playerName)
}

func TestDatabase_CommitTx_NilTransaction(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	db, err := InitDatabase(dbPath, logger)
	require.NoError(t, err)
	defer db.Close()

	// Try to commit nil transaction
	err = db.CommitTx(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot commit nil transaction")
}

func TestDatabase_RollbackTx_Success(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	db, err := InitDatabase(dbPath, logger)
	require.NoError(t, err)
	defer db.Close()

	// Begin transaction
	tx, err := db.BeginTx()
	require.NoError(t, err)

	// Insert a player within the transaction
	_, err = tx.Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES ('player1', 'TestPlayer', 'hash123', 1000, 1234567890, 0)
	`)
	require.NoError(t, err)

	// Rollback the transaction
	err = db.RollbackTx(tx)
	assert.NoError(t, err)

	// Verify the player was NOT inserted
	var count int
	err = db.conn.QueryRow("SELECT COUNT(*) FROM players WHERE player_id = 'player1'").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count, "Player should not exist after rollback")
}

func TestDatabase_RollbackTx_NilTransaction(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	db, err := InitDatabase(dbPath, logger)
	require.NoError(t, err)
	defer db.Close()

	// Try to rollback nil transaction
	err = db.RollbackTx(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot rollback nil transaction")
}

func TestDatabase_TransactionAtomicity(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	db, err := InitDatabase(dbPath, logger)
	require.NoError(t, err)
	defer db.Close()

	// Test 1: Multiple related writes in a successful transaction
	tx, err := db.BeginTx()
	require.NoError(t, err)

	// Insert player
	_, err = tx.Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES ('player1', 'TestPlayer', 'hash123', 1000, 1234567890, 0)
	`)
	require.NoError(t, err)

	// Insert session for that player
	_, err = tx.Exec(`
		INSERT INTO sessions (session_id, player_id, interface_mode, state, connected_at, last_activity_at)
		VALUES ('session1', 'player1', 'TEXT', 'CONNECTED', 1234567890, 1234567890)
	`)
	require.NoError(t, err)

	// Commit transaction
	err = db.CommitTx(tx)
	require.NoError(t, err)

	// Verify both records exist
	var playerName string
	err = db.conn.QueryRow("SELECT player_name FROM players WHERE player_id = 'player1'").Scan(&playerName)
	assert.NoError(t, err)
	assert.Equal(t, "TestPlayer", playerName)

	var sessionID string
	err = db.conn.QueryRow("SELECT session_id FROM sessions WHERE session_id = 'session1'").Scan(&sessionID)
	assert.NoError(t, err)
	assert.Equal(t, "session1", sessionID)

	// Test 2: Multiple writes in a rolled-back transaction
	tx2, err := db.BeginTx()
	require.NoError(t, err)

	// Insert another player
	_, err = tx2.Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES ('player2', 'TestPlayer2', 'hash456', 2000, 1234567890, 0)
	`)
	require.NoError(t, err)

	// Insert session for that player
	_, err = tx2.Exec(`
		INSERT INTO sessions (session_id, player_id, interface_mode, state, connected_at, last_activity_at)
		VALUES ('session2', 'player2', 'TEXT', 'CONNECTED', 1234567890, 1234567890)
	`)
	require.NoError(t, err)

	// Rollback transaction
	err = db.RollbackTx(tx2)
	require.NoError(t, err)

	// Verify neither record exists
	var count int
	err = db.conn.QueryRow("SELECT COUNT(*) FROM players WHERE player_id = 'player2'").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count, "Player2 should not exist after rollback")

	err = db.conn.QueryRow("SELECT COUNT(*) FROM sessions WHERE session_id = 'session2'").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count, "Session2 should not exist after rollback")
}

func TestDatabase_TransactionFailureRollback(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	db, err := InitDatabase(dbPath, logger)
	require.NoError(t, err)
	defer db.Close()

	// Begin transaction
	tx, err := db.BeginTx()
	require.NoError(t, err)

	// Insert a valid player
	_, err = tx.Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES ('player1', 'TestPlayer', 'hash123', 1000, 1234567890, 0)
	`)
	require.NoError(t, err)

	// Try to insert a session with invalid foreign key (should fail)
	_, err = tx.Exec(`
		INSERT INTO sessions (session_id, player_id, interface_mode, state, connected_at, last_activity_at)
		VALUES ('session1', 'nonexistent', 'TEXT', 'CONNECTED', 1234567890, 1234567890)
	`)
	assert.Error(t, err, "Foreign key constraint should fail")

	// Rollback the transaction due to error
	err = db.RollbackTx(tx)
	assert.NoError(t, err)

	// Verify the player was NOT inserted (entire transaction rolled back)
	var count int
	err = db.conn.QueryRow("SELECT COUNT(*) FROM players WHERE player_id = 'player1'").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count, "Player should not exist after rollback")
}

func TestDatabase_MultipleTransactions(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	db, err := InitDatabase(dbPath, logger)
	require.NoError(t, err)
	defer db.Close()

	// Transaction 1: Insert player1
	tx1, err := db.BeginTx()
	require.NoError(t, err)

	_, err = tx1.Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES ('player1', 'TestPlayer1', 'hash123', 1000, 1234567890, 0)
	`)
	require.NoError(t, err)

	err = db.CommitTx(tx1)
	require.NoError(t, err)

	// Transaction 2: Insert player2
	tx2, err := db.BeginTx()
	require.NoError(t, err)

	_, err = tx2.Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES ('player2', 'TestPlayer2', 'hash456', 2000, 1234567890, 0)
	`)
	require.NoError(t, err)

	err = db.CommitTx(tx2)
	require.NoError(t, err)

	// Verify both players exist
	var count int
	err = db.conn.QueryRow("SELECT COUNT(*) FROM players").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 2, count, "Both players should exist")
}
