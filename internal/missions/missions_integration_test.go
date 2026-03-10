package missions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BrianB-22/BlackSector/internal/db"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dbAdapter adapts db.Database to implement the missions.Database interface
type dbAdapter struct {
	db *db.Database
}

func newMissionDBAdapter(database *db.Database) *dbAdapter {
	return &dbAdapter{db: database}
}

// Mission instance operations
func (a *dbAdapter) CreateMissionInstance(instance *MissionInstance) error {
	dbInstance := &db.MissionInstance{
		InstanceID:    instance.InstanceID,
		MissionID:     instance.MissionID,
		PlayerID:      instance.PlayerID,
		Status:        string(instance.Status),
		AcceptedTick:  instance.AcceptedTick,
		StartedTick:   instance.StartedTick,
		CompletedTick: instance.CompletedTick,
		FailedReason:  instance.FailedReason,
		ExpiresAtTick: instance.ExpiresAtTick,
	}
	return a.db.CreateMissionInstance(dbInstance)
}

func (a *dbAdapter) GetMissionInstance(instanceID string) (*MissionInstance, error) {
	dbInstance, err := a.db.GetMissionInstance(instanceID)
	if err != nil || dbInstance == nil {
		return nil, err
	}
	return &MissionInstance{
		InstanceID:    dbInstance.InstanceID,
		MissionID:     dbInstance.MissionID,
		PlayerID:      dbInstance.PlayerID,
		Status:        MissionStatus(dbInstance.Status),
		AcceptedTick:  dbInstance.AcceptedTick,
		StartedTick:   dbInstance.StartedTick,
		CompletedTick: dbInstance.CompletedTick,
		FailedReason:  dbInstance.FailedReason,
		ExpiresAtTick: dbInstance.ExpiresAtTick,
	}, nil
}

func (a *dbAdapter) GetActiveMissionByPlayer(playerID string) (*MissionInstance, error) {
	dbInstance, err := a.db.GetActiveMissionByPlayer(playerID)
	if err != nil || dbInstance == nil {
		return nil, err
	}
	return &MissionInstance{
		InstanceID:    dbInstance.InstanceID,
		MissionID:     dbInstance.MissionID,
		PlayerID:      dbInstance.PlayerID,
		Status:        MissionStatus(dbInstance.Status),
		AcceptedTick:  dbInstance.AcceptedTick,
		StartedTick:   dbInstance.StartedTick,
		CompletedTick: dbInstance.CompletedTick,
		FailedReason:  dbInstance.FailedReason,
		ExpiresAtTick: dbInstance.ExpiresAtTick,
	}, nil
}

func (a *dbAdapter) GetAllInProgressMissions() ([]*MissionInstance, error) {
	dbInstances, err := a.db.GetAllInProgressMissions()
	if err != nil {
		return nil, err
	}
	instances := make([]*MissionInstance, len(dbInstances))
	for i, dbInstance := range dbInstances {
		instances[i] = &MissionInstance{
			InstanceID:    dbInstance.InstanceID,
			MissionID:     dbInstance.MissionID,
			PlayerID:      dbInstance.PlayerID,
			Status:        MissionStatus(dbInstance.Status),
			AcceptedTick:  dbInstance.AcceptedTick,
			StartedTick:   dbInstance.StartedTick,
			CompletedTick: dbInstance.CompletedTick,
			FailedReason:  dbInstance.FailedReason,
			ExpiresAtTick: dbInstance.ExpiresAtTick,
		}
	}
	return instances, nil
}

func (a *dbAdapter) GetCompletedMissionsByPlayer(playerID string) ([]*MissionInstance, error) {
	dbInstances, err := a.db.GetCompletedMissionsByPlayer(playerID)
	if err != nil {
		return nil, err
	}
	instances := make([]*MissionInstance, len(dbInstances))
	for i, dbInstance := range dbInstances {
		instances[i] = &MissionInstance{
			InstanceID:    dbInstance.InstanceID,
			MissionID:     dbInstance.MissionID,
			PlayerID:      dbInstance.PlayerID,
			Status:        MissionStatus(dbInstance.Status),
			AcceptedTick:  dbInstance.AcceptedTick,
			StartedTick:   dbInstance.StartedTick,
			CompletedTick: dbInstance.CompletedTick,
			FailedReason:  dbInstance.FailedReason,
			ExpiresAtTick: dbInstance.ExpiresAtTick,
		}
	}
	return instances, nil
}

func (a *dbAdapter) UpdateMissionStatus(instanceID string, status string, tick int64) error {
	return a.db.UpdateMissionStatus(instanceID, status, tick)
}

func (a *dbAdapter) UpdateMissionObjectiveIndex(instanceID string, objectiveIndex int) error {
	return a.db.UpdateMissionObjectiveIndex(instanceID, objectiveIndex)
}

func (a *dbAdapter) DeleteMissionInstance(instanceID string) error {
	return a.db.DeleteMissionInstance(instanceID)
}

// Objective progress operations
func (a *dbAdapter) CreateObjectiveProgress(progress *ObjectiveProgress) error {
	dbProgress := &db.ObjectiveProgress{
		InstanceID:     progress.InstanceID,
		ObjectiveIndex: progress.ObjectiveIndex,
		Status:         progress.Status,
		CurrentValue:   progress.CurrentValue,
		RequiredValue:  progress.RequiredValue,
	}
	return a.db.CreateObjectiveProgress(dbProgress)
}

func (a *dbAdapter) GetObjectiveProgress(instanceID string, objectiveIndex int) (*ObjectiveProgress, error) {
	dbProgress, err := a.db.GetObjectiveProgress(instanceID, objectiveIndex)
	if err != nil || dbProgress == nil {
		return nil, err
	}
	return &ObjectiveProgress{
		InstanceID:     dbProgress.InstanceID,
		ObjectiveIndex: dbProgress.ObjectiveIndex,
		Status:         dbProgress.Status,
		CurrentValue:   dbProgress.CurrentValue,
		RequiredValue:  dbProgress.RequiredValue,
	}, nil
}

