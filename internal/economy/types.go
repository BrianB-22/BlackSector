package economy

// Commodity represents a tradable good in the game
type Commodity struct {
	CommodityID  string  // "food_supplies", "fuel_cells", etc.
	Name         string  // Display name
	Category     string  // "essential", "industrial", "luxury"
	BasePrice    int     // Base price before zone adjustments
	Volatility   float64 // Price volatility (not used in Phase 1)
	IsContraband bool    // Whether commodity is illegal
}

// MarketPrice represents the current buy/sell prices at a port
type MarketPrice struct {
	CommodityID string // Commodity identifier
	Name        string // Display name
	BuyPrice    int    // Port sells to player at this price
	SellPrice   int    // Port buys from player at this price
	Quantity    int    // Available stock at port
}

// CargoManifestEntry represents a cargo slot with commodity name for display
type CargoManifestEntry struct {
	SlotIndex   int
	CommodityID string
	Name        string
	Quantity    int
}

// CargoCapacityInfo represents cargo usage information
type CargoCapacityInfo struct {
	CurrentCargo int
	MaxCargo     int
	Available    int
}
