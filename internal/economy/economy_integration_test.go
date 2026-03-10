package economy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BrianB-22/BlackSector/internal/db"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupIntegrationTest creates a real SQLite database and economy system for integration testing
func setupIntegrationTest(t *testing.T) (*db.Database, *EconomySystem, func()) {
	// Create temporary directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_economy.db")

	// Create logger
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	// Initialize real SQLite database
	database, err := db.InitDatabase(dbPath, logger)
	require.NoError(t, err, "Failed to initialize test database")

	// Populate required tables
	populateTestRegions(t, database)
	populateTestSystems(t, database)
	populateTestPorts(t, database)

	// Create economy system with default config
	cfg := DefaultConfig()
	economy := NewEconomySystem(cfg, database, logger)

	// Load commodities from test config
	err = economy.LoadCommodities("../../config/world/alpha_sector.json")
	require.NoError(t, err, "Failed to load commodities")

	// Populate commodities table in database (required for foreign key constraints)
	populateTestCommodities(t, database, economy)

	// Cleanup function
	cleanup := func() {
		database.Close()
	}

	return database, economy, cleanup
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
		{1, "Federated Space", 1, 2.0},
		{2, "High Security Zone", 1, 0.8},
		{3, "Low Security Zone", 1, 0.2},
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
		{100, 1, "Fed Space Trading Hub", "trading"},
		{200, 2, "High Sec Mining Station", "mining"},
		{300, 3, "Low Sec Trading Post", "trading"},
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

// populateTestCommodities inserts commodities into the database
func populateTestCommodities(t *testing.T, database *db.Database, economy *EconomySystem) {
	commodities := economy.GetAllCommodities()
	for _, commodity := range commodities {
		_, err := database.Conn().Exec(`
			INSERT INTO commodities (commodity_id, name, category, base_price, volatility, is_contraband)
			VALUES (?, ?, ?, ?, ?, ?)
		`, commodity.CommodityID, commodity.Name, commodity.Category, commodity.BasePrice, 0.0, 0)
		require.NoError(t, err, "Failed to insert commodity %s", commodity.CommodityID)
	}
}

// insertTestPlayer inserts a test player into the database
func insertTestPlayer(t *testing.T, database *db.Database, playerID, playerName string, credits int64) {
	player := &db.Player{
		PlayerID:   playerID,
		PlayerName: playerName,
		TokenHash:  "test_hash",
		Credits:    credits,
		CreatedAt:  1234567890,
		IsBanned:   false,
	}
	err := database.InsertPlayer(player)
	require.NoError(t, err, "Failed to insert test player")
}

// insertTestShip inserts a test ship into the database
func insertTestShip(t *testing.T, database *db.Database, shipID, playerID string, systemID int, dockedAtPortID *int) {
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
	require.NoError(t, err, "Failed to insert test ship")
}

// insertPortInventory inserts port inventory into the database
func insertPortInventory(t *testing.T, database *db.Database, economy *EconomySystem, portID int, systemID int, commodityID string, quantity int) {
	// Get system to determine security level
	system, err := database.Conn().Query("SELECT security_level FROM systems WHERE system_id = ?", systemID)
	require.NoError(t, err)
	require.True(t, system.Next())
	
	var securityLevel float64
	err = system.Scan(&securityLevel)
	require.NoError(t, err)
	system.Close()

	// Get commodity base price
	commodity, err := economy.GetCommodity(commodityID)
	require.NoError(t, err)

	// Calculate prices using economy system
	buyPrice := economy.CalculatePrice(commodity.BasePrice, securityLevel, true)
	sellPrice := economy.CalculatePrice(commodity.BasePrice, securityLevel, false)

	_, err = database.Conn().Exec(`
		INSERT INTO port_inventory (port_id, commodity_id, quantity, buy_price, sell_price, updated_tick)
		VALUES (?, ?, ?, ?, ?, ?)
	`, portID, commodityID, quantity, buyPrice, sellPrice, int64(0))
	require.NoError(t, err, "Failed to insert port inventory")
}

// TestIntegration_BuyAndSellCycle tests a complete buy→sell trade cycle
func TestIntegration_BuyAndSellCycle(t *testing.T) {
	database, economy, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Setup: Player with 10,000 credits, ship docked at Fed Space port
	playerID := "player-001"
	shipID := "ship-001"
	portID := 100

	insertTestPlayer(t, database, playerID, "TestTrader", 10000)
	insertTestShip(t, database, shipID, playerID, 1, &portID)
	insertPortInventory(t, database, economy, portID, 1, "food_supplies", 100)

	// Step 1: Buy 10 food_supplies at Fed Space (buy price = 110)
	err := economy.BuyCommodity(shipID, portID, "food_supplies", 10, 100)
	require.NoError(t, err, "Buy should succeed")

	// Verify credits deducted
	player, err := database.GetPlayerByID(playerID)
	require.NoError(t, err)
	assert.Equal(t, int64(10000-1100), player.Credits, "Credits should be deducted")

	// Verify cargo added
	cargo, err := database.GetShipCargo(shipID)
	require.NoError(t, err)
	require.Len(t, cargo, 1)
	assert.Equal(t, "food_supplies", cargo[0].CommodityID)
	assert.Equal(t, 10, cargo[0].Quantity)

	// Verify port inventory reduced
	portInv, err := database.GetPortInventory(portID, "food_supplies")
	require.NoError(t, err)
	assert.Equal(t, 90, portInv.Quantity)

	// Step 2: Sell 10 food_supplies back at Fed Space (sell price = 90)
	err = economy.SellCommodity(shipID, portID, "food_supplies", 10, 105)
	require.NoError(t, err, "Sell should succeed")

	// Verify credits added
	player, err = database.GetPlayerByID(playerID)
	require.NoError(t, err)
	assert.Equal(t, int64(10000-1100+900), player.Credits, "Credits should reflect buy-sell cycle")

	// Verify cargo removed
	cargo, err = database.GetShipCargo(shipID)
	require.NoError(t, err)
	assert.Len(t, cargo, 0, "Cargo should be empty")

	// Verify port inventory restored
	portInv, err = database.GetPortInventory(portID, "food_supplies")
	require.NoError(t, err)
	assert.Equal(t, 100, portInv.Quantity, "Port inventory should be restored")

	// Net result: Player lost 200 credits (10 * (110 - 90))
	assert.Equal(t, int64(9800), player.Credits, "Player should have lost 200 credits in spread")
}

// TestIntegration_ProfitableTradeRun tests buying low and selling high across zones
func TestIntegration_ProfitableTradeRun(t *testing.T) {
	database, economy, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Setup: Player with 10,000 credits
	playerID := "player-001"
	shipID := "ship-001"
	fedPortID := 100
	lowSecPortID := 300

	insertTestPlayer(t, database, playerID, "TestTrader", 10000)
	
	// Start docked at Fed Space
	insertTestShip(t, database, shipID, playerID, 1, &fedPortID)
	insertPortInventory(t, database, economy, fedPortID, 1, "food_supplies", 100)
	insertPortInventory(t, database, economy, lowSecPortID, 3, "food_supplies", 50)

	// Step 1: Buy 15 food_supplies at Fed Space (buy price = 110)
	err := economy.BuyCommodity(shipID, fedPortID, "food_supplies", 15, 100)
	require.NoError(t, err)

	player, err := database.GetPlayerByID(playerID)
	require.NoError(t, err)
	assert.Equal(t, int64(10000-1650), player.Credits, "Should spend 1650 credits")

	// Step 2: Simulate travel to Low Sec (update ship location and dock status)
	_, err = database.Conn().Exec(`
		UPDATE ships SET current_system_id = ?, docked_at_port_id = ?, last_updated_tick = ?
		WHERE ship_id = ?
	`, 3, lowSecPortID, int64(110), shipID)
	require.NoError(t, err)

	// Step 3: Sell 15 food_supplies at Low Sec (sell price = 106)
	err = economy.SellCommodity(shipID, lowSecPortID, "food_supplies", 15, 115)
	require.NoError(t, err)

	// Verify profit
	player, err = database.GetPlayerByID(playerID)
	require.NoError(t, err)
	expectedCredits := int64(10000 - 1650 + 1590) // Buy at 110, sell at 106
	assert.Equal(t, expectedCredits, player.Credits)

	// Net result: Lost 60 credits (15 * (110 - 106))
	// This demonstrates the spread makes it hard to profit without zone arbitrage
	assert.Equal(t, int64(9940), player.Credits)
}

// TestIntegration_MultipleTradesWithDifferentCommodities tests trading multiple commodities
func TestIntegration_MultipleTradesWithDifferentCommodities(t *testing.T) {
	database, economy, cleanup := setupIntegrationTest(t)
	defer cleanup()

	playerID := "player-001"
	shipID := "ship-001"
	portID := 100

	insertTestPlayer(t, database, playerID, "TestTrader", 20000)
	insertTestShip(t, database, shipID, playerID, 1, &portID)
	insertPortInventory(t, database, economy, portID, 1, "food_supplies", 100)
	insertPortInventory(t, database, economy, portID, 1, "fuel_cells", 100)
	insertPortInventory(t, database, economy, portID, 1, "electronics", 50)

	// Buy multiple commodities
	err := economy.BuyCommodity(shipID, portID, "food_supplies", 5, 100)
	require.NoError(t, err)

	err = economy.BuyCommodity(shipID, portID, "fuel_cells", 5, 105)
	require.NoError(t, err)

	err = economy.BuyCommodity(shipID, portID, "electronics", 3, 110)
	require.NoError(t, err)

	// Verify cargo has 3 different commodities
	cargo, err := database.GetShipCargo(shipID)
	require.NoError(t, err)
	assert.Len(t, cargo, 3, "Should have 3 different commodities")

	// Verify total cargo quantity
	totalQty := 0
	for _, slot := range cargo {
		totalQty += slot.Quantity
	}
	assert.Equal(t, 13, totalQty, "Total cargo should be 13 units")

	// Sell one commodity
	err = economy.SellCommodity(shipID, portID, "fuel_cells", 5, 115)
	require.NoError(t, err)

	// Verify cargo now has 2 commodities
	cargo, err = database.GetShipCargo(shipID)
	require.NoError(t, err)
	assert.Len(t, cargo, 2, "Should have 2 commodities after selling one")

	// Verify correct commodities remain
	commodityIDs := make(map[string]bool)
	for _, slot := range cargo {
		commodityIDs[slot.CommodityID] = true
	}
	assert.True(t, commodityIDs["food_supplies"])
	assert.True(t, commodityIDs["electronics"])
	assert.False(t, commodityIDs["fuel_cells"])
}

// TestIntegration_CargoCapacityEnforcement tests that cargo capacity is enforced across multiple trades
func TestIntegration_CargoCapacityEnforcement(t *testing.T) {
	database, economy, cleanup := setupIntegrationTest(t)
	defer cleanup()

	playerID := "player-001"
	shipID := "ship-001"
	portID := 100

	insertTestPlayer(t, database, playerID, "TestTrader", 50000)
	insertTestShip(t, database, shipID, playerID, 1, &portID) // Courier has 20 cargo capacity
	insertPortInventory(t, database, economy, portID, 1, "food_supplies", 100)

	// Fill cargo to capacity
	err := economy.BuyCommodity(shipID, portID, "food_supplies", 20, 100)
	require.NoError(t, err)

	// Attempt to buy more should fail
	err = economy.BuyCommodity(shipID, portID, "food_supplies", 1, 105)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInsufficientCargo)

	// Sell some cargo
	err = economy.SellCommodity(shipID, portID, "food_supplies", 10, 110)
	require.NoError(t, err)

	// Now should be able to buy again
	err = economy.BuyCommodity(shipID, portID, "food_supplies", 5, 115)
	require.NoError(t, err)

	// Verify final cargo
	cargo, err := database.GetShipCargo(shipID)
	require.NoError(t, err)
	require.Len(t, cargo, 1)
	assert.Equal(t, 15, cargo[0].Quantity, "Should have 15 units (20 - 10 + 5)")
}

