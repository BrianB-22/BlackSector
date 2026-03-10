package economy

import (
	"database/sql"
	"testing"

	"github.com/BrianB-22/BlackSector/internal/db"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDatabase implements the Database interface for testing
type mockDatabase struct {
	players       map[string]*db.Player
	ships         map[string]*db.Ship
	cargo         map[string][]db.CargoSlot
	portInventory map[int]map[string]*db.PortInventory
	txActive      bool
	txCommitted   bool
	txRolledBack  bool
}

func newMockDatabase() *mockDatabase {
	return &mockDatabase{
		players:       make(map[string]*db.Player),
		ships:         make(map[string]*db.Ship),
		cargo:         make(map[string][]db.CargoSlot),
		portInventory: make(map[int]map[string]*db.PortInventory),
	}
}

func (m *mockDatabase) GetPlayerByID(playerID string) (*db.Player, error) {
	return m.players[playerID], nil
}

func (m *mockDatabase) UpdatePlayerCredits(playerID string, credits int) error {
	if p := m.players[playerID]; p != nil {
		p.Credits = int64(credits)
	}
	return nil
}

func (m *mockDatabase) GetShipByID(shipID string) (*db.Ship, error) {
	return m.ships[shipID], nil
}

func (m *mockDatabase) GetShipCargo(shipID string) ([]db.CargoSlot, error) {
	return m.cargo[shipID], nil
}

func (m *mockDatabase) GetCargoTotalQuantity(shipID string) (int, error) {
	total := 0
	for _, slot := range m.cargo[shipID] {
		total += slot.Quantity
	}
	return total, nil
}

func (m *mockDatabase) GetCargoSlot(shipID string, commodityID string) (*db.CargoSlot, error) {
	for _, slot := range m.cargo[shipID] {
		if slot.CommodityID == commodityID {
			return &slot, nil
		}
	}
	return nil, nil
}

func (m *mockDatabase) AddCargo(shipID string, slotIndex int, commodityID string, quantity int) error {
	m.cargo[shipID] = append(m.cargo[shipID], db.CargoSlot{
		ShipID:      shipID,
		SlotIndex:   slotIndex,
		CommodityID: commodityID,
		Quantity:    quantity,
	})
	return nil
}

func (m *mockDatabase) UpdateCargoQuantity(shipID string, slotIndex int, quantity int) error {
	for i, slot := range m.cargo[shipID] {
		if slot.SlotIndex == slotIndex {
			m.cargo[shipID][i].Quantity = quantity
			return nil
		}
	}
	return nil
}

func (m *mockDatabase) RemoveCargo(shipID string, slotIndex int) error {
	newCargo := []db.CargoSlot{}
	for _, slot := range m.cargo[shipID] {
		if slot.SlotIndex != slotIndex {
			newCargo = append(newCargo, slot)
		}
	}
	m.cargo[shipID] = newCargo
	return nil
}

func (m *mockDatabase) GetPortInventory(portID int, commodityID string) (*db.PortInventory, error) {
	if portInv, ok := m.portInventory[portID]; ok {
		return portInv[commodityID], nil
	}
	return nil, nil
}

func (m *mockDatabase) GetAllPortInventory(portID int) ([]db.PortInventory, error) {
	result := []db.PortInventory{}
	if portInv, ok := m.portInventory[portID]; ok {
		for _, inv := range portInv {
			result = append(result, *inv)
		}
	}
	return result, nil
}

func (m *mockDatabase) UpdatePortInventory(portID int, commodityID string, quantity int, tick int64) error {
	if portInv, ok := m.portInventory[portID]; ok {
		if inv, ok := portInv[commodityID]; ok {
			inv.Quantity = quantity
			inv.UpdatedTick = tick
		}
	}
	return nil
}

func (m *mockDatabase) BeginTx() (*sql.Tx, error) {
	m.txActive = true
	m.txCommitted = false
	m.txRolledBack = false
	return nil, nil // Return nil for mock
}

func (m *mockDatabase) CommitTx(tx *sql.Tx) error {
	m.txCommitted = true
	m.txActive = false
	return nil
}

func (m *mockDatabase) RollbackTx(tx *sql.Tx) error {
	m.txRolledBack = true
	m.txActive = false
	return nil
}

// Transaction-aware methods (for mock, they just call the regular methods)
func (m *mockDatabase) TxGetPlayerByID(tx *sql.Tx, playerID string) (*db.Player, error) {
	return m.GetPlayerByID(playerID)
}

func (m *mockDatabase) TxUpdatePlayerCredits(tx *sql.Tx, playerID string, credits int) error {
	return m.UpdatePlayerCredits(playerID, credits)
}

func (m *mockDatabase) TxGetShipByID(tx *sql.Tx, shipID string) (*db.Ship, error) {
	return m.GetShipByID(shipID)
}

func (m *mockDatabase) TxGetCargoTotalQuantity(tx *sql.Tx, shipID string) (int, error) {
	return m.GetCargoTotalQuantity(shipID)
}

func (m *mockDatabase) TxGetCargoSlot(tx *sql.Tx, shipID string, commodityID string) (*db.CargoSlot, error) {
	return m.GetCargoSlot(shipID, commodityID)
}

func (m *mockDatabase) TxAddCargo(tx *sql.Tx, shipID string, slotIndex int, commodityID string, quantity int) error {
	return m.AddCargo(shipID, slotIndex, commodityID, quantity)
}

func (m *mockDatabase) TxUpdateCargoQuantity(tx *sql.Tx, shipID string, slotIndex int, quantity int) error {
	return m.UpdateCargoQuantity(shipID, slotIndex, quantity)
}

func (m *mockDatabase) TxRemoveCargo(tx *sql.Tx, shipID string, slotIndex int) error {
	return m.RemoveCargo(shipID, slotIndex)
}

func (m *mockDatabase) TxGetPortInventory(tx *sql.Tx, portID int, commodityID string) (*db.PortInventory, error) {
	return m.GetPortInventory(portID, commodityID)
}

func (m *mockDatabase) TxUpdatePortInventory(tx *sql.Tx, portID int, commodityID string, quantity int, tick int64) error {
	return m.UpdatePortInventory(portID, commodityID, quantity, tick)
}

// Test helpers
func setupTestEconomy() (*EconomySystem, *mockDatabase) {
	db := newMockDatabase()
	cfg := DefaultConfig()
	logger := zerolog.Nop()
	economy := NewEconomySystem(cfg, db, logger)
	
	// Load test commodities
	economy.LoadCommodities("../../config/world/alpha_sector.json")
	
	return economy, db
}

func setupTestPlayer(database *mockDatabase, playerID string, credits int64) {
	database.players[playerID] = &db.Player{
		PlayerID: playerID,
		Credits:  credits,
	}
}

func setupTestShip(database *mockDatabase, shipID string, playerID string, portID int, cargoCapacity int) {
	database.ships[shipID] = &db.Ship{
		ShipID:         shipID,
		PlayerID:       playerID,
		Status:         "DOCKED",
		DockedAtPortID: &portID,
		CargoCapacity:  cargoCapacity,
	}
}

func setupTestPortInventory(database *mockDatabase, portID int, commodityID string, quantity int, buyPrice int, sellPrice int) {
	if _, ok := database.portInventory[portID]; !ok {
		database.portInventory[portID] = make(map[string]*db.PortInventory)
	}
	database.portInventory[portID][commodityID] = &db.PortInventory{
		PortID:      portID,
		CommodityID: commodityID,
		Quantity:    quantity,
		BuyPrice:    buyPrice,
		SellPrice:   sellPrice,
	}
}

func TestBuyCommodity_Success(t *testing.T) {
	economy, testdb := setupTestEconomy()
	
	playerID := "player1"
	shipID := "ship1"
	portID := 1
	
	setupTestPlayer(testdb, playerID, 10000)
	setupTestShip(testdb, shipID, playerID, portID, 20)
	setupTestPortInventory(testdb, portID, "food_supplies", 100, 110, 90)
	
	err := economy.BuyCommodity(shipID, portID, "food_supplies", 10, 1)
	require.NoError(t, err)
	
	// Verify credits deducted
	assert.Equal(t, int64(10000-1100), testdb.players[playerID].Credits)
	
	// Verify port inventory reduced
	assert.Equal(t, 90, testdb.portInventory[portID]["food_supplies"].Quantity)
	
	// Verify cargo added
	cargo, _ := testdb.GetShipCargo(shipID)
	require.Len(t, cargo, 1)
	assert.Equal(t, "food_supplies", cargo[0].CommodityID)
	assert.Equal(t, 10, cargo[0].Quantity)
}

func TestBuyCommodity_InsufficientCredits(t *testing.T) {
	economy, testdb := setupTestEconomy()
	
	playerID := "player1"
	shipID := "ship1"
	portID := 1
	
	setupTestPlayer(testdb, playerID, 500) // Not enough for 10 units at 110 each
	setupTestShip(testdb, shipID, playerID, portID, 20)
	setupTestPortInventory(testdb, portID, "food_supplies", 100, 110, 90)
	
	err := economy.BuyCommodity(shipID, portID, "food_supplies", 10, 1)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInsufficientCredits)
	
	// Verify no changes
	assert.Equal(t, int64(500), testdb.players[playerID].Credits)
	assert.Equal(t, 100, testdb.portInventory[portID]["food_supplies"].Quantity)
}

