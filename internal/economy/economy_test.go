package economy

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestCalculatePrice(t *testing.T) {
	tests := []struct {
		name          string
		basePrice     int
		securityLevel float64
		isBuy         bool
		expected      int
	}{
		// Federated Space (SecurityLevel = 2.0) tests
		{
			name:          "fed_space_buy",
			basePrice:     100,
			securityLevel: 2.0,
			isBuy:         true,
			expected:      110, // 100 × 1.0 × 1.10 = 110
		},
		{
			name:          "fed_space_sell",
			basePrice:     100,
			securityLevel: 2.0,
			isBuy:         false,
			expected:      90, // 100 × 1.0 × 0.90 = 90
		},

		// High Security (0.7-1.0) tests
		{
			name:          "high_sec_buy_upper_bound",
			basePrice:     100,
			securityLevel: 1.0,
			isBuy:         true,
			expected:      110, // 100 × 1.0 × 1.10 = 110
		},
		{
			name:          "high_sec_sell_upper_bound",
			basePrice:     100,
			securityLevel: 1.0,
			isBuy:         false,
			expected:      90, // 100 × 1.0 × 0.90 = 90
		},
		{
			name:          "high_sec_buy_mid",
			basePrice:     100,
			securityLevel: 0.8,
			isBuy:         true,
			expected:      110, // 100 × 1.0 × 1.10 = 110
		},
		{
			name:          "high_sec_sell_mid",
			basePrice:     100,
			securityLevel: 0.8,
			isBuy:         false,
			expected:      90, // 100 × 1.0 × 0.90 = 90
		},
		{
			name:          "high_sec_buy_lower_bound",
			basePrice:     100,
			securityLevel: 0.7,
			isBuy:         true,
			expected:      110, // 100 × 1.0 × 1.10 = 110
		},
		{
			name:          "high_sec_sell_lower_bound",
			basePrice:     100,
			securityLevel: 0.7,
			isBuy:         false,
			expected:      90, // 100 × 1.0 × 0.90 = 90
		},

		// Low Security (0.0-0.4) tests
		{
			name:          "low_sec_buy",
			basePrice:     100,
			securityLevel: 0.2,
			isBuy:         true,
			expected:      129, // 100 × 1.18 × 1.10 = 129.8 → 129
		},
		{
			name:          "low_sec_sell",
			basePrice:     100,
			securityLevel: 0.2,
			isBuy:         false,
			expected:      106, // 100 × 1.18 × 0.90 = 106.2 → 106
		},
		{
			name:          "low_sec_buy_zero_security",
			basePrice:     100,
			securityLevel: 0.0,
			isBuy:         true,
			expected:      129, // 100 × 1.18 × 1.10 = 129.8 → 129
		},
		{
			name:          "low_sec_sell_zero_security",
			basePrice:     100,
			securityLevel: 0.0,
			isBuy:         false,
			expected:      106, // 100 × 1.18 × 0.90 = 106.2 → 106
		},
		{
			name:          "low_sec_buy_upper_bound",
			basePrice:     100,
			securityLevel: 0.39,
			isBuy:         true,
			expected:      129, // 100 × 1.18 × 1.10 = 129.8 → 129
		},
		{
			name:          "low_sec_sell_upper_bound",
			basePrice:     100,
			securityLevel: 0.39,
			isBuy:         false,
			expected:      106, // 100 × 1.18 × 0.90 = 106.2 → 106
		},

		// Medium Security (0.4-0.7) - Phase 2, treated as base prices in Phase 1
		{
			name:          "medium_sec_buy_lower_bound",
			basePrice:     100,
			securityLevel: 0.4,
			isBuy:         true,
			expected:      110, // 100 × 1.0 × 1.10 = 110
		},
		{
			name:          "medium_sec_sell_lower_bound",
			basePrice:     100,
			securityLevel: 0.4,
			isBuy:         false,
			expected:      90, // 100 × 1.0 × 0.90 = 90
		},
		{
			name:          "medium_sec_buy_mid",
			basePrice:     100,
			securityLevel: 0.5,
			isBuy:         true,
			expected:      110, // 100 × 1.0 × 1.10 = 110
		},
		{
			name:          "medium_sec_sell_mid",
			basePrice:     100,
			securityLevel: 0.5,
			isBuy:         false,
			expected:      90, // 100 × 1.0 × 0.90 = 90
		},
		{
			name:          "medium_sec_buy_upper_bound",
			basePrice:     100,
			securityLevel: 0.69,
			isBuy:         true,
			expected:      110, // 100 × 1.0 × 1.10 = 110
		},
		{
			name:          "medium_sec_sell_upper_bound",
			basePrice:     100,
			securityLevel: 0.69,
			isBuy:         false,
			expected:      90, // 100 × 1.0 × 0.90 = 90
		},

		// Different base prices
		{
			name:          "low_price_commodity_low_sec_buy",
			basePrice:     50,
			securityLevel: 0.2,
			isBuy:         true,
			expected:      64, // 50 × 1.18 × 1.10 = 64.9 → 64
		},
		{
			name:          "high_price_commodity_low_sec_buy",
			basePrice:     500,
			securityLevel: 0.2,
			isBuy:         true,
			expected:      649, // 500 × 1.18 × 1.10 = 649 → 649
		},
		{
			name:          "low_price_commodity_fed_space_sell",
			basePrice:     50,
			securityLevel: 2.0,
			isBuy:         false,
			expected:      45, // 50 × 1.0 × 0.90 = 45
		},
		{
			name:          "high_price_commodity_fed_space_sell",
			basePrice:     500,
			securityLevel: 2.0,
			isBuy:         false,
			expected:      450, // 500 × 1.0 × 0.90 = 450
		},

		// Edge cases
		{
			name:          "minimum_price_buy",
			basePrice:     1,
			securityLevel: 2.0,
			isBuy:         true,
			expected:      1, // 1 × 1.0 × 1.10 = 1.1 → 1
		},
		{
			name:          "minimum_price_sell",
			basePrice:     1,
			securityLevel: 2.0,
			isBuy:         false,
			expected:      0, // 1 × 1.0 × 0.90 = 0.9 → 0
		},
		{
			name:          "large_price_low_sec_buy",
			basePrice:     10000,
			securityLevel: 0.1,
			isBuy:         true,
			expected:      12980, // 10000 × 1.18 × 1.10 = 12980
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				LowSecPriceMultiplier: 1.18,
				BuyMarkup:             1.10,
				SellMarkdown:          0.90,
			}
			economy := NewEconomySystem(cfg, nil, zerolog.Nop())

			result := economy.CalculatePrice(tt.basePrice, tt.securityLevel, tt.isBuy)
			assert.Equal(t, tt.expected, result, "Price calculation mismatch")
		})
	}
}

