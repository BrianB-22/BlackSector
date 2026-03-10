package engine

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/BrianB-22/BlackSector/internal/session"
)

// TestErrorHandling_NavigationFailure tests that navigation errors are handled gracefully
func TestErrorHandling_NavigationFailure(t *testing.T) {
	config := Config{
		TickIntervalMs:        100,
		SnapshotIntervalTicks: 0,
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}

	engine, database, nav, _ := setupTestEngine(t, config)
	defer database.Close()

	// Configure navigator to return an error
	nav.jumpError = errors.New("insufficient fuel")

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

	// Create session and register for updates
	sessionMgr := session.NewSessionManager(database, zerolog.Nop())
	sess, err := sessionMgr.CreateSession(playerID, "TEXT")
	require.NoError(t, err)

	updateChan, err := sessionMgr.RegisterSessionForUpdates(sess.SessionID)
	require.NoError(t, err)

	// Replace engine's session manager with our test one
	engine.sessionMgr = sessionMgr

	// Enqueue a jump command
	jumpPayload := JumpPayload{TargetSystemID: 2}
	payloadBytes, err := json.Marshal(jumpPayload)
	require.NoError(t, err)

	cmd := Command{
		SessionID:   sess.SessionID,
		PlayerID:    playerID,
		CommandType: "jump",
		Payload:     payloadBytes,
		EnqueuedAt:  time.Now().Unix(),
	}

	engine.EnqueueCommand(cmd)

	// Process one tick
	engine.Start()
	engine.tickNumber = 1
	engine.drainCommandQueue()

	// Check that an error event was sent to the session
	select {
	case update := <-updateChan:
		errorEvent, ok := update.(ErrorEvent)
		require.True(t, ok, "Expected ErrorEvent, got %T", update)
		assert.Equal(t, "jump", errorEvent.CommandType)
		assert.Contains(t, errorEvent.ErrorMsg, "insufficient fuel")
		assert.Equal(t, int64(1), errorEvent.Tick)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected error event but got timeout")
	}
}

// TestErrorHandling_EconomyFailure tests that economy errors are handled gracefully
func TestErrorHandling_EconomyFailure(t *testing.T) {
	config := Config{
		TickIntervalMs:        100,
		SnapshotIntervalTicks: 0,
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}

	engine, database, _, trader := setupTestEngine(t, config)
	defer database.Close()

	// Configure trader to return an error
	trader.buyError = errors.New("insufficient credits")

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

	// Create session and register for updates
	sessionMgr := session.NewSessionManager(database, zerolog.Nop())
	sess, err := sessionMgr.CreateSession(playerID, "TEXT")
	require.NoError(t, err)

	updateChan, err := sessionMgr.RegisterSessionForUpdates(sess.SessionID)
	require.NoError(t, err)

	// Replace engine's session manager with our test one
	engine.sessionMgr = sessionMgr

	// Enqueue a buy command
	buyPayload := BuyPayload{
		PortID:      1,
		CommodityID: "food_supplies",
		Quantity:    10,
	}
	payloadBytes, err := json.Marshal(buyPayload)
	require.NoError(t, err)

	cmd := Command{
		SessionID:   sess.SessionID,
		PlayerID:    playerID,
		CommandType: "buy",
		Payload:     payloadBytes,
		EnqueuedAt:  time.Now().Unix(),
	}

	engine.EnqueueCommand(cmd)

	// Process one tick
	engine.Start()
	engine.tickNumber = 1
	engine.drainCommandQueue()

	// Check that an error event was sent to the session
	select {
	case update := <-updateChan:
		errorEvent, ok := update.(ErrorEvent)
		require.True(t, ok, "Expected ErrorEvent, got %T", update)
		assert.Equal(t, "buy", errorEvent.CommandType)
		assert.Contains(t, errorEvent.ErrorMsg, "insufficient credits")
		assert.Equal(t, int64(1), errorEvent.Tick)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected error event but got timeout")
	}
}

