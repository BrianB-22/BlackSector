package navigation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BrianB-22/BlackSector/internal/db"
	"github.com/BrianB-22/BlackSector/internal/world"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dbAdapter adapts db.Database to implement ShipRepository interface
type dbAdapter struct {
	db *db.Database
}

func (a *dbAdapter) GetShipByID(shipID string) (*Ship, error) {
	dbShip, err := a.db.GetShipByID(shipID)
	if err != nil {
		return nil, err
	}
	if dbShip == nil {
		return nil, nil
	}
	
	// Convert db.Ship to navigation.Ship
	return &Ship{
		ShipID:          dbShip.ShipID,
		PlayerID:        dbShip.PlayerID,
		ShipClass:       dbShip.ShipClass,
		HullPoints:      dbShip.HullPoints,
		MaxHullPoints:   dbShip.MaxHullPoints,
		ShieldPoints:    dbShip.ShieldPoints,
		MaxShieldPoints: dbShip.MaxShieldPoints,
		EnergyPoints:    dbShip.EnergyPoints,
		MaxEnergyPoints: dbShip.MaxEnergyPoints,
		CargoCapacity:   dbShip.CargoCapacity,
		MissilesCurrent: dbShip.MissilesCurrent,
		CurrentSystemID: dbShip.CurrentSystemID,
		PositionX:       dbShip.PositionX,
		PositionY:       dbShip.PositionY,
		Status:          ShipStatus(dbShip.Status),
		DockedAtPortID:  dbShip.DockedAtPortID,
		LastUpdatedTick: dbShip.LastUpdatedTick,
	}, nil
}

func (a *dbAdapter) UpdateShipPosition(shipID string, systemID int, tick int64) error {
	return a.db.UpdateShipPosition(shipID, systemID, tick)
}

func (a *dbAdapter) UpdateShipDockStatus(shipID string, status ShipStatus, dockedAtPortID *int, tick int64) error {
	return a.db.UpdateShipDockStatus(shipID, string(status), dockedAtPortID, tick)
}

// setupIntegrationTest creates a real SQLite database and navigation system for integration testing
func setupIntegrationTest(t *testing.T) (*db.Database, *NavigationSystem, func()) {
	// Create temporary directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_navigation.db")

	// Create logger
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	// Initialize real SQLite database
	database, err := db.InitDatabase(dbPath, logger)
	require.NoError(t, err, "Failed to initialize test database")

	// Populate regions table (required for systems foreign key)
	populateTestRegions(t, database)

	// Populate systems table (required for foreign key constraints)
	populateTestSystems(t, database)

	// Populate ports table (required for foreign key constraints)
	populateTestPorts(t, database)

	// Create test universe with connected systems and ports
	universe := createTestUniverseWithPorts()

	// Create adapter for database
	adapter := &dbAdapter{db: database}

	// Create navigation system with database adapter
	nav := NewNavigationSystem(universe, adapter, logger)

	// Cleanup function
	cleanup := func() {
		database.Close()
	}

	return database, nav, cleanup
}

// populateTestRegions inserts test regions into the database
func populateTestRegions(t *testing.T, database *db.Database) {
	_, err := database.Conn().Exec(`
		INSERT INTO regions (region_id, name, region_type, security_level)
		VALUES (1, 'Test Region', 'core', 1.0)
	`)
	require.NoError(t, err, "Failed to insert test region")
}

// populateTestSystems inserts test systems into the database
func populateTestSystems(t *testing.T, database *db.Database) {
	systems := []struct {
		systemID      int
		name          string
		regionID      int
		securityLevel float64
	}{
		{1, "Alpha Station", 1, 2.0},
		{2, "Beta Outpost", 1, 0.8},
		{3, "Gamma Sector", 1, 0.2},
	}

	for _, sys := range systems {
		_, err := database.Conn().Exec(`
			INSERT INTO systems (system_id, name, region_id, security_level, position_x, position_y, hazard_level)
			VALUES (?, ?, ?, ?, 0.0, 0.0, 0.0)
		`, sys.systemID, sys.name, sys.regionID, sys.securityLevel)
		require.NoError(t, err, "Failed to insert test system %d", sys.systemID)
	}
}

