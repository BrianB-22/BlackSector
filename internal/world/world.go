package world

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog"
)

// Generator defines the interface for loading and managing the static universe
type Generator interface {
	// LoadWorld loads the static alpha_sector.json world configuration
	LoadWorld(configPath string) (*Universe, error)

	// ValidateTopology ensures all systems are reachable and jump connections are valid
	ValidateTopology(u *Universe) error
}

// Universe represents the complete game world with all regions, systems, ports, and connections
type Universe struct {
	Regions         map[string]*Region
	Systems         map[string]*System
	Ports           map[string]*Port
	JumpConnections []*JumpConnection

	// Cache for quick lookups
	mu                sync.RWMutex
	systemsByRegion   map[string][]*System
	portsBySystem     map[string][]*Port
	connectionsByFrom map[string][]*JumpConnection
}

// Region represents a collection of star systems with shared characteristics
type Region struct {
	RegionID      string  `json:"region_id"`
	Name          string  `json:"name"`
	RegionType    string  `json:"region_type"` // "core", "industrial"
	SecurityLevel float64 `json:"security_level"`
}

// System represents a star system in the universe
type System struct {
	SystemID      string  `json:"system_id"`
	Name          string  `json:"name"`
	RegionID      string  `json:"region_id,omitempty"`
	SecurityLevel float64 `json:"security_rating"` // Note: JSON uses "security_rating"
	SecurityZone  string  `json:"security_zone"`   // "federated", "high", "low"
	PositionX     float64 `json:"x"`
	PositionY     float64 `json:"y"`
	Description   string  `json:"description,omitempty"`
}

// Port represents a trading station or starbase in a system
type Port struct {
	PortID           string         `json:"port_id"`
	SystemID         string         `json:"system_id"`
	Name             string         `json:"name"`
	PortType         string         `json:"port_type"` // "trading", "mining", "refueling"
	Description      string         `json:"description,omitempty"`
	Services         *PortServices  `json:"services,omitempty"`
	Commodities      *PortCommodityConfig `json:"commodities,omitempty"`
}

// PortServices defines what services are available at a port
type PortServices struct {
	HasBank          bool    `json:"has_bank"`
	InterestRate     float64 `json:"interest_rate_percent"`
	HasShipyard      bool    `json:"has_shipyard"`
	HasUpgradeMarket bool    `json:"has_upgrade_market"`
	HasDroneMarket   bool    `json:"has_drone_market"`
	HasMissileSupply bool    `json:"has_missile_supply"`
	HasRepair        bool    `json:"has_repair"`
	HasFuel          bool    `json:"has_fuel"`
}

// PortCommodityConfig defines which commodities a port produces/consumes
type PortCommodityConfig struct {
	Produces []string `json:"produces"`
	Consumes []string `json:"consumes"`
}

// JumpConnection represents a navigable route between two systems
type JumpConnection struct {
	FromSystemID     string  `json:"from_system_id"`
	ToSystemID       string  `json:"to_system_id"`
	FuelCost         int     `json:"fuel_cost"`
}

// PortCommodity represents commodity inventory at a port
type PortCommodity struct {
	PortID      string `json:"port_id"`
	CommodityID string `json:"commodity_id"`
	Quantity    int    `json:"quantity"`
	BuyPrice    int    `json:"buy_price"`
	SellPrice   int    `json:"sell_price"`
	UpdatedTick int64  `json:"updated_tick"`
}

// Commodity represents a tradable good
type Commodity struct {
	CommodityID string  `json:"commodity_id"`
	Name        string  `json:"name"`
	Category    string  `json:"category"` // "essential", "industrial", "luxury"
	BasePrice   int     `json:"base_price"`
	Unit        string  `json:"unit"`
}

// WorldGenerator implements the Generator interface for loading static universe data
type WorldGenerator struct {
	logger zerolog.Logger
}

// NewWorldGenerator creates a new world generator instance
func NewWorldGenerator(logger zerolog.Logger) *WorldGenerator {
	return &WorldGenerator{
		logger: logger,
	}
}

// GetOriginSystem returns the Federated Space origin system (security_zone = "federated")
// This is where new players start their journey
func (u *Universe) GetOriginSystem() (*System, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	// Find first system with security_zone = "federated"
	for _, system := range u.Systems {
		if system.SecurityZone == "federated" {
			return system, nil
		}
	}

	return nil, fmt.Errorf("no federated space origin system found in universe")
}

// GetSystemPorts returns all ports in a given system
func (u *Universe) GetSystemPorts(systemID string) []*Port {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if ports, ok := u.portsBySystem[systemID]; ok {
		return ports
	}
	return nil
}