func TestBuyCommodity_InsufficientCargo(t *testing.T) {
	economy, testdb := setupTestEconomy()
	
	playerID := "player1"
	shipID := "ship1"
	portID := 1
	
	setupTestPlayer(testdb, playerID, 10000)
	setupTestShip(testdb, shipID, playerID, portID, 5) // Only 5 cargo capacity
	setupTestPortInventory(testdb, portID, "food_supplies", 100, 110, 90)
	
	err := economy.BuyCommodity(shipID, portID, "food_supplies", 10, 1)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInsufficientCargo)
}

func TestBuyCommodity_InsufficientInventory(t *testing.T) {
	economy, testdb := setupTestEconomy()
	
	playerID := "player1"
	shipID := "ship1"
	portID := 1
	
	setupTestPlayer(testdb, playerID, 10000)
	setupTestShip(testdb, shipID, playerID, portID, 20)
	setupTestPortInventory(testdb, portID, "food_supplies", 5, 110, 90) // Only 5 available
	
	err := economy.BuyCommodity(shipID, portID, "food_supplies", 10, 1)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInsufficientInventory)
}

func TestBuyCommodity_ShipNotDocked(t *testing.T) {
	economy, testdb := setupTestEconomy()
	
	playerID := "player1"
	shipID := "ship1"
	portID := 1
	
	setupTestPlayer(testdb, playerID, 10000)
	setupTestShip(testdb, shipID, playerID, portID, 20)
	testdb.ships[shipID].Status = "IN_SPACE" // Not docked
	setupTestPortInventory(testdb, portID, "food_supplies", 100, 110, 90)
	
	err := economy.BuyCommodity(shipID, portID, "food_supplies", 10, 1)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrShipNotDocked)
}

