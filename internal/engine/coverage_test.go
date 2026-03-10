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

// TestEnqueueCommandQueueFull tests the default case when command queue is full
func TestEnqueueCommandQueueFull(t *testing.T) {
	config := Config{
		TickIntervalMs:        2000,
		SnapshotIntervalTicks: 100,
		ServerName:            "Black Sector",
		InitialTickNumber:     0,
	}

	engine, database, _, _ := setupTestEngine(t, config)
	defer database.Close()

	// Fill the queue to capacity (1000 commands)
	for i := 0; i < 1000; i++ {
		cmd := Command{
			SessionID:   "session-123",
			PlayerID:    "player-456",
			CommandType: "test_command",
			Payload:     []byte(`{"test": "data"}`),
			EnqueuedAt:  time.Now().Unix(),
		}
		engine.EnqueueCommand(cmd)
	}

	// Try to enqueue one more command - should hit the default case and be dropped
	droppedCmd := Command{
		SessionID:   "session-789",
		PlayerID:    "player-012",
		CommandType: "dropped_command",
		Payload:     []byte(`{"dropped": "true"}`),
		EnqueuedAt:  time.Now().Unix(),
	}
	engine.EnqueueCommand(droppedCmd)

	// Verify queue is still at capacity
	assert.Equal(t, 1000, len(engine.commandQueue))

	// Drain one command
	<-engine.commandQueue

	// Now we should be able to enqueue
	engine.EnqueueCommand(droppedCmd)
	assert.Equal(t, 1000, len(engine.commandQueue))
}

// TestShutdownChan tests the ShutdownChan method
func TestShutdownChan(t *testing.T) {
	config := Config{
		TickIntervalMs:        2000,
		SnapshotIntervalTicks: 100,
		ServerName:            "Black Sector",
		InitialTickNumber:     0,
	}

	engine, database, _, _ := setupTestEngine(t, config)
	defer database.Close()

	// Get shutdown channel
	shutdownChan := engine.ShutdownChan()
	assert.NotNil(t, shutdownChan)

	// Verify it's not closed initially
	select {
	case <-shutdownChan:
		t.Fatal("Shutdown channel should not be closed initially")
	default:
		// Expected - channel is open
	}

	// The shutdown channel is only closed during panic recovery,
	// not during normal shutdown. This test verifies the method exists
	// and returns a valid channel.
}

// TestCreateSnapshotGetAllSessionsError tests error handling in createSnapshot
func TestCreateSnapshotGetAllSessionsError(t *testing.T) {
	logger := zerolog.Nop()

	// Use a closed database to trigger errors
	database, err := db.InitDatabase(":memory:", logger)
	require.NoError(t, err)

	sessionMgr := session.NewSessionManager(database, logger)

	config := Config{
		TickIntervalMs:        2000,
		SnapshotIntervalTicks: 100,
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}

	engine := NewTickEngine(config, database, sessionMgr, &mockNavigator{}, &mockTrader{}, &mockCombatResolver{}, &mockMissionController{}, logger)

	// Close the database to trigger errors
	database.Close()

	// Try to create snapshot - should fail
	snap, err := engine.createSnapshot()
	assert.Error(t, err)
	assert.Nil(t, snap)
	assert.Contains(t, err.Error(), "failed to get")
}

// TestCreateFinalSnapshotError tests error handling in CreateFinalSnapshot
func TestCreateFinalSnapshotError(t *testing.T) {
	logger := zerolog.Nop()

	// Use a closed database to trigger errors
	database, err := db.InitDatabase(":memory:", logger)
	require.NoError(t, err)

	sessionMgr := session.NewSessionManager(database, logger)

	config := Config{
		TickIntervalMs:        2000,
		SnapshotIntervalTicks: 100,
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}

	engine := NewTickEngine(config, database, sessionMgr, &mockNavigator{}, &mockTrader{}, &mockCombatResolver{}, &mockMissionController{}, logger)

	// Close the database to trigger errors
	database.Close()

	// Try to create final snapshot - should fail
	err = engine.CreateFinalSnapshot()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create final snapshot")
}

