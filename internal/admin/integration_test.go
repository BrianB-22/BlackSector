package admin

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/BrianB-22/BlackSector/internal/db"
	"github.com/BrianB-22/BlackSector/internal/engine"
	"github.com/BrianB-22/BlackSector/internal/session"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing (if not already defined in handler_test.go)
type mockNavigator struct{}
func (m *mockNavigator) Jump(shipID string, targetSystemID int, currentTick int64) error { return nil }
func (m *mockNavigator) Dock(shipID string, portID int, currentTick int64) error { return nil }
func (m *mockNavigator) Undock(shipID string, currentTick int64) error { return nil }

type mockTrader struct{}
func (m *mockTrader) BuyCommodity(shipID string, portID int, commodityID string, quantity int, tick int64) error { return nil }
func (m *mockTrader) SellCommodity(shipID string, portID int, commodityID string, quantity int, tick int64) error { return nil }

type mockCombatResolver struct{}
func (m *mockCombatResolver) ProcessAttack(combatID string, attackerID string, tick int64) (interface{}, error) { return nil, nil }
func (m *mockCombatResolver) ProcessFlee(combatID string, playerID string, tick int64) (interface{}, error) { return nil, nil }
func (m *mockCombatResolver) ProcessSurrender(combatID string, playerID string, tick int64) error { return nil }
func (m *mockCombatResolver) ResolveCombatTick(tick int64) ([]interface{}, error) { return nil, nil }
func (m *mockCombatResolver) CheckPirateSpawns(tick int64) ([]interface{}, error) { return nil, nil }

type mockMissionController struct{}
func (m *mockMissionController) AcceptMission(missionID string, playerID string, tick int64) error { return nil }
func (m *mockMissionController) AbandonMission(playerID string, tick int64) error { return nil }
func (m *mockMissionController) EvaluateObjectives(tick int64) ([]interface{}, error) { return nil, nil }

// TestAdminCLI_Integration tests the full admin CLI workflow
func TestAdminCLI_Integration(t *testing.T) {
	logger := zerolog.Nop()
	
	// Create in-memory database
	database, err := db.InitDatabase(":memory:", logger)
	require.NoError(t, err)
	defer database.Close()
	
	// Create session manager
	sessionMgr := session.NewSessionManager(database, logger)
	
	// Create tick engine
	engineCfg := engine.Config{
		TickIntervalMs:        2000,
		SnapshotIntervalTicks: 100,
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}
	tickEngine := engine.NewTickEngine(engineCfg, database, sessionMgr, &mockNavigator{}, &mockTrader{}, &mockCombatResolver{}, &mockMissionController{}, logger)
	tickEngine.Start()
	
	// Start tick loop in background
	go tickEngine.RunTickLoop()
	
	// Give tick loop time to start
	time.Sleep(50 * time.Millisecond)
	
	// Create CLI with test input
	input := strings.NewReader("status\nsessions\n")
	output := &bytes.Buffer{}
	cli := NewCLI(input, output, logger)
	cli.Start()
	
	// Create and start handler
	handler := NewHandler(cli, tickEngine, sessionMgr, logger)
	handler.Start()
	
	// Give time for commands to process
	time.Sleep(100 * time.Millisecond)
	
	// Check output contains expected results
	result := output.String()
	assert.Contains(t, result, "Status:")
	assert.Contains(t, result, "Tick:")
	assert.Contains(t, result, "Active Sessions:")
	assert.Contains(t, result, "No active sessions")
	
	// Stop tick engine
	tickEngine.Stop()
	
	// Wait for tick engine to stop
	for tickEngine.IsRunning() {
		time.Sleep(10 * time.Millisecond)
	}
}

// TestAdminCLI_ShutdownCommand tests the shutdown command
func TestAdminCLI_ShutdownCommand(t *testing.T) {
	logger := zerolog.Nop()
	
	// Create in-memory database
	database, err := db.InitDatabase(":memory:", logger)
	require.NoError(t, err)
	defer database.Close()
	
	// Create session manager
	sessionMgr := session.NewSessionManager(database, logger)
	
	// Create tick engine
	engineCfg := engine.Config{
		TickIntervalMs:        2000,
		SnapshotIntervalTicks: 100,
		ServerName:            "Test Server",
		InitialTickNumber:     0,
	}
	tickEngine := engine.NewTickEngine(engineCfg, database, sessionMgr, &mockNavigator{}, &mockTrader{}, &mockCombatResolver{}, &mockMissionController{}, logger)
	tickEngine.Start()
	
	// Start tick loop in background
	go tickEngine.RunTickLoop()
	
	// Give tick loop time to start
	time.Sleep(50 * time.Millisecond)
	
	// Verify tick engine is running
	assert.True(t, tickEngine.IsRunning())
	
	// Create CLI with shutdown command
	input := strings.NewReader("shutdown\n")
	output := &bytes.Buffer{}
	cli := NewCLI(input, output, logger)
	cli.Start()
	
	// Create and start handler
	handler := NewHandler(cli, tickEngine, sessionMgr, logger)
	handler.Start()
	
	// Give time for shutdown command to process
	time.Sleep(100 * time.Millisecond)
	
	// Verify tick engine is stopped
	assert.False(t, tickEngine.IsRunning())
	
	// Check output
	result := output.String()
	assert.Contains(t, result, "Initiating graceful shutdown")
}