func TestBuyCommodity_AddToExistingCargo(t *testing.T) {
	economy, testdb := setupTestEconomy()
	
	playerID := "player1"
	shipID := "ship1"
	portID := 1
	
	setupTestPlayer(testdb, playerID, 10000)
	setupTestShip(testdb, shipID, playerID, portID, 20)
	setupTestPortInventory(testdb, portID, "food_supplies", 100, 110, 90)
	
	// Add initial cargo
	testdb.cargo[shipID] = []db.CargoSlot{
		{ShipID: shipID, SlotIndex: 0, CommodityID: "food_supplies", Quantity: 5},
	}
	
	err := economy.BuyCommodity(shipID, portID, "food_supplies", 10, 1)
	require.NoError(t, err)
	
	// Verify cargo updated
	cargo, _ := testdb.GetShipCargo(shipID)
	require.Len(t, cargo, 1)
	assert.Equal(t, 15, cargo[0].Quantity) // 5 + 10
}

func TestSellCommodity_Success(t *testing.T) {
	economy, testdb := setupTestEconomy()
	
	playerID := "player1"
	shipID := "ship1"
	portID := 1
	
	setupTestPlayer(testdb, playerID, 5000)
	setupTestShip(testdb, shipID, playerID, portID, 20)
	setupTestPortInventory(testdb, portID, "food_supplies", 50, 110, 90)
	
	// Add cargo to sell
	testdb.cargo[shipID] = []db.CargoSlot{
		{ShipID: shipID, SlotIndex: 0, CommodityID: "food_supplies", Quantity: 10},
	}
	
	err := economy.SellCommodity(shipID, portID, "food_supplies", 10, 1)
	require.NoError(t, err)
	
	// Verify credits added
	assert.Equal(t, int64(5000+900), testdb.players[playerID].Credits) // 10 * 90
	
	// Verify port inventory increased
	assert.Equal(t, 60, testdb.portInventory[portID]["food_supplies"].Quantity)
	
	// Verify cargo removed
	cargo, _ := testdb.GetShipCargo(shipID)
	assert.Len(t, cargo, 0)
}