func TestCalculatePriceWithCustomConfig(t *testing.T) {
	tests := []struct {
		name                  string
		lowSecMultiplier      float64
		buyMarkup             float64
		sellMarkdown          float64
		basePrice             int
		securityLevel         float64
		isBuy                 bool
		expected              int
	}{
		{
			name:             "custom_multiplier_higher",
			lowSecMultiplier: 1.25,
			buyMarkup:        1.10,
			sellMarkdown:     0.90,
			basePrice:        100,
			securityLevel:    0.2,
			isBuy:            true,
			expected:         137, // 100 × 1.25 × 1.10 = 137.5 → 137
		},
		{
			name:             "custom_spread_wider",
			lowSecMultiplier: 1.18,
			buyMarkup:        1.20,
			sellMarkdown:     0.80,
			basePrice:        100,
			securityLevel:    2.0,
			isBuy:            true,
			expected:         120, // 100 × 1.0 × 1.20 = 120
		},
		{
			name:             "custom_spread_wider_sell",
			lowSecMultiplier: 1.18,
			buyMarkup:        1.20,
			sellMarkdown:     0.80,
			basePrice:        100,
			securityLevel:    2.0,
			isBuy:            false,
			expected:         80, // 100 × 1.0 × 0.80 = 80
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				LowSecPriceMultiplier: tt.lowSecMultiplier,
				BuyMarkup:             tt.buyMarkup,
				SellMarkdown:          tt.sellMarkdown,
			}
			economy := NewEconomySystem(cfg, nil, zerolog.Nop())

			result := economy.CalculatePrice(tt.basePrice, tt.securityLevel, tt.isBuy)
			assert.Equal(t, tt.expected, result, "Price calculation with custom config mismatch")
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	
	assert.Equal(t, 1.18, cfg.LowSecPriceMultiplier, "Default low sec multiplier should be 1.18")
	assert.Equal(t, 1.10, cfg.BuyMarkup, "Default buy markup should be 1.10")
	assert.Equal(t, 0.90, cfg.SellMarkdown, "Default sell markdown should be 0.90")
}
