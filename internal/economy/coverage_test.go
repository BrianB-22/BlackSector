package economy

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/BrianB-22/BlackSector/internal/db"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDatabaseWithErrors extends mockDatabase to simulate error conditions
type mockDatabaseWithErrors struct {
	*mockDatabase
	getShipCargoError         error
	getCargoTotalQuantityErr  error
	getPlayerByIDError        error
	getShipByIDError          error
	getPortInventoryError     error
	getAllPortInventoryError  error
	beginTxError              error
	commitTxError             error
	updatePlayerCreditsError  error
	updatePortInventoryError  error
	addCargoError             error
	updateCargoQuantityError  error
	removeCargoError          error
	getCargoSlotError         error
}

func newMockDatabaseWithErrors() *mockDatabaseWithErrors {
	return &mockDatabaseWithErrors{
		mockDatabase: newMockDatabase(),
	}
}

func (m *mockDatabaseWithErrors) GetShipCargo(shipID string) ([]db.CargoSlot, error) {
	if m.getShipCargoError != nil {
		return nil, m.getShipCargoError
	}
	return m.mockDatabase.GetShipCargo(shipID)
}

func (m *mockDatabaseWithErrors) GetCargoTotalQuantity(shipID string) (int, error) {
	if m.getCargoTotalQuantityErr != nil {
		return 0, m.getCargoTotalQuantityErr
	}
	return m.mockDatabase.GetCargoTotalQuantity(shipID)
}

func (m *mockDatabaseWithErrors) GetPlayerByID(playerID string) (*db.Player, error) {
	if m.getPlayerByIDError != nil {
		return nil, m.getPlayerByIDError
	}
	return m.mockDatabase.GetPlayerByID(playerID)
}

func (m *mockDatabaseWithErrors) GetShipByID(shipID string) (*db.Ship, error) {
	if m.getShipByIDError != nil {
		return nil, m.getShipByIDError
	}
	return m.mockDatabase.GetShipByID(shipID)
}

func (m *mockDatabaseWithErrors) GetPortInventory(portID int, commodityID string) (*db.PortInventory, error) {
	if m.getPortInventoryError != nil {
		return nil, m.getPortInventoryError
	}
	return m.mockDatabase.GetPortInventory(portID, commodityID)
}

func (m *mockDatabaseWithErrors) GetAllPortInventory(portID int) ([]db.PortInventory, error) {
	if m.getAllPortInventoryError != nil {
		return nil, m.getAllPortInventoryError
	}
	return m.mockDatabase.GetAllPortInventory(portID)
}

func (m *mockDatabaseWithErrors) BeginTx() (*sql.Tx, error) {
	if m.beginTxError != nil {
		return nil, m.beginTxError
	}
	return m.mockDatabase.BeginTx()
}

func (m *mockDatabaseWithErrors) CommitTx(tx *sql.Tx) error {
	if m.commitTxError != nil {
		return m.commitTxError
	}
	return m.mockDatabase.CommitTx(tx)
}

func (m *mockDatabaseWithErrors) TxUpdatePlayerCredits(tx *sql.Tx, playerID string, credits int) error {
	if m.updatePlayerCreditsError != nil {
		return m.updatePlayerCreditsError
	}
	return m.mockDatabase.TxUpdatePlayerCredits(tx, playerID, credits)
}

func (m *mockDatabaseWithErrors) TxUpdatePortInventory(tx *sql.Tx, portID int, commodityID string, quantity int, tick int64) error {
	if m.updatePortInventoryError != nil {
		return m.updatePortInventoryError
	}
	return m.mockDatabase.TxUpdatePortInventory(tx, portID, commodityID, quantity, tick)
}

func (m *mockDatabaseWithErrors) TxAddCargo(tx *sql.Tx, shipID string, slotIndex int, commodityID string, quantity int) error {
	if m.addCargoError != nil {
		return m.addCargoError
	}
	return m.mockDatabase.TxAddCargo(tx, shipID, slotIndex, commodityID, quantity)
}

