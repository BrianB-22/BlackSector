package tui

import (
	"encoding/json"
	"testing"

	"github.com/BrianB-22/BlackSector/internal/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCommand_Jump(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
		wantType    string
		wantPayload engine.JumpPayload
	}{
		{
			name:     "valid jump command",
			input:    "jump 5",
			wantErr:  false,
			wantType: "jump",
			wantPayload: engine.JumpPayload{
				TargetSystemID: 5,
			},
		},
		{
			name:     "valid jump with uppercase",
			input:    "JUMP 10",
			wantErr:  false,
			wantType: "jump",
			wantPayload: engine.JumpPayload{
				TargetSystemID: 10,
			},
		},
		{
			name:     "valid jump with extra whitespace",
			input:    "  jump   15  ",
			wantErr:  false,
			wantType: "jump",
			wantPayload: engine.JumpPayload{
				TargetSystemID: 15,
			},
		},
		{
			name:        "jump with no arguments",
			input:       "jump",
			wantErr:     true,
			errContains: "requires exactly one argument",
		},
		{
			name:        "jump with too many arguments",
			input:       "jump 5 10",
			wantErr:     true,
			errContains: "requires exactly one argument",
		},
		{
			name:        "jump with non-integer argument",
			input:       "jump abc",
			wantErr:     true,
			errContains: "must be an integer",
		},
		{
			name:        "jump with negative system_id",
			input:       "jump -5",
			wantErr:     true,
			errContains: "must be positive",
		},
		{
			name:        "jump with zero system_id",
			input:       "jump 0",
			wantErr:     true,
			errContains: "must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := ParseCommand(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantType, cmd.Type)
			assert.False(t, cmd.IsLocal)

			// Unmarshal and verify payload
			var payload engine.JumpPayload
			err = json.Unmarshal(cmd.Payload, &payload)
			require.NoError(t, err)
			assert.Equal(t, tt.wantPayload, payload)
		})
	}
}

func TestParseCommand_Dock(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
		wantType    string
		wantPayload engine.DockPayload
	}{
		{
			name:     "valid dock command",
			input:    "dock 3",
			wantErr:  false,
			wantType: "dock",
			wantPayload: engine.DockPayload{
				PortID: 3,
			},
		},
		{
			name:     "valid dock with uppercase",
			input:    "DOCK 7",
			wantErr:  false,
			wantType: "dock",
			wantPayload: engine.DockPayload{
				PortID: 7,
			},
		},
		{
			name:        "dock with no arguments",
			input:       "dock",
			wantErr:     true,
			errContains: "requires exactly one argument",
		},
		{
			name:        "dock with too many arguments",
			input:       "dock 3 5",
			wantErr:     true,
			errContains: "requires exactly one argument",
		},
		{
			name:        "dock with non-integer argument",
			input:       "dock starbase",
			wantErr:     true,
			errContains: "must be an integer",
		},
		{
			name:        "dock with negative port_id",
			input:       "dock -3",
			wantErr:     true,
			errContains: "must be positive",
		},
		{
			name:        "dock with zero port_id",
			input:       "dock 0",
			wantErr:     true,
			errContains: "must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := ParseCommand(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantType, cmd.Type)
			assert.False(t, cmd.IsLocal)

			// Unmarshal and verify payload
			var payload engine.DockPayload
			err = json.Unmarshal(cmd.Payload, &payload)
			require.NoError(t, err)
			assert.Equal(t, tt.wantPayload, payload)
		})
	}
}

func TestParseCommand_Undock(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
		wantType    string
	}{
		{
			name:     "valid undock command",
			input:    "undock",
			wantErr:  false,
			wantType: "undock",
		},
		{
			name:     "valid undock with uppercase",
			input:    "UNDOCK",
			wantErr:  false,
			wantType: "undock",
		},
		{
			name:        "undock with arguments",
			input:       "undock now",
			wantErr:     true,
			errContains: "takes no arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := ParseCommand(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantType, cmd.Type)
			assert.False(t, cmd.IsLocal)

			// Verify payload is valid JSON (empty object)
			var payload engine.UndockPayload
			err = json.Unmarshal(cmd.Payload, &payload)
			require.NoError(t, err)
		})
	}
}