// TestProcessDockCommandPayloadError tests dock command with invalid payload
func TestProcessDockCommandPayloadError(t *testing.T) {
	config := Config{
		TickIntervalMs:        100,
		SnapshotIntervalTicks: 0,
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}

	engine, database, _, _ := setupTestEngine(t, config)
	defer database.Close()

	// Disable foreign key checks
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

	sessionMgr := session.NewSessionManager(database, zerolog.Nop())
	sess, err := sessionMgr.CreateSession(playerID, "TEXT")
	require.NoError(t, err)

	updateChan, err := sessionMgr.RegisterSessionForUpdates(sess.SessionID)
	require.NoError(t, err)

	engine.sessionMgr = sessionMgr

	// Enqueue dock command with invalid JSON payload
	cmd := Command{
		SessionID:   sess.SessionID,
		PlayerID:    playerID,
		CommandType: "dock",
		Payload:     []byte(`{invalid json`),
		EnqueuedAt:  time.Now().Unix(),
	}

	engine.EnqueueCommand(cmd)
	engine.Start()
	engine.tickNumber = 1
	engine.drainCommandQueue()

	// Should receive error event
	select {
	case update := <-updateChan:
		errorEvent, ok := update.(ErrorEvent)
		require.True(t, ok)
		assert.Equal(t, "dock", errorEvent.CommandType)
		assert.Contains(t, errorEvent.ErrorMsg, "invalid")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected error event")
	}
}

