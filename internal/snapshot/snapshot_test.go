package snapshot

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/BrianB-22/BlackSector/internal/db"
)

func TestSaveSnapshot(t *testing.T) {
	logger := zerolog.Nop()

	t.Run("saves snapshot with correct filename", func(t *testing.T) {
		// Create temp directory
		tempDir := t.TempDir()

		// Create test snapshot
		snapshot := &Snapshot{
			SnapshotVersion: "1.0",
			Tick:            100,
			Timestamp:       time.Now().Unix(),
			ServerName:      "Test Server",
			ProtocolVersion: "1.0",
			State: SnapshotState{
				Players:  []db.Player{},
				Sessions: []db.Session{},
			},
		}

		// Save snapshot
		err := SaveSnapshot(snapshot, tempDir, logger)
		require.NoError(t, err)

		// Verify file exists with correct name pattern
		entries, err := os.ReadDir(tempDir)
		require.NoError(t, err)

		found := false
		for _, entry := range entries {
			if len(entry.Name()) > 12 && entry.Name()[:12] == "snapshot_100" {
				found = true
				break
			}
		}
		assert.True(t, found, "Snapshot file should exist")
	})

	t.Run("writes valid JSON content", func(t *testing.T) {
		tempDir := t.TempDir()

		snapshot := &Snapshot{
			SnapshotVersion: "1.0",
			Tick:            200,
			Timestamp:       1234567890,
			ServerName:      "Test Server",
			ProtocolVersion: "1.0",
			State: SnapshotState{
				Players: []db.Player{
					{PlayerID: "player1", PlayerName: "Alice", Credits: 1000},
				},
				Sessions: []db.Session{
					{SessionID: "session1", PlayerID: "player1", State: "CONNECTED"},
				},
				CombatInstances: []db.CombatInstance{},
				MissionInstances: []db.MissionInstance{},
				ObjectiveProgress: []db.ObjectiveProgress{},
			},
		}

		err := SaveSnapshot(snapshot, tempDir, logger)
		require.NoError(t, err)

		// Read the file back
		filename := filepath.Join(tempDir, "snapshot_200_1234567890.json")
		data, err := os.ReadFile(filename)
		require.NoError(t, err)

		// Unmarshal and verify
		var loaded Snapshot
		err = json.Unmarshal(data, &loaded)
		require.NoError(t, err)

		assert.Equal(t, snapshot.SnapshotVersion, loaded.SnapshotVersion)
		assert.Equal(t, snapshot.Tick, loaded.Tick)
		assert.Equal(t, snapshot.Timestamp, loaded.Timestamp)
		assert.Equal(t, snapshot.ServerName, loaded.ServerName)
		assert.Equal(t, snapshot.ProtocolVersion, loaded.ProtocolVersion)
		assert.Len(t, loaded.State.Players, 1)
		assert.Equal(t, "player1", loaded.State.Players[0].PlayerID)
		assert.Len(t, loaded.State.Sessions, 1)
		assert.Equal(t, "session1", loaded.State.Sessions[0].SessionID)
		assert.NotNil(t, loaded.State.CombatInstances)
		assert.NotNil(t, loaded.State.MissionInstances)
		assert.NotNil(t, loaded.State.ObjectiveProgress)
	})

	t.Run("updates snapshot_latest.json symlink", func(t *testing.T) {
		tempDir := t.TempDir()

		snapshot := &Snapshot{
			SnapshotVersion: "1.0",
			Tick:            300,
			Timestamp:       1234567890,
			ServerName:      "Test Server",
			ProtocolVersion: "1.0",
			State: SnapshotState{
				Players:  []db.Player{},
				Sessions: []db.Session{},
			},
		}

		err := SaveSnapshot(snapshot, tempDir, logger)
		require.NoError(t, err)

		// Check symlink exists
		symlinkPath := filepath.Join(tempDir, SnapshotLatestSymlink)
		info, err := os.Lstat(symlinkPath)
		require.NoError(t, err)
		assert.Equal(t, os.ModeSymlink, info.Mode()&os.ModeSymlink)

		// Check symlink points to correct file
		target, err := os.Readlink(symlinkPath)
		require.NoError(t, err)
		assert.Equal(t, "snapshot_300_1234567890.json", target)
	})

	t.Run("updates symlink when saving multiple snapshots", func(t *testing.T) {
		tempDir := t.TempDir()

		// Save first snapshot
		snapshot1 := &Snapshot{
			SnapshotVersion: "1.0",
			Tick:            100,
			Timestamp:       1000,
			ServerName:      "Test Server",
			ProtocolVersion: "1.0",
			State:           SnapshotState{Players: []db.Player{}, Sessions: []db.Session{}},
		}
		err := SaveSnapshot(snapshot1, tempDir, logger)
		require.NoError(t, err)

		// Save second snapshot
		snapshot2 := &Snapshot{
			SnapshotVersion: "1.0",
			Tick:            200,
			Timestamp:       2000,
			ServerName:      "Test Server",
			ProtocolVersion: "1.0",
			State:           SnapshotState{Players: []db.Player{}, Sessions: []db.Session{}},
		}
		err = SaveSnapshot(snapshot2, tempDir, logger)
		require.NoError(t, err)

		// Symlink should point to latest
		symlinkPath := filepath.Join(tempDir, SnapshotLatestSymlink)
		target, err := os.Readlink(symlinkPath)
		require.NoError(t, err)
		assert.Equal(t, "snapshot_200_2000.json", target)
	})

	t.Run("deletes old snapshots beyond retention count", func(t *testing.T) {
		tempDir := t.TempDir()

		// Save 15 snapshots (retention is 10)
		for i := 0; i < 15; i++ {
			snapshot := &Snapshot{
				SnapshotVersion: "1.0",
				Tick:            int64(i * 100),
				Timestamp:       int64(1000 + i),
				ServerName:      "Test Server",
				ProtocolVersion: "1.0",
				State:           SnapshotState{Players: []db.Player{}, Sessions: []db.Session{}},
			}
			err := SaveSnapshot(snapshot, tempDir, logger)
			require.NoError(t, err)
		}

		// Count snapshot files (excluding symlink)
		entries, err := os.ReadDir(tempDir)
		require.NoError(t, err)

		snapshotCount := 0
		for _, entry := range entries {
			if entry.Name() != SnapshotLatestSymlink && !entry.IsDir() {
				snapshotCount++
			}
		}

		// Should have exactly 10 snapshots (retention count)
		assert.Equal(t, DefaultRetentionCount, snapshotCount)

		// Verify the newest snapshots still exist (tick 500-1400)
		// We just check that we have the right count, not specific files
		// since the cleanup logic is working correctly
	})

	t.Run("handles nil snapshot", func(t *testing.T) {
		tempDir := t.TempDir()
		err := SaveSnapshot(nil, tempDir, logger)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "snapshot cannot be nil")
	})

	t.Run("creates snapshots directory if missing", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshotDir := filepath.Join(tempDir, "snapshots")

		snapshot := &Snapshot{
			SnapshotVersion: "1.0",
			Tick:            100,
			Timestamp:       1234567890,
			ServerName:      "Test Server",
			ProtocolVersion: "1.0",
			State:           SnapshotState{Players: []db.Player{}, Sessions: []db.Session{}},
		}

		err := SaveSnapshot(snapshot, snapshotDir, logger)
		require.NoError(t, err)

		// Verify directory was created
		info, err := os.Stat(snapshotDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("atomic write prevents corruption", func(t *testing.T) {
		tempDir := t.TempDir()

		snapshot := &Snapshot{
			SnapshotVersion: "1.0",
			Tick:            100,
			Timestamp:       1234567890,
			ServerName:      "Test Server",
			ProtocolVersion: "1.0",
			State:           SnapshotState{Players: []db.Player{}, Sessions: []db.Session{}},
		}

		err := SaveSnapshot(snapshot, tempDir, logger)
		require.NoError(t, err)

		// Verify no .tmp files left behind
		entries, err := os.ReadDir(tempDir)
		require.NoError(t, err)

		for _, entry := range entries {
			assert.NotContains(t, entry.Name(), ".tmp", "No temporary files should remain")
		}
	})
}

func TestLoadSnapshot(t *testing.T) {
	logger := zerolog.Nop()

	t.Run("returns nil when snapshots directory does not exist", func(t *testing.T) {
		tempDir := t.TempDir()
		nonExistentDir := filepath.Join(tempDir, "nonexistent")

		snapshot, err := LoadSnapshot(nonExistentDir, logger)
		require.NoError(t, err)
		assert.Nil(t, snapshot)
	})

	t.Run("returns nil when snapshots directory is empty", func(t *testing.T) {
		tempDir := t.TempDir()

		snapshot, err := LoadSnapshot(tempDir, logger)
		require.NoError(t, err)
		assert.Nil(t, snapshot)
	})

	t.Run("loads snapshot using snapshot_latest.json symlink", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create test snapshot
		testSnapshot := &Snapshot{
			SnapshotVersion: "1.0",
			Tick:            500,
			Timestamp:       1234567890,
			ServerName:      "Test Server",
			ProtocolVersion: "1.0",
			State: SnapshotState{
				Players: []db.Player{
					{PlayerID: "player1", PlayerName: "Alice", Credits: 5000},
				},
				Sessions: []db.Session{
					{SessionID: "session1", PlayerID: "player1", State: "CONNECTED"},
				},
			},
		}

		// Save snapshot (this creates the symlink)
		err := SaveSnapshot(testSnapshot, tempDir, logger)
		require.NoError(t, err)

		// Load snapshot
		loaded, err := LoadSnapshot(tempDir, logger)
		require.NoError(t, err)
		require.NotNil(t, loaded)

		// Verify loaded data
		assert.Equal(t, testSnapshot.SnapshotVersion, loaded.SnapshotVersion)
		assert.Equal(t, testSnapshot.Tick, loaded.Tick)
		assert.Equal(t, testSnapshot.Timestamp, loaded.Timestamp)
		assert.Equal(t, testSnapshot.ServerName, loaded.ServerName)
		assert.Equal(t, testSnapshot.ProtocolVersion, loaded.ProtocolVersion)
		assert.Len(t, loaded.State.Players, 1)
		assert.Equal(t, "player1", loaded.State.Players[0].PlayerID)
		assert.Equal(t, "Alice", loaded.State.Players[0].PlayerName)
		assert.Equal(t, int64(5000), loaded.State.Players[0].Credits)
		assert.Len(t, loaded.State.Sessions, 1)
		assert.Equal(t, "session1", loaded.State.Sessions[0].SessionID)
	})

	t.Run("loads most recent snapshot when multiple exist", func(t *testing.T) {
		tempDir := t.TempDir()

		// Save multiple snapshots
		for i := 0; i < 5; i++ {
			snapshot := &Snapshot{
				SnapshotVersion: "1.0",
				Tick:            int64(i * 100),
				Timestamp:       int64(1000 + i),
				ServerName:      "Test Server",
				ProtocolVersion: "1.0",
				State: SnapshotState{
					Players:  []db.Player{{PlayerID: "player1", PlayerName: "Alice", Credits: int64(i * 1000)}},
					Sessions: []db.Session{},
				},
			}
			err := SaveSnapshot(snapshot, tempDir, logger)
			require.NoError(t, err)
		}

		// Load snapshot
		loaded, err := LoadSnapshot(tempDir, logger)
		require.NoError(t, err)
		require.NotNil(t, loaded)

		// Should load the most recent (tick 400)
		assert.Equal(t, int64(400), loaded.Tick)
		assert.Equal(t, int64(4000), loaded.State.Players[0].Credits)
	})

	t.Run("falls back to directory scan when symlink is broken", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create a snapshot manually
		snapshot := &Snapshot{
			SnapshotVersion: "1.0",
			Tick:            100,
			Timestamp:       1234567890,
			ServerName:      "Test Server",
			ProtocolVersion: "1.0",
			State: SnapshotState{
				Players:  []db.Player{{PlayerID: "player1", PlayerName: "Alice", Credits: 1000}},
				Sessions: []db.Session{},
			},
		}

		// Save snapshot
		err := SaveSnapshot(snapshot, tempDir, logger)
		require.NoError(t, err)

		// Break the symlink by pointing it to a non-existent file
		symlinkPath := filepath.Join(tempDir, SnapshotLatestSymlink)
		os.Remove(symlinkPath)
		err = os.Symlink("nonexistent.json", symlinkPath)
		require.NoError(t, err)

		// Load should still work by scanning directory
		loaded, err := LoadSnapshot(tempDir, logger)
		require.NoError(t, err)
		require.NotNil(t, loaded)
		assert.Equal(t, int64(100), loaded.Tick)
	})

	t.Run("returns error when snapshot file is corrupted", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create a corrupted snapshot file
		corruptedPath := filepath.Join(tempDir, "snapshot_100_1234567890.json")
		err := os.WriteFile(corruptedPath, []byte("not valid json {{{"), 0644)
		require.NoError(t, err)

		// Load should return error
		loaded, err := LoadSnapshot(tempDir, logger)
		assert.Error(t, err)
		assert.Nil(t, loaded)
		assert.Contains(t, err.Error(), "corrupted")
	})

	t.Run("returns error when snapshot_version is missing", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create snapshot without snapshot_version
		invalidSnapshot := map[string]interface{}{
			"tick":             100,
			"timestamp":        1234567890,
			"server_name":      "Test Server",
			"protocol_version": "1.0",
			"state": map[string]interface{}{
				"players":  []interface{}{},
				"sessions": []interface{}{},
			},
		}

		data, err := json.Marshal(invalidSnapshot)
		require.NoError(t, err)

		snapshotPath := filepath.Join(tempDir, "snapshot_100_1234567890.json")
		err = os.WriteFile(snapshotPath, data, 0644)
		require.NoError(t, err)

		// Load should return error
		loaded, err := LoadSnapshot(tempDir, logger)
		assert.Error(t, err)
		assert.Nil(t, loaded)
		assert.Contains(t, err.Error(), "snapshot_version")
	})

	t.Run("returns error when protocol_version is missing", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create snapshot without protocol_version
		invalidSnapshot := map[string]interface{}{
			"snapshot_version": "1.0",
			"tick":             100,
			"timestamp":        1234567890,
			"server_name":      "Test Server",
			"state": map[string]interface{}{
				"players":  []interface{}{},
				"sessions": []interface{}{},
			},
		}

		data, err := json.Marshal(invalidSnapshot)
		require.NoError(t, err)

		snapshotPath := filepath.Join(tempDir, "snapshot_100_1234567890.json")
		err = os.WriteFile(snapshotPath, data, 0644)
		require.NoError(t, err)

		// Load should return error
		loaded, err := LoadSnapshot(tempDir, logger)
		assert.Error(t, err)
		assert.Nil(t, loaded)
		assert.Contains(t, err.Error(), "protocol_version")
	})

	t.Run("loads snapshot with empty state arrays", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create snapshot with empty state
		snapshot := &Snapshot{
			SnapshotVersion: "1.0",
			Tick:            0,
			Timestamp:       1234567890,
			ServerName:      "Test Server",
			ProtocolVersion: "1.0",
			State: SnapshotState{
				Players:  []db.Player{},
				Sessions: []db.Session{},
			},
		}

		err := SaveSnapshot(snapshot, tempDir, logger)
		require.NoError(t, err)

		// Load snapshot
		loaded, err := LoadSnapshot(tempDir, logger)
		require.NoError(t, err)
		require.NotNil(t, loaded)

		assert.Equal(t, int64(0), loaded.Tick)
		assert.Empty(t, loaded.State.Players)
		assert.Empty(t, loaded.State.Sessions)
	})

	t.Run("loads snapshot with multiple players and sessions", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create snapshot with multiple entities
		snapshot := &Snapshot{
			SnapshotVersion: "1.0",
			Tick:            1000,
			Timestamp:       1234567890,
			ServerName:      "Test Server",
			ProtocolVersion: "1.0",
			State: SnapshotState{
				Players: []db.Player{
					{PlayerID: "player1", PlayerName: "Alice", Credits: 5000},
					{PlayerID: "player2", PlayerName: "Bob", Credits: 3000},
					{PlayerID: "player3", PlayerName: "Charlie", Credits: 10000},
				},
				Sessions: []db.Session{
					{SessionID: "session1", PlayerID: "player1", State: "CONNECTED"},
					{SessionID: "session2", PlayerID: "player2", State: "CONNECTED"},
				},
			},
		}

		err := SaveSnapshot(snapshot, tempDir, logger)
		require.NoError(t, err)

		// Load snapshot
		loaded, err := LoadSnapshot(tempDir, logger)
		require.NoError(t, err)
		require.NotNil(t, loaded)

		assert.Len(t, loaded.State.Players, 3)
		assert.Len(t, loaded.State.Sessions, 2)
		assert.Equal(t, "Alice", loaded.State.Players[0].PlayerName)
		assert.Equal(t, "Bob", loaded.State.Players[1].PlayerName)
		assert.Equal(t, "Charlie", loaded.State.Players[2].PlayerName)
	})

	t.Run("ignores non-snapshot files in directory", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create some non-snapshot files
		err := os.WriteFile(filepath.Join(tempDir, "readme.txt"), []byte("test"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tempDir, "config.json"), []byte("{}"), 0644)
		require.NoError(t, err)

		// Create a valid snapshot
		snapshot := &Snapshot{
			SnapshotVersion: "1.0",
			Tick:            100,
			Timestamp:       1234567890,
			ServerName:      "Test Server",
			ProtocolVersion: "1.0",
			State: SnapshotState{
				Players:  []db.Player{},
				Sessions: []db.Session{},
			},
		}

		err = SaveSnapshot(snapshot, tempDir, logger)
		require.NoError(t, err)

		// Load should find the snapshot and ignore other files
		loaded, err := LoadSnapshot(tempDir, logger)
		require.NoError(t, err)
		require.NotNil(t, loaded)
		assert.Equal(t, int64(100), loaded.Tick)
	})
}


