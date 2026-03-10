package navigation

import (
	"errors"
	"testing"

	"github.com/BrianB-22/BlackSector/internal/world"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockShipRepoForUndock is a mock implementation of ShipRepository for undock tests
type mockShipRepoForUndock struct {
	ships                  map[string]*Ship
	updateDockStatusCalled bool
	updateDockStatusError  error
}

func (m *mockShipRepoForUndock) GetShipByID(shipID string) (*Ship, error) {
	ship, exists := m.ships[shipID]
	if !exists {
		return nil, nil
	}
	return ship, nil
}

func (m *mockShipRepoForUndock) UpdateShipPosition(shipID string, systemID int, tick int64) error {
	return nil
}

func (m *mockShipRepoForUndock) UpdateShipDockStatus(shipID string, status ShipStatus, dockedAtPortID *int, tick int64) error {
	m.updateDockStatusCalled = true
	if m.updateDockStatusError != nil {
		return m.updateDockStatusError
	}
	if ship, exists := m.ships[shipID]; exists {
		ship.Status = status
		ship.DockedAtPortID = dockedAtPortID
		ship.LastUpdatedTick = tick
	}
	return nil
}

func TestUndock_Success(t *testing.T) {
	// Setup test universe
	universe := world.NewTestUniverse([]*world.JumpConnection{})
	universe.Systems = map[string]*world.System{
		"1": {
			SystemID:      "1",
			Name:          "Test System",
			SecurityLevel: 1.0,
		},
	}

	// Setup mock ship repository with docked ship
	portID := 100
	mockDB := &mockShipRepoForUndock{
		ships: map[string]*Ship{
			"ship-1": {
				ShipID:          "ship-1",
				PlayerID:        "player-1",
				CurrentSystemID: 1,
				Status:          StatusDocked,
				DockedAtPortID:  &portID,
			},
		},
	}

	// Create navigation system
	logger := zerolog.Nop()
	nav := NewNavigationSystem(universe, mockDB, logger)

	// Execute undock
	err := nav.Undock("ship-1", 1)

	// Verify success
	require.NoError(t, err)
	assert.True(t, mockDB.updateDockStatusCalled)
	assert.Equal(t, StatusInSpace, mockDB.ships["ship-1"].Status)
	assert.Nil(t, mockDB.ships["ship-1"].DockedAtPortID)
	assert.Equal(t, int64(1), mockDB.ships["ship-1"].LastUpdatedTick)
}

func TestUndock_ShipNotFound(t *testing.T) {
	universe := world.NewTestUniverse([]*world.JumpConnection{})
	mockDB := &mockShipRepoForUndock{
		ships: map[string]*Ship{},
	}

	logger := zerolog.Nop()
	nav := NewNavigationSystem(universe, mockDB, logger)

	err := nav.Undock("nonexistent-ship", 1)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrShipNotFound)
	assert.False(t, mockDB.updateDockStatusCalled)
}

func TestUndock_NotDocked(t *testing.T) {
	universe := world.NewTestUniverse([]*world.JumpConnection{})
	mockDB := &mockShipRepoForUndock{
		ships: map[string]*Ship{
			"ship-1": {
				ShipID:          "ship-1",
				PlayerID:        "player-1",
				CurrentSystemID: 1,
				Status:          StatusInSpace,
				DockedAtPortID:  nil,
			},
		},
	}

	logger := zerolog.Nop()
	nav := NewNavigationSystem(universe, mockDB, logger)

	err := nav.Undock("ship-1", 1)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotDocked)
	assert.False(t, mockDB.updateDockStatusCalled)
}

func TestUndock_ShipInCombat(t *testing.T) {
	universe := world.NewTestUniverse([]*world.JumpConnection{})
	portID := 100
	mockDB := &mockShipRepoForUndock{
		ships: map[string]*Ship{
			"ship-1": {
				ShipID:          "ship-1",
				PlayerID:        "player-1",
				CurrentSystemID: 1,
				Status:          StatusInCombat,
				DockedAtPortID:  &portID,
			},
		},
	}

	logger := zerolog.Nop()
	nav := NewNavigationSystem(universe, mockDB, logger)

	err := nav.Undock("ship-1", 1)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrShipInCombat)
	assert.False(t, mockDB.updateDockStatusCalled)
}