// TestProcessUndockCommandPayloadError tests undock command with missing ship
func TestProcessUndockCommandPayloadError(t *testing.T) {
	config := Config{
		TickIntervalMs:        100,
		SnapshotIntervalTicks: 0,
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}

	engine, database, _, _ := setupTestEngine(t, config)
	defer database.Close()

	// Disable foreign key checks
	_, err := database.Conn().Exec("PRAGMA foreign_keys = OFF")
	require.NoError(t, err)

	// Create test player WITHOUT ship
	playerID := uuid.New().String()

	_, err = database.Conn().Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES (?, 'TestPlayer', 'hash123', 10000, ?, 0)
	`, playerID, time.Now().Unix())
	require.NoError(t, err)

	sessionMgr := session.NewSessionManager(database, zerolog.Nop())
	sess, err := sessionMgr.CreateSession(playerID, "TEXT")
	require.NoError(t, err)

	updateChan, err := sessionMgr.RegisterSessionForUpdates(sess.SessionID)
	require.NoError(t, err)

	engine.sessionMgr = sessionMgr

	// Enqueue undock command
	cmd := Command{
		SessionID:   sess.SessionID,
		PlayerID:    playerID,
		CommandType: "undock",
		Payload:     []byte(`{}`),
		EnqueuedAt:  time.Now().Unix(),
	}

	engine.EnqueueCommand(cmd)
	engine.Start()
	engine.tickNumber = 1
	engine.drainCommandQueue()

	// Should receive error event
	select {
	case update := <-updateChan:
		errorEvent, ok := update.(ErrorEvent)
		require.True(t, ok)
		assert.Equal(t, "undock", errorEvent.CommandType)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected error event")
	}
}

// TestProcessSellCommandPayloadError tests sell command with invalid payload
func TestProcessSellCommandPayloadError(t *testing.T) {
	config := Config{
		TickIntervalMs:        100,
		SnapshotIntervalTicks: 0,
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}

	engine, database, _, _ := setupTestEngine(t, config)
	defer database.Close()

	// Disable foreign key checks
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

	sessionMgr := session.NewSessionManager(database, zerolog.Nop())
	sess, err := sessionMgr.CreateSession(playerID, "TEXT")
	require.NoError(t, err)

	updateChan, err := sessionMgr.RegisterSessionForUpdates(sess.SessionID)
	require.NoError(t, err)

	engine.sessionMgr = sessionMgr

	// Enqueue sell command with invalid JSON payload
	cmd := Command{
		SessionID:   sess.SessionID,
		PlayerID:    playerID,
		CommandType: "sell",
		Payload:     []byte(`{invalid json`),
		EnqueuedAt:  time.Now().Unix(),
	}

	engine.EnqueueCommand(cmd)
	engine.Start()
	engine.tickNumber = 1
	engine.drainCommandQueue()

	// Should receive error event
	select {
	case update := <-updateChan:
		errorEvent, ok := update.(ErrorEvent)
		require.True(t, ok)
		assert.Equal(t, "sell", errorEvent.CommandType)
		assert.Contains(t, errorEvent.ErrorMsg, "invalid")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected error event")
	}
}

// TestProcessSellCommandNavigatorError tests sell command when navigator returns error
func TestProcessSellCommandNavigatorError(t *testing.T) {
	config := Config{
		TickIntervalMs:        100,
		SnapshotIntervalTicks: 0,
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}

	engine, database, _, trader := setupTestEngine(t, config)
	defer database.Close()

	// Configure trader to return error
	trader.sellError = errors.New("port not found")

	// Disable foreign key checks
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

	sessionMgr := session.NewSessionManager(database, zerolog.Nop())
	sess, err := sessionMgr.CreateSession(playerID, "TEXT")
	require.NoError(t, err)

	updateChan, err := sessionMgr.RegisterSessionForUpdates(sess.SessionID)
	require.NoError(t, err)

	engine.sessionMgr = sessionMgr

	// Enqueue sell command
	sellPayload := SellPayload{
		PortID:      1,
		CommodityID: "ore",
		Quantity:    5,
	}
	payloadBytes, _ := json.Marshal(sellPayload)

	cmd := Command{
		SessionID:   sess.SessionID,
		PlayerID:    playerID,
		CommandType: "sell",
		Payload:     payloadBytes,
		EnqueuedAt:  time.Now().Unix(),
	}

	engine.EnqueueCommand(cmd)
	engine.Start()
	engine.tickNumber = 1
	engine.drainCommandQueue()

	// Should receive error event
	select {
	case update := <-updateChan:
		errorEvent, ok := update.(ErrorEvent)
		require.True(t, ok)
		assert.Equal(t, "sell", errorEvent.CommandType)
		assert.Contains(t, errorEvent.ErrorMsg, "port not found")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected error event")
	}
}

// TestBroadcastStateUpdateGetShipError tests broadcastStateUpdate when GetShipByPlayerID fails
func TestBroadcastStateUpdateGetShipError(t *testing.T) {
	logger := zerolog.Nop()

	database, err := db.InitDatabase(":memory:", logger)
	require.NoError(t, err)
	defer database.Close()

	sessionMgr := session.NewSessionManager(database, logger)

	// Create test player WITHOUT ship
	now := time.Now().Unix()
	_, err = database.Conn().Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES ('player-001', 'TestPlayer', 'hash123', 10000, ?, 0)
	`, now)
	require.NoError(t, err)

	// Create session
	sess, err := sessionMgr.CreateSession("player-001", "TEXT")
	require.NoError(t, err)

	// Register session for updates
	updateChan, err := sessionMgr.RegisterSessionForUpdates(sess.SessionID)
	require.NoError(t, err)

	config := Config{
		TickIntervalMs:        2000,
		SnapshotIntervalTicks: 100,
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}

	engine := NewTickEngine(config, database, sessionMgr, &mockNavigator{}, &mockTrader{}, &mockCombatResolver{}, &mockMissionController{}, logger)

	// Broadcast state update - should handle missing ship gracefully
	engine.broadcastStateUpdate()

	// Should still receive update, but with nil ship
	select {
	case update := <-updateChan:
		stateUpdate, ok := update.(StateUpdate)
		require.True(t, ok)
		assert.Equal(t, "player-001", stateUpdate.PlayerState.PlayerID)
		assert.Nil(t, stateUpdate.PlayerState.Ship)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Should receive state update even without ship")
	}
}