func TestSellCommodity_PartialSell(t *testing.T) {
	economy, testdb := setupTestEconomy()
	
	playerID := "player1"
	shipID := "ship1"
	portID := 1
	
	setupTestPlayer(testdb, playerID, 5000)
	setupTestShip(testdb, shipID, playerID, portID, 20)
	setupTestPortInventory(testdb, portID, "food_supplies", 50, 110, 90)
	
	// Add cargo to sell
	testdb.cargo[shipID] = []db.CargoSlot{
		{ShipID: shipID, SlotIndex: 0, CommodityID: "food_supplies", Quantity: 15},
	}
	
	err := economy.SellCommodity(shipID, portID, "food_supplies", 10, 1)
	require.NoError(t, err)
	
	// Verify cargo reduced but not removed
	cargo, _ := testdb.GetShipCargo(shipID)
	require.Len(t, cargo, 1)
	assert.Equal(t, 5, cargo[0].Quantity)
}

func TestSellCommodity_CommodityNotInCargo(t *testing.T) {
	economy, testdb := setupTestEconomy()
	
	playerID := "player1"
	shipID := "ship1"
	portID := 1
	
	setupTestPlayer(testdb, playerID, 5000)
	setupTestShip(testdb, shipID, playerID, portID, 20)
	setupTestPortInventory(testdb, portID, "food_supplies", 50, 110, 90)
	
	// No cargo
	
	err := economy.SellCommodity(shipID, portID, "food_supplies", 10, 1)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrCommodityNotInCargo)
}

func TestSellCommodity_InsufficientQuantity(t *testing.T) {
	economy, testdb := setupTestEconomy()
	
	playerID := "player1"
	shipID := "ship1"
	portID := 1
	
	setupTestPlayer(testdb, playerID, 5000)
	setupTestShip(testdb, shipID, playerID, portID, 20)
	setupTestPortInventory(testdb, portID, "food_supplies", 50, 110, 90)
	
	// Add cargo but not enough
	testdb.cargo[shipID] = []db.CargoSlot{
		{ShipID: shipID, SlotIndex: 0, CommodityID: "food_supplies", Quantity: 5},
	}
	
	err := economy.SellCommodity(shipID, portID, "food_supplies", 10, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient quantity in cargo")
}

func TestGetMarketPrices(t *testing.T) {
	economy, testdb := setupTestEconomy()
	
	portID := 1
	setupTestPortInventory(testdb, portID, "food_supplies", 100, 110, 90)
	setupTestPortInventory(testdb, portID, "fuel_cells", 50, 220, 180)
	
	prices, err := economy.GetMarketPrices(portID)
	require.NoError(t, err)
	require.Len(t, prices, 2)
	
	// Verify prices include commodity names
	for _, price := range prices {
		assert.NotEmpty(t, price.Name)
		assert.NotEmpty(t, price.CommodityID)
		assert.Greater(t, price.BuyPrice, 0)
		assert.Greater(t, price.SellPrice, 0)
	}
}

func TestGetCargoManifest(t *testing.T) {
	economy, testdb := setupTestEconomy()
	
	shipID := "ship1"
	testdb.cargo[shipID] = []db.CargoSlot{
		{ShipID: shipID, SlotIndex: 0, CommodityID: "food_supplies", Quantity: 10},
		{ShipID: shipID, SlotIndex: 1, CommodityID: "fuel_cells", Quantity: 5},
	}
	
	manifest, err := economy.GetCargoManifest(shipID)
	require.NoError(t, err)
	require.Len(t, manifest, 2)
	
	// Verify manifest includes commodity names
	for _, entry := range manifest {
		assert.NotEmpty(t, entry.Name)
		assert.NotEmpty(t, entry.CommodityID)
		assert.Greater(t, entry.Quantity, 0)
	}
}

func TestGetCargoCapacityInfo(t *testing.T) {
	economy, testdb := setupTestEconomy()
	
	shipID := "ship1"
	playerID := "player1"
	portID := 1
	
	setupTestShip(testdb, shipID, playerID, portID, 20)
	testdb.cargo[shipID] = []db.CargoSlot{
		{ShipID: shipID, SlotIndex: 0, CommodityID: "food_supplies", Quantity: 10},
		{ShipID: shipID, SlotIndex: 1, CommodityID: "fuel_cells", Quantity: 5},
	}
	
	info, err := economy.GetCargoCapacityInfo(shipID)
	require.NoError(t, err)
	assert.Equal(t, 15, info.CurrentCargo)
	assert.Equal(t, 20, info.MaxCargo)
	assert.Equal(t, 5, info.Available)
}
