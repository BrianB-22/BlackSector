package navigation

import (
	"errors"
	"testing"

	"github.com/BrianB-22/BlackSector/internal/world"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockShipRepository is a mock implementation of ShipRepository for testing
type mockShipRepository struct {
	ships              map[string]*mockShip
	updatePositionErr  error
	getShipErr         error
	updatedShips       []string // Track which ships were updated
}

// mockShip is a simplified ship for testing with string status
type mockShip struct {
	ShipID          string
	PlayerID        string
	ShipClass       string
	HullPoints      int
	MaxHullPoints   int
	ShieldPoints    int
	MaxShieldPoints int
	CurrentSystemID int
	Status          string
	DockedAtPortID  *int
	LastUpdatedTick int64
}

func newMockShipRepository() *mockShipRepository {
	return &mockShipRepository{
		ships:        make(map[string]*mockShip),
		updatedShips: make([]string, 0),
	}
}

func (m *mockShipRepository) GetShipByID(shipID string) (*Ship, error) {
	if m.getShipErr != nil {
		return nil, m.getShipErr
	}
	mockShip, exists := m.ships[shipID]
	if !exists {
		return nil, nil
	}
	// Convert mockShip to navigation.Ship
	return &Ship{
		ShipID:          mockShip.ShipID,
		PlayerID:        mockShip.PlayerID,
		ShipClass:       mockShip.ShipClass,
		HullPoints:      mockShip.HullPoints,
		MaxHullPoints:   mockShip.MaxHullPoints,
		ShieldPoints:    mockShip.ShieldPoints,
		MaxShieldPoints: mockShip.MaxShieldPoints,
		CurrentSystemID: mockShip.CurrentSystemID,
		Status:          ShipStatus(mockShip.Status),
		DockedAtPortID:  mockShip.DockedAtPortID,
		LastUpdatedTick: mockShip.LastUpdatedTick,
	}, nil
}

func (m *mockShipRepository) UpdateShipPosition(shipID string, systemID int, tick int64) error {
	if m.updatePositionErr != nil {
		return m.updatePositionErr
	}
	ship, exists := m.ships[shipID]
	if !exists {
		return errors.New("ship not found")
	}
	ship.CurrentSystemID = systemID
	ship.LastUpdatedTick = tick
	m.updatedShips = append(m.updatedShips, shipID)
	return nil
}

func (m *mockShipRepository) UpdateShipDockStatus(shipID string, status ShipStatus, dockedAtPortID *int, tick int64) error {
	ship, exists := m.ships[shipID]
	if !exists {
		return errors.New("ship not found")
	}
	ship.Status = string(status)
	ship.DockedAtPortID = dockedAtPortID
	ship.LastUpdatedTick = tick
	return nil
}

// createTestUniverse creates a simple test universe with connected systems
func createTestUniverse() *world.Universe {
	// Create jump connections: 1 <-> 2, 2 <-> 3 (no direct 1 <-> 3)
	connections := []*world.JumpConnection{
		{
			FromSystemID: "1",
			ToSystemID:   "2",
			FuelCost:     10,
		},
		{
			FromSystemID: "2",
			ToSystemID:   "3",
			FuelCost:     15,
		},
	}

	universe := world.NewTestUniverse(connections)

	// Add systems
	universe.Systems = map[string]*world.System{
		"1": {
			SystemID:      "1",
			Name:          "Alpha Station",
			RegionID:      "1",
			SecurityLevel: 2.0,
			PositionX:     0.0,
			PositionY:     0.0,
		},
		"2": {
			SystemID:      "2",
			Name:          "Beta Outpost",
			RegionID:      "1",
			SecurityLevel: 0.8,
			PositionX:     10.0,
			PositionY:     10.0,
		},
		"3": {
			SystemID:      "3",
			Name:          "Gamma Sector",
			RegionID:      "1",
			SecurityLevel: 0.2,
			PositionX:     20.0,
			PositionY:     20.0,
		},
	}

	return universe
}

func TestJump_Success(t *testing.T) {
	// Setup
	universe := createTestUniverse()
	mockDB := newMockShipRepository()
	logger := zerolog.Nop()
	nav := NewNavigationSystem(universe, mockDB, logger)

	// Create test ship in system 1, status IN_SPACE
	ship := &mockShip{
		ShipID:          "ship-001",
		PlayerID:        "player-001",
		ShipClass:       "courier",
		HullPoints:      100,
		MaxHullPoints:   100,
		ShieldPoints:    50,
		MaxShieldPoints: 50,
		CurrentSystemID: 1,
		Status:          "IN_SPACE",
		LastUpdatedTick: 100,
	}
	mockDB.ships[ship.ShipID] = ship

	// Execute jump from system 1 to system 2
	err := nav.Jump("ship-001", 2, 105)

	// Verify
	require.NoError(t, err)
	assert.Equal(t, 2, ship.CurrentSystemID, "ship should be in system 2")
	assert.Equal(t, int64(105), ship.LastUpdatedTick, "last updated tick should be updated")
	assert.Contains(t, mockDB.updatedShips, "ship-001", "ship should be marked as updated")
}

func TestJump_ShipNotFound(t *testing.T) {
	// Setup
	universe := createTestUniverse()
	mockDB := newMockShipRepository()
	logger := zerolog.Nop()
	nav := NewNavigationSystem(universe, mockDB, logger)

	// Execute jump with non-existent ship
	err := nav.Jump("nonexistent-ship", 2, 105)

	// Verify
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrShipNotFound)
}

