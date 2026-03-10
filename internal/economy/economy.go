package economy

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/BrianB-22/BlackSector/internal/db"
	"github.com/rs/zerolog"
)

// Common errors returned by the economy system
var (
	ErrInsufficientCredits   = errors.New("insufficient credits for purchase")
	ErrInsufficientCargo     = errors.New("insufficient cargo capacity")
	ErrInsufficientInventory = errors.New("port has insufficient inventory")
	ErrCommodityNotInCargo   = errors.New("commodity not found in cargo")
	ErrInvalidPort           = errors.New("invalid port")
	ErrInvalidCommodity      = errors.New("invalid commodity")
	ErrShipNotDocked         = errors.New("ship must be docked to trade")
)

// Database interface defines the required database operations for the economy system
type Database interface {
	// Player operations
	GetPlayerByID(playerID string) (*db.Player, error)
	UpdatePlayerCredits(playerID string, credits int) error

	// Ship operations
	GetShipByID(shipID string) (*db.Ship, error)

	// Cargo operations
	GetShipCargo(shipID string) ([]db.CargoSlot, error)
	GetCargoTotalQuantity(shipID string) (int, error)
	GetCargoSlot(shipID string, commodityID string) (*db.CargoSlot, error)
	AddCargo(shipID string, slotIndex int, commodityID string, quantity int) error
	UpdateCargoQuantity(shipID string, slotIndex int, quantity int) error
	RemoveCargo(shipID string, slotIndex int) error

	// Port inventory operations
	GetPortInventory(portID int, commodityID string) (*db.PortInventory, error)
	GetAllPortInventory(portID int) ([]db.PortInventory, error)
	UpdatePortInventory(portID int, commodityID string, quantity int, tick int64) error

	// Transaction management
	BeginTx() (*sql.Tx, error)
	CommitTx(tx *sql.Tx) error
	RollbackTx(tx *sql.Tx) error

	// Transaction-aware operations
	TxGetPlayerByID(tx *sql.Tx, playerID string) (*db.Player, error)
	TxUpdatePlayerCredits(tx *sql.Tx, playerID string, credits int) error
	TxGetShipByID(tx *sql.Tx, shipID string) (*db.Ship, error)
	TxGetCargoTotalQuantity(tx *sql.Tx, shipID string) (int, error)
	TxGetCargoSlot(tx *sql.Tx, shipID string, commodityID string) (*db.CargoSlot, error)
	TxAddCargo(tx *sql.Tx, shipID string, slotIndex int, commodityID string, quantity int) error
	TxUpdateCargoQuantity(tx *sql.Tx, shipID string, slotIndex int, quantity int) error
	TxRemoveCargo(tx *sql.Tx, shipID string, slotIndex int) error
	TxGetPortInventory(tx *sql.Tx, portID int, commodityID string) (*db.PortInventory, error)
	TxUpdatePortInventory(tx *sql.Tx, portID int, commodityID string, quantity int, tick int64) error
}

// Config holds economy system configuration
type Config struct {
	LowSecPriceMultiplier float64 // Price multiplier for Low Security zones (default 1.18)
	BuyMarkup             float64 // Markup when port sells to player (default 1.10)
	SellMarkdown          float64 // Markdown when port buys from player (default 0.90)
}

// DefaultConfig returns the default economy configuration
func DefaultConfig() *Config {
	return &Config{
		LowSecPriceMultiplier: 1.18,
		BuyMarkup:             1.10,
		SellMarkdown:          0.90,
	}
}

// EconomySystem is the concrete implementation of the Trader interface
type EconomySystem struct {
	cfg         *Config
	db          Database
	commodities *CommodityRegistry
	logger      zerolog.Logger
}

// NewEconomySystem creates a new economy system instance
func NewEconomySystem(cfg *Config, db Database, logger zerolog.Logger) *EconomySystem {
	return &EconomySystem{
		cfg:         cfg,
		db:          db,
		commodities: NewCommodityRegistry(),
		logger:      logger,
	}
}

// LoadCommodities loads commodity definitions from world configuration
func (e *EconomySystem) LoadCommodities(configPath string) error {
	if err := e.commodities.LoadFromFile(configPath); err != nil {
		return fmt.Errorf("load commodities: %w", err)
	}
	e.logger.Info().
		Int("count", e.commodities.Count()).
		Msg("commodities loaded")
	return nil
}

// GetCommodity retrieves a commodity definition by ID
func (e *EconomySystem) GetCommodity(commodityID string) (*CommodityDefinition, error) {
	return e.commodities.GetCommodity(commodityID)
}

// GetAllCommodities returns all registered commodity definitions
func (e *EconomySystem) GetAllCommodities() []*CommodityDefinition {
	return e.commodities.GetAllCommodities()
}