func TestParseCommand_Buy(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
		wantType    string
		commodity   string
		quantity    int
	}{
		{
			name:      "valid buy command",
			input:     "buy food_supplies 10",
			wantErr:   false,
			wantType:  "buy",
			commodity: "food_supplies",
			quantity:  10,
		},
		{
			name:      "valid buy with uppercase",
			input:     "BUY FUEL_CELLS 5",
			wantErr:   false,
			wantType:  "buy",
			commodity: "fuel_cells",
			quantity:  5,
		},
		{
			name:      "valid buy all commodities",
			input:     "buy raw_ore 1",
			wantErr:   false,
			wantType:  "buy",
			commodity: "raw_ore",
			quantity:  1,
		},
		{
			name:      "valid buy refined_ore",
			input:     "buy refined_ore 20",
			wantErr:   false,
			wantType:  "buy",
			commodity: "refined_ore",
			quantity:  20,
		},
		{
			name:      "valid buy machinery",
			input:     "buy machinery 15",
			wantErr:   false,
			wantType:  "buy",
			commodity: "machinery",
			quantity:  15,
		},
		{
			name:      "valid buy electronics",
			input:     "buy electronics 8",
			wantErr:   false,
			wantType:  "buy",
			commodity: "electronics",
			quantity:  8,
		},
		{
			name:      "valid buy luxury_goods",
			input:     "buy luxury_goods 3",
			wantErr:   false,
			wantType:  "buy",
			commodity: "luxury_goods",
			quantity:  3,
		},
		{
			name:        "buy with no arguments",
			input:       "buy",
			wantErr:     true,
			errContains: "requires two arguments",
		},
		{
			name:        "buy with one argument",
			input:       "buy food_supplies",
			wantErr:     true,
			errContains: "requires two arguments",
		},
		{
			name:        "buy with too many arguments",
			input:       "buy food_supplies 10 extra",
			wantErr:     true,
			errContains: "requires two arguments",
		},
		{
			name:        "buy with invalid commodity",
			input:       "buy invalid_commodity 10",
			wantErr:     true,
			errContains: "invalid commodity",
		},
		{
			name:        "buy with non-integer quantity",
			input:       "buy food_supplies abc",
			wantErr:     true,
			errContains: "must be an integer",
		},
		{
			name:        "buy with negative quantity",
			input:       "buy food_supplies -5",
			wantErr:     true,
			errContains: "must be positive",
		},
		{
			name:        "buy with zero quantity",
			input:       "buy food_supplies 0",
			wantErr:     true,
			errContains: "must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := ParseCommand(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantType, cmd.Type)
			assert.False(t, cmd.IsLocal)

			// Unmarshal and verify payload
			var payload engine.BuyPayload
			err = json.Unmarshal(cmd.Payload, &payload)
			require.NoError(t, err)
			assert.Equal(t, tt.commodity, payload.CommodityID)
			assert.Equal(t, tt.quantity, payload.Quantity)
			// PortID should be 0 (will be filled by session layer)
			assert.Equal(t, 0, payload.PortID)
		})
	}
}