func TestUndock_ShipDestroyed(t *testing.T) {
	universe := world.NewTestUniverse([]*world.JumpConnection{})
	portID := 100
	mockDB := &mockShipRepoForUndock{
		ships: map[string]*Ship{
			"ship-1": {
				ShipID:          "ship-1",
				PlayerID:        "player-1",
				CurrentSystemID: 1,
				Status:          StatusDestroyed,
				DockedAtPortID:  &portID,
			},
		},
	}

	logger := zerolog.Nop()
	nav := NewNavigationSystem(universe, mockDB, logger)

	err := nav.Undock("ship-1", 1)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrShipDestroyed)
	assert.False(t, mockDB.updateDockStatusCalled)
}

func TestUndock_DatabaseError(t *testing.T) {
	universe := world.NewTestUniverse([]*world.JumpConnection{})
	portID := 100
	dbError := errors.New("database connection lost")
	mockDB := &mockShipRepoForUndock{
		ships: map[string]*Ship{
			"ship-1": {
				ShipID:          "ship-1",
				PlayerID:        "player-1",
				CurrentSystemID: 1,
				Status:          StatusDocked,
				DockedAtPortID:  &portID,
			},
		},
		updateDockStatusError: dbError,
	}

	logger := zerolog.Nop()
	nav := NewNavigationSystem(universe, mockDB, logger)

	err := nav.Undock("ship-1", 1)

	require.Error(t, err)
	assert.ErrorIs(t, err, dbError)
	assert.True(t, mockDB.updateDockStatusCalled)
}

func TestUndock_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		shipID        string
		shipStatus    ShipStatus
		dockedAtPort  *int
		expectedError error
	}{
		{
			name:          "valid_undock",
			shipID:        "ship-1",
			shipStatus:    StatusDocked,
			dockedAtPort:  intPtr(100),
			expectedError: nil,
		},
		{
			name:          "not_docked",
			shipID:        "ship-1",
			shipStatus:    StatusInSpace,
			dockedAtPort:  nil,
			expectedError: ErrNotDocked,
		},
		{
			name:          "in_combat",
			shipID:        "ship-1",
			shipStatus:    StatusInCombat,
			dockedAtPort:  intPtr(100),
			expectedError: ErrShipInCombat,
		},
		{
			name:          "destroyed",
			shipID:        "ship-1",
			shipStatus:    StatusDestroyed,
			dockedAtPort:  intPtr(100),
			expectedError: ErrShipDestroyed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup universe
			universe := world.NewTestUniverse([]*world.JumpConnection{})
			universe.Systems = map[string]*world.System{
				"1": {SystemID: "1", Name: "System 1"},
			}

			// Setup mock DB
			mockDB := &mockShipRepoForUndock{
				ships: map[string]*Ship{
					tt.shipID: {
						ShipID:          tt.shipID,
						PlayerID:        "player-1",
						CurrentSystemID: 1,
						Status:          tt.shipStatus,
						DockedAtPortID:  tt.dockedAtPort,
					},
				},
			}

			logger := zerolog.Nop()
			nav := NewNavigationSystem(universe, mockDB, logger)

			// Execute
			err := nav.Undock(tt.shipID, 1)

			// Verify
			if tt.expectedError != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.True(t, mockDB.updateDockStatusCalled)
				assert.Equal(t, StatusInSpace, mockDB.ships[tt.shipID].Status)
				assert.Nil(t, mockDB.ships[tt.shipID].DockedAtPortID)
			}
		})
	}
}

// intPtr is a helper function to create int pointers for test data
func intPtr(i int) *int {
	return &i
}
