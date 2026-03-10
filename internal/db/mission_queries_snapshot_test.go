package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to insert test player for mission tests
func insertTestPlayerForMission(t *testing.T, db *Database, playerID, playerName string, credits int64) {
	_, err := db.conn.Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES (?, ?, ?, ?, ?, 0)
	`, playerID, playerName, "test-hash", credits, time.Now().Unix())
	require.NoError(t, err)
}

func TestGetAllMissionInstances(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	t.Run("returns empty array when no mission instances exist", func(t *testing.T) {
		missions, err := db.GetAllMissionInstances()
		require.NoError(t, err)
		assert.Empty(t, missions)
	})

	t.Run("returns all mission instances", func(t *testing.T) {
		// Create test players first (foreign key requirement)
		insertTestPlayerForMission(t, db, "player1", "Alice", 1000)
		insertTestPlayerForMission(t, db, "player2", "Bob", 2000)

		// Create test mission instances
		startedTick1 := int64(100)
		expiresAtTick1 := int64(1000)
		mission1 := &MissionInstance{
			InstanceID:    "mission1",
			MissionID:     "delivery",
			PlayerID:      "player1",
			Status:        "IN_PROGRESS",
			AcceptedTick:  90,
			StartedTick:   &startedTick1,
			ExpiresAtTick: &expiresAtTick1,
		}
		err := db.CreateMissionInstance(mission1)
		require.NoError(t, err)

		startedTick2 := int64(200)
		expiresAtTick2 := int64(2000)
		mission2 := &MissionInstance{
			InstanceID:    "mission2",
			MissionID:     "combat",
			PlayerID:      "player2",
			Status:        "IN_PROGRESS",
			AcceptedTick:  190,
			StartedTick:   &startedTick2,
			ExpiresAtTick: &expiresAtTick2,
		}
		err = db.CreateMissionInstance(mission2)
		require.NoError(t, err)

		// Get all mission instances
		missions, err := db.GetAllMissionInstances()
		require.NoError(t, err)
		assert.Len(t, missions, 2)

		// Verify data
		assert.Equal(t, "mission1", missions[0].InstanceID)
		assert.Equal(t, "delivery", missions[0].MissionID)
		assert.Equal(t, "player1", missions[0].PlayerID)
		assert.Equal(t, "IN_PROGRESS", missions[0].Status)
		assert.Equal(t, int64(90), missions[0].AcceptedTick)
		require.NotNil(t, missions[0].StartedTick)
		assert.Equal(t, int64(100), *missions[0].StartedTick)
		require.NotNil(t, missions[0].ExpiresAtTick)
		assert.Equal(t, int64(1000), *missions[0].ExpiresAtTick)

		assert.Equal(t, "mission2", missions[1].InstanceID)
		assert.Equal(t, "combat", missions[1].MissionID)
	})

	t.Run("returns mission instances with different statuses", func(t *testing.T) {
		// Create test players
		insertTestPlayerForMission(t, db, "player3", "Charlie", 3000)
		insertTestPlayerForMission(t, db, "player4", "Diana", 4000)

		// Create mission instances with different statuses
		startedTick3 := int64(300)
		expiresAtTick3 := int64(3000)
		mission3 := &MissionInstance{
			InstanceID:    "mission3",
			MissionID:     "exploration",
			PlayerID:      "player3",
			Status:        "IN_PROGRESS",
			AcceptedTick:  290,
			StartedTick:   &startedTick3,
			ExpiresAtTick: &expiresAtTick3,
		}
		err := db.CreateMissionInstance(mission3)
		require.NoError(t, err)

		startedTick4 := int64(400)
		completedTick4 := int64(500)
		expiresAtTick4 := int64(4000)
		mission4 := &MissionInstance{
			InstanceID:    "mission4",
			MissionID:     "trade",
			PlayerID:      "player4",
			Status:        "COMPLETED",
			AcceptedTick:  390,
			StartedTick:   &startedTick4,
			CompletedTick: &completedTick4,
			ExpiresAtTick: &expiresAtTick4,
		}
		err = db.CreateMissionInstance(mission4)
		require.NoError(t, err)

		// Get all mission instances (should include all statuses)
		missions, err := db.GetAllMissionInstances()
		require.NoError(t, err)

		// Find the specific missions
		var foundMission3, foundMission4 *MissionInstance
		for i := range missions {
			if missions[i].InstanceID == "mission3" {
				foundMission3 = &missions[i]
			}
			if missions[i].InstanceID == "mission4" {
				foundMission4 = &missions[i]
			}
		}

		require.NotNil(t, foundMission3)
		require.NotNil(t, foundMission4)
		assert.Equal(t, "IN_PROGRESS", foundMission3.Status)
		assert.Equal(t, "COMPLETED", foundMission4.Status)
		require.NotNil(t, foundMission4.CompletedTick)
		assert.Equal(t, int64(500), *foundMission4.CompletedTick)
	})

	t.Run("handles mission instances with null fields", func(t *testing.T) {
		// Create test player
		insertTestPlayerForMission(t, db, "player5", "Eve", 5000)

		// Create mission instance with minimal fields
		mission5 := &MissionInstance{
			InstanceID:    "mission5",
			MissionID:     "minimal",
			PlayerID:      "player5",
			Status:        "ACCEPTED",
			AcceptedTick:  500,
			StartedTick:   nil,
			CompletedTick: nil,
			FailedReason:  nil,
			ExpiresAtTick: nil,
		}
		err := db.CreateMissionInstance(mission5)
		require.NoError(t, err)

		// Get all mission instances
		missions, err := db.GetAllMissionInstances()
		require.NoError(t, err)

		// Find mission5
		var foundMission5 *MissionInstance
		for i := range missions {
			if missions[i].InstanceID == "mission5" {
				foundMission5 = &missions[i]
				break
			}
		}

		require.NotNil(t, foundMission5)
		assert.Equal(t, "ACCEPTED", foundMission5.Status)
		assert.Nil(t, foundMission5.StartedTick)
		assert.Nil(t, foundMission5.CompletedTick)
		assert.Nil(t, foundMission5.FailedReason)
		assert.Nil(t, foundMission5.ExpiresAtTick)
	})
}

func TestGetAllObjectiveProgressForSnapshot(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	t.Run("returns empty array when no objective progress exists", func(t *testing.T) {
		progress, err := db.GetAllObjectiveProgressForSnapshot()
		require.NoError(t, err)
		assert.Empty(t, progress)
	})

	t.Run("returns all objective progress", func(t *testing.T) {
		// Create test player first
		insertTestPlayerForMission(t, db, "player1", "Alice", 1000)

		// Create test mission instance first
		startedTick := int64(100)
		expiresAtTick := int64(1000)
		mission := &MissionInstance{
			InstanceID:    "mission1",
			MissionID:     "delivery",
			PlayerID:      "player1",
			Status:        "IN_PROGRESS",
			AcceptedTick:  90,
			StartedTick:   &startedTick,
			ExpiresAtTick: &expiresAtTick,
		}
		err := db.CreateMissionInstance(mission)
		require.NoError(t, err)

		// Create objective progress
		progress1 := &ObjectiveProgress{
			InstanceID:     "mission1",
			ObjectiveIndex: 0,
			Status:         "COMPLETED",
			CurrentValue:   5,
			RequiredValue:  5,
		}
		err = db.CreateObjectiveProgress(progress1)
		require.NoError(t, err)

		progress2 := &ObjectiveProgress{
			InstanceID:     "mission1",
			ObjectiveIndex: 1,
			Status:         "ACTIVE",
			CurrentValue:   2,
			RequiredValue:  10,
		}
		err = db.CreateObjectiveProgress(progress2)
		require.NoError(t, err)

		// Get all objective progress
		allProgress, err := db.GetAllObjectiveProgressForSnapshot()
		require.NoError(t, err)
		assert.Len(t, allProgress, 2)

		// Verify data (should be ordered by instance_id, objective_index)
		assert.Equal(t, "mission1", allProgress[0].InstanceID)
		assert.Equal(t, 0, allProgress[0].ObjectiveIndex)
		assert.Equal(t, "COMPLETED", allProgress[0].Status)
		assert.Equal(t, 5, allProgress[0].CurrentValue)
		assert.Equal(t, 5, allProgress[0].RequiredValue)

		assert.Equal(t, "mission1", allProgress[1].InstanceID)
		assert.Equal(t, 1, allProgress[1].ObjectiveIndex)
		assert.Equal(t, "ACTIVE", allProgress[1].Status)
		assert.Equal(t, 2, allProgress[1].CurrentValue)
		assert.Equal(t, 10, allProgress[1].RequiredValue)
	})

	t.Run("returns objective progress for multiple missions", func(t *testing.T) {
		// Create test player
		insertTestPlayerForMission(t, db, "player2", "Bob", 2000)

		// Create second mission instance
		startedTick := int64(200)
		expiresAtTick := int64(2000)
		mission2 := &MissionInstance{
			InstanceID:    "mission2",
			MissionID:     "combat",
			PlayerID:      "player2",
			Status:        "IN_PROGRESS",
			AcceptedTick:  190,
			StartedTick:   &startedTick,
			ExpiresAtTick: &expiresAtTick,
		}
		err := db.CreateMissionInstance(mission2)
		require.NoError(t, err)

		// Create objective progress for mission2
		progress3 := &ObjectiveProgress{
			InstanceID:     "mission2",
			ObjectiveIndex: 0,
			Status:         "ACTIVE",
			CurrentValue:   1,
			RequiredValue:  3,
		}
		err = db.CreateObjectiveProgress(progress3)
		require.NoError(t, err)

		// Get all objective progress
		allProgress, err := db.GetAllObjectiveProgressForSnapshot()
		require.NoError(t, err)

		// Should have progress from both missions
		mission1Count := 0
		mission2Count := 0
		for _, p := range allProgress {
			if p.InstanceID == "mission1" {
				mission1Count++
			}
			if p.InstanceID == "mission2" {
				mission2Count++
			}
		}

		assert.Equal(t, 2, mission1Count, "Should have 2 objectives for mission1")
		assert.Equal(t, 1, mission2Count, "Should have 1 objective for mission2")
	})

	t.Run("returns objective progress ordered by instance and index", func(t *testing.T) {
		// Create test player
		insertTestPlayerForMission(t, db, "player3", "Charlie", 3000)

		// Create third mission with multiple objectives
		startedTick := int64(300)
		expiresAtTick := int64(3000)
		mission3 := &MissionInstance{
			InstanceID:    "mission3",
			MissionID:     "exploration",
			PlayerID:      "player3",
			Status:        "IN_PROGRESS",
			AcceptedTick:  290,
			StartedTick:   &startedTick,
			ExpiresAtTick: &expiresAtTick,
		}
		err := db.CreateMissionInstance(mission3)
		require.NoError(t, err)

		// Create objectives in non-sequential order
		progress5 := &ObjectiveProgress{
			InstanceID:     "mission3",
			ObjectiveIndex: 2,
			Status:         "PENDING",
			CurrentValue:   0,
			RequiredValue:  1,
		}
		err = db.CreateObjectiveProgress(progress5)
		require.NoError(t, err)

		progress4 := &ObjectiveProgress{
			InstanceID:     "mission3",
			ObjectiveIndex: 0,
			Status:         "ACTIVE",
			CurrentValue:   1,
			RequiredValue:  5,
		}
		err = db.CreateObjectiveProgress(progress4)
		require.NoError(t, err)

		progress6 := &ObjectiveProgress{
			InstanceID:     "mission3",
			ObjectiveIndex: 1,
			Status:         "PENDING",
			CurrentValue:   0,
			RequiredValue:  3,
		}
		err = db.CreateObjectiveProgress(progress6)
		require.NoError(t, err)

		// Get all objective progress
		allProgress, err := db.GetAllObjectiveProgressForSnapshot()
		require.NoError(t, err)

		// Find mission3 objectives
		var mission3Progress []ObjectiveProgress
		for _, p := range allProgress {
			if p.InstanceID == "mission3" {
				mission3Progress = append(mission3Progress, p)
			}
		}

		// Should be ordered by objective_index
		require.Len(t, mission3Progress, 3)
		assert.Equal(t, 0, mission3Progress[0].ObjectiveIndex)
		assert.Equal(t, 1, mission3Progress[1].ObjectiveIndex)
		assert.Equal(t, 2, mission3Progress[2].ObjectiveIndex)
	})
}
