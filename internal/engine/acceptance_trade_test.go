package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BrianB-22/BlackSector/internal/db"
	"github.com/BrianB-22/BlackSector/internal/economy"
	"github.com/BrianB-22/BlackSector/internal/navigation"
	"github.com/BrianB-22/BlackSector/internal/session"
	"github.com/BrianB-22/BlackSector/internal/world"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAcceptance_CompleteTradeFlow tests the full trading workflow from player setup
// through buying and selling commodities across multiple systems.
// This is an end-to-end acceptance test using real components (no mocks).
func TestAcceptance_CompleteTradeFlow(t *testing.T) {
	// Setup: Create temporary database and real components
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_trade_flow.db")
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	// Initialize real SQLite database
	database, err := db.InitDatabase(dbPath, logger)
	require.NoError(t, err, "Failed to initialize database")
	defer database.Close()

	// Setup world data: regions, systems, ports, commodities
	setupWorldData(t, database)

	// Create real economy system
	economyCfg := economy.DefaultConfig()
	economySystem := economy.NewEconomySystem(economyCfg, database, logger)
	err = economySystem.LoadCommodities("../../config/world/alpha_sector.json")
	require.NoError(t, err, "Failed to load commodities")

	// Populate commodities table
	populateCommodities(t, database, economySystem)

	// Create test universe for navigation
	universe := createTestUniverse()

	// Create database adapter for navigation
	adapter := &dbAdapter{db: database}

	// Create real navigation system
	navSystem := navigation.NewNavigationSystem(universe, adapter, logger)

	// Create session manager
	sessionMgr := session.NewSessionManager(database, logger)

	// Create tick engine with real components
	engineCfg := Config{
		TickIntervalMs:        100,
		SnapshotIntervalTicks: 0,
		ServerName:            "Test Server",
		InitialTickNumber:     100,
	}
	engine := NewTickEngine(engineCfg, database, sessionMgr, navSystem, economySystem, &mockCombatResolver{}, &mockMissionController{}, logger)

	// Test scenario: Player buys food at Nexus Prime, travels to Vega Prime, sells it
	playerID := uuid.New().String()
	shipID := uuid.New().String()
	sessionID := "test-session-001"

	// Step 1: Create player with starting credits
	t.Log("Step 1: Creating player with 10,000 credits")
	createPlayer(t, database, playerID, "TraderJoe", 10000)

	// Step 2: Create ship docked at Nexus Prime (system_id=1, port_id=1)
	t.Log("Step 2: Creating ship docked at Nexus Prime Starbase")
	createShip(t, database, shipID, playerID, 1, intPtr(1))

	// Step 3: Setup port inventories
	t.Log("Step 3: Setting up port inventories")
	setupPortInventory(t, database, economySystem, 1, 1, "food_supplies", 100)  // Nexus Prime
	setupPortInventory(t, database, economySystem, 5, 4, "food_supplies", 50)   // Vega Prime

	// Verify initial state
	player, err := database.GetPlayerByID(playerID)
	require.NoError(t, err)
	assert.Equal(t, int64(10000), player.Credits, "Player should start with 10,000 credits")

	ship, err := database.GetShipByPlayerID(playerID)
	require.NoError(t, err)
	assert.Equal(t, 1, ship.CurrentSystemID, "Ship should be in system 1 (Nexus Prime)")
	assert.Equal(t, "DOCKED", ship.Status, "Ship should be docked")
	assert.Equal(t, intPtr(1), ship.DockedAtPortID, "Ship should be docked at port 1")

	// Step 4: Buy 10 units of food_supplies at Nexus Prime
	t.Log("Step 4: Buying 10 units of food_supplies at Nexus Prime")
	buyPayload := BuyPayload{
		PortID:      1,
		CommodityID: "food_supplies",
		Quantity:    10,
	}
	buyPayloadBytes, err := json.Marshal(buyPayload)
	require.NoError(t, err)

	buyCmd := Command{
		SessionID:   sessionID,
		PlayerID:    playerID,
		CommandType: "buy",
		Payload:     buyPayloadBytes,
		EnqueuedAt:  time.Now().Unix(),
	}
	engine.EnqueueCommand(buyCmd)
	engine.drainCommandQueue()

	// Verify purchase
	player, err = database.GetPlayerByID(playerID)
	require.NoError(t, err)
	expectedCreditsAfterBuy := int64(10000 - 1100) // 10 units * 110 (buy price at security 2.0)
	assert.Equal(t, expectedCreditsAfterBuy, player.Credits, "Credits should be deducted after purchase")

	cargo, err := database.GetShipCargo(shipID)
	require.NoError(t, err)
	require.Len(t, cargo, 1, "Ship should have 1 cargo item")
	assert.Equal(t, "food_supplies", cargo[0].CommodityID)
	assert.Equal(t, 10, cargo[0].Quantity)
	t.Logf("✓ Purchased 10 food_supplies, credits: %d", player.Credits)

	// Step 5: Undock from Nexus Prime
	t.Log("Step 5: Undocking from Nexus Prime")
	undockCmd := Command{
		SessionID:   sessionID,
		PlayerID:    playerID,
		CommandType: "undock",
		Payload:     []byte(`{}`),
		EnqueuedAt:  time.Now().Unix(),
	}
	engine.EnqueueCommand(undockCmd)
	engine.drainCommandQueue()

	// Verify undocked
	ship, err = database.GetShipByPlayerID(playerID)
	require.NoError(t, err)
	assert.Equal(t, "IN_SPACE", ship.Status, "Ship should be in space after undocking")
	assert.Nil(t, ship.DockedAtPortID, "Ship should not be docked at any port")
	t.Log("✓ Undocked successfully")

	// Step 6: Jump to Gateway Station (system_id=2)
	t.Log("Step 6: Jumping to Gateway Station (system 2)")
	jumpPayload := JumpPayload{TargetSystemID: 2}
	jumpPayloadBytes, err := json.Marshal(jumpPayload)
	require.NoError(t, err)

	jumpCmd := Command{
		SessionID:   sessionID,
		PlayerID:    playerID,
		CommandType: "jump",
		Payload:     jumpPayloadBytes,
		EnqueuedAt:  time.Now().Unix(),
	}
	engine.EnqueueCommand(jumpCmd)
	engine.drainCommandQueue()

	// Verify jump
	ship, err = database.GetShipByPlayerID(playerID)
	require.NoError(t, err)
	assert.Equal(t, 2, ship.CurrentSystemID, "Ship should be in system 2 (Gateway Station)")
	t.Log("✓ Jumped to Gateway Station")

	// Step 7: Jump to Vega Prime (system_id=4)
	t.Log("Step 7: Jumping to Vega Prime (system 4)")
	jumpPayload2 := JumpPayload{TargetSystemID: 4}
	jumpPayloadBytes2, err := json.Marshal(jumpPayload2)
	require.NoError(t, err)

	jumpCmd2 := Command{
		SessionID:   sessionID,
		PlayerID:    playerID,
		CommandType: "jump",
		Payload:     jumpPayloadBytes2,
		EnqueuedAt:  time.Now().Unix(),
	}
	engine.EnqueueCommand(jumpCmd2)
	engine.drainCommandQueue()

	// Verify jump
	ship, err = database.GetShipByPlayerID(playerID)
	require.NoError(t, err)
	assert.Equal(t, 4, ship.CurrentSystemID, "Ship should be in system 4 (Vega Prime)")
	t.Log("✓ Jumped to Vega Prime")

	// Step 8: Dock at Vega Prime Exchange (port_id=5)
	t.Log("Step 8: Docking at Vega Prime Exchange")
	dockPayload := DockPayload{PortID: 5}
	dockPayloadBytes, err := json.Marshal(dockPayload)
	require.NoError(t, err)

	dockCmd := Command{
		SessionID:   sessionID,
		PlayerID:    playerID,
		CommandType: "dock",
		Payload:     dockPayloadBytes,
		EnqueuedAt:  time.Now().Unix(),
	}
	engine.EnqueueCommand(dockCmd)
	engine.drainCommandQueue()

	// Verify docked
	ship, err = database.GetShipByPlayerID(playerID)
	require.NoError(t, err)
	assert.Equal(t, "DOCKED", ship.Status, "Ship should be docked")
	assert.Equal(t, intPtr(5), ship.DockedAtPortID, "Ship should be docked at port 5")
	t.Log("✓ Docked at Vega Prime Exchange")

	// Step 9: Sell 10 units of food_supplies at Vega Prime
	t.Log("Step 9: Selling 10 units of food_supplies at Vega Prime")
	sellPayload := SellPayload{
		PortID:      5,
		CommodityID: "food_supplies",
		Quantity:    10,
	}
	sellPayloadBytes, err := json.Marshal(sellPayload)
	require.NoError(t, err)

	sellCmd := Command{
		SessionID:   sessionID,
		PlayerID:    playerID,
		CommandType: "sell",
		Payload:     sellPayloadBytes,
		EnqueuedAt:  time.Now().Unix(),
	}
	engine.EnqueueCommand(sellCmd)
	engine.drainCommandQueue()

	// Verify sale
	player, err = database.GetPlayerByID(playerID)
	require.NoError(t, err)
	// Sell price at Vega Prime (security 0.8) should be ~90 per unit
	// Total received: 10 * 90 = 900
	expectedCreditsAfterSell := expectedCreditsAfterBuy + 900
	assert.Equal(t, expectedCreditsAfterSell, player.Credits, "Credits should increase after sale")

	cargo, err = database.GetShipCargo(shipID)
	require.NoError(t, err)
	assert.Len(t, cargo, 0, "Ship cargo should be empty after selling all goods")
	t.Logf("✓ Sold 10 food_supplies, final credits: %d", player.Credits)

	// Step 10: Verify port inventories updated correctly
	t.Log("Step 10: Verifying port inventory changes")
	nexusInv, err := database.GetPortInventory(1, "food_supplies")
	require.NoError(t, err)
	assert.Equal(t, 90, nexusInv.Quantity, "Nexus Prime should have 90 units (100 - 10)")

	vegaInv, err := database.GetPortInventory(5, "food_supplies")
	require.NoError(t, err)
	assert.Equal(t, 60, vegaInv.Quantity, "Vega Prime should have 60 units (50 + 10)")

	// Final summary
	netProfit := player.Credits - 10000
	t.Logf("\n=== Trade Flow Complete ===")
	t.Logf("Starting credits: 10,000")
	t.Logf("Final credits: %d", player.Credits)
	t.Logf("Net profit/loss: %d", netProfit)
	t.Logf("Trade route: Nexus Prime → Gateway Station → Vega Prime")
	t.Logf("Commodity: food_supplies (bought 10, sold 10)")

	// Verify net loss due to spread (buy high, sell low in same security zones)
	assert.Less(t, player.Credits, int64(10000), "Player should have net loss due to buy/sell spread")
}

