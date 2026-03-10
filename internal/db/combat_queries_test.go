package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDataForCombat inserts required foreign key data (regions, systems, players, ships)
// and creates the combat_instances table (from migration 003)
func setupTestDataForCombat(t *testing.T, db *Database) {
	// Create combat_instances table (from migration 003)
	_, err := db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS combat_instances (
			combat_id         TEXT PRIMARY KEY,
			player_ship_id    TEXT NOT NULL REFERENCES ships(ship_id),
			pirate_ship_id    TEXT NOT NULL,
			system_id         INTEGER NOT NULL REFERENCES systems(system_id),
			start_tick        INTEGER NOT NULL,
			status            TEXT NOT NULL CHECK (status IN ('ACTIVE', 'ENDED', 'FLED')),
			turn_number       INTEGER NOT NULL DEFAULT 0
		)
	`)
	require.NoError(t, err)

	// Create indexes
	_, err = db.conn.Exec(`CREATE INDEX IF NOT EXISTS idx_combat_player_ship ON combat_instances (player_ship_id)`)
	require.NoError(t, err)
	_, err = db.conn.Exec(`CREATE INDEX IF NOT EXISTS idx_combat_status ON combat_instances (status)`)
	require.NoError(t, err)
	_, err = db.conn.Exec(`CREATE INDEX IF NOT EXISTS idx_combat_system ON combat_instances (system_id)`)
	require.NoError(t, err)

	// Insert test region
	_, err = db.conn.Exec(`INSERT INTO regions (region_id, name, region_type, security_level) VALUES (1, 'Test Region', 'core', 2.0)`)
	require.NoError(t, err)

	// Insert test system
	_, err = db.conn.Exec(`INSERT INTO systems (system_id, name, region_id, security_level, position_x, position_y) VALUES (1, 'Test System', 1, 0.5, 0.0, 0.0)`)
	require.NoError(t, err)

	// Insert test player directly (avoiding InsertPlayer which expects migration 002 columns)
	_, err = db.conn.Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, "player-001", "TestPlayer", "test-hash", 10000, 1000)
	require.NoError(t, err)

	// Insert test ship
	_, err = db.conn.Exec(`
		INSERT INTO ships (
			ship_id, player_id, ship_class, hull_points, max_hull_points,
			shield_points, max_shield_points, energy_points, max_energy_points,
			cargo_capacity, missiles_current, current_system_id, position_x, position_y,
			status, docked_at_port_id, last_updated_tick
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "ship-001", "player-001", "courier", 100, 100, 50, 50, 100, 100, 20, 0, 1, 0.0, 0.0, "IN_SPACE", nil, 100)
	require.NoError(t, err)
}

func TestCreateCombatInstance(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	setupTestDataForCombat(t, db)

	tests := []struct {
		name          string
		combatID      string
		playerShipID  string
		pirateShipID  string
		systemID      int
		startTick     int64
		expectError   bool
	}{
		{
			name:         "valid combat instance",
			combatID:     "combat-001",
			playerShipID: "ship-001",
			pirateShipID: "pirate-001",
			systemID:     1,
			startTick:    1000,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.CreateCombatInstance(tt.combatID, tt.playerShipID, tt.pirateShipID, tt.systemID, tt.startTick)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Verify the combat was created
				combat, err := db.GetActiveCombatByPlayerShip(tt.playerShipID)
				require.NoError(t, err)
				require.NotNil(t, combat)
				assert.Equal(t, tt.combatID, combat.CombatID)
				assert.Equal(t, tt.playerShipID, combat.PlayerShipID)
				assert.Equal(t, tt.pirateShipID, combat.PirateShipID)
				assert.Equal(t, tt.systemID, combat.SystemID)
				assert.Equal(t, tt.startTick, combat.StartTick)
				assert.Equal(t, "ACTIVE", combat.Status)
				assert.Equal(t, 0, combat.TurnNumber)
			}
		})
	}
}

func TestCreateCombatInstance_DuplicateID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	setupTestDataForCombat(t, db)

	// Create first combat
	err := db.CreateCombatInstance("combat-001", "ship-001", "pirate-001", 1, 1000)
	require.NoError(t, err)

	// Try to create duplicate
	err = db.CreateCombatInstance("combat-001", "ship-001", "pirate-002", 1, 1050)
	require.Error(t, err, "should fail on duplicate combat_id")
}