func (m *mockDatabaseWithErrors) TxUpdateCargoQuantity(tx *sql.Tx, shipID string, slotIndex int, quantity int) error {
	if m.updateCargoQuantityError != nil {
		return m.updateCargoQuantityError
	}
	return m.mockDatabase.TxUpdateCargoQuantity(tx, shipID, slotIndex, quantity)
}

func (m *mockDatabaseWithErrors) TxRemoveCargo(tx *sql.Tx, shipID string, slotIndex int) error {
	if m.removeCargoError != nil {
		return m.removeCargoError
	}
	return m.mockDatabase.TxRemoveCargo(tx, shipID, slotIndex)
}

func (m *mockDatabaseWithErrors) TxGetCargoSlot(tx *sql.Tx, shipID string, commodityID string) (*db.CargoSlot, error) {
	if m.getCargoSlotError != nil {
		return nil, m.getCargoSlotError
	}
	return m.mockDatabase.TxGetCargoSlot(tx, shipID, commodityID)
}

func (m *mockDatabaseWithErrors) TxGetCargoTotalQuantity(tx *sql.Tx, shipID string) (int, error) {
	if m.getCargoTotalQuantityErr != nil {
		return 0, m.getCargoTotalQuantityErr
	}
	return m.mockDatabase.TxGetCargoTotalQuantity(tx, shipID)
}

// Test GetShipCargo - currently 0% coverage
func TestGetShipCargo_Success(t *testing.T) {
	economy, testdb := setupTestEconomy()
	
	shipID := "ship1"
	testdb.cargo[shipID] = []db.CargoSlot{
		{ShipID: shipID, SlotIndex: 0, CommodityID: "food_supplies", Quantity: 10},
		{ShipID: shipID, SlotIndex: 1, CommodityID: "fuel_cells", Quantity: 5},
	}
	
	cargo, err := economy.GetShipCargo(shipID)
	require.NoError(t, err)
	assert.Len(t, cargo, 2)
	assert.Equal(t, "food_supplies", cargo[0].CommodityID)
	assert.Equal(t, 10, cargo[0].Quantity)
}

func TestGetShipCargo_Empty(t *testing.T) {
	economy, _ := setupTestEconomy()
	
	cargo, err := economy.GetShipCargo("ship1")
	require.NoError(t, err)
	assert.Len(t, cargo, 0)
}

func TestGetShipCargo_DatabaseError(t *testing.T) {
	testdb := newMockDatabaseWithErrors()
	testdb.getShipCargoError = errors.New("database error")
	
	cfg := DefaultConfig()
	economy := NewEconomySystem(cfg, testdb, zerolog.Nop())
	economy.LoadCommodities("../../config/world/alpha_sector.json")
	
	cargo, err := economy.GetShipCargo("ship1")
	require.Error(t, err)
	assert.Nil(t, cargo)
	assert.Contains(t, err.Error(), "get ship cargo")
}

// Test BuyCommodity edge cases
func TestBuyCommodity_ZeroQuantity(t *testing.T) {
	economy, testdb := setupTestEconomy()
	
	playerID := "player1"
	shipID := "ship1"
	portID := 1
	
	setupTestPlayer(testdb, playerID, 10000)
	setupTestShip(testdb, shipID, playerID, portID, 20)
	setupTestPortInventory(testdb, portID, "food_supplies", 100, 110, 90)
	
	err := economy.BuyCommodity(shipID, portID, "food_supplies", 0, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "quantity must be positive")
}

func TestBuyCommodity_NegativeQuantity(t *testing.T) {
	economy, testdb := setupTestEconomy()
	
	playerID := "player1"
	shipID := "ship1"
	portID := 1
	
	setupTestPlayer(testdb, playerID, 10000)
	setupTestShip(testdb, shipID, playerID, portID, 20)
	setupTestPortInventory(testdb, portID, "food_supplies", 100, 110, 90)
	
	err := economy.BuyCommodity(shipID, portID, "food_supplies", -5, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "quantity must be positive")
}