// BuyCommodity purchases commodity from port to ship cargo
// Validates: ship docked, sufficient credits, sufficient cargo space, sufficient port inventory
// Performs atomic transaction: deduct credits, reduce port inventory, add cargo
func (e *EconomySystem) BuyCommodity(shipID string, portID int, commodityID string, quantity int, tick int64) error {
	if quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}

	// Start transaction
	tx, err := e.db.BeginTx()
	if err != nil {
		return fmt.Errorf("buy commodity: %w", err)
	}
	defer func() {
		if tx != nil {
			e.db.RollbackTx(tx)
		}
	}()

	// Get ship and validate
	ship, err := e.db.TxGetShipByID(tx, shipID)
	if err != nil {
		return fmt.Errorf("buy commodity: %w", err)
	}
	if ship == nil {
		return fmt.Errorf("buy commodity: ship not found")
	}

	// Validate ship is docked at the specified port
	if ship.Status != "DOCKED" {
		return fmt.Errorf("buy commodity: %w", ErrShipNotDocked)
	}
	if ship.DockedAtPortID == nil || *ship.DockedAtPortID != portID {
		return fmt.Errorf("buy commodity: ship not docked at port %d", portID)
	}

	// Get commodity definition (validate it exists)
	_, err = e.commodities.GetCommodity(commodityID)
	if err != nil {
		return fmt.Errorf("buy commodity: %w", err)
	}

	// Get port inventory
	portInv, err := e.db.TxGetPortInventory(tx, portID, commodityID)
	if err != nil {
		return fmt.Errorf("buy commodity: %w", err)
	}
	if portInv == nil {
		return fmt.Errorf("buy commodity: commodity not available at this port")
	}

	// Validate port has sufficient inventory
	if portInv.Quantity < quantity {
		return fmt.Errorf("buy commodity: %w (available: %d, requested: %d)", 
			ErrInsufficientInventory, portInv.Quantity, quantity)
	}

	// Calculate total cost (use port's buy price - what player pays)
	totalCost := portInv.BuyPrice * quantity

	// Get player and validate credits
	player, err := e.db.TxGetPlayerByID(tx, ship.PlayerID)
	if err != nil {
		return fmt.Errorf("buy commodity: %w", err)
	}
	if player == nil {
		return fmt.Errorf("buy commodity: player not found")
	}

	if player.Credits < int64(totalCost) {
		return fmt.Errorf("buy commodity: %w (have: %d, need: %d)", 
			ErrInsufficientCredits, player.Credits, totalCost)
	}

	// Validate cargo capacity
	currentCargo, err := e.db.TxGetCargoTotalQuantity(tx, shipID)
	if err != nil {
		return fmt.Errorf("buy commodity: %w", err)
	}

	if currentCargo+quantity > ship.CargoCapacity {
		return fmt.Errorf("buy commodity: %w (capacity: %d, current: %d, adding: %d)", 
			ErrInsufficientCargo, ship.CargoCapacity, currentCargo, quantity)
	}

	// All validations passed - perform atomic transaction

	// 1. Deduct credits from player
	newCredits := int(player.Credits) - totalCost
	if err := e.db.TxUpdatePlayerCredits(tx, player.PlayerID, newCredits); err != nil {
		return fmt.Errorf("buy commodity: %w", err)
	}

	// 2. Reduce port inventory
	newPortQuantity := portInv.Quantity - quantity
	if err := e.db.TxUpdatePortInventory(tx, portID, commodityID, newPortQuantity, tick); err != nil {
		return fmt.Errorf("buy commodity: %w", err)
	}

	// 3. Add to ship cargo
	existingSlot, err := e.db.TxGetCargoSlot(tx, shipID, commodityID)
	if err != nil {
		return fmt.Errorf("buy commodity: %w", err)
	}

	if existingSlot != nil {
		// Update existing slot
		newQuantity := existingSlot.Quantity + quantity
		if err := e.db.TxUpdateCargoQuantity(tx, shipID, existingSlot.SlotIndex, newQuantity); err != nil {
			return fmt.Errorf("buy commodity: %w", err)
		}
	} else {
		// Find next available slot index
		cargo, err := e.db.GetShipCargo(shipID)
		if err != nil {
			return fmt.Errorf("buy commodity: %w", err)
		}
		
		nextSlot := 0
		if len(cargo) > 0 {
			// Find highest slot index and add 1
			maxSlot := -1
			for _, slot := range cargo {
				if slot.SlotIndex > maxSlot {
					maxSlot = slot.SlotIndex
				}
			}
			nextSlot = maxSlot + 1
		}

		if err := e.db.TxAddCargo(tx, shipID, nextSlot, commodityID, quantity); err != nil {
			return fmt.Errorf("buy commodity: %w", err)
		}
	}

	// Commit transaction
	if err := e.db.CommitTx(tx); err != nil {
		return fmt.Errorf("buy commodity: %w", err)
	}

	e.logger.Info().
		Str("ship_id", shipID).
		Int("port_id", portID).
		Str("commodity_id", commodityID).
		Int("quantity", quantity).
		Int("total_cost", totalCost).
		Int64("tick", tick).
		Msg("Commodity purchased")

	return nil
}

