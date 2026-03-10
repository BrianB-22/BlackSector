package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDataForShips inserts required foreign key data (regions, systems, ports, players)
func setupTestDataForShips(t *testing.T, db *Database) {
	// Insert test region
	_, err := db.conn.Exec(`INSERT INTO regions (region_id, name, region_type, security_level) VALUES (1, 'Test Region', 'core', 2.0)`)
	require.NoError(t, err)
	
	// Insert test systems
	_, err = db.conn.Exec(`INSERT INTO systems (system_id, name, region_id, security_level, position_x, position_y) VALUES (1, 'Test System 1', 1, 2.0, 0.0, 0.0)`)
	require.NoError(t, err)
	_, err = db.conn.Exec(`INSERT INTO systems (system_id, name, region_id, security_level, position_x, position_y) VALUES (2, 'Test System 2', 1, 0.8, 10.0, 10.0)`)
	require.NoError(t, err)
	_, err = db.conn.Exec(`INSERT INTO systems (system_id, name, region_id, security_level, position_x, position_y) VALUES (3, 'Test System 3', 1, 0.2, 20.0, 20.0)`)
	require.NoError(t, err)
	_, err = db.conn.Exec(`INSERT INTO systems (system_id, name, region_id, security_level, position_x, position_y) VALUES (4, 'Test System 4', 1, 0.5, 30.0, 30.0)`)
	require.NoError(t, err)
	
	// Insert test port
	_, err = db.conn.Exec(`INSERT INTO ports (port_id, system_id, name, port_type, security_level) VALUES (1, 1, 'Test Port', 'trading', 2.0)`)
	require.NoError(t, err)
	
	// Insert test player
	player := &Player{
		PlayerID:   "player-001",
		PlayerName: "TestPlayer",
		TokenHash:  "test-hash",
		Credits:    10000,
		CreatedAt:  1000,
	}
	err = db.InsertPlayer(player)
	require.NoError(t, err)
}