func TestBuyCommodity_BeginTxError(t *testing.T) {
	testdb := newMockDatabaseWithErrors()
	testdb.beginTxError = errors.New("transaction error")
	
	cfg := DefaultConfig()
	economy := NewEconomySystem(cfg, testdb, zerolog.Nop())
	economy.LoadCommodities("../../config/world/alpha_sector.json")
	
	err := economy.BuyCommodity("ship1", 1, "food_supplies", 10, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "buy commodity")
}

func TestBuyCommodity_ShipNotFound(t *testing.T) {
	economy, _ := setupTestEconomy()
	
	err := economy.BuyCommodity("nonexistent", 1, "food_supplies", 10, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ship not found")
}

func TestBuyCommodity_WrongPort(t *testing.T) {
	economy, testdb := setupTestEconomy()
	
	playerID := "player1"
	shipID := "ship1"
	portID := 1
	
	setupTestPlayer(testdb, playerID, 10000)
	setupTestShip(testdb, shipID, playerID, portID, 20)
	setupTestPortInventory(testdb, portID, "food_supplies", 100, 110, 90)
	
	// Try to buy from port 2 while docked at port 1
	err := economy.BuyCommodity(shipID, 2, "food_supplies", 10, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ship not docked at port 2")
}

func TestBuyCommodity_InvalidCommodity(t *testing.T) {
	economy, testdb := setupTestEconomy()
	
	playerID := "player1"
	shipID := "ship1"
	portID := 1
	
	setupTestPlayer(testdb, playerID, 10000)
	setupTestShip(testdb, shipID, playerID, portID, 20)
	
	err := economy.BuyCommodity(shipID, portID, "invalid_commodity", 10, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "buy commodity")
}

func TestBuyCommodity_CommodityNotAtPort(t *testing.T) {
	economy, testdb := setupTestEconomy()
	
	playerID := "player1"
	shipID := "ship1"
	portID := 1
	
	setupTestPlayer(testdb, playerID, 10000)
	setupTestShip(testdb, shipID, playerID, portID, 20)
	// Don't setup port inventory for this commodity
	
	err := economy.BuyCommodity(shipID, portID, "food_supplies", 10, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "commodity not available at this port")
}

func TestBuyCommodity_PlayerNotFound(t *testing.T) {
	economy, testdb := setupTestEconomy()
	
	shipID := "ship1"
	portID := 1
	
	// Setup ship but no player
	setupTestShip(testdb, shipID, "player1", portID, 20)
	setupTestPortInventory(testdb, portID, "food_supplies", 100, 110, 90)
	
	err := economy.BuyCommodity(shipID, portID, "food_supplies", 10, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "player not found")
}

func TestBuyCommodity_CommitError(t *testing.T) {
	testdb := newMockDatabaseWithErrors()
	testdb.commitTxError = errors.New("commit failed")
	
	cfg := DefaultConfig()
	economy := NewEconomySystem(cfg, testdb, zerolog.Nop())
	economy.LoadCommodities("../../config/world/alpha_sector.json")
	
	playerID := "player1"
	shipID := "ship1"
	portID := 1
	
	setupTestPlayer(testdb.mockDatabase, playerID, 10000)
	setupTestShip(testdb.mockDatabase, shipID, playerID, portID, 20)
	setupTestPortInventory(testdb.mockDatabase, portID, "food_supplies", 100, 110, 90)
	
	err := economy.BuyCommodity(shipID, portID, "food_supplies", 10, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "buy commodity")
}

// Test SellCommodity edge cases
func TestSellCommodity_ZeroQuantity(t *testing.T) {
	economy, testdb := setupTestEconomy()
	
	playerID := "player1"
	shipID := "ship1"
	portID := 1
	
	setupTestPlayer(testdb, playerID, 5000)
	setupTestShip(testdb, shipID, playerID, portID, 20)
	setupTestPortInventory(testdb, portID, "food_supplies", 50, 110, 90)
	
	err := economy.SellCommodity(shipID, portID, "food_supplies", 0, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "quantity must be positive")
}

func TestSellCommodity_NegativeQuantity(t *testing.T) {
	economy, testdb := setupTestEconomy()
	
	playerID := "player1"
	shipID := "ship1"
	portID := 1
	
	setupTestPlayer(testdb, playerID, 5000)
	setupTestShip(testdb, shipID, playerID, portID, 20)
	setupTestPortInventory(testdb, portID, "food_supplies", 50, 110, 90)
	
	err := economy.SellCommodity(shipID, portID, "food_supplies", -5, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "quantity must be positive")
}

func TestSellCommodity_ShipNotFound(t *testing.T) {
	economy, _ := setupTestEconomy()
	
	err := economy.SellCommodity("nonexistent", 1, "food_supplies", 10, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ship not found")
}

func TestSellCommodity_WrongPort(t *testing.T) {
	economy, testdb := setupTestEconomy()
	
	playerID := "player1"
	shipID := "ship1"
	portID := 1
	
	setupTestPlayer(testdb, playerID, 5000)
	setupTestShip(testdb, shipID, playerID, portID, 20)
	setupTestPortInventory(testdb, portID, "food_supplies", 50, 110, 90)
	
	testdb.cargo[shipID] = []db.CargoSlot{
		{ShipID: shipID, SlotIndex: 0, CommodityID: "food_supplies", Quantity: 10},
	}
	
	// Try to sell at port 2 while docked at port 1
	err := economy.SellCommodity(shipID, 2, "food_supplies", 10, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ship not docked at port 2")
}

func TestSellCommodity_InvalidCommodity(t *testing.T) {
	economy, testdb := setupTestEconomy()
	
	playerID := "player1"
	shipID := "ship1"
	portID := 1
	
	setupTestPlayer(testdb, playerID, 5000)
	setupTestShip(testdb, shipID, playerID, portID, 20)
	
	err := economy.SellCommodity(shipID, portID, "invalid_commodity", 10, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sell commodity")
}

func TestSellCommodity_PortDoesNotBuy(t *testing.T) {
	economy, testdb := setupTestEconomy()
	
	playerID := "player1"
	shipID := "ship1"
	portID := 1
	
	setupTestPlayer(testdb, playerID, 5000)
	setupTestShip(testdb, shipID, playerID, portID, 20)
	// Don't setup port inventory - port doesn't buy this commodity
	
	testdb.cargo[shipID] = []db.CargoSlot{
		{ShipID: shipID, SlotIndex: 0, CommodityID: "food_supplies", Quantity: 10},
	}
	
	err := economy.SellCommodity(shipID, portID, "food_supplies", 10, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "port does not buy this commodity")
}

func TestSellCommodity_PlayerNotFound(t *testing.T) {
	economy, testdb := setupTestEconomy()
	
	shipID := "ship1"
	portID := 1
	
	// Setup ship but no player
	setupTestShip(testdb, shipID, "player1", portID, 20)
	setupTestPortInventory(testdb, portID, "food_supplies", 50, 110, 90)
	
	testdb.cargo[shipID] = []db.CargoSlot{
		{ShipID: shipID, SlotIndex: 0, CommodityID: "food_supplies", Quantity: 10},
	}
	
	err := economy.SellCommodity(shipID, portID, "food_supplies", 10, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "player not found")
}

func TestSellCommodity_CommitError(t *testing.T) {
	testdb := newMockDatabaseWithErrors()
	testdb.commitTxError = errors.New("commit failed")
	
	cfg := DefaultConfig()
	economy := NewEconomySystem(cfg, testdb, zerolog.Nop())
	economy.LoadCommodities("../../config/world/alpha_sector.json")
	
	playerID := "player1"
	shipID := "ship1"
	portID := 1
	
	setupTestPlayer(testdb.mockDatabase, playerID, 5000)
	setupTestShip(testdb.mockDatabase, shipID, playerID, portID, 20)
	setupTestPortInventory(testdb.mockDatabase, portID, "food_supplies", 50, 110, 90)
	
	testdb.cargo[shipID] = []db.CargoSlot{
		{ShipID: shipID, SlotIndex: 0, CommodityID: "food_supplies", Quantity: 10},
	}
	
	err := economy.SellCommodity(shipID, portID, "food_supplies", 10, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sell commodity")
}

// Test GetMarketPrices error cases
func TestGetMarketPrices_DatabaseError(t *testing.T) {
	testdb := newMockDatabaseWithErrors()
	testdb.getAllPortInventoryError = errors.New("database error")
	
	cfg := DefaultConfig()
	economy := NewEconomySystem(cfg, testdb, zerolog.Nop())
	
	prices, err := economy.GetMarketPrices(1)
	require.Error(t, err)
	assert.Nil(t, prices)
	assert.Contains(t, err.Error(), "get market prices")
}

func TestGetMarketPrices_EmptyPort(t *testing.T) {
	economy, _ := setupTestEconomy()
	
	prices, err := economy.GetMarketPrices(999)
	require.NoError(t, err)
	assert.Len(t, prices, 0)
}

// Test GetCargoManifest error cases
func TestGetCargoManifest_DatabaseError(t *testing.T) {
	testdb := newMockDatabaseWithErrors()
	testdb.getShipCargoError = errors.New("database error")
	
	cfg := DefaultConfig()
	economy := NewEconomySystem(cfg, testdb, zerolog.Nop())
	
	manifest, err := economy.GetCargoManifest("ship1")
	require.Error(t, err)
	assert.Nil(t, manifest)
	assert.Contains(t, err.Error(), "get cargo manifest")
}

func TestGetCargoManifest_EmptyCargo(t *testing.T) {
	economy, _ := setupTestEconomy()
	
	manifest, err := economy.GetCargoManifest("ship1")
	require.NoError(t, err)
	assert.Len(t, manifest, 0)
}

// Test GetCargoCapacityInfo error cases
func TestGetCargoCapacityInfo_ShipNotFound(t *testing.T) {
	economy, _ := setupTestEconomy()
	
	info, err := economy.GetCargoCapacityInfo("nonexistent")
	require.Error(t, err)
	assert.Nil(t, info)
	assert.Contains(t, err.Error(), "ship not found")
}

func TestGetCargoCapacityInfo_GetShipError(t *testing.T) {
	testdb := newMockDatabaseWithErrors()
	testdb.getShipByIDError = errors.New("database error")
	
	cfg := DefaultConfig()
	economy := NewEconomySystem(cfg, testdb, zerolog.Nop())
	
	info, err := economy.GetCargoCapacityInfo("ship1")
	require.Error(t, err)
	assert.Nil(t, info)
	assert.Contains(t, err.Error(), "get cargo capacity info")
}

func TestGetCargoCapacityInfo_GetCargoError(t *testing.T) {
	testdb := newMockDatabaseWithErrors()
	testdb.getCargoTotalQuantityErr = errors.New("database error")
	
	cfg := DefaultConfig()
	economy := NewEconomySystem(cfg, testdb, zerolog.Nop())
	
	playerID := "player1"
	shipID := "ship1"
	portID := 1
	
	setupTestShip(testdb.mockDatabase, shipID, playerID, portID, 20)
	
	info, err := economy.GetCargoCapacityInfo(shipID)
	require.Error(t, err)
	assert.Nil(t, info)
	assert.Contains(t, err.Error(), "get cargo capacity info")
}

// Test LoadCommodities error case
func TestLoadCommodities_FileNotFound(t *testing.T) {
	cfg := DefaultConfig()
	economy := NewEconomySystem(cfg, nil, zerolog.Nop())
	
	err := economy.LoadCommodities("/nonexistent/path/config.json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load commodities")
}

// Test CommodityRegistry LoadFromFile error case
func TestCommodityRegistry_LoadFromFile_FileNotFound(t *testing.T) {
	registry := NewCommodityRegistry()
	
	err := registry.LoadFromFile("/nonexistent/path/config.json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read world config")
}
