package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetShipByPlayerID(t *testing.T) {
	// Setup in-memory database
	db := setupTestDB(t)
	defer db.Close()

	// Create test player
	playerID := uuid.New().String()
	_, err := db.conn.Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES (?, 'TestPlayer', 'hash123', 10000, ?, 0)
	`, playerID, time.Now().Unix())
	require.NoError(t, err)

	// Disable foreign key checks for test
	_, err = db.conn.Exec("PRAGMA foreign_keys = OFF")
	require.NoError(t, err)

	// Create test ship
	shipID := uuid.New().String()
	_, err = db.conn.Exec(`
		INSERT INTO ships (ship_id, player_id, ship_class, hull_points, max_hull_points,
			shield_points, max_shield_points, energy_points, max_energy_points,
			cargo_capacity, missiles_current, current_system_id, position_x, position_y,
			status, docked_at_port_id, last_updated_tick)
		VALUES (?, ?, 'courier', 100, 100, 50, 50, 100, 100, 20, 0, 1, 0.0, 0.0, 'IN_SPACE', NULL, 0)
	`, shipID, playerID)
	require.NoError(t, err)

	// Test retrieval
	ship, err := db.GetShipByPlayerID(playerID)
	require.NoError(t, err)
	require.NotNil(t, ship)

	// Verify ship data
	assert.Equal(t, shipID, ship.ShipID)
	assert.Equal(t, playerID, ship.PlayerID)
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
	assert.Equal(t, "IN_SPACE", ship.Status)
	assert.Nil(t, ship.DockedAtPortID)
	assert.Equal(t, int64(0), ship.LastUpdatedTick)
}

func TestGetShipByPlayerID_NotFound(t *testing.T) {
	// Setup in-memory database
	db := setupTestDB(t)
	defer db.Close()

	// Try to get ship for non-existent player
	ship, err := db.GetShipByPlayerID("non-existent-player")
	require.NoError(t, err)
	assert.Nil(t, ship)
}

func TestGetShipByPlayerID_WithDockedPort(t *testing.T) {
	// Setup in-memory database
	db := setupTestDB(t)
	defer db.Close()

	// Create test player
	playerID := uuid.New().String()
	_, err := db.conn.Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES (?, 'TestPlayer', 'hash123', 10000, ?, 0)
	`, playerID, time.Now().Unix())
	require.NoError(t, err)

	// Disable foreign key checks for test
	_, err = db.conn.Exec("PRAGMA foreign_keys = OFF")
	require.NoError(t, err)

	// Create test ship docked at port
	shipID := uuid.New().String()
	portID := 5
	_, err = db.conn.Exec(`
		INSERT INTO ships (ship_id, player_id, ship_class, hull_points, max_hull_points,
			shield_points, max_shield_points, energy_points, max_energy_points,
			cargo_capacity, missiles_current, current_system_id, position_x, position_y,
			status, docked_at_port_id, last_updated_tick)
		VALUES (?, ?, 'courier', 100, 100, 50, 50, 100, 100, 20, 0, 1, 0.0, 0.0, 'DOCKED', ?, 0)
	`, shipID, playerID, portID)
	require.NoError(t, err)

	// Test retrieval
	ship, err := db.GetShipByPlayerID(playerID)
	require.NoError(t, err)
	require.NotNil(t, ship)

	// Verify ship data
	assert.Equal(t, shipID, ship.ShipID)
	assert.Equal(t, playerID, ship.PlayerID)
	assert.Equal(t, "DOCKED", ship.Status)
	require.NotNil(t, ship.DockedAtPortID)
	assert.Equal(t, portID, *ship.DockedAtPortID)
}

func TestGetShipByPlayerID_MultipleShips(t *testing.T) {
	// Setup in-memory database
	db := setupTestDB(t)
	defer db.Close()

	// Create test player
	playerID := uuid.New().String()
	_, err := db.conn.Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES (?, 'TestPlayer', 'hash123', 10000, ?, 0)
	`, playerID, time.Now().Unix())
	require.NoError(t, err)

	// Disable foreign key checks for test
	_, err = db.conn.Exec("PRAGMA foreign_keys = OFF")
	require.NoError(t, err)

	// Create multiple ships for the same player
	shipID1 := uuid.New().String()
	shipID2 := uuid.New().String()

	_, err = db.conn.Exec(`
		INSERT INTO ships (ship_id, player_id, ship_class, hull_points, max_hull_points,
			shield_points, max_shield_points, energy_points, max_energy_points,
			cargo_capacity, missiles_current, current_system_id, position_x, position_y,
			status, docked_at_port_id, last_updated_tick)
		VALUES (?, ?, 'courier', 100, 100, 50, 50, 100, 100, 20, 0, 1, 0.0, 0.0, 'IN_SPACE', NULL, 0)
	`, shipID1, playerID)
	require.NoError(t, err)

	_, err = db.conn.Exec(`
		INSERT INTO ships (ship_id, player_id, ship_class, hull_points, max_hull_points,
			shield_points, max_shield_points, energy_points, max_energy_points,
			cargo_capacity, missiles_current, current_system_id, position_x, position_y,
			status, docked_at_port_id, last_updated_tick)
		VALUES (?, ?, 'freighter', 150, 150, 75, 75, 120, 120, 50, 0, 2, 0.0, 0.0, 'IN_SPACE', NULL, 0)
	`, shipID2, playerID)
	require.NoError(t, err)

	// Test retrieval - should return first ship (LIMIT 1)
	ship, err := db.GetShipByPlayerID(playerID)
	require.NoError(t, err)
	require.NotNil(t, ship)

	// Verify we got one of the ships (either is valid due to LIMIT 1)
	assert.Equal(t, playerID, ship.PlayerID)
	assert.Contains(t, []string{shipID1, shipID2}, ship.ShipID)
}