func TestGetShipByID(t *testing.T) {
	// Setup in-memory database
	db := setupTestDB(t)
	defer db.Close()
	setupTestDataForShips(t, db)

	// Insert a test ship
	portID := 1
	_, err := db.conn.Exec(`
		INSERT INTO ships (
			ship_id, player_id, ship_class, hull_points, max_hull_points,
			shield_points, max_shield_points, energy_points, max_energy_points,
			cargo_capacity, missiles_current, current_system_id, position_x, position_y,
			status, docked_at_port_id, last_updated_tick
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "ship-001", "player-001", "courier", 100, 100, 50, 50, 100, 100, 20, 0, 1, 0.0, 0.0, "DOCKED", portID, 100)
	require.NoError(t, err)

	// Test: Retrieve ship by ID
	ship, err := db.GetShipByID("ship-001")
	require.NoError(t, err)
	require.NotNil(t, ship)

	// Verify ship data
	assert.Equal(t, "ship-001", ship.ShipID)
	assert.Equal(t, "player-001", ship.PlayerID)
	assert.Equal(t, "courier", ship.ShipClass)
	assert.Equal(t, 100, ship.HullPoints)
	assert.Equal(t, 100, ship.MaxHullPoints)
	assert.Equal(t, 50, ship.ShieldPoints)
	assert.Equal(t, 50, ship.MaxShieldPoints)
	assert.Equal(t, 100, ship.EnergyPoints)
	assert.Equal(t, 100, ship.MaxEnergyPoints)
	assert.Equal(t, 20, ship.CargoCapacity)
	assert.Equal(t, 0, ship.MissilesCurrent)
	assert.Equal(t, 1, ship.CurrentSystemID)
	assert.Equal(t, 0.0, ship.PositionX)
	assert.Equal(t, 0.0, ship.PositionY)
	assert.Equal(t, "DOCKED", ship.Status)
	assert.NotNil(t, ship.DockedAtPortID)
	assert.Equal(t, 1, *ship.DockedAtPortID)
	assert.Equal(t, int64(100), ship.LastUpdatedTick)
}

func TestGetShipByID_NotFound(t *testing.T) {
	// Setup in-memory database
	db := setupTestDB(t)
	defer db.Close()

	// Test: Retrieve non-existent ship
	ship, err := db.GetShipByID("nonexistent-ship")
	require.NoError(t, err)
	assert.Nil(t, ship)
}

func TestGetShipByID_NullDockedAtPortID(t *testing.T) {
	// Setup in-memory database
	db := setupTestDB(t)
	defer db.Close()
	setupTestDataForShips(t, db)

	// Insert a ship with NULL docked_at_port_id (in space)
	_, err := db.conn.Exec(`
		INSERT INTO ships (
			ship_id, player_id, ship_class, hull_points, max_hull_points,
			shield_points, max_shield_points, energy_points, max_energy_points,
			cargo_capacity, missiles_current, current_system_id, position_x, position_y,
			status, docked_at_port_id, last_updated_tick
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "ship-002", "player-001", "courier", 100, 100, 50, 50, 100, 100, 20, 0, 1, 0.0, 0.0, "IN_SPACE", nil, 100)
	require.NoError(t, err)

	// Test: Retrieve ship with NULL docked_at_port_id
	ship, err := db.GetShipByID("ship-002")
	require.NoError(t, err)
	require.NotNil(t, ship)

	assert.Equal(t, "ship-002", ship.ShipID)
	assert.Equal(t, "IN_SPACE", ship.Status)
	assert.Nil(t, ship.DockedAtPortID)
}

func TestUpdateShipPosition(t *testing.T) {
	// Setup in-memory database
	db := setupTestDB(t)
	defer db.Close()
	setupTestDataForShips(t, db)

	// Insert a test ship in system 1
	_, err := db.conn.Exec(`
		INSERT INTO ships (
			ship_id, player_id, ship_class, hull_points, max_hull_points,
			shield_points, max_shield_points, energy_points, max_energy_points,
			cargo_capacity, missiles_current, current_system_id, position_x, position_y,
			status, docked_at_port_id, last_updated_tick
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "ship-001", "player-001", "courier", 100, 100, 50, 50, 100, 100, 20, 0, 1, 0.0, 0.0, "IN_SPACE", nil, 100)
	require.NoError(t, err)

	// Test: Update ship position to system 2 at tick 105
	err = db.UpdateShipPosition("ship-001", 2, 105)
	require.NoError(t, err)

	// Verify: Retrieve ship and check updated values
	ship, err := db.GetShipByID("ship-001")
	require.NoError(t, err)
	require.NotNil(t, ship)

	assert.Equal(t, 2, ship.CurrentSystemID, "ship should be in system 2")
	assert.Equal(t, int64(105), ship.LastUpdatedTick, "last updated tick should be 105")
}

func TestUpdateShipPosition_NotFound(t *testing.T) {
	// Setup in-memory database
	db := setupTestDB(t)
	defer db.Close()

	// Test: Update non-existent ship
	err := db.UpdateShipPosition("nonexistent-ship", 2, 105)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ship not found")
}

func TestUpdateShipPosition_MultipleUpdates(t *testing.T) {
	// Setup in-memory database
	db := setupTestDB(t)
	defer db.Close()
	setupTestDataForShips(t, db)

	// Insert a test ship
	_, err := db.conn.Exec(`
		INSERT INTO ships (
			ship_id, player_id, ship_class, hull_points, max_hull_points,
			shield_points, max_shield_points, energy_points, max_energy_points,
			cargo_capacity, missiles_current, current_system_id, position_x, position_y,
			status, docked_at_port_id, last_updated_tick
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "ship-001", "player-001", "courier", 100, 100, 50, 50, 100, 100, 20, 0, 1, 0.0, 0.0, "IN_SPACE", nil, 100)
	require.NoError(t, err)

	// Test: Multiple position updates
	updates := []struct {
		systemID int
		tick     int64
	}{
		{2, 105},
		{3, 110},
		{4, 115},
	}

	for _, update := range updates {
		err = db.UpdateShipPosition("ship-001", update.systemID, update.tick)
		require.NoError(t, err)

		// Verify each update
		ship, err := db.GetShipByID("ship-001")
		require.NoError(t, err)
		assert.Equal(t, update.systemID, ship.CurrentSystemID)
		assert.Equal(t, update.tick, ship.LastUpdatedTick)
	}
}