func (a *dbAdapter) GetAllObjectiveProgress(instanceID string) ([]*ObjectiveProgress, error) {
	dbProgressList, err := a.db.GetAllObjectiveProgress(instanceID)
	if err != nil {
		return nil, err
	}
	progressList := make([]*ObjectiveProgress, len(dbProgressList))
	for i, dbProgress := range dbProgressList {
		progressList[i] = &ObjectiveProgress{
			InstanceID:     dbProgress.InstanceID,
			ObjectiveIndex: dbProgress.ObjectiveIndex,
			Status:         dbProgress.Status,
			CurrentValue:   dbProgress.CurrentValue,
			RequiredValue:  dbProgress.RequiredValue,
		}
	}
	return progressList, nil
}

func (a *dbAdapter) UpdateObjectiveProgress(instanceID string, objectiveIndex int, status string, currentValue int) error {
	return a.db.UpdateObjectiveProgress(instanceID, objectiveIndex, status, currentValue)
}

func (a *dbAdapter) DeleteObjectiveProgress(instanceID string) error {
	return a.db.DeleteObjectiveProgress(instanceID)
}

// Player and ship operations
func (a *dbAdapter) GetPlayerByID(playerID string) (*Player, error) {
	dbPlayer, err := a.db.GetPlayerByID(playerID)
	if err != nil || dbPlayer == nil {
		return nil, err
	}
	return &Player{
		PlayerID: dbPlayer.PlayerID,
		Credits:  dbPlayer.Credits,
	}, nil
}

func (a *dbAdapter) GetShipByPlayerID(playerID string) (*Ship, error) {
	dbShip, err := a.db.GetShipByPlayerID(playerID)
	if err != nil || dbShip == nil {
		return nil, err
	}
	return &Ship{
		ShipID:          dbShip.ShipID,
		PlayerID:        dbShip.PlayerID,
		CurrentSystemID: dbShip.CurrentSystemID,
		Status:          dbShip.Status,
		DockedAtPortID:  dbShip.DockedAtPortID,
	}, nil
}

func (a *dbAdapter) UpdatePlayerCredits(playerID string, credits int) error {
	return a.db.UpdatePlayerCredits(playerID, credits)
}

func (a *dbAdapter) GetCargoByShipID(shipID string) ([]*CargoSlot, error) {
	dbCargo, err := a.db.GetCargoByShipID(shipID)
	if err != nil {
		return nil, err
	}
	cargo := make([]*CargoSlot, len(dbCargo))
	for i, dbSlot := range dbCargo {
		cargo[i] = &CargoSlot{
			ShipID:      dbSlot.ShipID,
			CommodityID: dbSlot.CommodityID,
			Quantity:    dbSlot.Quantity,
		}
	}
	return cargo, nil
}

// World queries
func (a *dbAdapter) GetPortByID(portID int) (*Port, error) {
	dbPort, err := a.db.GetPortByID(portID)
	if err != nil || dbPort == nil {
		return nil, err
	}
	return &Port{
		PortID:   dbPort.PortID,
		SystemID: dbPort.SystemID,
	}, nil
}

func (a *dbAdapter) GetSystemSecurityLevel(systemID int) (float64, error) {
	return a.db.GetSystemSecurityLevel(systemID)
}

// setupIntegrationTest creates a real SQLite database and mission system for integration testing
func setupMissionIntegrationTest(t *testing.T) (*db.Database, *MissionManager, func()) {
	// Create temporary directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_missions.db")

	// Create logger
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	// Initialize real SQLite database
	database, err := db.InitDatabase(dbPath, logger)
	require.NoError(t, err, "Failed to initialize test database")

	// Populate required tables
	populateMissionTestData(t, database)

	// Create adapter
	adapter := newMissionDBAdapter(database)

	// Create test mission config directory
	missionDir := filepath.Join(tmpDir, "missions")
	err = os.MkdirAll(missionDir, 0755)
	require.NoError(t, err)

	// Create test mission file
	createTestMissionFile(t, missionDir)

	// Create mission manager with test config
	cfg := DefaultConfig()
	cfg.MissionConfigPath = missionDir
	
	// Use a logger that outputs to test log
	testLogger := zerolog.New(zerolog.NewTestWriter(t)).Level(zerolog.DebugLevel)
	manager := NewMissionManager(cfg, adapter, testLogger)

	// Load missions from test config
	err = manager.LoadMissions(cfg.MissionConfigPath)
	require.NoError(t, err, "Failed to load missions")

	// Cleanup function
	cleanup := func() {
		database.Close()
	}

	return database, manager, cleanup
}

