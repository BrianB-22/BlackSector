package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/BrianB-22/BlackSector/internal/config"
	"github.com/BrianB-22/BlackSector/internal/db"
	"github.com/BrianB-22/BlackSector/internal/snapshot"
)

func TestLoadSnapshotOrInitialize_NoSnapshot(t *testing.T) {
	// Setup: Create temp directory for test
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	snapshotsDir := filepath.Join(tempDir, "snapshots")
	
	// Create config
	cfg := &config.Config{
		Server: config.ServerConfig{
			DBPath: dbPath,
		},
	}
	
	// Initialize database
	logger := zerolog.New(os.Stderr).Level(zerolog.Disabled)
	database, err := db.InitDatabase(dbPath, logger)
	require.NoError(t, err)
	defer database.Close()
	
	// Change to temp directory so LoadSnapshot looks in the right place
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)
	
	// Test: Load snapshot when none exists
	tickNumber, err := loadSnapshotOrInitialize(cfg, database)
	
	// Assert: Should start at tick 0 with no error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), tickNumber)
	
	// Verify snapshots directory doesn't exist yet
	_, err = os.Stat(snapshotsDir)
	assert.True(t, os.IsNotExist(err))
}

func TestLoadSnapshotOrInitialize_WithSnapshot(t *testing.T) {
	// Setup: Create temp directory for test
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	snapshotsDir := filepath.Join(tempDir, "snapshots")
	
	// Create config
	cfg := &config.Config{
		Server: config.ServerConfig{
			DBPath: dbPath,
		},
	}
	
	// Initialize database
	logger := zerolog.New(os.Stderr).Level(zerolog.Disabled)
	database, err := db.InitDatabase(dbPath, logger)
	require.NoError(t, err)
	defer database.Close()
	
	// Create a snapshot
	testSnapshot := &snapshot.Snapshot{
		SnapshotVersion: "1.0",
		Tick:            100,
		Timestamp:       time.Now().Unix(),
		ServerName:      "Test Server",
		ProtocolVersion: "1.0",
		State: snapshot.SnapshotState{
			Players: []db.Player{
				{
					PlayerID:   "player-1",
					PlayerName: "TestPlayer",
					TokenHash:  "hash123",
					Credits:    1000,
					CreatedAt:  time.Now().Unix(),
				},
			},
			Sessions: []db.Session{
				{
					SessionID:     "session-1",
					PlayerID:      "player-1",
					InterfaceMode: "TEXT",
					State:         db.SessionConnected,
					ConnectedAt:   time.Now().Unix(),
				},
			},
		},
	}
	
	// Save the snapshot
	err = snapshot.SaveSnapshot(testSnapshot, snapshotsDir, logger)
	require.NoError(t, err)
	
	// Change to temp directory so LoadSnapshot looks in the right place
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)
	
	// Test: Load snapshot
	tickNumber, err := loadSnapshotOrInitialize(cfg, database)
	
	// Assert: Should start at snapshot.Tick + 1
	assert.NoError(t, err)
	assert.Equal(t, int64(101), tickNumber, "Tick number should be snapshot.Tick + 1")
}

func TestLoadSnapshotOrInitialize_MultipleSnapshots(t *testing.T) {
	// Setup: Create temp directory for test
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	snapshotsDir := filepath.Join(tempDir, "snapshots")
	
	// Create config
	cfg := &config.Config{
		Server: config.ServerConfig{
			DBPath: dbPath,
		},
	}
	
	// Initialize database
	logger := zerolog.New(os.Stderr).Level(zerolog.Disabled)
	database, err := db.InitDatabase(dbPath, logger)
	require.NoError(t, err)
	defer database.Close()
	
	// Create multiple snapshots at different ticks
	for _, tick := range []int64{50, 100, 150} {
		testSnapshot := &snapshot.Snapshot{
			SnapshotVersion: "1.0",
			Tick:            tick,
			Timestamp:       time.Now().Unix(),
			ServerName:      "Test Server",
			ProtocolVersion: "1.0",
			State: snapshot.SnapshotState{
				Players:  []db.Player{},
				Sessions: []db.Session{},
			},
		}
		
		err = snapshot.SaveSnapshot(testSnapshot, snapshotsDir, logger)
		require.NoError(t, err)
		
		// Small delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}
	
	// Change to temp directory so LoadSnapshot looks in the right place
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)
	
	// Test: Load snapshot (should load the most recent one)
	tickNumber, err := loadSnapshotOrInitialize(cfg, database)
	
	// Assert: Should start at the most recent snapshot's tick + 1
	assert.NoError(t, err)
	assert.Equal(t, int64(151), tickNumber, "Should load most recent snapshot (tick 150) and start at 151")
}