// TestIntegration_ConcurrentTrades tests multiple players trading simultaneously
func TestIntegration_ConcurrentTrades(t *testing.T) {
	database, economy, cleanup := setupIntegrationTest(t)
	defer cleanup()

	portID := 100

	// Setup 3 players with ships
	for i := 1; i <= 3; i++ {
		playerID := "player-00" + string(rune('0'+i))
		shipID := "ship-00" + string(rune('0'+i))
		
		insertTestPlayer(t, database, playerID, "Player"+string(rune('0'+i)), 10000)
		insertTestShip(t, database, shipID, playerID, 1, &portID)
	}

	insertPortInventory(t, database, economy, portID, 1, "food_supplies", 100)

	// All players buy commodities
	for i := 1; i <= 3; i++ {
		shipID := "ship-00" + string(rune('0'+i))
		err := economy.BuyCommodity(shipID, portID, "food_supplies", 5, 100)
		require.NoError(t, err, "Player %d buy should succeed", i)
	}

	// Verify port inventory reduced correctly
	portInv, err := database.GetPortInventory(portID, "food_supplies")
	require.NoError(t, err)
	assert.Equal(t, 85, portInv.Quantity, "Port should have 85 units left (100 - 15)")

	// Verify each player has cargo
	for i := 1; i <= 3; i++ {
		shipID := "ship-00" + string(rune('0'+i))
		cargo, err := database.GetShipCargo(shipID)
		require.NoError(t, err)
		require.Len(t, cargo, 1)
		assert.Equal(t, 5, cargo[0].Quantity)
	}

	// All players sell back
	for i := 1; i <= 3; i++ {
		shipID := "ship-00" + string(rune('0'+i))
		err := economy.SellCommodity(shipID, portID, "food_supplies", 5, 105)
		require.NoError(t, err, "Player %d sell should succeed", i)
	}

	// Verify port inventory restored
	portInv, err = database.GetPortInventory(portID, "food_supplies")
	require.NoError(t, err)
	assert.Equal(t, 100, portInv.Quantity, "Port inventory should be restored")
}