// SellCommodity sells commodity from ship cargo to port
// Validates: ship docked, commodity in cargo
// Performs atomic transaction: add credits, increase port inventory, remove cargo
func (e *EconomySystem) SellCommodity(shipID string, portID int, commodityID string, quantity int, tick int64) error {
	if quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}

	// Start transaction
	tx, err := e.db.BeginTx()
	if err != nil {
		return fmt.Errorf("sell commodity: %w", err)
	}
	defer func() {
		if tx != nil {
			e.db.RollbackTx(tx)
		}
	}()

	// Get ship and validate
	ship, err := e.db.TxGetShipByID(tx, shipID)
	if err != nil {
		return fmt.Errorf("sell commodity: %w", err)
	}
	if ship == nil {
		return fmt.Errorf("sell commodity: ship not found")
	}

	// Validate ship is docked at the specified port
	if ship.Status != "DOCKED" {
		return fmt.Errorf("sell commodity: %w", ErrShipNotDocked)
	}
	if ship.DockedAtPortID == nil || *ship.DockedAtPortID != portID {
		return fmt.Errorf("sell commodity: ship not docked at port %d", portID)
	}

	// Get commodity definition (validate it exists)
	_, err = e.commodities.GetCommodity(commodityID)
	if err != nil {
		return fmt.Errorf("sell commodity: %w", err)
	}

	// Get port inventory (must exist for port to buy)
	portInv, err := e.db.TxGetPortInventory(tx, portID, commodityID)
	if err != nil {
		return fmt.Errorf("sell commodity: %w", err)
	}
	if portInv == nil {
		return fmt.Errorf("sell commodity: port does not buy this commodity")
	}

	// Check if player has commodity in cargo
	cargoSlot, err := e.db.TxGetCargoSlot(tx, shipID, commodityID)
	if err != nil {
		return fmt.Errorf("sell commodity: %w", err)
	}
	if cargoSlot == nil {
		return fmt.Errorf("sell commodity: %w", ErrCommodityNotInCargo)
	}

	// Validate player has sufficient quantity
	if cargoSlot.Quantity < quantity {
		return fmt.Errorf("sell commodity: insufficient quantity in cargo (have: %d, selling: %d)", 
			cargoSlot.Quantity, quantity)
	}

	// Calculate total revenue (use port's sell price - what player receives)
	totalRevenue := portInv.SellPrice * quantity

	// Get player
	player, err := e.db.TxGetPlayerByID(tx, ship.PlayerID)
	if err != nil {
		return fmt.Errorf("sell commodity: %w", err)
	}
	if player == nil {
		return fmt.Errorf("sell commodity: player not found")
	}

	// All validations passed - perform atomic transaction

	// 1. Add credits to player
	newCredits := int(player.Credits) + totalRevenue
	if err := e.db.TxUpdatePlayerCredits(tx, player.PlayerID, newCredits); err != nil {
		return fmt.Errorf("sell commodity: %w", err)
	}

	// 2. Increase port inventory
	newPortQuantity := portInv.Quantity + quantity
	if err := e.db.TxUpdatePortInventory(tx, portID, commodityID, newPortQuantity, tick); err != nil {
		return fmt.Errorf("sell commodity: %w", err)
	}

	// 3. Remove from ship cargo
	newCargoQuantity := cargoSlot.Quantity - quantity
	if newCargoQuantity == 0 {
		// Remove slot entirely
		if err := e.db.TxRemoveCargo(tx, shipID, cargoSlot.SlotIndex); err != nil {
			return fmt.Errorf("sell commodity: %w", err)
		}
	} else {
		// Update quantity
		if err := e.db.TxUpdateCargoQuantity(tx, shipID, cargoSlot.SlotIndex, newCargoQuantity); err != nil {
			return fmt.Errorf("sell commodity: %w", err)
		}
	}

	// Commit transaction
	if err := e.db.CommitTx(tx); err != nil {
		return fmt.Errorf("sell commodity: %w", err)
	}

	e.logger.Info().
		Str("ship_id", shipID).
		Int("port_id", portID).
		Str("commodity_id", commodityID).
		Int("quantity", quantity).
		Int("total_revenue", totalRevenue).
		Int64("tick", tick).
		Msg("Commodity sold")

	return nil
}

