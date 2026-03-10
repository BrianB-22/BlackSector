package economy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testLogger returns a no-op logger for testing
func testLogger() zerolog.Logger {
	return zerolog.Nop()
}

func TestCommodityRegistry_LoadFromFile(t *testing.T) {
	tests := []struct {
		name        string
		configJSON  string
		expectError bool
		expectCount int
	}{
		{
			name: "valid_7_commodities",
			configJSON: `{
				"commodities": [
					{"commodity_id": "food_supplies", "name": "Food Supplies", "category": "essential", "base_price": 100, "unit": "crate"},
					{"commodity_id": "fuel_cells", "name": "Fuel Cells", "category": "essential", "base_price": 150, "unit": "cell"},
					{"commodity_id": "raw_ore", "name": "Raw Ore", "category": "industrial", "base_price": 80, "unit": "ton"},
					{"commodity_id": "refined_ore", "name": "Refined Ore", "category": "industrial", "base_price": 240, "unit": "ton"},
					{"commodity_id": "machinery", "name": "Machinery", "category": "industrial", "base_price": 600, "unit": "unit"},
					{"commodity_id": "electronics", "name": "Electronics", "category": "industrial", "base_price": 800, "unit": "unit"},
					{"commodity_id": "luxury_goods", "name": "Luxury Goods", "category": "luxury", "base_price": 1500, "unit": "crate"}
				]
			}`,
			expectError: false,
			expectCount: 7,
		},
		{
			name: "empty_commodities",
			configJSON: `{
				"commodities": []
			}`,
			expectError: false,
			expectCount: 0,
		},
		{
			name: "invalid_json",
			configJSON: `{
				"commodities": [
					{"commodity_id": "food_supplies", "name": "Food Supplies"
				]
			}`,
			expectError: true,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "test_config.json")
			err := os.WriteFile(configPath, []byte(tt.configJSON), 0644)
			require.NoError(t, err)

			// Load commodities
			registry := NewCommodityRegistry()
			err = registry.LoadFromFile(configPath)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectCount, registry.Count())
			}
		})
	}
}

func TestCommodityRegistry_GetCommodity(t *testing.T) {
	// Setup registry with test commodities
	registry := NewCommodityRegistry()
	registry.commodities["food_supplies"] = &CommodityDefinition{
		CommodityID: "food_supplies",
		Name:        "Food Supplies",
		Category:    "essential",
		BasePrice:   100,
		Unit:        "crate",
	}
	registry.commodities["fuel_cells"] = &CommodityDefinition{
		CommodityID: "fuel_cells",
		Name:        "Fuel Cells",
		Category:    "essential",
		BasePrice:   150,
		Unit:        "cell",
	}

	tests := []struct {
		name        string
		commodityID string
		expectError bool
		expectName  string
		expectPrice int
	}{
		{
			name:        "valid_food_supplies",
			commodityID: "food_supplies",
			expectError: false,
			expectName:  "Food Supplies",
			expectPrice: 100,
		},
		{
			name:        "valid_fuel_cells",
			commodityID: "fuel_cells",
			expectError: false,
			expectName:  "Fuel Cells",
			expectPrice: 150,
		},
		{
			name:        "invalid_commodity",
			commodityID: "invalid_commodity",
			expectError: true,
		},
		{
			name:        "empty_commodity_id",
			commodityID: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commodity, err := registry.GetCommodity(tt.commodityID)

			if tt.expectError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidCommodity)
				assert.Nil(t, commodity)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, commodity)
				assert.Equal(t, tt.expectName, commodity.Name)
				assert.Equal(t, tt.expectPrice, commodity.BasePrice)
			}
		})
	}
}

func TestCommodityRegistry_GetAllCommodities(t *testing.T) {
	tests := []struct {
		name          string
		commodities   map[string]*CommodityDefinition
		expectedCount int
	}{
		{
			name: "multiple_commodities",
			commodities: map[string]*CommodityDefinition{
				"food_supplies": {
					CommodityID: "food_supplies",
					Name:        "Food Supplies",
					Category:    "essential",
					BasePrice:   100,
				},
				"fuel_cells": {
					CommodityID: "fuel_cells",
					Name:        "Fuel Cells",
					Category:    "essential",
					BasePrice:   150,
				},
				"raw_ore": {
					CommodityID: "raw_ore",
					Name:        "Raw Ore",
					Category:    "industrial",
					BasePrice:   80,
				},
			},
			expectedCount: 3,
		},
		{
			name:          "empty_registry",
			commodities:   map[string]*CommodityDefinition{},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewCommodityRegistry()
			registry.commodities = tt.commodities

			commodities := registry.GetAllCommodities()

			assert.Len(t, commodities, tt.expectedCount)

			// Verify all commodities are present
			commodityIDs := make(map[string]bool)
			for _, commodity := range commodities {
				commodityIDs[commodity.CommodityID] = true
			}

			for expectedID := range tt.commodities {
				assert.True(t, commodityIDs[expectedID], "expected commodity %s not found", expectedID)
			}
		})
	}
}

