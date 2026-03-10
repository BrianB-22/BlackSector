// Package economy implements the trading and commodity management system for BlackSector.
//
// The economy system handles:
//   - Commodity trading (buy/sell operations)
//   - Price calculations with zone multipliers and buy/sell spreads
//   - Cargo management
//   - Port inventory tracking
//   - Market price display
//
// Phase 1 Implementation:
//   - 7 commodities: food_supplies, fuel_cells, raw_ore, refined_ore, machinery, electronics, luxury_goods
//   - Static base prices (no dynamic pricing)
//   - Zone price multipliers: Federated Space = 1.0x, High Security = 1.0x, Low Security = 1.18x
//   - Buy/sell spread: buy price = zone_price × 1.10, sell price = zone_price × 0.90
//   - Port type restrictions: trading/mining/refueling ports stock different commodities
//
// The EconomySystem struct provides the concrete implementation of the Trader interface.
// All trading operations are atomic and validate credits, cargo capacity, and inventory
// before committing transactions to the database.
package economy
