package economy

import (
	"encoding/json"
	"fmt"
	"os"
)

// CommodityDefinition represents a commodity loaded from world configuration
type CommodityDefinition struct {
	CommodityID string `json:"commodity_id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	BasePrice   int    `json:"base_price"`
	Unit        string `json:"unit"`
}

// WorldConfig represents the structure of alpha_sector.json
type WorldConfig struct {
	Commodities []CommodityDefinition `json:"commodities"`
}

// CommodityRegistry holds all commodity definitions
type CommodityRegistry struct {
	commodities map[string]*CommodityDefinition
}

// NewCommodityRegistry creates a new commodity registry
func NewCommodityRegistry() *CommodityRegistry {
	return &CommodityRegistry{
		commodities: make(map[string]*CommodityDefinition),
	}
}

// LoadFromFile loads commodity definitions from a world configuration file
func (r *CommodityRegistry) LoadFromFile(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("read world config: %w", err)
	}

	var config WorldConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("parse world config: %w", err)
	}

	for i := range config.Commodities {
		commodity := &config.Commodities[i]
		r.commodities[commodity.CommodityID] = commodity
	}

	return nil
}

// GetCommodity retrieves a commodity by ID
func (r *CommodityRegistry) GetCommodity(commodityID string) (*CommodityDefinition, error) {
	commodity, exists := r.commodities[commodityID]
	if !exists {
		return nil, fmt.Errorf("commodity %s: %w", commodityID, ErrInvalidCommodity)
	}
	return commodity, nil
}

// GetAllCommodities returns all registered commodities
func (r *CommodityRegistry) GetAllCommodities() []*CommodityDefinition {
	commodities := make([]*CommodityDefinition, 0, len(r.commodities))
	for _, commodity := range r.commodities {
		commodities = append(commodities, commodity)
	}
	return commodities
}

// Count returns the number of registered commodities
func (r *CommodityRegistry) Count() int {
	return len(r.commodities)
}
