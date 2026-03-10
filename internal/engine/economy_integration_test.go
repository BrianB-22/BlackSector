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

// TestEconomyIntegration_BuyCommand tests buy command processing
func TestEconomyIntegration_BuyCommand(t *testing.T) {
	config := Config{
		TickIntervalMs:        100,
		SnapshotIntervalTicks: 0,
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}

	engine, database, _, trader := setupTestEngine(t, config)
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

	// Create buy command
	payload := BuyPayload{
		PortID:      1,
		CommodityID: "food",
		Quantity:    10,
	}
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	cmd := Command{
		SessionID:   "session-123",
		PlayerID:    playerID,
		CommandType: "buy",
		Payload:     payloadBytes,
		EnqueuedAt:  time.Now().Unix(),
	}

	// Enqueue command
	engine.EnqueueCommand(cmd)

	// Process commands
	engine.drainCommandQueue()

	// Verify trader.BuyCommodity was called
	assert.True(t, trader.buyCalled, "Trader.BuyCommodity should have been called")
}

// TestEconomyIntegration_SellCommand tests sell command processing
func TestEconomyIntegration_SellCommand(t *testing.T) {
	config := Config{
		TickIntervalMs:        100,
		SnapshotIntervalTicks: 0,
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}

	engine, database, _, trader := setupTestEngine(t, config)
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

	// Create sell command
	payload := SellPayload{
		PortID:      1,
		CommodityID: "food",
		Quantity:    5,
	}
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	cmd := Command{
		SessionID:   "session-123",
		PlayerID:    playerID,
		CommandType: "sell",
		Payload:     payloadBytes,
		EnqueuedAt:  time.Now().Unix(),
	}

	// Enqueue command
	engine.EnqueueCommand(cmd)

	// Process commands
	engine.drainCommandQueue()

	// Verify trader.SellCommodity was called
	assert.True(t, trader.sellCalled, "Trader.SellCommodity should have been called")
}

// TestEconomyIntegration_NoShipFound tests error handling when player has no ship
func TestEconomyIntegration_NoShipFound(t *testing.T) {
	config := Config{
		TickIntervalMs:        100,
		SnapshotIntervalTicks: 0,
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}

	engine, database, _, trader := setupTestEngine(t, config)
	defer database.Close()

	// Create test player without ship
	playerID := uuid.New().String()

	_, err := database.Conn().Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES (?, 'TestPlayer', 'hash123', 10000, ?, 0)
	`, playerID, time.Now().Unix())
	require.NoError(t, err)

	// Create buy command
	payload := BuyPayload{
		PortID:      1,
		CommodityID: "food",
		Quantity:    10,
	}
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	cmd := Command{
		SessionID:   "session-123",
		PlayerID:    playerID,
		CommandType: "buy",
		Payload:     payloadBytes,
		EnqueuedAt:  time.Now().Unix(),
	}

	// Enqueue command
	engine.EnqueueCommand(cmd)

	// Process commands - should handle error gracefully
	engine.drainCommandQueue()

	// Verify trader.BuyCommodity was NOT called (error occurred before)
	assert.False(t, trader.buyCalled, "Trader.BuyCommodity should not have been called when ship not found")
}

// TestEconomyIntegration_TraderError tests error handling from economy subsystem
func TestEconomyIntegration_TraderError(t *testing.T) {
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

	// Create mock trader that returns error
	trader := &mockTrader{
		buyError: errors.New("insufficient credits"),
	}

	sessionMgr := session.NewSessionManager(database, logger)
	engine := NewTickEngine(config, database, sessionMgr, &mockNavigator{}, trader, &mockCombatResolver{}, &mockMissionController{}, logger)

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

	// Create buy command
	payload := BuyPayload{
		PortID:      1,
		CommodityID: "food",
		Quantity:    10,
	}
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	cmd := Command{
		SessionID:   "session-123",
		PlayerID:    playerID,
		CommandType: "buy",
		Payload:     payloadBytes,
		EnqueuedAt:  time.Now().Unix(),
	}

	// Enqueue command
	engine.EnqueueCommand(cmd)

	// Process commands - should handle error gracefully
	engine.drainCommandQueue()

	// Verify trader.BuyCommodity was called (error occurred in trader layer)
	assert.True(t, trader.buyCalled, "Trader.BuyCommodity should have been called")
}

// TestEconomyIntegration_InvalidPayload tests error handling for malformed JSON
func TestEconomyIntegration_InvalidPayload(t *testing.T) {
	config := Config{
		TickIntervalMs:        100,
		SnapshotIntervalTicks: 0,
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}

	engine, database, _, trader := setupTestEngine(t, config)
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

	// Create buy command with invalid JSON
	cmd := Command{
		SessionID:   "session-123",
		PlayerID:    playerID,
		CommandType: "buy",
		Payload:     []byte(`{invalid json`),
		EnqueuedAt:  time.Now().Unix(),
	}

	// Enqueue command
	engine.EnqueueCommand(cmd)

	// Process commands - should handle error gracefully
	engine.drainCommandQueue()

	// Verify trader.BuyCommodity was NOT called (JSON parse error)
	assert.False(t, trader.buyCalled, "Trader.BuyCommodity should not have been called with invalid JSON")
}

// TestEconomyIntegration_BuyAndSellCommands tests both buy and sell in sequence
func TestEconomyIntegration_BuyAndSellCommands(t *testing.T) {
	config := Config{
		TickIntervalMs:        100,
		SnapshotIntervalTicks: 0,
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}

	engine, database, _, trader := setupTestEngine(t, config)
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

	// Create buy command
	buyPayload := BuyPayload{
		PortID:      1,
		CommodityID: "food",
		Quantity:    10,
	}
	buyPayloadBytes, err := json.Marshal(buyPayload)
	require.NoError(t, err)

	buyCmd := Command{
		SessionID:   "session-123",
		PlayerID:    playerID,
		CommandType: "buy",
		Payload:     buyPayloadBytes,
		EnqueuedAt:  time.Now().Unix(),
	}

	// Create sell command
	sellPayload := SellPayload{
		PortID:      1,
		CommodityID: "food",
		Quantity:    5,
	}
	sellPayloadBytes, err := json.Marshal(sellPayload)
	require.NoError(t, err)

	sellCmd := Command{
		SessionID:   "session-123",
		PlayerID:    playerID,
		CommandType: "sell",
		Payload:     sellPayloadBytes,
		EnqueuedAt:  time.Now().Unix() + 1,
	}

	// Enqueue both commands
	engine.EnqueueCommand(buyCmd)
	engine.EnqueueCommand(sellCmd)

	// Process commands
	engine.drainCommandQueue()

	// Verify both methods were called
	assert.True(t, trader.buyCalled, "Trader.BuyCommodity should have been called")
	assert.True(t, trader.sellCalled, "Trader.SellCommodity should have been called")
}