// createTestMissionFile creates a test mission configuration file
func createTestMissionFile(t *testing.T, missionDir string) {
	missionConfig := `{
  "missions": [
    {
      "mission_id": "test_delivery_1",
      "name": "Test Delivery Mission",
      "description": "Deliver cargo to test port",
      "version": "1.0.0",
      "author": "Test",
      "enabled": true,
      "repeatable": true,
      "repeat_cooldown_ticks": 600,
      "security_zones": ["high_security"],
      "expiry_ticks": 800,
      "objectives": [
        {
          "objective_id": "deliver_cargo",
          "type": "deliver_commodity",
          "description": "Deliver 10 units of food_supplies to port 200",
          "parameters": {
            "commodity_id": "food_supplies",
            "quantity": 10,
            "destination_port_id": 200
          }
        }
      ],
      "rewards": {
        "credits": 1500,
        "items": []
      }
    },
    {
      "mission_id": "test_delivery_2",
      "name": "Test Ore Delivery",
      "description": "Deliver ore to test port",
      "version": "1.0.0",
      "author": "Test",
      "enabled": true,
      "repeatable": true,
      "repeat_cooldown_ticks": 700,
      "security_zones": ["high_security"],
      "expiry_ticks": 900,
      "objectives": [
        {
          "objective_id": "deliver_ore",
          "type": "deliver_commodity",
          "description": "Deliver 15 units of raw_ore to port 300",
          "parameters": {
            "commodity_id": "raw_ore",
            "quantity": 15,
            "destination_port_id": 300
          }
        }
      ],
      "rewards": {
        "credits": 2000,
        "items": []
      }
    },
    {
      "mission_id": "test_delivery_3",
      "name": "Test Electronics Rush",
      "description": "Rush delivery of electronics",
      "version": "1.0.0",
      "author": "Test",
      "enabled": true,
      "repeatable": true,
      "repeat_cooldown_ticks": 800,
      "security_zones": ["high_security"],
      "expiry_ticks": 600,
      "objectives": [
        {
          "objective_id": "deliver_electronics",
          "type": "deliver_commodity",
          "description": "Deliver 8 units of electronics to port 200",
          "parameters": {
            "commodity_id": "electronics",
            "quantity": 8,
            "destination_port_id": 200
          }
        }
      ],
      "rewards": {
        "credits": 3500,
        "items": []
      }
    }
  ]
}`

	missionFile := filepath.Join(missionDir, "test_missions.json")
	err := os.WriteFile(missionFile, []byte(missionConfig), 0644)
	require.NoError(t, err)
}

// populateMissionTestData inserts test data into the database
func populateMissionTestData(t *testing.T, database *db.Database) {
	// Insert test region
	_, err := database.Conn().Exec(`
		INSERT INTO regions (region_id, name, region_type, security_level)
		VALUES (1, 'Test Region', 'core', 1.0)
	`)
	require.NoError(t, err)

	// Insert test systems
	systems := []struct {
		systemID      int
		name          string
		securityLevel float64
	}{
		{1, "High Security System", 0.8},
		{2, "Low Security System", 0.2},
	}

	for _, sys := range systems {
		_, err := database.Conn().Exec(`
			INSERT INTO systems (system_id, name, region_id, security_level, position_x, position_y, hazard_level)
			VALUES (?, ?, 1, ?, 0.0, 0.0, 0.0)
		`, sys.systemID, sys.name, sys.securityLevel)
		require.NoError(t, err)
	}

	// Insert test ports
	ports := []struct {
		portID   int
		systemID int
		name     string
	}{
		{100, 1, "Test Port Alpha"},
		{200, 1, "Test Port Beta"},
		{300, 2, "Test Port Gamma"},
	}

	for _, port := range ports {
		_, err := database.Conn().Exec(`
			INSERT INTO ports (port_id, system_id, name, port_type, security_level, docking_fee,
				has_bank, has_shipyard, has_upgrade_market, has_repair, has_fuel)
			VALUES (?, ?, ?, 'trading', 1.0, 0, 0, 0, 0, 1, 1)
		`, port.portID, port.systemID, port.name)
		require.NoError(t, err)
	}

	// Insert test commodities
	commodities := []struct {
		commodityID string
		name        string
		basePrice   int
	}{
		{"food_supplies", "Food Supplies", 100},
		{"raw_ore", "Raw Ore", 80},
		{"electronics", "Electronics", 200},
	}

	for _, commodity := range commodities {
		_, err := database.Conn().Exec(`
			INSERT INTO commodities (commodity_id, name, category, base_price, volatility, is_contraband)
			VALUES (?, ?, 'basic', ?, 0.1, 0)
		`, commodity.commodityID, commodity.name, commodity.basePrice)
		require.NoError(t, err)
	}
}

// insertMissionTestPlayer inserts a test player
func insertMissionTestPlayer(t *testing.T, database *db.Database, playerID, playerName string, credits int64) {
	player := &db.Player{
		PlayerID:   playerID,
		PlayerName: playerName,
		TokenHash:  "test_hash",
		Credits:    credits,
		CreatedAt:  1234567890,
		IsBanned:   false,
	}
	err := database.InsertPlayer(player)
	require.NoError(t, err)
}

