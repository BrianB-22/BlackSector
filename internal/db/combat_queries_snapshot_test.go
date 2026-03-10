package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to insert test player
func insertTestPlayerForSnapshot(t *testing.T, db *Database, playerID, playerName string, credits int64) {
	_, err := db.conn.Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES (?, ?, ?, ?, ?, 0)
	`, playerID, playerName, "test-hash", credits, time.Now().Unix())
	require.NoError(t, err)
}

// Helper function to insert test system
func insertTestSystemForSnapshot(t *testing.T, db *Database) {
	// Insert region first
	_, err := db.conn.Exec(`
		INSERT INTO regions (region_id, name, region_type, security_level) 
		VALUES (1, 'Test Region', 'core', 0.8)
	`)
	require.NoError(t, err)

	// Insert system
	_, err = db.conn.Exec(`
		INSERT INTO systems (system_id, name, region_id, security_level, position_x, position_y)
		VALUES (1, 'Test System', 1, 0.5, 0.0, 0.0)
	`)
	require.NoError(t, err)
}

// Helper function to insert test ship
func insertTestShipForSnapshot(t *testing.T, db *Database, shipID, playerID, shipClass string) {
	_, err := db.conn.Exec(`
		INSERT INTO ships (ship_id, player_id, ship_class, hull_points, max_hull_points, 
			shield_points, max_shield_points, energy_points, max_energy_points, 
			cargo_capacity, current_system_id, status, last_updated_tick)
		VALUES (?, ?, ?, 100, 100, 50, 50, 100, 100, 20, 1, 'IN_SPACE', 0)
	`, shipID, playerID, shipClass)
	require.NoError(t, err)
}

func TestGetAllCombatInstances(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Common setup for all sub-tests
	insertTestSystemForSnapshot(t, db)
	insertTestPlayerForSnapshot(t, db, "player1", "Alice", 1000)

	t.Run("returns empty array when no combat instances exist", func(t *testing.T) {
		combats, err := db.GetAllCombatInstances()
		require.NoError(t, err)
		assert.Empty(t, combats)
	})

	t.Run("returns all combat instances", func(t *testing.T) {
		// Create test ships
		insertTestShipForSnapshot(t, db, "ship1", "player1", "courier")
		insertTestShipForSnapshot(t, db, "ship2", "player1", "courier")
		insertTestShipForSnapshot(t, db, "ship3", "player1", "courier")

		// Create test combat instances (pirate ships don't need to exist in DB)
		// Use system_id 1 which we created in setup
		err := db.CreateCombatInstance("combat1", "ship1", "pirate1", 1, 100)
		require.NoError(t, err)

		err = db.CreateCombatInstance("combat2", "ship2", "pirate2", 1, 150)
		require.NoError(t, err)

		err = db.CreateCombatInstance("combat3", "ship3", "pirate3", 1, 200)
		require.NoError(t, err)

		// Get all combat instances
		combats, err := db.GetAllCombatInstances()
		require.NoError(t, err)
		assert.Len(t, combats, 3)

		// Verify data
		assert.Equal(t, "combat1", combats[0].CombatID)
		assert.Equal(t, "ship1", combats[0].PlayerShipID)
		assert.Equal(t, "pirate1", combats[0].PirateShipID)
		assert.Equal(t, 1, combats[0].SystemID)
		assert.Equal(t, int64(100), combats[0].StartTick)
		assert.Equal(t, "ACTIVE", combats[0].Status)
		assert.Equal(t, 0, combats[0].TurnNumber)

		assert.Equal(t, "combat2", combats[1].CombatID)
		assert.Equal(t, "combat3", combats[2].CombatID)
	})

	t.Run("returns combat instances with different statuses", func(t *testing.T) {
		// Create test ships (reuse player1 and system from previous test, or create if needed)
		insertTestShipForSnapshot(t, db, "ship4", "player1", "courier")
		insertTestShipForSnapshot(t, db, "ship5", "player1", "courier")

		// Create combat instances with different statuses (pirate ships don't need to exist)
		err := db.CreateCombatInstance("combat4", "ship4", "pirate4", 1, 250)
		require.NoError(t, err)

		err = db.CreateCombatInstance("combat5", "ship5", "pirate5", 1, 300)
		require.NoError(t, err)

		// Update one to ended status
		err = db.UpdateCombatStatus("combat5", "ENDED")
		require.NoError(t, err)

		// Get all combat instances (should include both active and ended)
		combats, err := db.GetAllCombatInstances()
		require.NoError(t, err)
		
		// Find the specific combats
		var combat4, combat5 *CombatInstance
		for i := range combats {
			if combats[i].CombatID == "combat4" {
				combat4 = &combats[i]
			}
			if combats[i].CombatID == "combat5" {
				combat5 = &combats[i]
			}
		}

		require.NotNil(t, combat4)
		require.NotNil(t, combat5)
		assert.Equal(t, "ACTIVE", combat4.Status)
		assert.Equal(t, "ENDED", combat5.Status)
	})

	t.Run("returns combat instances with updated turn numbers", func(t *testing.T) {
		// Create test ship (reuse player1 and system from previous tests)
		insertTestShipForSnapshot(t, db, "ship6", "player1", "courier")

		// Create combat instance (pirate ship doesn't need to exist)
		err := db.CreateCombatInstance("combat6", "ship6", "pirate6", 1, 350)
		require.NoError(t, err)

		// Update turn number
		err = db.UpdateCombatTurn("combat6", 5)
		require.NoError(t, err)

		// Get all combat instances
		combats, err := db.GetAllCombatInstances()
		require.NoError(t, err)

		// Find combat6
		var combat6 *CombatInstance
		for i := range combats {
			if combats[i].CombatID == "combat6" {
				combat6 = &combats[i]
				break
			}
		}

		require.NotNil(t, combat6)
		assert.Equal(t, 5, combat6.TurnNumber)
	})
}