// populateTestPorts inserts test ports into the database
func populateTestPorts(t *testing.T, database *db.Database) {
	ports := []struct {
		portID   int
		systemID int
		name     string
		portType string
	}{
		{100, 1, "Alpha Trading Hub", "trading"},
		{200, 2, "Beta Mining Station", "mining"},
		{300, 3, "Gamma Refueling Depot", "refueling"},
	}

	for _, port := range ports {
		_, err := database.Conn().Exec(`
			INSERT INTO ports (port_id, system_id, name, port_type, security_level, docking_fee, 
				has_bank, has_shipyard, has_upgrade_market, has_repair, has_fuel)
			VALUES (?, ?, ?, ?, 1.0, 0, 0, 0, 0, 1, 1)
		`, port.portID, port.systemID, port.name, port.portType)
		require.NoError(t, err, "Failed to insert test port %d", port.portID)
	}
}

// createTestUniverseWithPorts creates a test universe with systems and ports
func createTestUniverseWithPorts() *world.Universe {
	// Create jump connections: 1 <-> 2, 2 <-> 3 (no direct 1 <-> 3)
	connections := []*world.JumpConnection{
		{
			FromSystemID: "1",
			ToSystemID:   "2",
			FuelCost:     10,
		},
		{
			FromSystemID: "2",
			ToSystemID:   "3",
			FuelCost:     15,
		},
	}

	universe := world.NewTestUniverse(connections)

	// Add systems
	universe.Systems = map[string]*world.System{
		"1": {
			SystemID:      "1",
			Name:          "Alpha Station",
			RegionID:      "1",
			SecurityLevel: 2.0,
			PositionX:     0.0,
			PositionY:     0.0,
		},
		"2": {
			SystemID:      "2",
			Name:          "Beta Outpost",
			RegionID:      "1",
			SecurityLevel: 0.8,
			PositionX:     10.0,
			PositionY:     10.0,
		},
		"3": {
			SystemID:      "3",
			Name:          "Gamma Sector",
			RegionID:      "1",
			SecurityLevel: 0.2,
			PositionX:     20.0,
			PositionY:     20.0,
		},
	}

	// Add ports
	universe.Ports = map[string]*world.Port{
		"100": {
			PortID:   "100",
			SystemID: "1",
			Name:     "Alpha Trading Hub",
			PortType: "trading",
		},
		"200": {
			PortID:   "200",
			SystemID: "2",
			Name:     "Beta Mining Station",
			PortType: "mining",
		},
		"300": {
			PortID:   "300",
			SystemID: "3",
			Name:     "Gamma Refueling Depot",
			PortType: "refueling",
		},
	}

	return universe
}

// insertTestPlayer inserts a test player into the database
func insertTestPlayer(t *testing.T, database *db.Database, playerID, playerName string) {
	player := &db.Player{
		PlayerID:   playerID,
		PlayerName: playerName,
		TokenHash:  "test_hash",
		Credits:    10000,
		CreatedAt:  1234567890,
		IsBanned:   false,
	}
	err := database.InsertPlayer(player)
	require.NoError(t, err, "Failed to insert test player")
}

// insertTestShip inserts a test ship into the database
func insertTestShip(t *testing.T, database *db.Database, shipID, playerID string, systemID int, status ShipStatus, dockedAtPortID *int) {
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
		string(status), portID, int64(100),
	)
	require.NoError(t, err, "Failed to insert test ship")
}