// TestErrorHandling_MultipleCommandsWithFailures tests that the engine continues processing
// commands even when some fail
func TestErrorHandling_MultipleCommandsWithFailures(t *testing.T) {
	config := Config{
		TickIntervalMs:        100,
		SnapshotIntervalTicks: 0,
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}

	engine, database, nav, trader := setupTestEngine(t, config)
	defer database.Close()

	// Configure subsystems: jump fails, dock succeeds, buy fails, sell succeeds
	nav.jumpError = errors.New("jump failed")
	nav.dockError = nil
	trader.buyError = errors.New("buy failed")
	trader.sellError = nil

	// Disable foreign key checks for test
	_, err := database.Conn().Exec("PRAGMA foreign_keys = OFF")
	require.NoError(t, err)

	// Create test players and ships
	player1ID := uuid.New().String()
	ship1ID := uuid.New().String()
	player2ID := uuid.New().String()
	ship2ID := uuid.New().String()

	_, err = database.Conn().Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES (?, 'Player1', 'hash1', 10000, ?, 0), (?, 'Player2', 'hash2', 10000, ?, 0)
	`, player1ID, time.Now().Unix(), player2ID, time.Now().Unix())
	require.NoError(t, err)

	_, err = database.Conn().Exec(`
		INSERT INTO ships (ship_id, player_id, ship_class, hull_points, max_hull_points,
			shield_points, max_shield_points, energy_points, max_energy_points,
			cargo_capacity, missiles_current, current_system_id, position_x, position_y,
			status, docked_at_port_id, last_updated_tick)
		VALUES 
			(?, ?, 'courier', 100, 100, 50, 50, 100, 100, 20, 0, 1, 0.0, 0.0, 'IN_SPACE', NULL, 0),
			(?, ?, 'courier', 100, 100, 50, 50, 100, 100, 20, 0, 1, 0.0, 0.0, 'IN_SPACE', NULL, 0)
	`, ship1ID, player1ID, ship2ID, player2ID)
	require.NoError(t, err)

	// Create sessions
	sessionMgr := session.NewSessionManager(database, zerolog.Nop())
	sess1, err := sessionMgr.CreateSession(player1ID, "TEXT")
	require.NoError(t, err)
	updateChan1, err := sessionMgr.RegisterSessionForUpdates(sess1.SessionID)
	require.NoError(t, err)

	sess2, err := sessionMgr.CreateSession(player2ID, "TEXT")
	require.NoError(t, err)
	updateChan2, err := sessionMgr.RegisterSessionForUpdates(sess2.SessionID)
	require.NoError(t, err)

	// Replace engine's session manager
	engine.sessionMgr = sessionMgr

	// Enqueue commands: jump (fail), dock (success), buy (fail), sell (success)
	
	// Command 1: Jump (will fail)
	jumpPayload := JumpPayload{TargetSystemID: 2}
	payloadBytes, _ := json.Marshal(jumpPayload)
	engine.EnqueueCommand(Command{
		SessionID:   sess1.SessionID,
		PlayerID:    player1ID,
		CommandType: "jump",
		Payload:     payloadBytes,
		EnqueuedAt:  time.Now().Unix(),
	})

	// Command 2: Dock (will succeed)
	dockPayload := DockPayload{PortID: 1}
	payloadBytes, _ = json.Marshal(dockPayload)
	engine.EnqueueCommand(Command{
		SessionID:   sess2.SessionID,
		PlayerID:    player2ID,
		CommandType: "dock",
		Payload:     payloadBytes,
		EnqueuedAt:  time.Now().Unix(),
	})

	// Command 3: Buy (will fail)
	buyPayload := BuyPayload{PortID: 1, CommodityID: "food", Quantity: 5}
	payloadBytes, _ = json.Marshal(buyPayload)
	engine.EnqueueCommand(Command{
		SessionID:   sess1.SessionID,
		PlayerID:    player1ID,
		CommandType: "buy",
		Payload:     payloadBytes,
		EnqueuedAt:  time.Now().Unix(),
	})

	// Command 4: Sell (will succeed)
	sellPayload := SellPayload{PortID: 1, CommodityID: "ore", Quantity: 3}
	payloadBytes, _ = json.Marshal(sellPayload)
	engine.EnqueueCommand(Command{
		SessionID:   sess2.SessionID,
		PlayerID:    player2ID,
		CommandType: "sell",
		Payload:     payloadBytes,
		EnqueuedAt:  time.Now().Unix(),
	})

	// Process one tick
	engine.Start()
	engine.tickNumber = 1
	engine.drainCommandQueue()

	// Verify that player1 received 2 error events (jump and buy failed)
	errorCount1 := 0
	for i := 0; i < 2; i++ {
		select {
		case update := <-updateChan1:
			if _, ok := update.(ErrorEvent); ok {
				errorCount1++
			}
		case <-time.After(100 * time.Millisecond):
			break
		}
	}
	assert.Equal(t, 2, errorCount1, "Player1 should receive 2 error events")

	// Verify that player2 received no error events (dock and sell succeeded)
	select {
	case update := <-updateChan2:
		if _, ok := update.(ErrorEvent); ok {
			t.Fatal("Player2 should not receive error events")
		}
	case <-time.After(50 * time.Millisecond):
		// Expected - no error events
	}

	// Verify that both subsystems were called (proving the engine didn't stop after first error)
	assert.True(t, nav.jumpCalled, "Jump should have been called")
	assert.True(t, nav.dockCalled, "Dock should have been called")
	assert.True(t, trader.buyCalled, "Buy should have been called")
	assert.True(t, trader.sellCalled, "Sell should have been called")
}

// TestErrorHandling_InvalidCommandStructure tests that invalid commands are rejected
// and the engine continues processing other commands
func TestErrorHandling_InvalidCommandStructure(t *testing.T) {
	config := Config{
		TickIntervalMs:        100,
		SnapshotIntervalTicks: 0,
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}

	engine, database, _, _ := setupTestEngine(t, config)
	defer database.Close()

	// Disable foreign key checks for test
	_, err := database.Conn().Exec("PRAGMA foreign_keys = OFF")
	require.NoError(t, err)

	// Create test player
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

	// Create session and register for updates
	sessionMgr := session.NewSessionManager(database, zerolog.Nop())
	sess, err := sessionMgr.CreateSession(playerID, "TEXT")
	require.NoError(t, err)

	_, err = sessionMgr.RegisterSessionForUpdates(sess.SessionID)
	require.NoError(t, err)

	// Replace engine's session manager
	engine.sessionMgr = sessionMgr

	// Enqueue an invalid command (missing SessionID)
	cmd := Command{
		SessionID:   "", // Invalid - empty
		PlayerID:    playerID,
		CommandType: "jump",
		Payload:     []byte("{}"),
		EnqueuedAt:  time.Now().Unix(),
	}

	engine.EnqueueCommand(cmd)

	// Enqueue a valid command after the invalid one
	jumpPayload := JumpPayload{TargetSystemID: 2}
	payloadBytes, _ := json.Marshal(jumpPayload)
	validCmd := Command{
		SessionID:   sess.SessionID,
		PlayerID:    playerID,
		CommandType: "jump",
		Payload:     payloadBytes,
		EnqueuedAt:  time.Now().Unix(),
	}
	engine.EnqueueCommand(validCmd)

	// Process one tick
	engine.Start()
	engine.tickNumber = 1
	engine.drainCommandQueue()

	// The invalid command should be rejected, but the valid command should still be processed
	// Just verify the engine didn't crash and is still running
	assert.True(t, engine.IsRunning())

	// The valid command should have been processed (even though it might fail in the navigator)
	// We're just testing that the engine continues after an invalid command
}