// TestIntegration_ErrorRecovery tests that failed transactions don't corrupt state
func TestIntegration_ErrorRecovery(t *testing.T) {
	database, economy, cleanup := setupIntegrationTest(t)
	defer cleanup()

	playerID := "player-001"
	shipID := "ship-001"
	portID := 100

	insertTestPlayer(t, database, playerID, "TestTrader", 1000)
	insertTestShip(t, database, shipID, playerID, 1, &portID)
	insertPortInventory(t, database, economy, portID, 1, "food_supplies", 100)

	// Attempt to buy more than player can afford
	err := economy.BuyCommodity(shipID, portID, "food_supplies", 20, 100)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInsufficientCredits)

	// Verify state unchanged
	player, err := database.GetPlayerByID(playerID)
	require.NoError(t, err)
	assert.Equal(t, int64(1000), player.Credits, "Credits should be unchanged")

	cargo, err := database.GetShipCargo(shipID)
	require.NoError(t, err)
	assert.Len(t, cargo, 0, "Cargo should be empty")

	portInv, err := database.GetPortInventory(portID, "food_supplies")
	require.NoError(t, err)
	assert.Equal(t, 100, portInv.Quantity, "Port inventory should be unchanged")

	// Now perform valid trade
	err = economy.BuyCommodity(shipID, portID, "food_supplies", 5, 105)
	require.NoError(t, err, "Valid trade should succeed after failed attempt")

	player, err = database.GetPlayerByID(playerID)
	require.NoError(t, err)
	assert.Equal(t, int64(1000-550), player.Credits)
}

