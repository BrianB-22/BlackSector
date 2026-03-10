package navigation

import (
	"errors"
	"testing"

	"github.com/BrianB-22/BlackSector/internal/world"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockShipRepoForDock is a mock implementation of ShipRepository for dock tests
type mockShipRepoForDock struct {
	ships                  map[string]*Ship
	updateDockStatusCalled bool
	updateDockStatusError  error
}

func (m *mockShipRepoForDock) GetShipByID(shipID string) (*Ship, error) {
	ship, exists := m.ships[shipID]
	if !exists {
		return nil, nil
	}
	return ship, nil
}

func (m *mockShipRepoForDock) UpdateShipPosition(shipID string, systemID int, tick int64) error {
	return nil
}

func (m *mockShipRepoForDock) UpdateShipDockStatus(shipID string, status ShipStatus, dockedAtPortID *int, tick int64) error {
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

func TestDock_Success(t *testing.T) {
	// Setup test universe with system and port
	universe := world.NewTestUniverse([]*world.JumpConnection{})
	universe.Systems = map[string]*world.System{
		"1": {
			SystemID:      "1",
			Name:          "Test System",
			SecurityLevel: 1.0,
		},
	}
	universe.Ports = map[string]*world.Port{
		"100": {
			PortID:   "100",
			SystemID: "1",
			Name:     "Test Station",
			PortType: "trading",
		},
	}

	// Setup mock ship repository
	mockDB := &mockShipRepoForDock{
		ships: map[string]*Ship{
			"ship-1": {
				ShipID:          "ship-1",
				PlayerID:        "player-1",
				CurrentSystemID: 1,
				Status:          StatusInSpace,
			},
		},
	}

	// Create navigation system
	logger := zerolog.Nop()
	nav := NewNavigationSystem(universe, mockDB, logger)

	// Execute dock
	err := nav.Dock("ship-1", 100, 1)

	// Verify success
	require.NoError(t, err)
	assert.True(t, mockDB.updateDockStatusCalled)
	assert.Equal(t, StatusDocked, mockDB.ships["ship-1"].Status)
	assert.NotNil(t, mockDB.ships["ship-1"].DockedAtPortID)
	assert.Equal(t, 100, *mockDB.ships["ship-1"].DockedAtPortID)
	assert.Equal(t, int64(1), mockDB.ships["ship-1"].LastUpdatedTick)
}

func TestDock_ShipNotFound(t *testing.T) {
	universe := world.NewTestUniverse([]*world.JumpConnection{})
	mockDB := &mockShipRepoForDock{
		ships: map[string]*Ship{},
	}

	logger := zerolog.Nop()
	nav := NewNavigationSystem(universe, mockDB, logger)

	err := nav.Dock("nonexistent-ship", 100, 1)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrShipNotFound)
	assert.False(t, mockDB.updateDockStatusCalled)
}

func TestDock_AlreadyDocked(t *testing.T) {
	universe := world.NewTestUniverse([]*world.JumpConnection{})
	portID := 50
	mockDB := &mockShipRepoForDock{
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

	logger := zerolog.Nop()
	nav := NewNavigationSystem(universe, mockDB, logger)

	err := nav.Dock("ship-1", 100, 1)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrAlreadyDocked)
	assert.False(t, mockDB.updateDockStatusCalled)
}

func TestDock_ShipInCombat(t *testing.T) {
	universe := world.NewTestUniverse([]*world.JumpConnection{})
	mockDB := &mockShipRepoForDock{
		ships: map[string]*Ship{
			"ship-1": {
				ShipID:          "ship-1",
				PlayerID:        "player-1",
				CurrentSystemID: 1,
				Status:          StatusInCombat,
			},
		},
	}

	logger := zerolog.Nop()
	nav := NewNavigationSystem(universe, mockDB, logger)

	err := nav.Dock("ship-1", 100, 1)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrShipInCombat)
	assert.False(t, mockDB.updateDockStatusCalled)
}

func TestDock_ShipDestroyed(t *testing.T) {
	universe := world.NewTestUniverse([]*world.JumpConnection{})
	mockDB := &mockShipRepoForDock{
		ships: map[string]*Ship{
			"ship-1": {
				ShipID:          "ship-1",
				PlayerID:        "player-1",
				CurrentSystemID: 1,
				Status:          StatusDestroyed,
			},
		},
	}

	logger := zerolog.Nop()
	nav := NewNavigationSystem(universe, mockDB, logger)

	err := nav.Dock("ship-1", 100, 1)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrShipDestroyed)
	assert.False(t, mockDB.updateDockStatusCalled)
}

func TestDock_PortNotFound(t *testing.T) {
	universe := world.NewTestUniverse([]*world.JumpConnection{})
	universe.Systems = map[string]*world.System{
		"1": {
			SystemID:      "1",
			Name:          "Test System",
			SecurityLevel: 1.0,
		},
	}
	universe.Ports = map[string]*world.Port{} // No ports

	mockDB := &mockShipRepoForDock{
		ships: map[string]*Ship{
			"ship-1": {
				ShipID:          "ship-1",
				PlayerID:        "player-1",
				CurrentSystemID: 1,
				Status:          StatusInSpace,
			},
		},
	}

	logger := zerolog.Nop()
	nav := NewNavigationSystem(universe, mockDB, logger)

	err := nav.Dock("ship-1", 100, 1)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPortNotFound)
	assert.False(t, mockDB.updateDockStatusCalled)
}

