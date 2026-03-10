package engine

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/BrianB-22/BlackSector/internal/db"
	"github.com/BrianB-22/BlackSector/internal/session"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNavigationIntegration_JumpCommand tests jump command processing
func TestNavigationIntegration_JumpCommand(t *testing.T) {
	config := Config{
		TickIntervalMs:        100,
		SnapshotIntervalTicks: 0,
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}

	engine, database, nav, _ := setupTestEngine(t, config)
	defer database.Close()

	// Disable foreign key checks for test
	_, err := database.Conn().Exec("PRAGMA foreign_keys = OFF")
	require.NoError(t, err)

	// Create test player and ship
	playerID := uuid.New().String()
	shipID := uuid.New().String()

	_, err = database.Conn().Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES (?, 'TestPlayer', 'hash123', 10000, ?, 0)
	`, playerID, time.Now().Unix())
	require.NoError(t, err)

	_, err = database.Conn().Exec(`
		INSERT INTO ships (ship_id, player_id, ship_class, hull_points, max_hull_points,
			shield_points, max_shield_points, energy_points, max_energy_points,
			cargo_capacity, missiles_current, current_system_id, position_x, position_y,
			status, docked_at_port_id, last_updated_tick)
		VALUES (?, ?, 'courier', 100, 100, 50, 50, 100, 100, 20, 0, 1, 0.0, 0.0, 'IN_SPACE', NULL, 0)
	`, shipID, playerID)
	require.NoError(t, err)

	// Create jump command
	payload := JumpPayload{TargetSystemID: 2}
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	cmd := Command{
		SessionID:   "session-123",
		PlayerID:    playerID,
		CommandType: "jump",
		Payload:     payloadBytes,
		EnqueuedAt:  time.Now().Unix(),
	}

	// Enqueue command
	engine.EnqueueCommand(cmd)

	// Process commands
	engine.drainCommandQueue()

	// Verify navigation.Jump was called
	assert.True(t, nav.jumpCalled, "Navigation.Jump should have been called")
}

// TestNavigationIntegration_DockCommand tests dock command processing
func TestNavigationIntegration_DockCommand(t *testing.T) {
	config := Config{
		TickIntervalMs:        100,
		SnapshotIntervalTicks: 0,
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}

	engine, database, nav, _ := setupTestEngine(t, config)
	defer database.Close()

	// Disable foreign key checks for test
	_, err := database.Conn().Exec("PRAGMA foreign_keys = OFF")
	require.NoError(t, err)

	// Create test player and ship
	playerID := uuid.New().String()
	shipID := uuid.New().String()

	_, err = database.Conn().Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES (?, 'TestPlayer', 'hash123', 10000, ?, 0)
	`, playerID, time.Now().Unix())
	require.NoError(t, err)

	_, err = database.Conn().Exec(`
		INSERT INTO ships (ship_id, player_id, ship_class, hull_points, max_hull_points,
			shield_points, max_shield_points, energy_points, max_energy_points,
			cargo_capacity, missiles_current, current_system_id, position_x, position_y,
			status, docked_at_port_id, last_updated_tick)
		VALUES (?, ?, 'courier', 100, 100, 50, 50, 100, 100, 20, 0, 1, 0.0, 0.0, 'IN_SPACE', NULL, 0)
	`, shipID, playerID)
	require.NoError(t, err)

	// Create dock command
	payload := DockPayload{PortID: 1}
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	cmd := Command{
		SessionID:   "session-123",
		PlayerID:    playerID,
		CommandType: "dock",
		Payload:     payloadBytes,
		EnqueuedAt:  time.Now().Unix(),
	}

	// Enqueue command
	engine.EnqueueCommand(cmd)

	// Process commands
	engine.drainCommandQueue()

	// Verify navigation.Dock was called
	assert.True(t, nav.dockCalled, "Navigation.Dock should have been called")
}