// TestIntegration_JumpSequence tests a sequence of jumps between systems
func TestIntegration_JumpSequence(t *testing.T) {
	database, nav, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Insert test player and ship
	insertTestPlayer(t, database, "player-001", "TestPlayer")
	insertTestShip(t, database, "ship-001", "player-001", 1, StatusInSpace, nil)

	// Execute jump sequence: 1 -> 2 -> 3
	// Jump from system 1 to system 2
	err := nav.Jump("ship-001", 2, 105)
	require.NoError(t, err, "First jump should succeed")

	// Verify ship is in system 2
	ship, err := database.GetShipByID("ship-001")
	require.NoError(t, err)
	assert.Equal(t, 2, ship.CurrentSystemID, "Ship should be in system 2")
	assert.Equal(t, int64(105), ship.LastUpdatedTick, "Tick should be updated")

	// Jump from system 2 to system 3
	err = nav.Jump("ship-001", 3, 110)
	require.NoError(t, err, "Second jump should succeed")

	// Verify ship is in system 3
	ship, err = database.GetShipByID("ship-001")
	require.NoError(t, err)
	assert.Equal(t, 3, ship.CurrentSystemID, "Ship should be in system 3")
	assert.Equal(t, int64(110), ship.LastUpdatedTick, "Tick should be updated")

	// Jump back from system 3 to system 2 (bidirectional)
	err = nav.Jump("ship-001", 2, 115)
	require.NoError(t, err, "Return jump should succeed")

	// Verify ship is back in system 2
	ship, err = database.GetShipByID("ship-001")
	require.NoError(t, err)
	assert.Equal(t, 2, ship.CurrentSystemID, "Ship should be back in system 2")
	assert.Equal(t, int64(115), ship.LastUpdatedTick, "Tick should be updated")
}

// TestIntegration_DockUndockCycle tests docking and undocking at a port
func TestIntegration_DockUndockCycle(t *testing.T) {
	database, nav, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Insert test player and ship in system 1
	insertTestPlayer(t, database, "player-001", "TestPlayer")
	insertTestShip(t, database, "ship-001", "player-001", 1, StatusInSpace, nil)

	// Dock at port 100 in system 1
	err := nav.Dock("ship-001", 100, 105)
	require.NoError(t, err, "Dock should succeed")

	// Verify ship is docked
	ship, err := database.GetShipByID("ship-001")
	require.NoError(t, err)
	assert.Equal(t, StatusDocked, ShipStatus(ship.Status), "Ship should be docked")
	require.NotNil(t, ship.DockedAtPortID, "DockedAtPortID should be set")
	assert.Equal(t, 100, *ship.DockedAtPortID, "Ship should be docked at port 100")
	assert.Equal(t, int64(105), ship.LastUpdatedTick, "Tick should be updated")

	// Undock from port
	err = nav.Undock("ship-001", 110)
	require.NoError(t, err, "Undock should succeed")

	// Verify ship is in space
	ship, err = database.GetShipByID("ship-001")
	require.NoError(t, err)
	assert.Equal(t, StatusInSpace, ShipStatus(ship.Status), "Ship should be in space")
	assert.Nil(t, ship.DockedAtPortID, "DockedAtPortID should be cleared")
	assert.Equal(t, int64(110), ship.LastUpdatedTick, "Tick should be updated")
}