// GetMarketPrices returns current buy/sell prices at a port
func (e *EconomySystem) GetMarketPrices(portID int) ([]*MarketPrice, error) {
	// Get all port inventory
	inventory, err := e.db.GetAllPortInventory(portID)
	if err != nil {
		return nil, fmt.Errorf("get market prices: %w", err)
	}

	// Convert to MarketPrice with commodity names
	prices := make([]*MarketPrice, 0, len(inventory))
	for _, inv := range inventory {
		commodity, err := e.commodities.GetCommodity(inv.CommodityID)
		if err != nil {
			e.logger.Warn().
				Str("commodity_id", inv.CommodityID).
				Err(err).
				Msg("Failed to get commodity definition for market price")
			continue
		}

		prices = append(prices, &MarketPrice{
			CommodityID: inv.CommodityID,
			Name:        commodity.Name,
			BuyPrice:    inv.BuyPrice,
			SellPrice:   inv.SellPrice,
			Quantity:    inv.Quantity,
		})
	}

	return prices, nil
}

// CalculatePrice computes zone-adjusted price for a commodity
// Applies security level multiplier and buy/sell spread according to:
// 1. Determine zone multiplier based on security level:
//    - SecurityLevel = 2.0 (Federated Space) → 1.0x
//    - SecurityLevel 0.7-1.0 (High Security) → 1.0x
//    - SecurityLevel 0.0-0.4 (Low Security) → 1.18x (configurable)
// 2. Apply zone multiplier: zone_price = basePrice × zone_multiplier
// 3. Apply buy/sell spread:
//    - If buying (port sells to player): final_price = zone_price × 1.10
//    - If selling (port buys from player): final_price = zone_price × 0.90
func (e *EconomySystem) CalculatePrice(basePrice int, securityLevel float64, isBuy bool) int {
	// Determine zone multiplier based on security level
	var zoneMultiplier float64
	if securityLevel == 2.0 {
		// Federated Space - base prices
		zoneMultiplier = 1.0
	} else if securityLevel >= 0.7 {
		// High Security - base prices
		zoneMultiplier = 1.0
	} else if securityLevel < 0.4 {
		// Low Security - premium pricing
		zoneMultiplier = e.cfg.LowSecPriceMultiplier
	} else {
		// Medium Security (Phase 2) - base prices
		zoneMultiplier = 1.0
	}

	// Apply zone adjustment
	zonePrice := int(float64(basePrice) * zoneMultiplier)

	// Apply buy/sell spread
	var finalPrice int
	if isBuy {
		// Port sells to player - markup
		finalPrice = int(float64(zonePrice) * e.cfg.BuyMarkup)
	} else {
		// Port buys from player - markdown
		finalPrice = int(float64(zonePrice) * e.cfg.SellMarkdown)
	}

	return finalPrice
}

// GetShipCargo retrieves all cargo slots for a ship
func (e *EconomySystem) GetShipCargo(shipID string) ([]db.CargoSlot, error) {
	cargo, err := e.db.GetShipCargo(shipID)
	if err != nil {
		return nil, fmt.Errorf("get ship cargo: %w", err)
	}
	return cargo, nil
}

// GetCargoManifest returns cargo with commodity names for display
func (e *EconomySystem) GetCargoManifest(shipID string) ([]*CargoManifestEntry, error) {
	cargo, err := e.db.GetShipCargo(shipID)
	if err != nil {
		return nil, fmt.Errorf("get cargo manifest: %w", err)
	}

	manifest := make([]*CargoManifestEntry, 0, len(cargo))
	for _, slot := range cargo {
		commodity, err := e.commodities.GetCommodity(slot.CommodityID)
		if err != nil {
			e.logger.Warn().
				Str("commodity_id", slot.CommodityID).
				Err(err).
				Msg("Failed to get commodity definition for cargo manifest")
			continue
		}

		manifest = append(manifest, &CargoManifestEntry{
			SlotIndex:   slot.SlotIndex,
			CommodityID: slot.CommodityID,
			Name:        commodity.Name,
			Quantity:    slot.Quantity,
		})
	}

	return manifest, nil
}

// GetCargoCapacityInfo returns cargo usage information
func (e *EconomySystem) GetCargoCapacityInfo(shipID string) (*CargoCapacityInfo, error) {
	ship, err := e.db.GetShipByID(shipID)
	if err != nil {
		return nil, fmt.Errorf("get cargo capacity info: %w", err)
	}
	if ship == nil {
		return nil, fmt.Errorf("get cargo capacity info: ship not found")
	}

	currentCargo, err := e.db.GetCargoTotalQuantity(shipID)
	if err != nil {
		return nil, fmt.Errorf("get cargo capacity info: %w", err)
	}

	return &CargoCapacityInfo{
		CurrentCargo: currentCargo,
		MaxCargo:     ship.CargoCapacity,
		Available:    ship.CargoCapacity - currentCargo,
	}, nil
}