// TestNavigationIntegration_UndockCommand tests undock command processing
func TestNavigationIntegration_UndockCommand(t *testing.T) {
	config := Config{
		TickIntervalMs:        100,
		SnapshotIntervalTicks: 0,
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}

	engine, database, nav, _ := setupTestEngine(t, config)
	defer database.Close()

	// Disable foreign key checks for test
	_, err := database.Conn().Exec("PRAGMA foreign_keys = OFF")
	require.NoError(t, err)

	// Create test player and ship
	playerID := uuid.New().String()
	shipID := uuid.New().String()

	_, err = database.Conn().Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES (?, 'TestPlayer', 'hash123', 10000, ?, 0)
	`, playerID, time.Now().Unix())
	require.NoError(t, err)

	_, err = database.Conn().Exec(`
		INSERT INTO ships (ship_id, player_id, ship_class, hull_points, max_hull_points,
			shield_points, max_shield_points, energy_points, max_energy_points,
			cargo_capacity, missiles_current, current_system_id, position_x, position_y,
			status, docked_at_port_id, last_updated_tick)
		VALUES (?, ?, 'courier', 100, 100, 50, 50, 100, 100, 20, 0, 1, 0.0, 0.0, 'DOCKED', 1, 0)
	`, shipID, playerID)
	require.NoError(t, err)

	// Create undock command
	cmd := Command{
		SessionID:   "session-123",
		PlayerID:    playerID,
		CommandType: "undock",
		Payload:     []byte(`{}`),
		EnqueuedAt:  time.Now().Unix(),
	}

	// Enqueue command
	engine.EnqueueCommand(cmd)

	// Process commands
	engine.drainCommandQueue()

	// Verify navigation.Undock was called
	assert.True(t, nav.undockCalled, "Navigation.Undock should have been called")
}

// TestNavigationIntegration_NoShipFound tests error handling when player has no ship
func TestNavigationIntegration_NoShipFound(t *testing.T) {
	config := Config{
		TickIntervalMs:        100,
		SnapshotIntervalTicks: 0,
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}

	engine, database, nav, _ := setupTestEngine(t, config)
	defer database.Close()

	// Create test player without ship
	playerID := uuid.New().String()

	_, err := database.Conn().Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES (?, 'TestPlayer', 'hash123', 10000, ?, 0)
	`, playerID, time.Now().Unix())
	require.NoError(t, err)

	// Create jump command
	payload := JumpPayload{TargetSystemID: 2}
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	cmd := Command{
		SessionID:   "session-123",
		PlayerID:    playerID,
		CommandType: "jump",
		Payload:     payloadBytes,
		EnqueuedAt:  time.Now().Unix(),
	}

	// Enqueue command
	engine.EnqueueCommand(cmd)

	// Process commands - should handle error gracefully
	engine.drainCommandQueue()

	// Verify navigation.Jump was NOT called (error occurred before)
	assert.False(t, nav.jumpCalled, "Navigation.Jump should not have been called when ship not found")
}

