package navigation

import (
	"testing"

	"github.com/BrianB-22/BlackSector/internal/world"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSystemMap_Success(t *testing.T) {
	logger := zerolog.Nop()

	// Create test universe with multiple systems and connections
	universe := createTestUniverseForMap()
	navSys := NewNavigationSystem(universe, nil, logger)

	// Test getting map for system 1 (Federated Space origin)
	mapData, err := navSys.GetSystemMap(1)
	require.NoError(t, err)
	require.NotNil(t, mapData)

	// Verify current system info
	assert.NotNil(t, mapData.CurrentSystem)
	assert.Equal(t, 1, mapData.CurrentSystem.SystemID)
	assert.Equal(t, "Alpha Prime", mapData.CurrentSystem.Name)
	assert.Equal(t, 2.0, mapData.CurrentSystem.SecurityLevel)
	assert.Equal(t, "federated", mapData.CurrentSystem.SecurityZone)

	// Verify jump connections
	assert.Len(t, mapData.JumpConnections, 2)
	
	// Check first connection (to system 2)
	conn1 := findConnectionByDestID(mapData.JumpConnections, 2)
	require.NotNil(t, conn1, "connection to system 2 should exist")
	assert.Equal(t, "Beta Station", conn1.DestinationSystemName)
	assert.Equal(t, 1.0, conn1.SecurityLevel)
	assert.Equal(t, "high", conn1.SecurityZone)
	assert.Equal(t, 10, conn1.FuelCost)

	// Check second connection (to system 3)
	conn2 := findConnectionByDestID(mapData.JumpConnections, 3)
	require.NotNil(t, conn2, "connection to system 3 should exist")
	assert.Equal(t, "Gamma Outpost", conn2.DestinationSystemName)
	assert.Equal(t, 0.3, conn2.SecurityLevel)
	assert.Equal(t, "low", conn2.SecurityZone)
	assert.Equal(t, 15, conn2.FuelCost)

	// Verify ports
	assert.Len(t, mapData.Ports, 2)
	
	// Check first port
	port1 := findPortByID(mapData.Ports, 101)
	require.NotNil(t, port1, "port 101 should exist")
	assert.Equal(t, "Alpha Prime Starbase", port1.Name)
	assert.Equal(t, "trading", port1.PortType)

	// Check second port
	port2 := findPortByID(mapData.Ports, 102)
	require.NotNil(t, port2, "port 102 should exist")
	assert.Equal(t, "Alpha Prime Refueling", port2.Name)
	assert.Equal(t, "refueling", port2.PortType)
}

func TestGetSystemMap_SystemWithNoConnections(t *testing.T) {
	logger := zerolog.Nop()

	// Create universe with isolated system
	universe := createTestUniverseForMap()
	navSys := NewNavigationSystem(universe, nil, logger)

	// System 4 has connection back to system 2 (bidirectional from 2->4)
	mapData, err := navSys.GetSystemMap(4)
	require.NoError(t, err)
	require.NotNil(t, mapData)

	assert.Equal(t, 4, mapData.CurrentSystem.SystemID)
	assert.Equal(t, "Delta Isolated", mapData.CurrentSystem.Name)
	assert.Len(t, mapData.JumpConnections, 1, "system 4 should have one connection back to system 2")
	assert.Len(t, mapData.Ports, 1, "system should have one port")
	
	// Verify the connection is to system 2
	conn := mapData.JumpConnections[0]
	assert.Equal(t, 2, conn.DestinationSystemID)
	assert.Equal(t, "Beta Station", conn.DestinationSystemName)
}

func TestGetSystemMap_SystemWithNoPorts(t *testing.T) {
	logger := zerolog.Nop()

	// Create universe with system that has no ports
	universe := createTestUniverseForMap()
	navSys := NewNavigationSystem(universe, nil, logger)

	// System 5 has no ports
	mapData, err := navSys.GetSystemMap(5)
	require.NoError(t, err)
	require.NotNil(t, mapData)

	assert.Equal(t, 5, mapData.CurrentSystem.SystemID)
	assert.Equal(t, "Epsilon Empty", mapData.CurrentSystem.Name)
	assert.Empty(t, mapData.Ports, "system should have no ports")
	assert.Len(t, mapData.JumpConnections, 1, "system should have one connection")
}

func TestGetSystemMap_InvalidSystemID(t *testing.T) {
	logger := zerolog.Nop()

	universe := createTestUniverseForMap()
	navSys := NewNavigationSystem(universe, nil, logger)

	// Test with non-existent system ID
	mapData, err := navSys.GetSystemMap(999)
	assert.Error(t, err)
	assert.Nil(t, mapData)
	assert.ErrorIs(t, err, ErrInvalidSystemID)
}

func TestGetSystemMap_LowSecuritySystem(t *testing.T) {
	logger := zerolog.Nop()

	universe := createTestUniverseForMap()
	navSys := NewNavigationSystem(universe, nil, logger)

	// Test low security system (system 3)
	mapData, err := navSys.GetSystemMap(3)
	require.NoError(t, err)
	require.NotNil(t, mapData)

	assert.Equal(t, 3, mapData.CurrentSystem.SystemID)
	assert.Equal(t, "Gamma Outpost", mapData.CurrentSystem.Name)
	assert.Equal(t, 0.3, mapData.CurrentSystem.SecurityLevel)
	assert.Equal(t, "low", mapData.CurrentSystem.SecurityZone)
	
	// Should have connections to system 1 (bidirectional) and system 5 (reverse of 5->3)
	assert.Len(t, mapData.JumpConnections, 2)
	
	// Check connection to system 1
	conn1 := findConnectionByDestID(mapData.JumpConnections, 1)
	require.NotNil(t, conn1, "connection to system 1 should exist")
	assert.Equal(t, "Alpha Prime", conn1.DestinationSystemName)
	
	// Check connection to system 5
	conn5 := findConnectionByDestID(mapData.JumpConnections, 5)
	require.NotNil(t, conn5, "connection to system 5 should exist")
	assert.Equal(t, "Epsilon Empty", conn5.DestinationSystemName)
}

func TestGetSystemMap_HighSecuritySystem(t *testing.T) {
	logger := zerolog.Nop()

	universe := createTestUniverseForMap()
	navSys := NewNavigationSystem(universe, nil, logger)

	// Test high security system (system 2)
	mapData, err := navSys.GetSystemMap(2)
	require.NoError(t, err)
	require.NotNil(t, mapData)

	assert.Equal(t, 2, mapData.CurrentSystem.SystemID)
	assert.Equal(t, "Beta Station", mapData.CurrentSystem.Name)
	assert.Equal(t, 1.0, mapData.CurrentSystem.SecurityLevel)
	assert.Equal(t, "high", mapData.CurrentSystem.SecurityZone)
	
	// Should have connections to systems 1 and 4
	assert.Len(t, mapData.JumpConnections, 2)
}

// Helper function to create a test universe with multiple systems, ports, and connections
func createTestUniverseForMap() *world.Universe {
	// Systems
	systems := map[string]*world.System{
		"1": {
			SystemID:      "1",
			Name:          "Alpha Prime",
			SecurityLevel: 2.0,
			SecurityZone:  "federated",
			PositionX:     0.0,
			PositionY:     0.0,
		},
		"2": {
			SystemID:      "2",
			Name:          "Beta Station",
			SecurityLevel: 1.0,
			SecurityZone:  "high",
			PositionX:     10.0,
			PositionY:     5.0,
		},
		"3": {
			SystemID:      "3",
			Name:          "Gamma Outpost",
			SecurityLevel: 0.3,
			SecurityZone:  "low",
			PositionX:     -5.0,
			PositionY:     10.0,
		},
		"4": {
			SystemID:      "4",
			Name:          "Delta Isolated",
			SecurityLevel: 0.8,
			SecurityZone:  "high",
			PositionX:     15.0,
			PositionY:     15.0,
		},
		"5": {
			SystemID:      "5",
			Name:          "Epsilon Empty",
			SecurityLevel: 0.5,
			SecurityZone:  "low",
			PositionX:     -10.0,
			PositionY:     -10.0,
		},
	}

	// Ports
	ports := map[string]*world.Port{
		"101": {
			PortID:   "101",
			SystemID: "1",
			Name:     "Alpha Prime Starbase",
			PortType: "trading",
		},
		"102": {
			PortID:   "102",
			SystemID: "1",
			Name:     "Alpha Prime Refueling",
			PortType: "refueling",
		},
		"201": {
			PortID:   "201",
			SystemID: "2",
			Name:     "Beta Mining Station",
			PortType: "mining",
		},
		"301": {
			PortID:   "301",
			SystemID: "3",
			Name:     "Gamma Trading Post",
			PortType: "trading",
		},
		"401": {
			PortID:   "401",
			SystemID: "4",
			Name:     "Delta Starbase",
			PortType: "trading",
		},
	}

	// Jump connections (only one direction - bidirectional will be added automatically)
	connections := []*world.JumpConnection{
		{FromSystemID: "1", ToSystemID: "2", FuelCost: 10},
		{FromSystemID: "1", ToSystemID: "3", FuelCost: 15},
		{FromSystemID: "2", ToSystemID: "4", FuelCost: 20},
		{FromSystemID: "5", ToSystemID: "3", FuelCost: 12},
	}

	return world.NewTestUniverseComplete(systems, ports, connections)
}

// Helper function to find a connection by destination system ID
func findConnectionByDestID(connections []*JumpConnectionInfo, destID int) *JumpConnectionInfo {
	for _, conn := range connections {
		if conn.DestinationSystemID == destID {
			return conn
		}
	}
	return nil
}

// Helper function to find a port by ID
func findPortByID(ports []*PortInfo, portID int) *PortInfo {
	for _, port := range ports {
		if port.PortID == portID {
			return port
		}
	}
	return nil
}
