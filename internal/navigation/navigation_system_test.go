package navigation

import (
	"fmt"
	"testing"

	"github.com/BrianB-22/BlackSector/internal/world"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNavigationSystem(t *testing.T) {
	logger := zerolog.Nop()
	universe := &world.Universe{}
	mockDB := newMockShipRepository()
	nav := NewNavigationSystem(universe, mockDB, logger)

	require.NotNil(t, nav)
	assert.IsType(t, &NavigationSystem{}, nav)
}

func TestNavigationSystem_ImplementsNavigator(t *testing.T) {
	logger := zerolog.Nop()
	universe := &world.Universe{}
	mockDB := newMockShipRepository()
	var _ Navigator = NewNavigationSystem(universe, mockDB, logger)
}

func TestShipStatus_Constants(t *testing.T) {
	tests := []struct {
		name     string
		status   ShipStatus
		expected string
	}{
		{"docked", StatusDocked, "DOCKED"},
		{"in_space", StatusInSpace, "IN_SPACE"},
		{"in_combat", StatusInCombat, "IN_COMBAT"},
		{"destroyed", StatusDestroyed, "DESTROYED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.status))
		})
	}
}

func TestShip_Structure(t *testing.T) {
	portID := 1
	ship := &Ship{
		ShipID:          "test-ship-id",
		PlayerID:        "test-player-id",
		ShipClass:       "courier",
		HullPoints:      100,
		MaxHullPoints:   100,
		ShieldPoints:    50,
		MaxShieldPoints: 50,
		EnergyPoints:    100,
		MaxEnergyPoints: 100,
		CargoCapacity:   20,
		MissilesCurrent: 0,
		CurrentSystemID: 1,
		PositionX:       0.0,
		PositionY:       0.0,
		Status:          StatusDocked,
		DockedAtPortID:  &portID,
		LastUpdatedTick: 0,
	}

	assert.Equal(t, "test-ship-id", ship.ShipID)
	assert.Equal(t, "test-player-id", ship.PlayerID)
	assert.Equal(t, "courier", ship.ShipClass)
	assert.Equal(t, StatusDocked, ship.Status)
	assert.NotNil(t, ship.DockedAtPortID)
	assert.Equal(t, 1, *ship.DockedAtPortID)
}

func TestJumpConnection_Structure(t *testing.T) {
	conn := &JumpConnection{
		ConnectionID:     1,
		FromSystemID:     1,
		ToSystemID:       2,
		Bidirectional:    true,
		FuelCostModifier: 1.0,
	}

	assert.Equal(t, 1, conn.ConnectionID)
	assert.Equal(t, 1, conn.FromSystemID)
	assert.Equal(t, 2, conn.ToSystemID)
	assert.True(t, conn.Bidirectional)
	assert.Equal(t, 1.0, conn.FuelCostModifier)
}

func TestGetJumpConnections(t *testing.T) {
	tests := []struct {
		name             string
		systemID         int
		worldConnections []*world.JumpConnection
		expectedCount    int
		expectError      bool
	}{
		{
			name:     "system_with_connections",
			systemID: 1,
			worldConnections: []*world.JumpConnection{
				{FromSystemID: "1", ToSystemID: "2", FuelCost: 10},
				{FromSystemID: "1", ToSystemID: "3", FuelCost: 15},
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:             "system_with_no_connections",
			systemID:         99,
			worldConnections: []*world.JumpConnection{},
			expectedCount:    0,
			expectError:      false,
		},
		{
			name:     "single_connection",
			systemID: 5,
			worldConnections: []*world.JumpConnection{
				{FromSystemID: "5", ToSystemID: "6", FuelCost: 20},
			},
			expectedCount: 1,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test universe with connection cache
			universe := world.NewTestUniverse(tt.worldConnections)
			logger := zerolog.Nop()
			mockDB := newMockShipRepository()
			nav := NewNavigationSystem(universe, mockDB, logger)

			// Call GetJumpConnections
			connections, err := nav.GetJumpConnections(tt.systemID)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, connections, tt.expectedCount)

			// Verify connection properties
			for i, conn := range connections {
				assert.NotZero(t, conn.ConnectionID)
				assert.Equal(t, tt.systemID, conn.FromSystemID)
				assert.True(t, conn.Bidirectional, "all connections should be bidirectional in Phase 1")
				
				// Verify ToSystemID matches world connection
				if i < len(tt.worldConnections) {
					expectedToID := mustAtoi(tt.worldConnections[i].ToSystemID)
					assert.Equal(t, expectedToID, conn.ToSystemID)
					assert.Equal(t, float64(tt.worldConnections[i].FuelCost), conn.FuelCostModifier)
				}
			}
		})
	}
}

func TestGetJumpConnections_InvalidSystemID(t *testing.T) {
	// Create universe with no connections for system 999
	universe := world.NewTestUniverse([]*world.JumpConnection{
		{FromSystemID: "1", ToSystemID: "2", FuelCost: 10},
	})
	
	logger := zerolog.Nop()
	mockDB := newMockShipRepository()
	nav := NewNavigationSystem(universe, mockDB, logger)

	// Query non-existent system - should return empty list, not error
	connections, err := nav.GetJumpConnections(999)
	require.NoError(t, err)
	assert.Empty(t, connections)
}

func TestGetJumpConnections_VerifyBidirectional(t *testing.T) {
	// Test that all connections are marked as bidirectional
	worldConnections := []*world.JumpConnection{
		{FromSystemID: "1", ToSystemID: "2", FuelCost: 10},
		{FromSystemID: "1", ToSystemID: "3", FuelCost: 15},
		{FromSystemID: "1", ToSystemID: "4", FuelCost: 20},
	}
	
	universe := world.NewTestUniverse(worldConnections)
	logger := zerolog.Nop()
	mockDB := newMockShipRepository()
	nav := NewNavigationSystem(universe, mockDB, logger)

	connections, err := nav.GetJumpConnections(1)
	require.NoError(t, err)
	require.Len(t, connections, 3)

	for _, conn := range connections {
		assert.True(t, conn.Bidirectional, "Phase 1 requires all connections to be bidirectional")
	}
}

// Helper to convert string to int (panics on error, for test data only)
func mustAtoi(s string) int {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		panic(err)
	}
	return result
}