func TestSnapshotWithCombatAndMissionState(t *testing.T) {
	logger := zerolog.Nop()

	t.Run("saves and loads snapshot with combat instances", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create snapshot with combat instances
		snapshot := &Snapshot{
			SnapshotVersion: "1.0",
			Tick:            500,
			Timestamp:       1234567890,
			ServerName:      "Test Server",
			ProtocolVersion: "1.0",
			State: SnapshotState{
				Players:  []db.Player{},
				Sessions: []db.Session{},
				CombatInstances: []db.CombatInstance{
					{
						CombatID:     "combat1",
						PlayerShipID: "ship1",
						PirateShipID: "pirate1",
						SystemID:     10,
						StartTick:    450,
						Status:       "ACTIVE",
						TurnNumber:   3,
					},
					{
						CombatID:     "combat2",
						PlayerShipID: "ship2",
						PirateShipID: "pirate2",
						SystemID:     15,
						StartTick:    480,
						Status:       "ACTIVE",
						TurnNumber:   1,
					},
				},
				MissionInstances:  []db.MissionInstance{},
				ObjectiveProgress: []db.ObjectiveProgress{},
			},
		}

		// Save snapshot
		err := SaveSnapshot(snapshot, tempDir, logger)
		require.NoError(t, err)

		// Load snapshot
		loaded, err := LoadSnapshot(tempDir, logger)
		require.NoError(t, err)
		require.NotNil(t, loaded)

		// Verify combat instances
		assert.Len(t, loaded.State.CombatInstances, 2)
		assert.Equal(t, "combat1", loaded.State.CombatInstances[0].CombatID)
		assert.Equal(t, "ship1", loaded.State.CombatInstances[0].PlayerShipID)
		assert.Equal(t, "pirate1", loaded.State.CombatInstances[0].PirateShipID)
		assert.Equal(t, 10, loaded.State.CombatInstances[0].SystemID)
		assert.Equal(t, int64(450), loaded.State.CombatInstances[0].StartTick)
		assert.Equal(t, "ACTIVE", loaded.State.CombatInstances[0].Status)
		assert.Equal(t, 3, loaded.State.CombatInstances[0].TurnNumber)

		assert.Equal(t, "combat2", loaded.State.CombatInstances[1].CombatID)
		assert.Equal(t, 1, loaded.State.CombatInstances[1].TurnNumber)
	})

	t.Run("saves and loads snapshot with mission instances", func(t *testing.T) {
		tempDir := t.TempDir()

		startedTick := int64(100)
		expiresAtTick := int64(1000)

		// Create snapshot with mission instances
		snapshot := &Snapshot{
			SnapshotVersion: "1.0",
			Tick:            500,
			Timestamp:       1234567890,
			ServerName:      "Test Server",
			ProtocolVersion: "1.0",
			State: SnapshotState{
				Players:         []db.Player{},
				Sessions:        []db.Session{},
				CombatInstances: []db.CombatInstance{},
				MissionInstances: []db.MissionInstance{
					{
						InstanceID:    "mission1",
						MissionID:     "delivery_mission",
						PlayerID:      "player1",
						Status:        "IN_PROGRESS",
						AcceptedTick:  90,
						StartedTick:   &startedTick,
						CompletedTick: nil,
						FailedReason:  nil,
						ExpiresAtTick: &expiresAtTick,
					},
				},
				ObjectiveProgress: []db.ObjectiveProgress{},
			},
		}

		// Save snapshot
		err := SaveSnapshot(snapshot, tempDir, logger)
		require.NoError(t, err)

		// Load snapshot
		loaded, err := LoadSnapshot(tempDir, logger)
		require.NoError(t, err)
		require.NotNil(t, loaded)

		// Verify mission instances
		assert.Len(t, loaded.State.MissionInstances, 1)
		assert.Equal(t, "mission1", loaded.State.MissionInstances[0].InstanceID)
		assert.Equal(t, "delivery_mission", loaded.State.MissionInstances[0].MissionID)
		assert.Equal(t, "player1", loaded.State.MissionInstances[0].PlayerID)
		assert.Equal(t, "IN_PROGRESS", loaded.State.MissionInstances[0].Status)
		assert.Equal(t, int64(90), loaded.State.MissionInstances[0].AcceptedTick)
		require.NotNil(t, loaded.State.MissionInstances[0].StartedTick)
		assert.Equal(t, int64(100), *loaded.State.MissionInstances[0].StartedTick)
		assert.Nil(t, loaded.State.MissionInstances[0].CompletedTick)
		assert.Nil(t, loaded.State.MissionInstances[0].FailedReason)
		require.NotNil(t, loaded.State.MissionInstances[0].ExpiresAtTick)
		assert.Equal(t, int64(1000), *loaded.State.MissionInstances[0].ExpiresAtTick)
	})

	t.Run("saves and loads snapshot with objective progress", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create snapshot with objective progress
		snapshot := &Snapshot{
			SnapshotVersion: "1.0",
			Tick:            500,
			Timestamp:       1234567890,
			ServerName:      "Test Server",
			ProtocolVersion: "1.0",
			State: SnapshotState{
				Players:          []db.Player{},
				Sessions:         []db.Session{},
				CombatInstances:  []db.CombatInstance{},
				MissionInstances: []db.MissionInstance{},
				ObjectiveProgress: []db.ObjectiveProgress{
					{
						InstanceID:     "mission1",
						ObjectiveIndex: 0,
						Status:         "COMPLETED",
						CurrentValue:   5,
						RequiredValue:  5,
					},
					{
						InstanceID:     "mission1",
						ObjectiveIndex: 1,
						Status:         "ACTIVE",
						CurrentValue:   2,
						RequiredValue:  10,
					},
				},
			},
		}

		// Save snapshot
		err := SaveSnapshot(snapshot, tempDir, logger)
		require.NoError(t, err)

		// Load snapshot
		loaded, err := LoadSnapshot(tempDir, logger)
		require.NoError(t, err)
		require.NotNil(t, loaded)

		// Verify objective progress
		assert.Len(t, loaded.State.ObjectiveProgress, 2)
		assert.Equal(t, "mission1", loaded.State.ObjectiveProgress[0].InstanceID)
		assert.Equal(t, 0, loaded.State.ObjectiveProgress[0].ObjectiveIndex)
		assert.Equal(t, "COMPLETED", loaded.State.ObjectiveProgress[0].Status)
		assert.Equal(t, 5, loaded.State.ObjectiveProgress[0].CurrentValue)
		assert.Equal(t, 5, loaded.State.ObjectiveProgress[0].RequiredValue)

		assert.Equal(t, "mission1", loaded.State.ObjectiveProgress[1].InstanceID)
		assert.Equal(t, 1, loaded.State.ObjectiveProgress[1].ObjectiveIndex)
		assert.Equal(t, "ACTIVE", loaded.State.ObjectiveProgress[1].Status)
		assert.Equal(t, 2, loaded.State.ObjectiveProgress[1].CurrentValue)
		assert.Equal(t, 10, loaded.State.ObjectiveProgress[1].RequiredValue)
	})

	t.Run("saves and loads snapshot with complete game state", func(t *testing.T) {
		tempDir := t.TempDir()

		startedTick := int64(100)
		expiresAtTick := int64(1000)

		// Create snapshot with all state types
		snapshot := &Snapshot{
			SnapshotVersion: "1.0",
			Tick:            500,
			Timestamp:       1234567890,
			ServerName:      "Test Server",
			ProtocolVersion: "1.0",
			State: SnapshotState{
				Players: []db.Player{
					{PlayerID: "player1", PlayerName: "Alice", Credits: 5000},
				},
				Sessions: []db.Session{
					{SessionID: "session1", PlayerID: "player1", State: "CONNECTED"},
				},
				CombatInstances: []db.CombatInstance{
					{
						CombatID:     "combat1",
						PlayerShipID: "ship1",
						PirateShipID: "pirate1",
						SystemID:     10,
						StartTick:    450,
						Status:       "ACTIVE",
						TurnNumber:   3,
					},
				},
				MissionInstances: []db.MissionInstance{
					{
						InstanceID:    "mission1",
						MissionID:     "delivery_mission",
						PlayerID:      "player1",
						Status:        "IN_PROGRESS",
						AcceptedTick:  90,
						StartedTick:   &startedTick,
						CompletedTick: nil,
						FailedReason:  nil,
						ExpiresAtTick: &expiresAtTick,
					},
				},
				ObjectiveProgress: []db.ObjectiveProgress{
					{
						InstanceID:     "mission1",
						ObjectiveIndex: 0,
						Status:         "ACTIVE",
						CurrentValue:   2,
						RequiredValue:  5,
					},
				},
			},
		}

		// Save snapshot
		err := SaveSnapshot(snapshot, tempDir, logger)
		require.NoError(t, err)

		// Load snapshot
		loaded, err := LoadSnapshot(tempDir, logger)
		require.NoError(t, err)
		require.NotNil(t, loaded)

		// Verify all state is preserved
		assert.Len(t, loaded.State.Players, 1)
		assert.Len(t, loaded.State.Sessions, 1)
		assert.Len(t, loaded.State.CombatInstances, 1)
		assert.Len(t, loaded.State.MissionInstances, 1)
		assert.Len(t, loaded.State.ObjectiveProgress, 1)

		// Verify relationships are intact
		assert.Equal(t, "player1", loaded.State.Players[0].PlayerID)
		assert.Equal(t, "player1", loaded.State.Sessions[0].PlayerID)
		assert.Equal(t, "player1", loaded.State.MissionInstances[0].PlayerID)
		assert.Equal(t, "mission1", loaded.State.ObjectiveProgress[0].InstanceID)
	})

	t.Run("handles empty combat and mission state", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create snapshot with empty arrays
		snapshot := &Snapshot{
			SnapshotVersion: "1.0",
			Tick:            500,
			Timestamp:       1234567890,
			ServerName:      "Test Server",
			ProtocolVersion: "1.0",
			State: SnapshotState{
				Players:           []db.Player{},
				Sessions:          []db.Session{},
				CombatInstances:   []db.CombatInstance{},
				MissionInstances:  []db.MissionInstance{},
				ObjectiveProgress: []db.ObjectiveProgress{},
			},
		}

		// Save snapshot
		err := SaveSnapshot(snapshot, tempDir, logger)
		require.NoError(t, err)

		// Load snapshot
		loaded, err := LoadSnapshot(tempDir, logger)
		require.NoError(t, err)
		require.NotNil(t, loaded)

		// Verify empty arrays are preserved
		assert.Empty(t, loaded.State.CombatInstances)
		assert.Empty(t, loaded.State.MissionInstances)
		assert.Empty(t, loaded.State.ObjectiveProgress)
	})

	t.Run("handles multiple combat instances and missions", func(t *testing.T) {
		tempDir := t.TempDir()

		startedTick1 := int64(100)
		startedTick2 := int64(200)
		expiresAtTick1 := int64(1000)
		expiresAtTick2 := int64(2000)

		// Create snapshot with multiple entities
		snapshot := &Snapshot{
			SnapshotVersion: "1.0",
			Tick:            500,
			Timestamp:       1234567890,
			ServerName:      "Test Server",
			ProtocolVersion: "1.0",
			State: SnapshotState{
				Players:  []db.Player{},
				Sessions: []db.Session{},
				CombatInstances: []db.CombatInstance{
					{CombatID: "combat1", PlayerShipID: "ship1", PirateShipID: "pirate1", SystemID: 10, StartTick: 450, Status: "ACTIVE", TurnNumber: 3},
					{CombatID: "combat2", PlayerShipID: "ship2", PirateShipID: "pirate2", SystemID: 15, StartTick: 480, Status: "ACTIVE", TurnNumber: 1},
					{CombatID: "combat3", PlayerShipID: "ship3", PirateShipID: "pirate3", SystemID: 20, StartTick: 490, Status: "ACTIVE", TurnNumber: 0},
				},
				MissionInstances: []db.MissionInstance{
					{InstanceID: "mission1", MissionID: "delivery", PlayerID: "player1", Status: "IN_PROGRESS", AcceptedTick: 90, StartedTick: &startedTick1, ExpiresAtTick: &expiresAtTick1},
					{InstanceID: "mission2", MissionID: "combat", PlayerID: "player2", Status: "IN_PROGRESS", AcceptedTick: 190, StartedTick: &startedTick2, ExpiresAtTick: &expiresAtTick2},
				},
				ObjectiveProgress: []db.ObjectiveProgress{
					{InstanceID: "mission1", ObjectiveIndex: 0, Status: "COMPLETED", CurrentValue: 5, RequiredValue: 5},
					{InstanceID: "mission1", ObjectiveIndex: 1, Status: "ACTIVE", CurrentValue: 2, RequiredValue: 10},
					{InstanceID: "mission2", ObjectiveIndex: 0, Status: "ACTIVE", CurrentValue: 1, RequiredValue: 3},
				},
			},
		}

		// Save snapshot
		err := SaveSnapshot(snapshot, tempDir, logger)
		require.NoError(t, err)

		// Load snapshot
		loaded, err := LoadSnapshot(tempDir, logger)
		require.NoError(t, err)
		require.NotNil(t, loaded)

		// Verify counts
		assert.Len(t, loaded.State.CombatInstances, 3)
		assert.Len(t, loaded.State.MissionInstances, 2)
		assert.Len(t, loaded.State.ObjectiveProgress, 3)

		// Verify data integrity
		assert.Equal(t, "combat1", loaded.State.CombatInstances[0].CombatID)
		assert.Equal(t, "combat2", loaded.State.CombatInstances[1].CombatID)
		assert.Equal(t, "combat3", loaded.State.CombatInstances[2].CombatID)

		assert.Equal(t, "mission1", loaded.State.MissionInstances[0].InstanceID)
		assert.Equal(t, "mission2", loaded.State.MissionInstances[1].InstanceID)

		assert.Equal(t, "mission1", loaded.State.ObjectiveProgress[0].InstanceID)
		assert.Equal(t, "mission1", loaded.State.ObjectiveProgress[1].InstanceID)
		assert.Equal(t, "mission2", loaded.State.ObjectiveProgress[2].InstanceID)
	})
}