// TestIntegration_CompleteNavigationFlow tests a complete navigation workflow
func TestIntegration_CompleteNavigationFlow(t *testing.T) {
	database, nav, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Insert test player and ship
	insertTestPlayer(t, database, "player-001", "TestPlayer")
	insertTestShip(t, database, "ship-001", "player-001", 1, StatusInSpace, nil)

	// Complete flow: Jump -> Dock -> Undock -> Jump
	
	// Step 1: Jump from system 1 to system 2
	err := nav.Jump("ship-001", 2, 100)
	require.NoError(t, err, "Jump to system 2 should succeed")

	ship, err := database.GetShipByID("ship-001")
	require.NoError(t, err)
	assert.Equal(t, 2, ship.CurrentSystemID, "Ship should be in system 2")
	assert.Equal(t, StatusInSpace, ShipStatus(ship.Status), "Ship should be in space")

	// Step 2: Dock at port 200 in system 2
	err = nav.Dock("ship-001", 200, 105)
	require.NoError(t, err, "Dock at port 200 should succeed")

	ship, err = database.GetShipByID("ship-001")
	require.NoError(t, err)
	assert.Equal(t, StatusDocked, ShipStatus(ship.Status), "Ship should be docked")
	require.NotNil(t, ship.DockedAtPortID)
	assert.Equal(t, 200, *ship.DockedAtPortID, "Ship should be docked at port 200")

	// Step 3: Undock from port
	err = nav.Undock("ship-001", 110)
	require.NoError(t, err, "Undock should succeed")

	ship, err = database.GetShipByID("ship-001")
	require.NoError(t, err)
	assert.Equal(t, StatusInSpace, ShipStatus(ship.Status), "Ship should be in space")
	assert.Nil(t, ship.DockedAtPortID, "DockedAtPortID should be cleared")

	// Step 4: Jump from system 2 to system 3
	err = nav.Jump("ship-001", 3, 115)
	require.NoError(t, err, "Jump to system 3 should succeed")

	ship, err = database.GetShipByID("ship-001")
	require.NoError(t, err)
	assert.Equal(t, 3, ship.CurrentSystemID, "Ship should be in system 3")
	assert.Equal(t, StatusInSpace, ShipStatus(ship.Status), "Ship should be in space")
	assert.Equal(t, int64(115), ship.LastUpdatedTick, "Tick should be updated")
}

// TestIntegration_MultipleShips tests navigation with multiple ships
func TestIntegration_MultipleShips(t *testing.T) {
	database, nav, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Insert two players with ships
	insertTestPlayer(t, database, "player-001", "Player1")
	insertTestPlayer(t, database, "player-002", "Player2")
	insertTestShip(t, database, "ship-001", "player-001", 1, StatusInSpace, nil)
	insertTestShip(t, database, "ship-002", "player-002", 2, StatusInSpace, nil)

	// Ship 1 jumps from system 1 to system 2
	err := nav.Jump("ship-001", 2, 100)
	require.NoError(t, err, "Ship 1 jump should succeed")

	// Ship 2 jumps from system 2 to system 3
	err = nav.Jump("ship-002", 3, 100)
	require.NoError(t, err, "Ship 2 jump should succeed")

	// Verify both ships are in correct positions
	ship1, err := database.GetShipByID("ship-001")
	require.NoError(t, err)
	assert.Equal(t, 2, ship1.CurrentSystemID, "Ship 1 should be in system 2")

	ship2, err := database.GetShipByID("ship-002")
	require.NoError(t, err)
	assert.Equal(t, 3, ship2.CurrentSystemID, "Ship 2 should be in system 3")

	// Ship 1 docks at port 200
	err = nav.Dock("ship-001", 200, 105)
	require.NoError(t, err, "Ship 1 dock should succeed")

	// Ship 2 docks at port 300
	err = nav.Dock("ship-002", 300, 105)
	require.NoError(t, err, "Ship 2 dock should succeed")

	// Verify both ships are docked
	ship1, err = database.GetShipByID("ship-001")
	require.NoError(t, err)
	assert.Equal(t, StatusDocked, ShipStatus(ship1.Status), "Ship 1 should be docked")
	require.NotNil(t, ship1.DockedAtPortID)
	assert.Equal(t, 200, *ship1.DockedAtPortID)

	ship2, err = database.GetShipByID("ship-002")
	require.NoError(t, err)
	assert.Equal(t, StatusDocked, ShipStatus(ship2.Status), "Ship 2 should be docked")
	require.NotNil(t, ship2.DockedAtPortID)
	assert.Equal(t, 300, *ship2.DockedAtPortID)
}