// insertMissionTestShip inserts a test ship
func insertMissionTestShip(t *testing.T, database *db.Database, shipID, playerID string, systemID int, dockedAtPortID *int) {
	status := "IN_SPACE"
	if dockedAtPortID != nil {
		status = "DOCKED"
	}

	query := `
		INSERT INTO ships (
			ship_id, player_id, ship_class, hull_points, max_hull_points,
			shield_points, max_shield_points, energy_points, max_energy_points,
			cargo_capacity, missiles_current, current_system_id, position_x, position_y,
			status, docked_at_port_id, last_updated_tick
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var portID interface{}
	if dockedAtPortID != nil {
		portID = *dockedAtPortID
	}

	_, err := database.Conn().Exec(
		query,
		shipID, playerID, "courier", 100, 100,
		50, 50, 100, 100,
		20, 0, systemID, 0.0, 0.0,
		status, portID, int64(100),
	)
	require.NoError(t, err)
}

// insertShipCargo adds cargo to a ship
func insertShipCargo(t *testing.T, database *db.Database, shipID, commodityID string, quantity int) {
	_, err := database.Conn().Exec(`
		INSERT INTO ship_cargo (ship_id, slot_index, commodity_id, quantity)
		VALUES (?, 0, ?, ?)
	`, shipID, commodityID, quantity)
	require.NoError(t, err)
}

// TestIntegration_CompleteMissionLifecycle_Success tests a complete mission from accept to completion
func TestIntegration_CompleteMissionLifecycle_Success(t *testing.T) {
	database, manager, cleanup := setupMissionIntegrationTest(t)
	defer cleanup()

	// Setup: Player with ship docked at port
	playerID := "player-001"
	shipID := "ship-001"
	portID := 100

	insertMissionTestPlayer(t, database, playerID, "TestPlayer", 10000)
	insertMissionTestShip(t, database, shipID, playerID, 1, &portID)

	// Step 1: Get available missions at port
	t.Log("Step 1: Get available missions")
	missions, err := manager.GetAvailableMissions(portID, playerID)
	require.NoError(t, err)
	assert.Greater(t, len(missions), 0, "Should have available missions")

	// Find a simple delivery mission
	var missionID string
	for _, mission := range missions {
		if len(mission.Objectives) == 1 && mission.Objectives[0].Type == "deliver_commodity" {
			missionID = mission.MissionID
			t.Logf("  Selected mission: %s - %s", mission.MissionID, mission.Name)
			break
		}
	}
	require.NotEmpty(t, missionID, "Should find a delivery mission")

	// Step 2: Accept mission
	t.Log("Step 2: Accept mission")
	err = manager.AcceptMission(missionID, playerID, 100)
	require.NoError(t, err, "Mission acceptance should succeed")

	// Verify mission instance created
	activeMission, err := manager.GetActiveMission(playerID)
	require.NoError(t, err)
	require.NotNil(t, activeMission)
	assert.Equal(t, missionID, activeMission.MissionID)
	assert.Equal(t, MissionInProgress, activeMission.Status)

	// Verify objective progress created
	progress, err := database.GetAllObjectiveProgress(activeMission.InstanceID)
	require.NoError(t, err)
	require.Len(t, progress, 1)
	assert.Equal(t, string(ObjectiveActive), progress[0].Status)

	// Step 3: Get mission definition to see requirements
	t.Log("Step 3: Check mission requirements")
	missionDef, err := manager.GetMissionDefinition(missionID)
	require.NoError(t, err)

	objective := missionDef.Objectives[0]
	commodityID := objective.Parameters["commodity_id"].(string)
	quantity := int(objective.Parameters["quantity"].(float64))
	destPortID := int(objective.Parameters["destination_port_id"].(float64))

	t.Logf("  Need to deliver %d units of %s to port %d", quantity, commodityID, destPortID)

	// Step 4: Add required cargo to ship
	t.Log("Step 4: Load cargo")
	insertShipCargo(t, database, shipID, commodityID, quantity)

	// Step 5: Move ship to destination port
	t.Log("Step 5: Travel to destination")
	_, err = database.Conn().Exec(`
		UPDATE ships SET docked_at_port_id = ?, status = 'DOCKED' WHERE ship_id = ?
	`, destPortID, shipID)
	require.NoError(t, err)

	// Step 6: Evaluate objectives (mission system checks progress)
	t.Log("Step 6: Evaluate mission progress")
	events, err := manager.EvaluateObjectives(200)
	require.NoError(t, err)

	// Should have completion event
	require.Len(t, events, 1)
	assert.Equal(t, "completed", events[0].Type)
	assert.Equal(t, playerID, events[0].PlayerID)
	assert.Equal(t, missionID, events[0].MissionID)

	// Save instance ID before checking completion
	instanceID := activeMission.InstanceID

	// Step 7: Verify mission completed
	t.Log("Step 7: Verify mission completion")
	activeMission, err = manager.GetActiveMission(playerID)
	require.NoError(t, err)
	assert.Nil(t, activeMission, "Should have no active mission")

	// Verify mission status updated
	completedMission, err := database.GetMissionInstance(instanceID)
	require.NoError(t, err)
	assert.Equal(t, string(MissionCompleted), completedMission.Status)
	assert.NotNil(t, completedMission.CompletedTick)

	// Step 8: Verify rewards distributed
	t.Log("Step 8: Verify rewards")
	player, err := database.GetPlayerByID(playerID)
	require.NoError(t, err)
	expectedCredits := int64(10000 + missionDef.Rewards.Credits)
	assert.Equal(t, expectedCredits, player.Credits, "Should receive mission reward")

	t.Log("=== Mission lifecycle completed successfully ===")
}

// TestIntegration_MissionExpiry tests mission expiration after time limit
func TestIntegration_MissionExpiry(t *testing.T) {
	database, manager, cleanup := setupMissionIntegrationTest(t)
	defer cleanup()

	// Setup
	playerID := "player-001"
	shipID := "ship-001"
	portID := 100

	insertMissionTestPlayer(t, database, playerID, "TestPlayer", 10000)
	insertMissionTestShip(t, database, shipID, playerID, 1, &portID)

	// Find a mission with expiry
	missions, err := manager.GetAvailableMissions(portID, playerID)
	require.NoError(t, err)

	var missionID string
	for _, mission := range missions {
		if mission.ExpiryTicks != nil {
			missionID = mission.MissionID
			t.Logf("Selected mission with expiry: %s (%d ticks)", mission.Name, *mission.ExpiryTicks)
			break
		}
	}
	require.NotEmpty(t, missionID, "Should find a mission with expiry")

	// Accept mission at tick 100
	err = manager.AcceptMission(missionID, playerID, 100)
	require.NoError(t, err)

	activeMission, err := manager.GetActiveMission(playerID)
	require.NoError(t, err)
	require.NotNil(t, activeMission.ExpiresAtTick)

	expiryTick := *activeMission.ExpiresAtTick
	t.Logf("Mission expires at tick %d", expiryTick)

	// Evaluate before expiry - should not expire
	events, err := manager.EvaluateObjectives(expiryTick - 1)
	require.NoError(t, err)
	assert.Len(t, events, 0, "Mission should not expire yet")

	// Evaluate at expiry tick - should expire
	events, err = manager.EvaluateObjectives(expiryTick)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "expired", events[0].Type)
	assert.Equal(t, playerID, events[0].PlayerID)

	// Verify mission status
	expiredMission, err := database.GetMissionInstance(activeMission.InstanceID)
	require.NoError(t, err)
	assert.Equal(t, string(MissionExpired), expiredMission.Status)

	// Verify no active mission
	activeMission, err = manager.GetActiveMission(playerID)
	require.NoError(t, err)
	assert.Nil(t, activeMission)

	// Verify no rewards given
	player, err := database.GetPlayerByID(playerID)
	require.NoError(t, err)
	assert.Equal(t, int64(10000), player.Credits, "Credits should be unchanged")
}

// TestIntegration_MissionAbandon tests player abandoning an active mission
func TestIntegration_MissionAbandon(t *testing.T) {
	database, manager, cleanup := setupMissionIntegrationTest(t)
	defer cleanup()

	// Setup
	playerID := "player-001"
	shipID := "ship-001"
	portID := 100

	insertMissionTestPlayer(t, database, playerID, "TestPlayer", 10000)
	insertMissionTestShip(t, database, shipID, playerID, 1, &portID)

	// Accept a mission
	missions, err := manager.GetAvailableMissions(portID, playerID)
	require.NoError(t, err)
	require.Greater(t, len(missions), 0)

	missionID := missions[0].MissionID
	err = manager.AcceptMission(missionID, playerID, 100)
	require.NoError(t, err)

	// Verify mission active
	activeMission, err := manager.GetActiveMission(playerID)
	require.NoError(t, err)
	require.NotNil(t, activeMission)

	// Save instance ID for later verification
	instanceID := activeMission.InstanceID

	// Abandon mission
	err = manager.AbandonMission(playerID, 150)
	require.NoError(t, err)

	// Verify no active mission
	activeMission, err = manager.GetActiveMission(playerID)
	require.NoError(t, err)
	assert.Nil(t, activeMission, "Should have no active mission after abandoning")

	// Verify mission status updated
	abandonedMission, err := database.GetMissionInstance(instanceID)
	require.NoError(t, err)
	assert.Equal(t, string(MissionAbandoned), abandonedMission.Status)
	assert.NotNil(t, abandonedMission.CompletedTick)
	assert.Equal(t, int64(150), *abandonedMission.CompletedTick)

	// Verify no rewards given
	player, err := database.GetPlayerByID(playerID)
	require.NoError(t, err)
	assert.Equal(t, int64(10000), player.Credits, "Credits should be unchanged")

	// Verify player can accept new mission
	missions, err = manager.GetAvailableMissions(portID, playerID)
	require.NoError(t, err)
	assert.Greater(t, len(missions), 0, "Should be able to see missions again")
}

// TestIntegration_MissionStatePersistence tests that mission state persists across database reopens
func TestIntegration_MissionStatePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_persistence.db")
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	// Initialize database and mission system
	database, err := db.InitDatabase(dbPath, logger)
	require.NoError(t, err)

	populateMissionTestData(t, database)

	adapter := newMissionDBAdapter(database)

	// Create test mission config directory
	missionDir := filepath.Join(tmpDir, "missions")
	err = os.MkdirAll(missionDir, 0755)
	require.NoError(t, err)

	// Create test mission file
	createTestMissionFile(t, missionDir)

	cfg := DefaultConfig()
	cfg.MissionConfigPath = missionDir
	
	// Use a logger that outputs to test log
	testLogger := zerolog.New(zerolog.NewTestWriter(t)).Level(zerolog.DebugLevel)
	manager := NewMissionManager(cfg, adapter, testLogger)
	err = manager.LoadMissions(cfg.MissionConfigPath)
	require.NoError(t, err)

	// Setup player and accept mission
	playerID := "player-001"
	shipID := "ship-001"
	portID := 100

	insertMissionTestPlayer(t, database, playerID, "TestPlayer", 10000)
	insertMissionTestShip(t, database, shipID, playerID, 1, &portID)

	missions, err := manager.GetAvailableMissions(portID, playerID)
	require.NoError(t, err)
	require.Greater(t, len(missions), 0)

	missionID := missions[0].MissionID
	err = manager.AcceptMission(missionID, playerID, 100)
	require.NoError(t, err)

	// Get mission state before closing
	activeMissionBefore, err := manager.GetActiveMission(playerID)
	require.NoError(t, err)
	require.NotNil(t, activeMissionBefore)

	progressBefore, err := database.GetAllObjectiveProgress(activeMissionBefore.InstanceID)
	require.NoError(t, err)

	// Close database
	database.Close()

	// Reopen database
	database2, err := db.InitDatabase(dbPath, logger)
	require.NoError(t, err)
	defer database2.Close()

	// Verify mission state persisted
	activeMissionAfter, err := database2.GetActiveMissionByPlayer(playerID)
	require.NoError(t, err)
	require.NotNil(t, activeMissionAfter)

	assert.Equal(t, activeMissionBefore.InstanceID, activeMissionAfter.InstanceID)
	assert.Equal(t, activeMissionBefore.MissionID, activeMissionAfter.MissionID)
	assert.Equal(t, activeMissionBefore.PlayerID, activeMissionAfter.PlayerID)
	assert.Equal(t, string(activeMissionBefore.Status), activeMissionAfter.Status)
	assert.Equal(t, activeMissionBefore.AcceptedTick, activeMissionAfter.AcceptedTick)

	// Verify objective progress persisted
	progressAfter, err := database2.GetAllObjectiveProgress(activeMissionAfter.InstanceID)
	require.NoError(t, err)
	require.Len(t, progressAfter, len(progressBefore))

	for i := range progressBefore {
		assert.Equal(t, progressBefore[i].Status, progressAfter[i].Status)
		assert.Equal(t, progressBefore[i].CurrentValue, progressAfter[i].CurrentValue)
		assert.Equal(t, progressBefore[i].RequiredValue, progressAfter[i].RequiredValue)
	}
}

// TestIntegration_OnlyOneActiveMission tests that players can only have one active mission
func TestIntegration_OnlyOneActiveMission(t *testing.T) {
	database, manager, cleanup := setupMissionIntegrationTest(t)
	defer cleanup()

	// Setup
	playerID := "player-001"
	shipID := "ship-001"
	portID := 100

	insertMissionTestPlayer(t, database, playerID, "TestPlayer", 10000)
	insertMissionTestShip(t, database, shipID, playerID, 1, &portID)

	// Get available missions
	missions, err := manager.GetAvailableMissions(portID, playerID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(missions), 2, "Need at least 2 missions for test")

	// Accept first mission
	err = manager.AcceptMission(missions[0].MissionID, playerID, 100)
	require.NoError(t, err)

	// Verify mission active
	activeMission, err := manager.GetActiveMission(playerID)
	require.NoError(t, err)
	require.NotNil(t, activeMission)
	assert.Equal(t, missions[0].MissionID, activeMission.MissionID)

	// Try to accept second mission - should fail
	err = manager.AcceptMission(missions[1].MissionID, playerID, 105)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPlayerHasActiveMission)

	// Verify still only one active mission
	activeMission, err = manager.GetActiveMission(playerID)
	require.NoError(t, err)
	require.NotNil(t, activeMission)
	assert.Equal(t, missions[0].MissionID, activeMission.MissionID)

	// Verify no missions available while one is active
	availableMissions, err := manager.GetAvailableMissions(portID, playerID)
	require.NoError(t, err)
	assert.Len(t, availableMissions, 0, "Should have no available missions while one is active")

	// Abandon first mission
	err = manager.AbandonMission(playerID, 110)
	require.NoError(t, err)

	// Now should be able to accept second mission
	err = manager.AcceptMission(missions[1].MissionID, playerID, 115)
	require.NoError(t, err)

	activeMission, err = manager.GetActiveMission(playerID)
	require.NoError(t, err)
	require.NotNil(t, activeMission)
	assert.Equal(t, missions[1].MissionID, activeMission.MissionID)
}

// TestIntegration_PartialProgress tests mission with incomplete objectives
func TestIntegration_PartialProgress(t *testing.T) {
	database, manager, cleanup := setupMissionIntegrationTest(t)
	defer cleanup()

	// Setup
	playerID := "player-001"
	shipID := "ship-001"
	portID := 100

	insertMissionTestPlayer(t, database, playerID, "TestPlayer", 10000)
	insertMissionTestShip(t, database, shipID, playerID, 1, &portID)

	// Accept a delivery mission
	missions, err := manager.GetAvailableMissions(portID, playerID)
	require.NoError(t, err)

	var missionID string
	var missionDef *MissionDefinition
	for _, mission := range missions {
		if len(mission.Objectives) == 1 && mission.Objectives[0].Type == "deliver_commodity" {
			missionID = mission.MissionID
			missionDef, err = manager.GetMissionDefinition(missionID)
			require.NoError(t, err)
			break
		}
	}
	require.NotEmpty(t, missionID)

	err = manager.AcceptMission(missionID, playerID, 100)
	require.NoError(t, err)

	activeMission, err := manager.GetActiveMission(playerID)
	require.NoError(t, err)

	// Get objective requirements
	objective := missionDef.Objectives[0]
	commodityID := objective.Parameters["commodity_id"].(string)
	requiredQty := int(objective.Parameters["quantity"].(float64))
	destPortID := int(objective.Parameters["destination_port_id"].(float64))

	// Scenario 1: At destination but no cargo
	_, err = database.Conn().Exec(`
		UPDATE ships SET docked_at_port_id = ?, status = 'DOCKED' WHERE ship_id = ?
	`, destPortID, shipID)
	require.NoError(t, err)

	events, err := manager.EvaluateObjectives(150)
	require.NoError(t, err)
	assert.Len(t, events, 0, "Should not complete without cargo")

	// Scenario 2: Has cargo but wrong location
	_, err = database.Conn().Exec(`
		UPDATE ships SET docked_at_port_id = ?, status = 'DOCKED' WHERE ship_id = ?
	`, portID, shipID)
	require.NoError(t, err)

	insertShipCargo(t, database, shipID, commodityID, requiredQty)

	events, err = manager.EvaluateObjectives(160)
	require.NoError(t, err)
	assert.Len(t, events, 0, "Should not complete at wrong location")

	// Scenario 3: Insufficient cargo at destination
	_, err = database.Conn().Exec(`
		UPDATE ships SET docked_at_port_id = ?, status = 'DOCKED' WHERE ship_id = ?
	`, destPortID, shipID)
	require.NoError(t, err)

	_, err = database.Conn().Exec(`
		UPDATE ship_cargo SET quantity = ? WHERE ship_id = ?
	`, requiredQty-1, shipID)
	require.NoError(t, err)

	events, err = manager.EvaluateObjectives(170)
	require.NoError(t, err)
	assert.Len(t, events, 0, "Should not complete with insufficient cargo")

	// Scenario 4: Correct cargo and location
	_, err = database.Conn().Exec(`
		UPDATE ship_cargo SET quantity = ? WHERE ship_id = ?
	`, requiredQty, shipID)
	require.NoError(t, err)

	events, err = manager.EvaluateObjectives(180)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "completed", events[0].Type)

	// Verify mission completed
	activeMission, err = manager.GetActiveMission(playerID)
	require.NoError(t, err)
	assert.Nil(t, activeMission)
}

// TestIntegration_ErrorRecovery tests that failed operations don't corrupt mission state
func TestIntegration_ErrorRecovery(t *testing.T) {
	database, manager, cleanup := setupMissionIntegrationTest(t)
	defer cleanup()

	// Setup
	playerID := "player-001"
	shipID := "ship-001"
	portID := 100

	insertMissionTestPlayer(t, database, playerID, "TestPlayer", 10000)
	insertMissionTestShip(t, database, shipID, playerID, 1, &portID)

	missions, err := manager.GetAvailableMissions(portID, playerID)
	require.NoError(t, err)
	require.Greater(t, len(missions), 0)

	missionID := missions[0].MissionID

	// Try to accept mission for non-existent player
	err = manager.AcceptMission(missionID, "non-existent-player", 100)
	require.Error(t, err)

	// Verify no mission instance created
	noMission, err := manager.GetActiveMission("non-existent-player")
	require.NoError(t, err)
	assert.Nil(t, noMission)

	// Accept mission successfully
	err = manager.AcceptMission(missionID, playerID, 100)
	require.NoError(t, err)

	activeMission, err := manager.GetActiveMission(playerID)
	require.NoError(t, err)
	require.NotNil(t, activeMission)

	// Try to accept same mission again - should fail
	err = manager.AcceptMission(missionID, playerID, 105)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPlayerHasActiveMission)

	// Verify original mission still active and unchanged
	activeMission2, err := manager.GetActiveMission(playerID)
	require.NoError(t, err)
	require.NotNil(t, activeMission2)
	assert.Equal(t, activeMission.InstanceID, activeMission2.InstanceID)
	assert.Equal(t, int64(100), activeMission2.AcceptedTick)

	// Try to abandon mission for wrong player
	err = manager.AbandonMission("wrong-player", 110)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNoActiveMission)

	// Verify mission still active
	activeMission3, err := manager.GetActiveMission(playerID)
	require.NoError(t, err)
	require.NotNil(t, activeMission3)
	assert.Equal(t, MissionInProgress, activeMission3.Status)

	// Abandon successfully
	err = manager.AbandonMission(playerID, 115)
	require.NoError(t, err)

	// Verify mission abandoned
	activeMission4, err := manager.GetActiveMission(playerID)
	require.NoError(t, err)
	assert.Nil(t, activeMission4)
}

// TestIntegration_MultiplePlayers tests multiple players with missions simultaneously
func TestIntegration_MultiplePlayers(t *testing.T) {
	database, manager, cleanup := setupMissionIntegrationTest(t)
	defer cleanup()

	portID := 100
	numPlayers := 3

	// Setup multiple players with ships
	for i := 1; i <= numPlayers; i++ {
		playerID := "player-00" + string(rune('0'+i))
		shipID := "ship-00" + string(rune('0'+i))

		insertMissionTestPlayer(t, database, playerID, "Player"+string(rune('0'+i)), 10000)
		insertMissionTestShip(t, database, shipID, playerID, 1, &portID)
	}

	// Get available missions
	missions, err := manager.GetAvailableMissions(portID, "player-001")
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(missions), numPlayers, "Need enough missions for all players")

	// All players accept different missions
	for i := 1; i <= numPlayers; i++ {
		playerID := "player-00" + string(rune('0'+i))
		missionID := missions[i-1].MissionID

		err := manager.AcceptMission(missionID, playerID, 100)
		require.NoError(t, err, "Player %d should accept mission", i)
	}

	// Verify all players have active missions
	for i := 1; i <= numPlayers; i++ {
		playerID := "player-00" + string(rune('0'+i))

		activeMission, err := manager.GetActiveMission(playerID)
		require.NoError(t, err)
		require.NotNil(t, activeMission, "Player %d should have active mission", i)
		assert.Equal(t, missions[i-1].MissionID, activeMission.MissionID)
	}

	// Verify all missions are in progress
	allMissions, err := database.GetAllInProgressMissions()
	require.NoError(t, err)
	assert.Len(t, allMissions, numPlayers, "Should have %d in-progress missions", numPlayers)

	// Player 1 abandons mission
	err = manager.AbandonMission("player-001", 150)
	require.NoError(t, err)

	// Player 2 completes mission (simulate)
	player2Mission, err := manager.GetActiveMission("player-002")
	require.NoError(t, err)
	missionDef, err := manager.GetMissionDefinition(player2Mission.MissionID)
	require.NoError(t, err)

	// Setup completion conditions for player 2
	objective := missionDef.Objectives[0]
	if objective.Type == "deliver_commodity" {
		commodityID := objective.Parameters["commodity_id"].(string)
		quantity := int(objective.Parameters["quantity"].(float64))
		destPortID := int(objective.Parameters["destination_port_id"].(float64))

		insertShipCargo(t, database, "ship-002", commodityID, quantity)
		_, err = database.Conn().Exec(`
			UPDATE ships SET docked_at_port_id = ?, status = 'DOCKED' WHERE ship_id = ?
		`, destPortID, "ship-002")
		require.NoError(t, err)
	}

	// Evaluate objectives
	events, err := manager.EvaluateObjectives(200)
	require.NoError(t, err)

	// Should have completion event for player 2
	completionFound := false
	for _, event := range events {
		if event.Type == "completed" && event.PlayerID == "player-002" {
			completionFound = true
			break
		}
	}
	assert.True(t, completionFound, "Player 2 should have completed mission")

	// Verify final state
	// Player 1: no active mission (abandoned)
	activeMission1, err := manager.GetActiveMission("player-001")
	require.NoError(t, err)
	assert.Nil(t, activeMission1)

	// Player 2: no active mission (completed)
	activeMission2, err := manager.GetActiveMission("player-002")
	require.NoError(t, err)
	assert.Nil(t, activeMission2)

	// Player 3: still has active mission
	activeMission3, err := manager.GetActiveMission("player-003")
	require.NoError(t, err)
	assert.NotNil(t, activeMission3)
}

// TestIntegration_RewardDistribution tests that rewards are correctly distributed
func TestIntegration_RewardDistribution(t *testing.T) {
	database, manager, cleanup := setupMissionIntegrationTest(t)
	defer cleanup()

	// Setup
	playerID := "player-001"
	shipID := "ship-001"
	portID := 100
	initialCredits := int64(5000)

	insertMissionTestPlayer(t, database, playerID, "TestPlayer", initialCredits)
	insertMissionTestShip(t, database, shipID, playerID, 1, &portID)

	// Find mission with known reward
	missions, err := manager.GetAvailableMissions(portID, playerID)
	require.NoError(t, err)

	var missionID string
	var expectedReward int
	for _, mission := range missions {
		if mission.Rewards != nil && mission.Rewards.Credits > 0 {
			missionID = mission.MissionID
			expectedReward = mission.Rewards.Credits
			t.Logf("Selected mission with %d credit reward", expectedReward)
			break
		}
	}
	require.NotEmpty(t, missionID)

	// Accept mission
	err = manager.AcceptMission(missionID, playerID, 100)
	require.NoError(t, err)

	// Get mission details
	missionDef, err := manager.GetMissionDefinition(missionID)
	require.NoError(t, err)

	// Setup completion conditions
	_, err = manager.GetActiveMission(playerID)
	require.NoError(t, err)

	objective := missionDef.Objectives[0]
	if objective.Type == "deliver_commodity" {
		commodityID := objective.Parameters["commodity_id"].(string)
		quantity := int(objective.Parameters["quantity"].(float64))
		destPortID := int(objective.Parameters["destination_port_id"].(float64))

		insertShipCargo(t, database, shipID, commodityID, quantity)
		_, err = database.Conn().Exec(`
			UPDATE ships SET docked_at_port_id = ?, status = 'DOCKED' WHERE ship_id = ?
		`, destPortID, shipID)
		require.NoError(t, err)
	}

	// Complete mission
	events, err := manager.EvaluateObjectives(200)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "completed", events[0].Type)

	// Verify reward in event
	rewardInEvent, ok := events[0].Details["reward_credits"].(int)
	require.True(t, ok)
	assert.Equal(t, expectedReward, rewardInEvent)

	// Verify credits updated
	player, err := database.GetPlayerByID(playerID)
	require.NoError(t, err)
	expectedCredits := initialCredits + int64(expectedReward)
	assert.Equal(t, expectedCredits, player.Credits, "Should receive mission reward")

	t.Logf("Player credits: %d -> %d (reward: %d)", initialCredits, player.Credits, expectedReward)
}

// TestIntegration_CompleteGameplayScenario tests a realistic mission gameplay scenario
func TestIntegration_CompleteGameplayScenario(t *testing.T) {
	database, manager, cleanup := setupMissionIntegrationTest(t)
	defer cleanup()

	// Scenario: New player accepts first mission, travels, completes, gets reward
	playerID := "player-001"
	shipID := "ship-001"
	startPortID := 100
	initialCredits := int64(1000)

	insertMissionTestPlayer(t, database, playerID, "NewTrader", initialCredits)
	insertMissionTestShip(t, database, shipID, playerID, 1, &startPortID)

	t.Log("=== Scenario Start: New player at starting port ===")

	// Step 1: Check available missions
	t.Log("Step 1: Browse available missions")
	missions, err := manager.GetAvailableMissions(startPortID, playerID)
	require.NoError(t, err)
	t.Logf("  Found %d available missions", len(missions))

	// Find a simple delivery mission
	var selectedMission *MissionListing
	for _, mission := range missions {
		if len(mission.Objectives) == 1 && mission.Objectives[0].Type == "deliver_commodity" {
			selectedMission = mission
			break
		}
	}
	require.NotNil(t, selectedMission, "Should find a delivery mission")
	t.Logf("  Selected: %s - %s", selectedMission.Name, selectedMission.Description)
	t.Logf("  Reward: %d credits", selectedMission.Rewards.Credits)

	// Step 2: Accept mission
	t.Log("Step 2: Accept mission")
	err = manager.AcceptMission(selectedMission.MissionID, playerID, 100)
	require.NoError(t, err)
	t.Log("  Mission accepted!")

	// Verify mission active
	activeMission, err := manager.GetActiveMission(playerID)
	require.NoError(t, err)
	require.NotNil(t, activeMission)

	// Step 3: Check mission requirements
	t.Log("Step 3: Review mission objectives")
	missionDef, err := manager.GetMissionDefinition(selectedMission.MissionID)
	require.NoError(t, err)

	objective := missionDef.Objectives[0]
	commodityID := objective.Parameters["commodity_id"].(string)
	quantity := int(objective.Parameters["quantity"].(float64))
	destPortID := int(objective.Parameters["destination_port_id"].(float64))

	t.Logf("  Objective: Deliver %d units of %s to port %d", quantity, commodityID, destPortID)

	// Step 4: Acquire cargo (simulate buying)
	t.Log("Step 4: Load cargo")
	insertShipCargo(t, database, shipID, commodityID, quantity)
	t.Logf("  Loaded %d units of %s", quantity, commodityID)

	// Step 5: Travel to destination
	t.Log("Step 5: Travel to destination port")
	_, err = database.Conn().Exec(`
		UPDATE ships SET current_system_id = 1, docked_at_port_id = ?, status = 'DOCKED' WHERE ship_id = ?
	`, destPortID, shipID)
	require.NoError(t, err)
	t.Logf("  Arrived at port %d", destPortID)

	// Step 6: Mission system evaluates progress
	t.Log("Step 6: Mission system checks progress")
	events, err := manager.EvaluateObjectives(200)
	require.NoError(t, err)

	// Should complete
	require.Len(t, events, 1)
	assert.Equal(t, "completed", events[0].Type)
	t.Log("  >>> Mission completed! <<<")

	// Step 7: Verify rewards
	t.Log("Step 7: Collect rewards")
	player, err := database.GetPlayerByID(playerID)
	require.NoError(t, err)

	expectedCredits := initialCredits + int64(missionDef.Rewards.Credits)
	assert.Equal(t, expectedCredits, player.Credits)
	t.Logf("  Credits: %d -> %d (+%d)", initialCredits, player.Credits, missionDef.Rewards.Credits)

	// Step 8: Verify can accept new mission
	t.Log("Step 8: Check for new missions")
	newMissions, err := manager.GetAvailableMissions(destPortID, playerID)
	require.NoError(t, err)
	t.Logf("  %d new missions available", len(newMissions))

	t.Log("=== Scenario Complete: Player successfully completed first mission ===")
}