// Helper functions

// dbAdapter adapts db.Database to implement navigation.ShipRepository interface
type dbAdapter struct {
	db *db.Database
}

func (a *dbAdapter) GetShipByID(shipID string) (*navigation.Ship, error) {
	dbShip, err := a.db.GetShipByID(shipID)
	if err != nil {
		return nil, err
	}
	if dbShip == nil {
		return nil, nil
	}

	// Convert db.Ship to navigation.Ship
	return &navigation.Ship{
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
		Status:          navigation.ShipStatus(dbShip.Status),
		DockedAtPortID:  dbShip.DockedAtPortID,
		LastUpdatedTick: dbShip.LastUpdatedTick,
	}, nil
}

func (a *dbAdapter) UpdateShipPosition(shipID string, systemID int, tick int64) error {
	return a.db.UpdateShipPosition(shipID, systemID, tick)
}

func (a *dbAdapter) UpdateShipDockStatus(shipID string, status navigation.ShipStatus, dockedAtPortID *int, tick int64) error {
	return a.db.UpdateShipDockStatus(shipID, string(status), dockedAtPortID, tick)
}

func createTestUniverse() *world.Universe {
	// Create jump connections
	connections := []*world.JumpConnection{
		{FromSystemID: "1", ToSystemID: "2", FuelCost: 5},
		{FromSystemID: "2", ToSystemID: "1", FuelCost: 5},
		{FromSystemID: "2", ToSystemID: "4", FuelCost: 10},
		{FromSystemID: "4", ToSystemID: "2", FuelCost: 10},
	}

	universe := world.NewTestUniverse(connections)

	// Add systems (using string IDs matching database integer IDs)
	universe.Systems = map[string]*world.System{
		"1": {
			SystemID:      "1",
			Name:          "Nexus Prime",
			SecurityLevel: 2.0,
			SecurityZone:  "federated",
		},
		"2": {
			SystemID:      "2",
			Name:          "Gateway Station",
			SecurityLevel: 2.0,
			SecurityZone:  "federated",
		},
		"3": {
			SystemID:      "3",
			Name:          "New Haven",
			SecurityLevel: 0.8,
			SecurityZone:  "high",
		},
		"4": {
			SystemID:      "4",
			Name:          "Vega Prime",
			SecurityLevel: 0.8,
			SecurityZone:  "high",
		},
	}

	// Add ports
	universe.Ports = map[string]*world.Port{
		"1": {
			PortID:   "1",
			SystemID: "1",
			Name:     "Nexus Prime Starbase",
			PortType: "trading",
		},
		"2": {
			PortID:   "2",
			SystemID: "2",
			Name:     "Gateway Trading Post",
			PortType: "trading",
		},
		"3": {
			PortID:   "3",
			SystemID: "3",
			Name:     "New Haven Central Market",
			PortType: "trading",
		},
		"5": {
			PortID:   "5",
			SystemID: "4",
			Name:     "Vega Prime Exchange",
			PortType: "trading",
		},
	}

	return universe
}