func TestJump_ShipDocked(t *testing.T) {
	// Setup
	universe := createTestUniverse()
	mockDB := newMockShipRepository()
	logger := zerolog.Nop()
	nav := NewNavigationSystem(universe, mockDB, logger)

	portID := 1
	ship := &mockShip{
		ShipID:          "ship-001",
		PlayerID:        "player-001",
		CurrentSystemID: 1,
		Status:          "DOCKED",
		DockedAtPortID:  &portID,
		LastUpdatedTick: 100,
	}
	mockDB.ships[ship.ShipID] = ship

	// Execute jump while docked
	err := nav.Jump("ship-001", 2, 105)

	// Verify
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrShipDocked)
	assert.Equal(t, 1, ship.CurrentSystemID, "ship should not have moved")
}

func TestJump_ShipInCombat(t *testing.T) {
	// Setup
	universe := createTestUniverse()
	mockDB := newMockShipRepository()
	logger := zerolog.Nop()
	nav := NewNavigationSystem(universe, mockDB, logger)

	ship := &mockShip{
		ShipID:          "ship-001",
		PlayerID:        "player-001",
		CurrentSystemID: 1,
		Status:          "IN_COMBAT",
		LastUpdatedTick: 100,
	}
	mockDB.ships[ship.ShipID] = ship

	// Execute jump while in combat
	err := nav.Jump("ship-001", 2, 105)

	// Verify
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrShipInCombat)
	assert.Equal(t, 1, ship.CurrentSystemID, "ship should not have moved")
}

func TestJump_ShipDestroyed(t *testing.T) {
	// Setup
	universe := createTestUniverse()
	mockDB := newMockShipRepository()
	logger := zerolog.Nop()
	nav := NewNavigationSystem(universe, mockDB, logger)

	ship := &mockShip{
		ShipID:          "ship-001",
		PlayerID:        "player-001",
		CurrentSystemID: 1,
		Status:          "DESTROYED",
		LastUpdatedTick: 100,
	}
	mockDB.ships[ship.ShipID] = ship

	// Execute jump while destroyed
	err := nav.Jump("ship-001", 2, 105)

	// Verify
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrShipDestroyed)
	assert.Equal(t, 1, ship.CurrentSystemID, "ship should not have moved")
}