// TestRunTickLoopSnapshotError tests snapshot creation error handling
func TestRunTickLoopSnapshotError(t *testing.T) {
	logger := zerolog.Nop()

	database, err := db.InitDatabase(":memory:", logger)
	require.NoError(t, err)

	sessionMgr := session.NewSessionManager(database, logger)

	config := Config{
		TickIntervalMs:        50,
		SnapshotIntervalTicks: 2, // Snapshot every 2 ticks
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}

	engine := NewTickEngine(config, database, sessionMgr, &mockNavigator{}, &mockTrader{}, &mockCombatResolver{}, &mockMissionController{}, logger)

	// Start the engine
	engine.Start()

	// Run tick loop in a goroutine
	go engine.RunTickLoop()

	// Let it run for enough ticks to trigger snapshot
	time.Sleep(200 * time.Millisecond)

	// Close database to cause snapshot errors on next trigger
	database.Close()

	// Let it run a bit more to trigger snapshot with closed DB
	time.Sleep(150 * time.Millisecond)

	// Stop the engine
	engine.Stop()

	// Wait for loop to exit
	time.Sleep(150 * time.Millisecond)

	// Verify engine stopped gracefully despite snapshot errors
	assert.False(t, engine.IsRunning())
	assert.Greater(t, engine.TickNumber(), int64(0))
}

// TestProcessJumpCommandGetShipError tests jump command when ship lookup fails
func TestProcessJumpCommandGetShipError(t *testing.T) {
	config := Config{
		TickIntervalMs:        100,
		SnapshotIntervalTicks: 0,
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}

	engine, database, _, _ := setupTestEngine(t, config)
	defer database.Close()

	// Create test player WITHOUT ship
	playerID := uuid.New().String()

	_, err := database.Conn().Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES (?, 'TestPlayer', 'hash123', 10000, ?, 0)
	`, playerID, time.Now().Unix())
	require.NoError(t, err)

	sessionMgr := session.NewSessionManager(database, zerolog.Nop())
	sess, err := sessionMgr.CreateSession(playerID, "TEXT")
	require.NoError(t, err)

	updateChan, err := sessionMgr.RegisterSessionForUpdates(sess.SessionID)
	require.NoError(t, err)

	engine.sessionMgr = sessionMgr

	// Enqueue jump command
	jumpPayload := JumpPayload{TargetSystemID: 2}
	payloadBytes, _ := json.Marshal(jumpPayload)

	cmd := Command{
		SessionID:   sess.SessionID,
		PlayerID:    playerID,
		CommandType: "jump",
		Payload:     payloadBytes,
		EnqueuedAt:  time.Now().Unix(),
	}

	engine.EnqueueCommand(cmd)
	engine.Start()
	engine.tickNumber = 1
	engine.drainCommandQueue()

	// Should receive error event
	select {
	case update := <-updateChan:
		errorEvent, ok := update.(ErrorEvent)
		require.True(t, ok)
		assert.Equal(t, "jump", errorEvent.CommandType)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected error event")
	}
}

// TestProcessBuyCommandGetShipError tests buy command when ship lookup fails
func TestProcessBuyCommandGetShipError(t *testing.T) {
	config := Config{
		TickIntervalMs:        100,
		SnapshotIntervalTicks: 0,
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}

	engine, database, _, _ := setupTestEngine(t, config)
	defer database.Close()

	// Create test player WITHOUT ship
	playerID := uuid.New().String()

	_, err := database.Conn().Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES (?, 'TestPlayer', 'hash123', 10000, ?, 0)
	`, playerID, time.Now().Unix())
	require.NoError(t, err)

	sessionMgr := session.NewSessionManager(database, zerolog.Nop())
	sess, err := sessionMgr.CreateSession(playerID, "TEXT")
	require.NoError(t, err)

	updateChan, err := sessionMgr.RegisterSessionForUpdates(sess.SessionID)
	require.NoError(t, err)

	engine.sessionMgr = sessionMgr

	// Enqueue buy command
	buyPayload := BuyPayload{
		PortID:      1,
		CommodityID: "ore",
		Quantity:    5,
	}
	payloadBytes, _ := json.Marshal(buyPayload)

	cmd := Command{
		SessionID:   sess.SessionID,
		PlayerID:    playerID,
		CommandType: "buy",
		Payload:     payloadBytes,
		EnqueuedAt:  time.Now().Unix(),
	}

	engine.EnqueueCommand(cmd)
	engine.Start()
	engine.tickNumber = 1
	engine.drainCommandQueue()

	// Should receive error event
	select {
	case update := <-updateChan:
		errorEvent, ok := update.(ErrorEvent)
		require.True(t, ok)
		assert.Equal(t, "buy", errorEvent.CommandType)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected error event")
	}
}