func setupWorldData(t *testing.T, database *db.Database) {
	// Insert regions
	_, err := database.Conn().Exec(`
		INSERT INTO regions (region_id, name, region_type, security_level)
		VALUES 
			(1, 'Alpha Sector Core', 'core', 2.0),
			(2, 'Alpha Sector High Security', 'high', 0.8)
	`)
	require.NoError(t, err)

	// Insert systems
	systems := []struct {
		id       int
		name     string
		regionID int
		security float64
	}{
		{1, "Nexus Prime", 1, 2.0},
		{2, "Gateway Station", 1, 2.0},
		{3, "New Haven", 2, 0.8},
		{4, "Vega Prime", 2, 0.8},
	}

	for _, sys := range systems {
		_, err := database.Conn().Exec(`
			INSERT INTO systems (system_id, name, region_id, security_level, position_x, position_y, hazard_level)
			VALUES (?, ?, ?, ?, 0.0, 0.0, 0.0)
		`, sys.id, sys.name, sys.regionID, sys.security)
		require.NoError(t, err)
	}

	// Insert jump connections
	connections := []struct {
		from, to int
	}{
		{1, 2}, // Nexus Prime → Gateway Station
		{2, 1}, // Gateway Station → Nexus Prime (bidirectional)
		{2, 4}, // Gateway Station → Vega Prime
		{4, 2}, // Vega Prime → Gateway Station (bidirectional)
	}

	for _, conn := range connections {
		_, err := database.Conn().Exec(`
			INSERT INTO jump_connections (from_system_id, to_system_id, bidirectional, fuel_cost_modifier)
			VALUES (?, ?, 1, 1.0)
		`, conn.from, conn.to)
		require.NoError(t, err)
	}

	// Insert ports
	ports := []struct {
		id       int
		systemID int
		name     string
		portType string
	}{
		{1, 1, "Nexus Prime Starbase", "trading"},
		{2, 2, "Gateway Trading Post", "trading"},
		{3, 3, "New Haven Market", "trading"},
		{5, 4, "Vega Prime Exchange", "trading"},
	}

	for _, port := range ports {
		_, err := database.Conn().Exec(`
			INSERT INTO ports (port_id, system_id, name, port_type, security_level, docking_fee,
				has_bank, has_shipyard, has_upgrade_market, has_repair, has_fuel)
			VALUES (?, ?, ?, ?, 1.0, 0, 1, 0, 0, 1, 1)
		`, port.id, port.systemID, port.name, port.portType)
		require.NoError(t, err)
	}
}