func TestParseCommand_Sell(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
		wantType    string
		commodity   string
		quantity    int
	}{
		{
			name:      "valid sell command",
			input:     "sell food_supplies 10",
			wantErr:   false,
			wantType:  "sell",
			commodity: "food_supplies",
			quantity:  10,
		},
		{
			name:      "valid sell with uppercase",
			input:     "SELL FUEL_CELLS 5",
			wantErr:   false,
			wantType:  "sell",
			commodity: "fuel_cells",
			quantity:  5,
		},
		{
			name:        "sell with no arguments",
			input:       "sell",
			wantErr:     true,
			errContains: "requires two arguments",
		},
		{
			name:        "sell with one argument",
			input:       "sell food_supplies",
			wantErr:     true,
			errContains: "requires two arguments",
		},
		{
			name:        "sell with invalid commodity",
			input:       "sell invalid_commodity 10",
			wantErr:     true,
			errContains: "invalid commodity",
		},
		{
			name:        "sell with non-integer quantity",
			input:       "sell food_supplies abc",
			wantErr:     true,
			errContains: "must be an integer",
		},
		{
			name:        "sell with negative quantity",
			input:       "sell food_supplies -5",
			wantErr:     true,
			errContains: "must be positive",
		},
		{
			name:        "sell with zero quantity",
			input:       "sell food_supplies 0",
			wantErr:     true,
			errContains: "must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := ParseCommand(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantType, cmd.Type)
			assert.False(t, cmd.IsLocal)

			// Unmarshal and verify payload
			var payload engine.SellPayload
			err = json.Unmarshal(cmd.Payload, &payload)
			require.NoError(t, err)
			assert.Equal(t, tt.commodity, payload.CommodityID)
			assert.Equal(t, tt.quantity, payload.Quantity)
			// PortID should be 0 (will be filled by session layer)
			assert.Equal(t, 0, payload.PortID)
		})
	}
}

func TestParseCommand_LocalCommands(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
		wantType    string
		wantLocal   bool
	}{
		{
			name:      "valid system command",
			input:     "system",
			wantErr:   false,
			wantType:  "system",
			wantLocal: true,
		},
		{
			name:      "valid market command",
			input:     "market",
			wantErr:   false,
			wantType:  "market",
			wantLocal: true,
		},
		{
			name:      "valid cargo command",
			input:     "cargo",
			wantErr:   false,
			wantType:  "cargo",
			wantLocal: true,
		},
		{
			name:      "valid help command",
			input:     "help",
			wantErr:   false,
			wantType:  "help",
			wantLocal: true,
		},
		{
			name:        "system with arguments",
			input:       "system 5",
			wantErr:     true,
			errContains: "takes no arguments",
		},
		{
			name:        "market with arguments",
			input:       "market all",
			wantErr:     true,
			errContains: "takes no arguments",
		},
		{
			name:        "cargo with arguments",
			input:       "cargo list",
			wantErr:     true,
			errContains: "takes no arguments",
		},
		{
			name:        "help with arguments",
			input:       "help me",
			wantErr:     true,
			errContains: "takes no arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := ParseCommand(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantType, cmd.Type)
			assert.Equal(t, tt.wantLocal, cmd.IsLocal)
			assert.Nil(t, cmd.Payload)
		})
	}
}

func TestParseCommand_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty command",
			input:       "",
			wantErr:     true,
			errContains: "empty command",
		},
		{
			name:        "whitespace only",
			input:       "   ",
			wantErr:     true,
			errContains: "empty command",
		},
		{
			name:        "unknown command",
			input:       "unknown",
			wantErr:     true,
			errContains: "unknown command",
		},
		{
			name:        "typo in command",
			input:       "jmp 5",
			wantErr:     true,
			errContains: "unknown command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := ParseCommand(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, cmd)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestIsValidCommodity(t *testing.T) {
	tests := []struct {
		name       string
		commodity  string
		wantValid  bool
	}{
		{"food_supplies", "food_supplies", true},
		{"fuel_cells", "fuel_cells", true},
		{"raw_ore", "raw_ore", true},
		{"refined_ore", "refined_ore", true},
		{"machinery", "machinery", true},
		{"electronics", "electronics", true},
		{"luxury_goods", "luxury_goods", true},
		{"invalid", "invalid", false},
		{"FOOD_SUPPLIES", "FOOD_SUPPLIES", false}, // Case sensitive
		{"food", "food", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidCommodity(tt.commodity)
			assert.Equal(t, tt.wantValid, result)
		})
	}
}
