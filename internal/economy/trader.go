package economy

// Trader defines the interface for economy and trading operations
type Trader interface {
	// BuyCommodity purchases commodity from port to ship cargo
	// Validates credits, cargo capacity, and port inventory before transaction
	BuyCommodity(shipID string, portID int, commodityID string, quantity int, tick int64) error

	// SellCommodity sells commodity from ship cargo to port
	// Validates player has commodity in cargo before transaction
	SellCommodity(shipID string, portID int, commodityID string, quantity int, tick int64) error

	// GetMarketPrices returns current buy/sell prices at a port
	// Filters commodities based on port type (trading/mining/refueling)
	GetMarketPrices(portID int) ([]*MarketPrice, error)

	// CalculatePrice computes zone-adjusted price for a commodity
	// Applies security level multiplier and buy/sell spread
	CalculatePrice(basePrice int, securityLevel float64, isBuy bool) int
}