// TestNavigationIntegration_NavigationError tests error handling from navigation subsystem
func TestNavigationIntegration_NavigationError(t *testing.T) {
	config := Config{
		TickIntervalMs:        100,
		SnapshotIntervalTicks: 0,
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}

	logger := zerolog.Nop()
	database, err := db.InitDatabase(":memory:", logger)
	require.NoError(t, err)
	defer database.Close()

	// Create mock navigator that returns error
	nav := &mockNavigator{
		jumpError: errors.New("ship is docked"),
	}

	sessionMgr := session.NewSessionManager(database, logger)
	engine := NewTickEngine(config, database, sessionMgr, nav, &mockTrader{}, &mockCombatResolver{}, &mockMissionController{}, logger)

	// Disable foreign key checks for test
	_, err = database.Conn().Exec("PRAGMA foreign_keys = OFF")
	require.NoError(t, err)

	// Create test player and ship
	playerID := uuid.New().String()
	shipID := uuid.New().String()

	_, err = database.Conn().Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES (?, 'TestPlayer', 'hash123', 10000, ?, 0)
	`, playerID, time.Now().Unix())
	require.NoError(t, err)

	_, err = database.Conn().Exec(`
		INSERT INTO ships (ship_id, player_id, ship_class, hull_points, max_hull_points,
			shield_points, max_shield_points, energy_points, max_energy_points,
			cargo_capacity, missiles_current, current_system_id, position_x, position_y,
			status, docked_at_port_id, last_updated_tick)
		VALUES (?, ?, 'courier', 100, 100, 50, 50, 100, 100, 20, 0, 1, 0.0, 0.0, 'DOCKED', 1, 0)
	`, shipID, playerID)
	require.NoError(t, err)

	// Create jump command
	payload := JumpPayload{TargetSystemID: 2}
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	cmd := Command{
		SessionID:   "session-123",
		PlayerID:    playerID,
		CommandType: "jump",
		Payload:     payloadBytes,
		EnqueuedAt:  time.Now().Unix(),
	}

	// Enqueue command
	engine.EnqueueCommand(cmd)

	// Process commands - should handle error gracefully
	engine.drainCommandQueue()

	// Verify navigation.Jump was called (error occurred in navigation layer)
	assert.True(t, nav.jumpCalled, "Navigation.Jump should have been called")
}

// TestNavigationIntegration_InvalidPayload tests error handling for malformed JSON
func TestNavigationIntegration_InvalidPayload(t *testing.T) {
	config := Config{
		TickIntervalMs:        100,
		SnapshotIntervalTicks: 0,
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}

	engine, database, nav, _ := setupTestEngine(t, config)
	defer database.Close()

	// Disable foreign key checks for test
	_, err := database.Conn().Exec("PRAGMA foreign_keys = OFF")
	require.NoError(t, err)

	// Create test player and ship
	playerID := uuid.New().String()
	shipID := uuid.New().String()

	_, err = database.Conn().Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES (?, 'TestPlayer', 'hash123', 10000, ?, 0)
	`, playerID, time.Now().Unix())
	require.NoError(t, err)

	_, err = database.Conn().Exec(`
		INSERT INTO ships (ship_id, player_id, ship_class, hull_points, max_hull_points,
			shield_points, max_shield_points, energy_points, max_energy_points,
			cargo_capacity, missiles_current, current_system_id, position_x, position_y,
			status, docked_at_port_id, last_updated_tick)
		VALUES (?, ?, 'courier', 100, 100, 50, 50, 100, 100, 20, 0, 1, 0.0, 0.0, 'IN_SPACE', NULL, 0)
	`, shipID, playerID)
	require.NoError(t, err)

	// Create jump command with invalid JSON
	cmd := Command{
		SessionID:   "session-123",
		PlayerID:    playerID,
		CommandType: "jump",
		Payload:     []byte(`{invalid json`),
		EnqueuedAt:  time.Now().Unix(),
	}

	// Enqueue command
	engine.EnqueueCommand(cmd)

	// Process commands - should handle error gracefully
	engine.drainCommandQueue()

	// Verify navigation.Jump was NOT called (JSON parse error)
	assert.False(t, nav.jumpCalled, "Navigation.Jump should not have been called with invalid JSON")
}

// TestNavigationIntegration_UnknownCommandType tests error handling for unknown commands
func TestNavigationIntegration_UnknownCommandType(t *testing.T) {
	config := Config{
		TickIntervalMs:        100,
		SnapshotIntervalTicks: 0,
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}

	engine, database, nav, _ := setupTestEngine(t, config)
	defer database.Close()

	// Create test player and ship
	playerID := uuid.New().String()

	_, err := database.Conn().Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES (?, 'TestPlayer', 'hash123', 10000, ?, 0)
	`, playerID, time.Now().Unix())
	require.NoError(t, err)

	// Create command with unknown type
	cmd := Command{
		SessionID:   "session-123",
		PlayerID:    playerID,
		CommandType: "unknown_command",
		Payload:     []byte(`{}`),
		EnqueuedAt:  time.Now().Unix(),
	}

	// Enqueue command
	engine.EnqueueCommand(cmd)

	// Process commands - should handle error gracefully
	engine.drainCommandQueue()

	// Verify no navigation methods were called
	assert.False(t, nav.jumpCalled, "Navigation.Jump should not have been called")
	assert.False(t, nav.dockCalled, "Navigation.Dock should not have been called")
	assert.False(t, nav.undockCalled, "Navigation.Undock should not have been called")
}
