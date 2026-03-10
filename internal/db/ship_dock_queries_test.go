package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateShipDockStatus_Dock(t *testing.T) {
	db := setupTestDB(t)

	// Insert required foreign key data
	_, err := db.conn.Exec(`INSERT INTO regions (region_id, name, region_type, security_level) VALUES (1, 'Test Region', 'core', 1.0)`)
	require.NoError(t, err)
	_, err = db.conn.Exec(`INSERT INTO systems (system_id, name, region_id, security_level, position_x, position_y) VALUES (1, 'Test System', 1, 1.0, 0.0, 0.0)`)
	require.NoError(t, err)

	// Insert test player
	player := &Player{
		PlayerID:   "test-player-1",
		PlayerName: "TestPlayer",
		TokenHash:  "hash123",
		Credits:    10000,
		CreatedAt:  1000,
	}
	err = db.InsertPlayer(player)
	require.NoError(t, err)

	// Insert test ship (IN_SPACE)
	query := `
		INSERT INTO ships (
			ship_id, player_id, ship_class, hull_points, max_hull_points,
			shield_points, max_shield_points, energy_points, max_energy_points,
			cargo_capacity, missiles_current, current_system_id, position_x, position_y,
			status, docked_at_port_id, last_updated_tick
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = db.conn.Exec(query,
		"test-ship-1", "test-player-1", "courier", 100, 100,
		50, 50, 100, 100,
		20, 0, 1, 0.0, 0.0,
		"IN_SPACE", nil, int64(0),
	)
	require.NoError(t, err)

	// Test: Update ship to docked status
	portID := 100
	err = db.UpdateShipDockStatus("test-ship-1", "DOCKED", &portID, 10)
	require.NoError(t, err)

	// Verify ship was updated
	updatedShip, err := db.GetShipByID("test-ship-1")
	require.NoError(t, err)
	require.NotNil(t, updatedShip)
	assert.Equal(t, "DOCKED", updatedShip.Status)
	assert.NotNil(t, updatedShip.DockedAtPortID)
	assert.Equal(t, 100, *updatedShip.DockedAtPortID)
	assert.Equal(t, int64(10), updatedShip.LastUpdatedTick)
}

func TestUpdateShipDockStatus_Undock(t *testing.T) {
	db := setupTestDB(t)

	// Insert required foreign key data
	_, err := db.conn.Exec(`INSERT INTO regions (region_id, name, region_type, security_level) VALUES (1, 'Test Region', 'core', 1.0)`)
	require.NoError(t, err)
	_, err = db.conn.Exec(`INSERT INTO systems (system_id, name, region_id, security_level, position_x, position_y) VALUES (1, 'Test System', 1, 1.0, 0.0, 0.0)`)
	require.NoError(t, err)
	_, err = db.conn.Exec(`INSERT INTO ports (port_id, system_id, name, port_type) VALUES (100, 1, 'Test Port', 'trading')`)
	require.NoError(t, err)

	// Insert test player
	player := &Player{
		PlayerID:   "test-player-1",
		PlayerName: "TestPlayer",
		TokenHash:  "hash123",
		Credits:    10000,
		CreatedAt:  1000,
	}
	err = db.InsertPlayer(player)
	require.NoError(t, err)

	// Insert test ship (DOCKED)
	portID := 100
	query := `
		INSERT INTO ships (
			ship_id, player_id, ship_class, hull_points, max_hull_points,
			shield_points, max_shield_points, energy_points, max_energy_points,
			cargo_capacity, missiles_current, current_system_id, position_x, position_y,
			status, docked_at_port_id, last_updated_tick
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = db.conn.Exec(query,
		"test-ship-1", "test-player-1", "courier", 100, 100,
		50, 50, 100, 100,
		20, 0, 1, 0.0, 0.0,
		"DOCKED", portID, int64(0),
	)
	require.NoError(t, err)

	// Test: Update ship to undocked status
	err = db.UpdateShipDockStatus("test-ship-1", "IN_SPACE", nil, 20)
	require.NoError(t, err)

	// Verify ship was updated
	updatedShip, err := db.GetShipByID("test-ship-1")
	require.NoError(t, err)
	require.NotNil(t, updatedShip)
	assert.Equal(t, "IN_SPACE", updatedShip.Status)
	assert.Nil(t, updatedShip.DockedAtPortID)
	assert.Equal(t, int64(20), updatedShip.LastUpdatedTick)
}

func TestUpdateShipDockStatus_ShipNotFound(t *testing.T) {
	db := setupTestDB(t)

	// Test: Update non-existent ship
	portID := 100
	err := db.UpdateShipDockStatus("nonexistent-ship", "DOCKED", &portID, 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ship not found")
}