// TestIntegration_ErrorRecovery tests error handling and state consistency
func TestIntegration_ErrorRecovery(t *testing.T) {
	database, nav, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Insert test player and ship
	insertTestPlayer(t, database, "player-001", "TestPlayer")
	insertTestShip(t, database, "ship-001", "player-001", 1, StatusInSpace, nil)

	// Attempt invalid jump (no connection from 1 to 3)
	err := nav.Jump("ship-001", 3, 100)
	require.Error(t, err, "Invalid jump should fail")
	assert.ErrorIs(t, err, ErrNoConnection)

	// Verify ship is still in system 1 (state unchanged)
	ship, err := database.GetShipByID("ship-001")
	require.NoError(t, err)
	assert.Equal(t, 1, ship.CurrentSystemID, "Ship should still be in system 1")
	assert.Equal(t, int64(100), ship.LastUpdatedTick, "Tick should not be updated")

	// Perform valid jump
	err = nav.Jump("ship-001", 2, 105)
	require.NoError(t, err, "Valid jump should succeed")

	ship, err = database.GetShipByID("ship-001")
	require.NoError(t, err)
	assert.Equal(t, 2, ship.CurrentSystemID, "Ship should be in system 2")

	// Attempt to dock at port in wrong system
	err = nav.Dock("ship-001", 100, 110)
	require.Error(t, err, "Dock at wrong system should fail")
	assert.ErrorIs(t, err, ErrPortNotInSystem)

	// Verify ship is still in space (state unchanged)
	ship, err = database.GetShipByID("ship-001")
	require.NoError(t, err)
	assert.Equal(t, StatusInSpace, ShipStatus(ship.Status), "Ship should still be in space")
	assert.Nil(t, ship.DockedAtPortID, "DockedAtPortID should still be nil")

	// Dock at correct port
	err = nav.Dock("ship-001", 200, 115)
	require.NoError(t, err, "Valid dock should succeed")

	ship, err = database.GetShipByID("ship-001")
	require.NoError(t, err)
	assert.Equal(t, StatusDocked, ShipStatus(ship.Status), "Ship should be docked")
}

// TestIntegration_DatabasePersistence tests that navigation state persists correctly
func TestIntegration_DatabasePersistence(t *testing.T) {
	// Create temporary directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_persistence.db")
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	// Initialize database and navigation system
	database, err := db.InitDatabase(dbPath, logger)
	require.NoError(t, err)

	// Populate database tables
	populateTestRegions(t, database)
	populateTestSystems(t, database)
	populateTestPorts(t, database)

	universe := createTestUniverseWithPorts()
	adapter := &dbAdapter{db: database}
	nav := NewNavigationSystem(universe, adapter, logger)

	// Insert test player and ship
	insertTestPlayer(t, database, "player-001", "TestPlayer")
	insertTestShip(t, database, "ship-001", "player-001", 1, StatusInSpace, nil)

	// Perform navigation operations
	err = nav.Jump("ship-001", 2, 100)
	require.NoError(t, err)

	err = nav.Dock("ship-001", 200, 105)
	require.NoError(t, err)

	// Close database
	database.Close()

	// Reopen database
	database2, err := db.InitDatabase(dbPath, logger)
	require.NoError(t, err)
	defer database2.Close()

	// Verify ship state persisted
	ship, err := database2.GetShipByID("ship-001")
	require.NoError(t, err)
	assert.Equal(t, 2, ship.CurrentSystemID, "Ship system should persist")
	assert.Equal(t, StatusDocked, ShipStatus(ship.Status), "Ship status should persist")
	require.NotNil(t, ship.DockedAtPortID)
	assert.Equal(t, 200, *ship.DockedAtPortID, "Docked port should persist")
	assert.Equal(t, int64(105), ship.LastUpdatedTick, "Tick should persist")
}