func TestJump_InvalidTargetSystem(t *testing.T) {
	// Setup
	universe := createTestUniverse()
	mockDB := newMockShipRepository()
	logger := zerolog.Nop()
	nav := NewNavigationSystem(universe, mockDB, logger)

	ship := &mockShip{
		ShipID:          "ship-001",
		PlayerID:        "player-001",
		CurrentSystemID: 1,
		Status:          "IN_SPACE",
		LastUpdatedTick: 100,
	}
	mockDB.ships[ship.ShipID] = ship

	// Execute jump to non-existent system
	err := nav.Jump("ship-001", 999, 105)

	// Verify
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidSystemID)
	assert.Equal(t, 1, ship.CurrentSystemID, "ship should not have moved")
}

func TestJump_NoConnection(t *testing.T) {
	// Setup
	universe := createTestUniverse()
	mockDB := newMockShipRepository()
	logger := zerolog.Nop()
	nav := NewNavigationSystem(universe, mockDB, logger)

	ship := &mockShip{
		ShipID:          "ship-001",
		PlayerID:        "player-001",
		CurrentSystemID: 1,
		Status:          "IN_SPACE",
		LastUpdatedTick: 100,
	}
	mockDB.ships[ship.ShipID] = ship

	// Execute jump from system 1 to system 3 (no direct connection)
	err := nav.Jump("ship-001", 3, 105)

	// Verify
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNoConnection)
	assert.Equal(t, 1, ship.CurrentSystemID, "ship should not have moved")
}

func TestJump_DatabaseError(t *testing.T) {
	// Setup
	universe := createTestUniverse()
	mockDB := newMockShipRepository()
	logger := zerolog.Nop()
	nav := NewNavigationSystem(universe, mockDB, logger)

	ship := &mockShip{
		ShipID:          "ship-001",
		PlayerID:        "player-001",
		CurrentSystemID: 1,
		Status:          "IN_SPACE",
		LastUpdatedTick: 100,
	}
	mockDB.ships[ship.ShipID] = ship

	// Simulate database error on update
	mockDB.updatePositionErr = errors.New("database connection lost")

	// Execute jump
	err := nav.Jump("ship-001", 2, 105)

	// Verify
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database connection lost")
}

func TestJump_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		shipStatus     string
		fromSystem     int
		toSystem       int
		expectError    error
		expectMove     bool
	}{
		{
			name:        "valid_jump_in_space",
			shipStatus:  "IN_SPACE",
			fromSystem:  1,
			toSystem:    2,
			expectError: nil,
			expectMove:  true,
		},
		{
			name:        "cannot_jump_while_docked",
			shipStatus:  "DOCKED",
			fromSystem:  1,
			toSystem:    2,
			expectError: ErrShipDocked,
			expectMove:  false,
		},
		{
			name:        "cannot_jump_in_combat",
			shipStatus:  "IN_COMBAT",
			fromSystem:  1,
			toSystem:    2,
			expectError: ErrShipInCombat,
			expectMove:  false,
		},
		{
			name:        "cannot_jump_when_destroyed",
			shipStatus:  "DESTROYED",
			fromSystem:  1,
			toSystem:    2,
			expectError: ErrShipDestroyed,
			expectMove:  false,
		},
		{
			name:        "cannot_jump_without_connection",
			shipStatus:  "IN_SPACE",
			fromSystem:  1,
			toSystem:    3,
			expectError: ErrNoConnection,
			expectMove:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			universe := createTestUniverse()
			mockDB := newMockShipRepository()
			logger := zerolog.Nop()
			nav := NewNavigationSystem(universe, mockDB, logger)

			ship := &mockShip{
				ShipID:          "ship-001",
				PlayerID:        "player-001",
				CurrentSystemID: tt.fromSystem,
				Status:          tt.shipStatus,
				LastUpdatedTick: 100,
			}
			mockDB.ships[ship.ShipID] = ship

			// Execute
			err := nav.Jump("ship-001", tt.toSystem, 105)

			// Verify
			if tt.expectError != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectError)
			} else {
				require.NoError(t, err)
			}

			if tt.expectMove {
				assert.Equal(t, tt.toSystem, ship.CurrentSystemID, "ship should have moved")
				assert.Equal(t, int64(105), ship.LastUpdatedTick, "tick should be updated")
			} else {
				assert.Equal(t, tt.fromSystem, ship.CurrentSystemID, "ship should not have moved")
			}
		})
	}
}