// TestIntegration_DatabasePersistence tests that trade state persists across database reopens
func TestIntegration_DatabasePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_persistence.db")
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	// Initialize database and economy
	database, err := db.InitDatabase(dbPath, logger)
	require.NoError(t, err)

	populateTestRegions(t, database)
	populateTestSystems(t, database)
	populateTestPorts(t, database)

	cfg := DefaultConfig()
	economy := NewEconomySystem(cfg, database, logger)
	err = economy.LoadCommodities("../../config/world/alpha_sector.json")
	require.NoError(t, err)

	// Populate commodities table
	populateTestCommodities(t, database, economy)

	playerID := "player-001"
	shipID := "ship-001"
	portID := 100

	insertTestPlayer(t, database, playerID, "TestTrader", 10000)
	insertTestShip(t, database, shipID, playerID, 1, &portID)
	insertPortInventory(t, database, economy, portID, 1, "food_supplies", 100)

	// Perform trade
	err = economy.BuyCommodity(shipID, portID, "food_supplies", 10, 100)
	require.NoError(t, err)

	// Close database
	database.Close()

	// Reopen database
	database2, err := db.InitDatabase(dbPath, logger)
	require.NoError(t, err)
	defer database2.Close()

	// Verify trade persisted
	player, err := database2.GetPlayerByID(playerID)
	require.NoError(t, err)
	assert.Equal(t, int64(10000-1100), player.Credits, "Credits should persist")

	cargo, err := database2.GetShipCargo(shipID)
	require.NoError(t, err)
	require.Len(t, cargo, 1)
	assert.Equal(t, "food_supplies", cargo[0].CommodityID)
	assert.Equal(t, 10, cargo[0].Quantity)

	portInv, err := database2.GetPortInventory(portID, "food_supplies")
	require.NoError(t, err)
	assert.Equal(t, 90, portInv.Quantity, "Port inventory should persist")
}