func TestLoadSnapshotOrInitialize_CorruptedSnapshot(t *testing.T) {
	// Setup: Create temp directory for test
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	snapshotsDir := filepath.Join(tempDir, "snapshots")
	
	// Create config
	cfg := &config.Config{
		Server: config.ServerConfig{
			DBPath: dbPath,
		},
	}
	
	// Initialize database
	logger := zerolog.New(os.Stderr).Level(zerolog.Disabled)
	database, err := db.InitDatabase(dbPath, logger)
	require.NoError(t, err)
	defer database.Close()
	
	// Create snapshots directory and write corrupted snapshot
	err = os.MkdirAll(snapshotsDir, 0755)
	require.NoError(t, err)
	
	corruptedPath := filepath.Join(snapshotsDir, "snapshot_100_123456.json")
	err = os.WriteFile(corruptedPath, []byte("{ invalid json }"), 0644)
	require.NoError(t, err)
	
	// Change to temp directory so LoadSnapshot looks in the right place
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)
	
	// Test: Load corrupted snapshot
	tickNumber, err := loadSnapshotOrInitialize(cfg, database)
	
	// Assert: Should return error for corrupted snapshot
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load snapshot")
	assert.Equal(t, int64(0), tickNumber)
}

func TestInitDatabase_Success(t *testing.T) {
	// Setup: Create temp directory for test
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	
	// Create config
	cfg := &config.Config{
		Server: config.ServerConfig{
			DBPath: dbPath,
		},
	}
	
	// Test: Initialize database
	database, err := initDatabase(cfg)
	
	// Assert: Should succeed
	require.NoError(t, err)
	require.NotNil(t, database)
	defer database.Close()
	
	// Verify database file was created
	_, err = os.Stat(dbPath)
	assert.NoError(t, err)
}

func TestInitDatabase_InvalidPath(t *testing.T) {
	// Create config with invalid path (read-only directory)
	cfg := &config.Config{
		Server: config.ServerConfig{
			DBPath: "/invalid/readonly/path/test.db",
		},
	}
	
	// Test: Initialize database with invalid path
	database, err := initDatabase(cfg)
	
	// Assert: Should fail
	assert.Error(t, err)
	assert.Nil(t, database)
	assert.Contains(t, err.Error(), "failed to initialize database")
}

func TestSignalHandling(t *testing.T) {
	// This test verifies that signal handling is properly configured
	// We test the signal channel setup and notification behavior
	
	// Setup: Create a signal channel like in main()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)
	
	// Test: Send SIGINT signal to current process
	process, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	
	err = process.Signal(syscall.SIGINT)
	require.NoError(t, err)
	
	// Assert: Signal should be received on channel
	select {
	case sig := <-sigChan:
		assert.Equal(t, syscall.SIGINT, sig, "Should receive SIGINT signal")
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for signal")
	}
}

func TestSignalHandling_SIGTERM(t *testing.T) {
	// This test verifies SIGTERM is also handled
	
	// Setup: Create a signal channel like in main()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)
	
	// Test: Send SIGTERM signal to current process
	process, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	
	err = process.Signal(syscall.SIGTERM)
	require.NoError(t, err)
	
	// Assert: Signal should be received on channel
	select {
	case sig := <-sigChan:
		assert.Equal(t, syscall.SIGTERM, sig, "Should receive SIGTERM signal")
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for signal")
	}
}