func TestGetActiveCombatByPlayerShip(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	setupTestDataForCombat(t, db)

	// Create a combat instance
	err := db.CreateCombatInstance("combat-001", "ship-001", "pirate-001", 1, 1000)
	require.NoError(t, err)

	// Test: Retrieve active combat
	combat, err := db.GetActiveCombatByPlayerShip("ship-001")
	require.NoError(t, err)
	require.NotNil(t, combat)

	assert.Equal(t, "combat-001", combat.CombatID)
	assert.Equal(t, "ship-001", combat.PlayerShipID)
	assert.Equal(t, "pirate-001", combat.PirateShipID)
	assert.Equal(t, 1, combat.SystemID)
	assert.Equal(t, int64(1000), combat.StartTick)
	assert.Equal(t, "ACTIVE", combat.Status)
	assert.Equal(t, 0, combat.TurnNumber)
}

func TestGetActiveCombatByPlayerShip_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	setupTestDataForCombat(t, db)

	// Test: Retrieve combat for ship with no active combat
	combat, err := db.GetActiveCombatByPlayerShip("ship-001")
	require.NoError(t, err)
	assert.Nil(t, combat)
}

func TestGetActiveCombatByPlayerShip_OnlyActive(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	setupTestDataForCombat(t, db)

	// Create a combat and end it
	err := db.CreateCombatInstance("combat-001", "ship-001", "pirate-001", 1, 1000)
	require.NoError(t, err)

	err = db.UpdateCombatStatus("combat-001", "ENDED")
	require.NoError(t, err)

	// Test: Should not find ended combat
	combat, err := db.GetActiveCombatByPlayerShip("ship-001")
	require.NoError(t, err)
	assert.Nil(t, combat, "should not return ended combat")
}