func TestCommodityRegistry_Count(t *testing.T) {
	registry := NewCommodityRegistry()

	// Initially empty
	assert.Equal(t, 0, registry.Count())

	// Add commodities
	registry.commodities["food_supplies"] = &CommodityDefinition{
		CommodityID: "food_supplies",
		Name:        "Food Supplies",
		BasePrice:   100,
	}
	assert.Equal(t, 1, registry.Count())

	registry.commodities["fuel_cells"] = &CommodityDefinition{
		CommodityID: "fuel_cells",
		Name:        "Fuel Cells",
		BasePrice:   150,
	}
	assert.Equal(t, 2, registry.Count())
}

func TestEconomySystem_LoadCommodities(t *testing.T) {
	// Create temporary config file with all 7 commodities
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "alpha_sector.json")
	configJSON := `{
		"commodities": [
			{"commodity_id": "food_supplies", "name": "Food Supplies", "category": "essential", "base_price": 100, "unit": "crate"},
			{"commodity_id": "fuel_cells", "name": "Fuel Cells", "category": "essential", "base_price": 150, "unit": "cell"},
			{"commodity_id": "raw_ore", "name": "Raw Ore", "category": "industrial", "base_price": 80, "unit": "ton"},
			{"commodity_id": "refined_ore", "name": "Refined Ore", "category": "industrial", "base_price": 240, "unit": "ton"},
			{"commodity_id": "machinery", "name": "Machinery", "category": "industrial", "base_price": 600, "unit": "unit"},
			{"commodity_id": "electronics", "name": "Electronics", "category": "industrial", "base_price": 800, "unit": "unit"},
			{"commodity_id": "luxury_goods", "name": "Luxury Goods", "category": "luxury", "base_price": 1500, "unit": "crate"}
		]
	}`
	err := os.WriteFile(configPath, []byte(configJSON), 0644)
	require.NoError(t, err)

	// Create economy system
	cfg := &Config{
		LowSecPriceMultiplier: 1.18,
		BuyMarkup:             1.10,
		SellMarkdown:          0.90,
	}
	economy := NewEconomySystem(cfg, nil, testLogger())

	// Load commodities
	err = economy.LoadCommodities(configPath)
	assert.NoError(t, err)

	// Verify all 7 commodities are loaded
	commodities := economy.GetAllCommodities()
	assert.Len(t, commodities, 7)

	// Verify specific commodities
	foodSupplies, err := economy.GetCommodity("food_supplies")
	assert.NoError(t, err)
	assert.Equal(t, "Food Supplies", foodSupplies.Name)
	assert.Equal(t, 100, foodSupplies.BasePrice)

	luxuryGoods, err := economy.GetCommodity("luxury_goods")
	assert.NoError(t, err)
	assert.Equal(t, "Luxury Goods", luxuryGoods.Name)
	assert.Equal(t, 1500, luxuryGoods.BasePrice)
}

func TestEconomySystem_GetCommodity(t *testing.T) {
	// Create economy system with test commodities
	economy := NewEconomySystem(&Config{}, nil, testLogger())
	economy.commodities.commodities["food_supplies"] = &CommodityDefinition{
		CommodityID: "food_supplies",
		Name:        "Food Supplies",
		Category:    "essential",
		BasePrice:   100,
	}

	tests := []struct {
		name        string
		commodityID string
		expectError bool
	}{
		{
			name:        "valid_commodity",
			commodityID: "food_supplies",
			expectError: false,
		},
		{
			name:        "invalid_commodity",
			commodityID: "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commodity, err := economy.GetCommodity(tt.commodityID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, commodity)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, commodity)
			}
		})
	}
}

func TestEconomySystem_GetAllCommodities(t *testing.T) {
	economy := NewEconomySystem(&Config{}, nil, testLogger())

	// Initially empty
	commodities := economy.GetAllCommodities()
	assert.Len(t, commodities, 0)

	// Add commodities
	economy.commodities.commodities["food_supplies"] = &CommodityDefinition{
		CommodityID: "food_supplies",
		Name:        "Food Supplies",
		BasePrice:   100,
	}
	economy.commodities.commodities["fuel_cells"] = &CommodityDefinition{
		CommodityID: "fuel_cells",
		Name:        "Fuel Cells",
		BasePrice:   150,
	}

	commodities = economy.GetAllCommodities()
	assert.Len(t, commodities, 2)
}

func TestCommodityDefinitions_Phase1Requirements(t *testing.T) {
	// Test that all 7 Phase 1 commodities are defined with correct base prices
	expectedCommodities := map[string]int{
		"food_supplies": 100,
		"fuel_cells":    150,
		"raw_ore":       80,
		"refined_ore":   240,
		"machinery":     600,
		"electronics":   800,
		"luxury_goods":  1500,
	}

	// Load from actual alpha_sector.json
	registry := NewCommodityRegistry()
	err := registry.LoadFromFile("../../config/world/alpha_sector.json")
	require.NoError(t, err)

	// Verify count
	assert.Equal(t, 7, registry.Count(), "Phase 1 requires exactly 7 commodities")

	// Verify each commodity
	for commodityID, expectedPrice := range expectedCommodities {
		commodity, err := registry.GetCommodity(commodityID)
		assert.NoError(t, err, "commodity %s should exist", commodityID)
		if commodity != nil {
			assert.Equal(t, expectedPrice, commodity.BasePrice,
				"commodity %s should have base price %d", commodityID, expectedPrice)
		}
	}
}