func populateCommodities(t *testing.T, database *db.Database, economySystem *economy.EconomySystem) {
	commodities := economySystem.GetAllCommodities()
	for _, commodity := range commodities {
		_, err := database.Conn().Exec(`
			INSERT INTO commodities (commodity_id, name, category, base_price, volatility, is_contraband)
			VALUES (?, ?, ?, ?, 0.0, 0)
		`, commodity.CommodityID, commodity.Name, commodity.Category, commodity.BasePrice)
		require.NoError(t, err)
	}
}

func createPlayer(t *testing.T, database *db.Database, playerID, playerName string, credits int64) {
	// Insert directly using SQL to match the actual schema
	_, err := database.Conn().Exec(`
		INSERT INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
		VALUES (?, ?, ?, ?, ?, 0)
	`, playerID, playerName, "test_hash_"+playerID, credits, time.Now().Unix())
	require.NoError(t, err)
}

func createShip(t *testing.T, database *db.Database, shipID, playerID string, systemID int, dockedAtPortID *int) {
	status := "IN_SPACE"
	if dockedAtPortID != nil {
		status = "DOCKED"
	}

	var portID interface{}
	if dockedAtPortID != nil {
		portID = *dockedAtPortID
	}

	_, err := database.Conn().Exec(`
		INSERT INTO ships (
			ship_id, player_id, ship_class, hull_points, max_hull_points,
			shield_points, max_shield_points, energy_points, max_energy_points,
			cargo_capacity, missiles_current, current_system_id, position_x, position_y,
			status, docked_at_port_id, last_updated_tick
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, shipID, playerID, "courier", 100, 100, 50, 50, 100, 100, 20, 0, systemID, 0.0, 0.0, status, portID, int64(100))
	require.NoError(t, err)
}

func setupPortInventory(t *testing.T, database *db.Database, economySystem *economy.EconomySystem, portID, systemID int, commodityID string, quantity int) {
	// Get system security level
	var securityLevel float64
	err := database.Conn().QueryRow("SELECT security_level FROM systems WHERE system_id = ?", systemID).Scan(&securityLevel)
	require.NoError(t, err)

	// Get commodity
	commodity, err := economySystem.GetCommodity(commodityID)
	require.NoError(t, err)

	// Calculate prices
	buyPrice := economySystem.CalculatePrice(commodity.BasePrice, securityLevel, true)
	sellPrice := economySystem.CalculatePrice(commodity.BasePrice, securityLevel, false)

	_, err = database.Conn().Exec(`
		INSERT INTO port_inventory (port_id, commodity_id, quantity, buy_price, sell_price, updated_tick)
		VALUES (?, ?, ?, ?, ?, ?)
	`, portID, commodityID, quantity, buyPrice, sellPrice, int64(0))
	require.NoError(t, err)
}

func intPtr(i int) *int {
	return &i
}