func TestGetActiveCombats(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	setupTestDataForCombat(t, db)

	// Insert additional ships for testing
	_, err := db.conn.Exec(`
		INSERT INTO ships (
			ship_id, player_id, ship_class, hull_points, max_hull_points,
			shield_points, max_shield_points, energy_points, max_energy_points,
			cargo_capacity, missiles_current, current_system_id, position_x, position_y,
			status, docked_at_port_id, last_updated_tick
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "ship-002", "player-001", "courier", 100, 100, 50, 50, 100, 100, 20, 0, 1, 0.0, 0.0, "IN_SPACE", nil, 100)
	require.NoError(t, err)

	// Create multiple combat instances
	err = db.CreateCombatInstance("combat-001", "ship-001", "pirate-001", 1, 1000)
	require.NoError(t, err)

	err = db.CreateCombatInstance("combat-002", "ship-002", "pirate-002", 1, 1050)
	require.NoError(t, err)

	// Create one ended combat
	err = db.CreateCombatInstance("combat-003", "ship-001", "pirate-003", 1, 1100)
	require.NoError(t, err)
	err = db.UpdateCombatStatus("combat-003", "ENDED")
	require.NoError(t, err)

	// Test: Get all active combats
	combats, err := db.GetActiveCombats()
	require.NoError(t, err)
	require.Len(t, combats, 2, "should return only active combats")

	// Verify combat IDs
	combatIDs := make(map[string]bool)
	for _, combat := range combats {
		combatIDs[combat.CombatID] = true
		assert.Equal(t, "ACTIVE", combat.Status)
	}

	assert.True(t, combatIDs["combat-001"])
	assert.True(t, combatIDs["combat-002"])
	assert.False(t, combatIDs["combat-003"], "ended combat should not be included")
}

func TestGetActiveCombats_Empty(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	setupTestDataForCombat(t, db)

	// Test: Get active combats when none exist
	combats, err := db.GetActiveCombats()
	require.NoError(t, err)
	assert.Empty(t, combats)
}

func TestUpdateCombatStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	setupTestDataForCombat(t, db)

	// Create a combat instance
	err := db.CreateCombatInstance("combat-001", "ship-001", "pirate-001", 1, 1000)
	require.NoError(t, err)

	tests := []struct {
		name        string
		combatID    string
		status      string
		expectError bool
	}{
		{
			name:        "update to ENDED",
			combatID:    "combat-001",
			status:      "ENDED",
			expectError: false,
		},
		{
			name:        "update to FLED",
			combatID:    "combat-001",
			status:      "FLED",
			expectError: false,
		},
		{
			name:        "update back to ACTIVE",
			combatID:    "combat-001",
			status:      "ACTIVE",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.UpdateCombatStatus(tt.combatID, tt.status)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Verify status was updated
				var status string
				err = db.conn.QueryRow("SELECT status FROM combat_instances WHERE combat_id = ?", tt.combatID).Scan(&status)
				require.NoError(t, err)
				assert.Equal(t, tt.status, status)
			}
		})
	}
}

func TestUpdateCombatStatus_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	setupTestDataForCombat(t, db)

	// Test: Update non-existent combat
	err := db.UpdateCombatStatus("nonexistent-combat", "ENDED")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "combat instance not found")
}

func TestUpdateCombatTurn(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	setupTestDataForCombat(t, db)

	// Create a combat instance
	err := db.CreateCombatInstance("combat-001", "ship-001", "pirate-001", 1, 1000)
	require.NoError(t, err)

	tests := []struct {
		name       string
		combatID   string
		turnNumber int
	}{
		{
			name:       "increment to turn 1",
			combatID:   "combat-001",
			turnNumber: 1,
		},
		{
			name:       "increment to turn 2",
			combatID:   "combat-001",
			turnNumber: 2,
		},
		{
			name:       "increment to turn 10",
			combatID:   "combat-001",
			turnNumber: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.UpdateCombatTurn(tt.combatID, tt.turnNumber)
			require.NoError(t, err)

			// Verify turn was updated
			var turnNumber int
			err = db.conn.QueryRow("SELECT turn_number FROM combat_instances WHERE combat_id = ?", tt.combatID).Scan(&turnNumber)
			require.NoError(t, err)
			assert.Equal(t, tt.turnNumber, turnNumber)
		})
	}
}

func TestUpdateCombatTurn_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	setupTestDataForCombat(t, db)

	// Test: Update non-existent combat
	err := db.UpdateCombatTurn("nonexistent-combat", 5)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "combat instance not found")
}

func TestDeleteCombat(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	setupTestDataForCombat(t, db)

	// Create a combat instance
	err := db.CreateCombatInstance("combat-001", "ship-001", "pirate-001", 1, 1000)
	require.NoError(t, err)

	// Verify it exists
	combat, err := db.GetActiveCombatByPlayerShip("ship-001")
	require.NoError(t, err)
	require.NotNil(t, combat)

	// Test: Delete combat
	err = db.DeleteCombat("combat-001")
	require.NoError(t, err)

	// Verify it's gone
	combat, err = db.GetActiveCombatByPlayerShip("ship-001")
	require.NoError(t, err)
	assert.Nil(t, combat)
}

func TestDeleteCombat_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	setupTestDataForCombat(t, db)

	// Test: Delete non-existent combat
	err := db.DeleteCombat("nonexistent-combat")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "combat instance not found")
}

func TestCombatLifecycle(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	setupTestDataForCombat(t, db)

	// Create combat
	err := db.CreateCombatInstance("combat-001", "ship-001", "pirate-001", 1, 1000)
	require.NoError(t, err)

	// Verify initial state
	combat, err := db.GetActiveCombatByPlayerShip("ship-001")
	require.NoError(t, err)
	require.NotNil(t, combat)
	assert.Equal(t, "ACTIVE", combat.Status)
	assert.Equal(t, 0, combat.TurnNumber)

	// Simulate combat turns
	for turn := 1; turn <= 5; turn++ {
		err = db.UpdateCombatTurn("combat-001", turn)
		require.NoError(t, err)
	}

	// Verify turn progression
	combat, err = db.GetActiveCombatByPlayerShip("ship-001")
	require.NoError(t, err)
	assert.Equal(t, 5, combat.TurnNumber)

	// End combat
	err = db.UpdateCombatStatus("combat-001", "ENDED")
	require.NoError(t, err)

	// Verify no longer active
	combat, err = db.GetActiveCombatByPlayerShip("ship-001")
	require.NoError(t, err)
	assert.Nil(t, combat, "ended combat should not be returned by GetActiveCombatByPlayerShip")

	// Clean up
	err = db.DeleteCombat("combat-001")
	require.NoError(t, err)
}