// TestIntegration_ConcurrentOperations tests navigation operations on different ships
func TestIntegration_ConcurrentOperations(t *testing.T) {
	database, nav, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Insert multiple players and ships
	for i := 1; i <= 5; i++ {
		playerID := "player-" + string(rune('0'+i))
		shipID := "ship-" + string(rune('0'+i))
		insertTestPlayer(t, database, playerID, "Player"+string(rune('0'+i)))
		insertTestShip(t, database, shipID, playerID, 1, StatusInSpace, nil)
	}

	// Perform operations on all ships
	for i := 1; i <= 5; i++ {
		shipID := "ship-" + string(rune('0'+i))
		
		// Jump to system 2
		err := nav.Jump(shipID, 2, 100)
		require.NoError(t, err, "Jump should succeed for ship %s", shipID)

		// Dock at port 200
		err = nav.Dock(shipID, 200, 105)
		require.NoError(t, err, "Dock should succeed for ship %s", shipID)
	}

	// Verify all ships are in correct state
	for i := 1; i <= 5; i++ {
		shipID := "ship-" + string(rune('0'+i))
		ship, err := database.GetShipByID(shipID)
		require.NoError(t, err)
		assert.Equal(t, 2, ship.CurrentSystemID, "Ship %s should be in system 2", shipID)
		assert.Equal(t, StatusDocked, ShipStatus(ship.Status), "Ship %s should be docked", shipID)
		require.NotNil(t, ship.DockedAtPortID)
		assert.Equal(t, 200, *ship.DockedAtPortID, "Ship %s should be docked at port 200", shipID)
	}
}

// TestIntegration_JumpDockUndockJump tests the full cycle multiple times
func TestIntegration_JumpDockUndockJump(t *testing.T) {
	database, nav, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Insert test player and ship
	insertTestPlayer(t, database, "player-001", "TestPlayer")
	insertTestShip(t, database, "ship-001", "player-001", 1, StatusInSpace, nil)

	// Cycle 1: System 1 -> 2, dock, undock
	err := nav.Jump("ship-001", 2, 100)
	require.NoError(t, err)

	err = nav.Dock("ship-001", 200, 105)
	require.NoError(t, err)

	err = nav.Undock("ship-001", 110)
	require.NoError(t, err)

	// Cycle 2: System 2 -> 3, dock, undock
	err = nav.Jump("ship-001", 3, 115)
	require.NoError(t, err)

	err = nav.Dock("ship-001", 300, 120)
	require.NoError(t, err)

	err = nav.Undock("ship-001", 125)
	require.NoError(t, err)

	// Cycle 3: System 3 -> 2, dock, undock
	err = nav.Jump("ship-001", 2, 130)
	require.NoError(t, err)

	err = nav.Dock("ship-001", 200, 135)
	require.NoError(t, err)

	err = nav.Undock("ship-001", 140)
	require.NoError(t, err)

	// Verify final state
	ship, err := database.GetShipByID("ship-001")
	require.NoError(t, err)
	assert.Equal(t, 2, ship.CurrentSystemID, "Ship should be in system 2")
	assert.Equal(t, StatusInSpace, ShipStatus(ship.Status), "Ship should be in space")
	assert.Nil(t, ship.DockedAtPortID, "Ship should not be docked")
	assert.Equal(t, int64(140), ship.LastUpdatedTick, "Tick should be updated")
}

// TestIntegration_InvalidOperationSequences tests that invalid sequences are rejected
func TestIntegration_InvalidOperationSequences(t *testing.T) {
	database, nav, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Insert test player and ship
	insertTestPlayer(t, database, "player-001", "TestPlayer")
	insertTestShip(t, database, "ship-001", "player-001", 1, StatusInSpace, nil)

	// Try to undock when not docked
	err := nav.Undock("ship-001", 100)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotDocked)

	// Dock at port
	err = nav.Dock("ship-001", 100, 105)
	require.NoError(t, err)

	// Try to jump while docked
	err = nav.Jump("ship-001", 2, 110)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrShipDocked)

	// Try to dock again while already docked
	err = nav.Dock("ship-001", 100, 115)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrAlreadyDocked)

	// Undock
	err = nav.Undock("ship-001", 120)
	require.NoError(t, err)

	// Now jump should work
	err = nav.Jump("ship-001", 2, 125)
	require.NoError(t, err)

	// Verify final state
	ship, err := database.GetShipByID("ship-001")
	require.NoError(t, err)
	assert.Equal(t, 2, ship.CurrentSystemID, "Ship should be in system 2")
	assert.Equal(t, StatusInSpace, ShipStatus(ship.Status), "Ship should be in space")
}