// TestIntegration_PartialSellAndRebuy tests selling part of cargo and buying more
func TestIntegration_PartialSellAndRebuy(t *testing.T) {
	database, economy, cleanup := setupIntegrationTest(t)
	defer cleanup()

	playerID := "player-001"
	shipID := "ship-001"
	portID := 100

	insertTestPlayer(t, database, playerID, "TestTrader", 20000)
	insertTestShip(t, database, shipID, playerID, 1, &portID)
	insertPortInventory(t, database, economy, portID, 1, "food_supplies", 100)

	// Buy 15 units
	err := economy.BuyCommodity(shipID, portID, "food_supplies", 15, 100)
	require.NoError(t, err)

	// Sell 10 units (partial)
	err = economy.SellCommodity(shipID, portID, "food_supplies", 10, 105)
	require.NoError(t, err)

	// Verify 5 units remain
	cargo, err := database.GetShipCargo(shipID)
	require.NoError(t, err)
	require.Len(t, cargo, 1)
	assert.Equal(t, 5, cargo[0].Quantity)

	// Buy 10 more units (should add to existing slot)
	err = economy.BuyCommodity(shipID, portID, "food_supplies", 10, 110)
	require.NoError(t, err)

	// Verify total is 15 units in same slot
	cargo, err = database.GetShipCargo(shipID)
	require.NoError(t, err)
	require.Len(t, cargo, 1, "Should still be one cargo slot")
	assert.Equal(t, 15, cargo[0].Quantity, "Should have 15 units (5 + 10)")
}

// TestIntegration_CompleteTradeFlow tests a realistic multi-step trade scenario
func TestIntegration_CompleteTradeFlow(t *testing.T) {
	database, economy, cleanup := setupIntegrationTest(t)
	defer cleanup()

	playerID := "player-001"
	shipID := "ship-001"
	fedPortID := 100
	lowSecPortID := 300

	insertTestPlayer(t, database, playerID, "TestTrader", 15000)
	insertTestShip(t, database, shipID, playerID, 1, &fedPortID)
	
	// Setup inventories at both ports
	insertPortInventory(t, database, economy, fedPortID, 1, "food_supplies", 100)
	insertPortInventory(t, database, economy, fedPortID, 1, "electronics", 50)
	insertPortInventory(t, database, economy, lowSecPortID, 3, "food_supplies", 50)
	insertPortInventory(t, database, economy, lowSecPortID, 3, "raw_ore", 100)

	// Scenario: Buy supplies at Fed Space, travel to Low Sec, sell and buy ore, return
	
	// Step 1: Buy food and electronics at Fed Space
	err := economy.BuyCommodity(shipID, fedPortID, "food_supplies", 10, 100)
	require.NoError(t, err)
	
	err = economy.BuyCommodity(shipID, fedPortID, "electronics", 5, 105)
	require.NoError(t, err)

	initialCredits := int64(15000)
	player, _ := database.GetPlayerByID(playerID)
	spent := initialCredits - player.Credits
	t.Logf("Spent at Fed Space: %d credits", spent)

	// Step 2: Travel to Low Sec
	_, err = database.Conn().Exec(`
		UPDATE ships SET current_system_id = ?, docked_at_port_id = ?
		WHERE ship_id = ?
	`, 3, lowSecPortID, shipID)
	require.NoError(t, err)

	// Step 3: Sell food at Low Sec (higher prices)
	err = economy.SellCommodity(shipID, lowSecPortID, "food_supplies", 10, 110)
	require.NoError(t, err)

	// Step 4: Buy raw ore at Low Sec
	err = economy.BuyCommodity(shipID, lowSecPortID, "raw_ore", 10, 115)
	require.NoError(t, err)

	// Verify final state
	cargo, err := database.GetShipCargo(shipID)
	require.NoError(t, err)
	assert.Len(t, cargo, 2, "Should have electronics and raw_ore")

	// Verify cargo contents
	cargoMap := make(map[string]int)
	for _, slot := range cargo {
		cargoMap[slot.CommodityID] = slot.Quantity
	}
	assert.Equal(t, 5, cargoMap["electronics"])
	assert.Equal(t, 10, cargoMap["raw_ore"])

	// Verify player still has credits
	player, err = database.GetPlayerByID(playerID)
	require.NoError(t, err)
	assert.Greater(t, player.Credits, int64(0), "Player should still have credits")
	
	t.Logf("Final credits: %d (started with %d)", player.Credits, initialCredits)
}