func TestDock_PortNotInSystem(t *testing.T) {
	universe := world.NewTestUniverse([]*world.JumpConnection{})
	universe.Systems = map[string]*world.System{
		"1": {
			SystemID:      "1",
			Name:          "Test System",
			SecurityLevel: 1.0,
		},
		"2": {
			SystemID:      "2",
			Name:          "Other System",
			SecurityLevel: 1.0,
		},
	}
	universe.Ports = map[string]*world.Port{
		"100": {
			PortID:   "100",
			SystemID: "2", // Port is in system 2
			Name:     "Test Station",
			PortType: "trading",
		},
	}

	mockDB := &mockShipRepoForDock{
		ships: map[string]*Ship{
			"ship-1": {
				ShipID:          "ship-1",
				PlayerID:        "player-1",
				CurrentSystemID: 1, // Ship is in system 1
				Status:          StatusInSpace,
			},
		},
	}

	logger := zerolog.Nop()
	nav := NewNavigationSystem(universe, mockDB, logger)

	err := nav.Dock("ship-1", 100, 1)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPortNotInSystem)
	assert.False(t, mockDB.updateDockStatusCalled)
}

func TestDock_DatabaseError(t *testing.T) {
	universe := world.NewTestUniverse([]*world.JumpConnection{})
	universe.Systems = map[string]*world.System{
		"1": {
			SystemID:      "1",
			Name:          "Test System",
			SecurityLevel: 1.0,
		},
	}
	universe.Ports = map[string]*world.Port{
		"100": {
			PortID:   "100",
			SystemID: "1",
			Name:     "Test Station",
			PortType: "trading",
		},
	}

	dbError := errors.New("database connection lost")
	mockDB := &mockShipRepoForDock{
		ships: map[string]*Ship{
			"ship-1": {
				ShipID:          "ship-1",
				PlayerID:        "player-1",
				CurrentSystemID: 1,
				Status:          StatusInSpace,
			},
		},
		updateDockStatusError: dbError,
	}

	logger := zerolog.Nop()
	nav := NewNavigationSystem(universe, mockDB, logger)

	err := nav.Dock("ship-1", 100, 1)

	require.Error(t, err)
	assert.ErrorIs(t, err, dbError)
	assert.True(t, mockDB.updateDockStatusCalled)
}

func TestDock_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		shipID        string
		portID        int
		shipStatus    ShipStatus
		shipSystemID  int
		portSystemID  string
		portExists    bool
		expectedError error
	}{
		{
			name:          "valid_dock",
			shipID:        "ship-1",
			portID:        100,
			shipStatus:    StatusInSpace,
			shipSystemID:  1,
			portSystemID:  "1",
			portExists:    true,
			expectedError: nil,
		},
		{
			name:          "already_docked",
			shipID:        "ship-1",
			portID:        100,
			shipStatus:    StatusDocked,
			shipSystemID:  1,
			portSystemID:  "1",
			portExists:    true,
			expectedError: ErrAlreadyDocked,
		},
		{
			name:          "in_combat",
			shipID:        "ship-1",
			portID:        100,
			shipStatus:    StatusInCombat,
			shipSystemID:  1,
			portSystemID:  "1",
			portExists:    true,
			expectedError: ErrShipInCombat,
		},
		{
			name:          "destroyed",
			shipID:        "ship-1",
			portID:        100,
			shipStatus:    StatusDestroyed,
			shipSystemID:  1,
			portSystemID:  "1",
			portExists:    true,
			expectedError: ErrShipDestroyed,
		},
		{
			name:          "port_not_found",
			shipID:        "ship-1",
			portID:        100,
			shipStatus:    StatusInSpace,
			shipSystemID:  1,
			portSystemID:  "1",
			portExists:    false,
			expectedError: ErrPortNotFound,
		},
		{
			name:          "port_wrong_system",
			shipID:        "ship-1",
			portID:        100,
			shipStatus:    StatusInSpace,
			shipSystemID:  1,
			portSystemID:  "2",
			portExists:    true,
			expectedError: ErrPortNotInSystem,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup universe
			universe := world.NewTestUniverse([]*world.JumpConnection{})
			universe.Systems = map[string]*world.System{
				"1": {SystemID: "1", Name: "System 1"},
				"2": {SystemID: "2", Name: "System 2"},
			}
			universe.Ports = map[string]*world.Port{}
			if tt.portExists {
				universe.Ports["100"] = &world.Port{
					PortID:   "100",
					SystemID: tt.portSystemID,
					Name:     "Test Port",
					PortType: "trading",
				}
			}

			// Setup mock DB
			mockDB := &mockShipRepoForDock{
				ships: map[string]*Ship{
					tt.shipID: {
						ShipID:          tt.shipID,
						PlayerID:        "player-1",
						CurrentSystemID: tt.shipSystemID,
						Status:          tt.shipStatus,
					},
				},
			}

			logger := zerolog.Nop()
			nav := NewNavigationSystem(universe, mockDB, logger)

			// Execute
			err := nav.Dock(tt.shipID, tt.portID, 1)

			// Verify
			if tt.expectedError != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.True(t, mockDB.updateDockStatusCalled)
			}
		})
	}
}
