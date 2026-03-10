package admin

import (
	"bytes"
	"testing"
	"time"

	"github.com/BrianB-22/BlackSector/internal/db"
	"github.com/BrianB-22/BlackSector/internal/engine"
	"github.com/BrianB-22/BlackSector/internal/session"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing
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

// setupTestHandler creates a test handler with in-memory database
func setupTestHandler(t *testing.T) (*Handler, *CLI, *bytes.Buffer, *db.Database) {
	logger := zerolog.Nop()
	
	// Create in-memory database
	database, err := db.InitDatabase(":memory:", logger)
	require.NoError(t, err)
	
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
	
	// Create CLI
	input := bytes.NewReader([]byte{})
	output := &bytes.Buffer{}
	cli := NewCLI(input, output, logger)
	
	// Create handler
	handler := NewHandler(cli, tickEngine, sessionMgr, logger)
	
	return handler, cli, output, database
}

func TestHandler_HandleStatus(t *testing.T) {
	handler, _, output, database := setupTestHandler(t)
	defer database.Close()
	
	// Execute status command
	handler.handleStatus()
	
	// Check output
	result := output.String()
	assert.Contains(t, result, "Status:")
	assert.Contains(t, result, "Tick:")
	assert.Contains(t, result, "Active Sessions:")
}

func TestHandler_HandleSessions_NoSessions(t *testing.T) {
	handler, _, output, database := setupTestHandler(t)
	defer database.Close()
	
	// Execute sessions command with no active sessions
	handler.handleSessions()
	
	// Check output
	result := output.String()
	assert.Contains(t, result, "No active sessions")
}

func TestHandler_HandleSessions_WithSessions(t *testing.T) {
	handler, _, output, database := setupTestHandler(t)
	defer database.Close()
	
	// Create a test player
	player := &db.Player{
		PlayerID:   "player-1",
		PlayerName: "TestPlayer",
		TokenHash:  "test-token-hash",
		Credits:    1000,
		CreatedAt:  time.Now().Unix(),
		IsBanned:   false,
	}
	
	// Insert player into database
	err := database.InsertPlayer(player)
	require.NoError(t, err)
	
	// Create a session
	sess := &db.Session{
		SessionID:      "session-1",
		PlayerID:       player.PlayerID,
		InterfaceMode:  "TEXT",
		State:          db.SessionConnected,
		ConnectedAt:    time.Now().Unix(),
		LastActivityAt: time.Now().Unix(),
	}
	
	err = database.InsertSession(sess)
	require.NoError(t, err)
	
	// Add session to session manager's active sessions
	handler.sessionMgr.CreateSession(player.PlayerID, "TEXT")
	
	// Execute sessions command
	handler.handleSessions()
	
	// Check output
	result := output.String()
	assert.Contains(t, result, "Active Sessions")
	assert.Contains(t, result, "TestPlayer")
	assert.Contains(t, result, player.PlayerID)
}

func TestHandler_HandleShutdown(t *testing.T) {
	handler, _, output, database := setupTestHandler(t)
	defer database.Close()
	
	// Verify tick engine is running
	assert.True(t, handler.tickEngine.IsRunning())
	
	// Execute shutdown command
	handler.handleShutdown()
	
	// Check output
	result := output.String()
	assert.Contains(t, result, "Initiating graceful shutdown")
	
	// Verify tick engine is stopped
	assert.False(t, handler.tickEngine.IsRunning())
}

func TestHandler_HandleCommand(t *testing.T) {
	tests := []struct {
		name           string
		command        Command
		expectedOutput string
	}{
		{
			name:           "status command",
			command:        Command{Type: CommandStatus},
			expectedOutput: "Status:",
		},
		{
			name:           "sessions command",
			command:        Command{Type: CommandSessions},
			expectedOutput: "No active sessions",
		},
		{
			name:           "shutdown command",
			command:        Command{Type: CommandShutdown},
			expectedOutput: "Initiating graceful shutdown",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _, output, database := setupTestHandler(t)
			defer database.Close()
			
			// Execute command
			handler.handleCommand(tt.command)
			
			// Check output
			result := output.String()
			assert.Contains(t, result, tt.expectedOutput)
		})
	}
}
