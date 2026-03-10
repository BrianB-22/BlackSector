package world

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWorldGenerator(t *testing.T) {
	logger := zerolog.Nop()
	wg := NewWorldGenerator(logger)
	
	assert.NotNil(t, wg)
}

func TestUniverseGetters(t *testing.T) {
	u := &Universe{
		Regions: map[string]*Region{
			"test_region": {RegionID: "test_region", Name: "Test Region"},
		},
		Systems: map[string]*System{
			"test_system": {SystemID: "test_system", Name: "Test System", RegionID: "test_region"},
		},
		Ports: map[string]*Port{
			"test_port": {PortID: "test_port", Name: "Test Port", SystemID: "test_system"},
		},
		systemsByRegion: map[string][]*System{
			"test_region": {{SystemID: "test_system", Name: "Test System", RegionID: "test_region"}},
		},
		portsBySystem: map[string][]*Port{
			"test_system": {{PortID: "test_port", Name: "Test Port", SystemID: "test_system"}},
		},
		connectionsByFrom: map[string][]*JumpConnection{},
	}

	// Test GetSystem
	system := u.GetSystem("test_system")
	require.NotNil(t, system)
	assert.Equal(t, "Test System", system.Name)

	// Test GetPort
	port := u.GetPort("test_port")
	require.NotNil(t, port)
	assert.Equal(t, "Test Port", port.Name)

	// Test GetRegion
	region := u.GetRegion("test_region")
	require.NotNil(t, region)
	assert.Equal(t, "Test Region", region.Name)

	// Test GetSystemsByRegion
	systems := u.GetSystemsByRegion("test_region")
	assert.Len(t, systems, 1)
	assert.Equal(t, "Test System", systems[0].Name)

	// Test GetPortsBySystem
	ports := u.GetPortsBySystem("test_system")
	assert.Len(t, ports, 1)
	assert.Equal(t, "Test Port", ports[0].Name)
}

func TestFindNearestPort(t *testing.T) {
	u := &Universe{
		Systems: map[string]*System{
			"system_1": {SystemID: "system_1", Name: "System 1"},
			"system_2": {SystemID: "system_2", Name: "System 2"},
		},
		Ports: map[string]*Port{
			"port_1": {PortID: "port_1", Name: "Port 1", SystemID: "system_1"},
			"port_2": {PortID: "port_2", Name: "Port 2", SystemID: "system_2"},
		},
		portsBySystem: map[string][]*Port{
			"system_1": {{PortID: "port_1", Name: "Port 1", SystemID: "system_1"}},
			"system_2": {{PortID: "port_2", Name: "Port 2", SystemID: "system_2"}},
		},
		connectionsByFrom: map[string][]*JumpConnection{
			"system_1": {{FromSystemID: "system_1", ToSystemID: "system_2", FuelCost: 10}},
			"system_2": {{FromSystemID: "system_2", ToSystemID: "system_1", FuelCost: 10}},
		},
	}

	// Test finding port in same system
	port := u.FindNearestPort("system_1")
	require.NotNil(t, port)
	assert.Equal(t, "system_1", port.SystemID)

	// Test finding port in connected system
	u.portsBySystem["system_1"] = nil // Remove ports from system 1
	port = u.FindNearestPort("system_1")
	require.NotNil(t, port)
	// Should find port in connected system 2
	assert.Equal(t, "system_2", port.SystemID)
}

func TestLoadWorldFromFile(t *testing.T) {
	logger := zerolog.Nop()
	wg := NewWorldGenerator(logger)

	// Load the actual alpha_sector.json file
	universe, err := wg.LoadWorld("../../config/world/alpha_sector.json")
	require.NoError(t, err)
	require.NotNil(t, universe)

	// Verify systems were loaded
	assert.Greater(t, len(universe.Systems), 0, "should have loaded systems")
	assert.GreaterOrEqual(t, len(universe.Systems), 15, "should have at least 15 systems")
	assert.LessOrEqual(t, len(universe.Systems), 20, "should have at most 20 systems")

	// Verify ports were loaded
	assert.Greater(t, len(universe.Ports), 0, "should have loaded ports")

	// Verify jump connections were loaded
	assert.Greater(t, len(universe.JumpConnections), 0, "should have loaded jump connections")

	// Verify Federated Space origin exists
	nexusPrime := universe.GetSystem("nexus_prime")
	require.NotNil(t, nexusPrime, "Nexus Prime (origin) should exist")
	assert.Equal(t, 2.0, nexusPrime.SecurityLevel, "Nexus Prime should be Federated Space (2.0)")

	// Verify origin starbase exists
	starbase := universe.GetPort("nexus_prime_starbase")
	require.NotNil(t, starbase, "Nexus Prime Starbase should exist")
	assert.Equal(t, "nexus_prime", starbase.SystemID, "Starbase should be in Nexus Prime")
	assert.Equal(t, "trading", starbase.PortType, "Starbase should be a trading port")

	// Verify indexes were built correctly
	portsInNexus := universe.GetPortsBySystem("nexus_prime")
	assert.Greater(t, len(portsInNexus), 0, "Nexus Prime should have ports")

	// Verify jump connections from Nexus Prime
	connections := universe.GetJumpConnections("nexus_prime")
	assert.Greater(t, len(connections), 0, "Nexus Prime should have jump connections")

	// Verify bidirectional connections work
	gatewayStation := universe.GetSystem("gateway_station")
	require.NotNil(t, gatewayStation, "Gateway Station should exist")
	
	connectionsFromGateway := universe.GetJumpConnections("gateway_station")
	assert.Greater(t, len(connectionsFromGateway), 0, "Gateway Station should have connections")
	
	// Check that reverse connection exists (bidirectional)
	hasReverseConnection := false
	for _, conn := range connectionsFromGateway {
		if conn.ToSystemID == "nexus_prime" {
			hasReverseConnection = true
			break
		}
	}
	assert.True(t, hasReverseConnection, "Should have bidirectional connection back to Nexus Prime")
}

func TestLoadWorldValidation(t *testing.T) {
	logger := zerolog.Nop()
	wg := NewWorldGenerator(logger)

	// Load the actual alpha_sector.json file
	universe, err := wg.LoadWorld("../../config/world/alpha_sector.json")
	require.NoError(t, err)
	require.NotNil(t, universe)

	// Validate topology
	err = wg.ValidateTopology(universe)
	assert.NoError(t, err, "World topology should be valid")
}

func TestLoadWorldInvalidPath(t *testing.T) {
	logger := zerolog.Nop()
	wg := NewWorldGenerator(logger)

	// Try to load non-existent file
	universe, err := wg.LoadWorld("nonexistent.json")
	assert.Error(t, err)
	assert.Nil(t, universe)
	assert.Contains(t, err.Error(), "failed to read world config")
}

func TestLoadWorldInvalidJSON(t *testing.T) {
	logger := zerolog.Nop()
	wg := NewWorldGenerator(logger)

	// Create a temporary invalid JSON file
	tmpFile := t.TempDir() + "/invalid.json"
	err := os.WriteFile(tmpFile, []byte("{invalid json"), 0644)
	require.NoError(t, err)

	// Try to load invalid JSON
	universe, err := wg.LoadWorld(tmpFile)
	assert.Error(t, err)
	assert.Nil(t, universe)
	assert.Contains(t, err.Error(), "failed to parse world config JSON")
}

func TestValidateTopologyMissingSystem(t *testing.T) {
	logger := zerolog.Nop()
	wg := NewWorldGenerator(logger)

	// Create universe with port referencing non-existent system
	universe := &Universe{
		Systems: map[string]*System{
			"system_1": {SystemID: "system_1", Name: "System 1", SecurityLevel: 2.0},
		},
		Ports: map[string]*Port{
			"port_1": {PortID: "port_1", Name: "Port 1", SystemID: "nonexistent"},
		},
		JumpConnections:   []*JumpConnection{},
		systemsByRegion:   make(map[string][]*System),
		portsBySystem:     make(map[string][]*Port),
		connectionsByFrom: make(map[string][]*JumpConnection),
	}

	err := wg.ValidateTopology(universe)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "references non-existent system")
}

func TestValidateTopologyInvalidSecurityLevel(t *testing.T) {
	logger := zerolog.Nop()
	wg := NewWorldGenerator(logger)

	// Create universe with invalid security level
	universe := &Universe{
		Systems: map[string]*System{
			"system_1": {SystemID: "system_1", Name: "System 1", SecurityLevel: 3.0}, // Invalid
		},
		Ports:             map[string]*Port{},
		JumpConnections:   []*JumpConnection{},
		systemsByRegion:   make(map[string][]*System),
		portsBySystem:     make(map[string][]*Port),
		connectionsByFrom: make(map[string][]*JumpConnection),
	}

	err := wg.ValidateTopology(universe)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid security level")
}

func TestValidateTopologyNoFederatedSpace(t *testing.T) {
	logger := zerolog.Nop()
	wg := NewWorldGenerator(logger)

	// Create universe without Federated Space
	universe := &Universe{
		Systems: map[string]*System{
			"system_1": {SystemID: "system_1", Name: "System 1", SecurityLevel: 0.8},
		},
		Ports:             map[string]*Port{},
		JumpConnections:   []*JumpConnection{},
		systemsByRegion:   make(map[string][]*System),
		portsBySystem:     make(map[string][]*Port),
		connectionsByFrom: make(map[string][]*JumpConnection),
	}

	err := wg.ValidateTopology(universe)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no Federated Space system found")
}

func TestValidateTopologyUnreachableSystems(t *testing.T) {
	logger := zerolog.Nop()
	wg := NewWorldGenerator(logger)

	// Create universe with isolated system
	universe := &Universe{
		Systems: map[string]*System{
			"system_1": {SystemID: "system_1", Name: "System 1", SecurityLevel: 2.0},
			"system_2": {SystemID: "system_2", Name: "System 2", SecurityLevel: 0.8},
		},
		Ports: map[string]*Port{},
		JumpConnections: []*JumpConnection{
			// No connection to system_2
		},
		systemsByRegion:   make(map[string][]*System),
		portsBySystem:     make(map[string][]*Port),
		connectionsByFrom: make(map[string][]*JumpConnection),
	}

	err := wg.ValidateTopology(universe)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unreachable systems detected")
}

func TestGetJumpConnections(t *testing.T) {
	u := &Universe{
		Systems: map[string]*System{
			"system_1": {SystemID: "system_1", Name: "System 1"},
			"system_2": {SystemID: "system_2", Name: "System 2"},
		},
		connectionsByFrom: map[string][]*JumpConnection{
			"system_1": {
				{FromSystemID: "system_1", ToSystemID: "system_2", FuelCost: 10},
			},
		},
	}

	connections := u.GetJumpConnections("system_1")
	assert.Len(t, connections, 1)
	assert.Equal(t, "system_2", connections[0].ToSystemID)

	// Test system with no connections
	connections = u.GetJumpConnections("system_2")
	assert.Len(t, connections, 0)
}

func TestValidateTopologyComplexGraph(t *testing.T) {
	logger := zerolog.Nop()
	wg := NewWorldGenerator(logger)

	// Create a complex graph with multiple paths
	// Graph structure:
	//   system_1 (Fed Space) <-> system_2 <-> system_3
	//        |                                    |
	//        v                                    v
	//   system_4 <----------> system_5 <-----> system_6
	universe := &Universe{
		Systems: map[string]*System{
			"system_1": {SystemID: "system_1", Name: "System 1", SecurityLevel: 2.0},
			"system_2": {SystemID: "system_2", Name: "System 2", SecurityLevel: 0.8},
			"system_3": {SystemID: "system_3", Name: "System 3", SecurityLevel: 0.5},
			"system_4": {SystemID: "system_4", Name: "System 4", SecurityLevel: 0.3},
			"system_5": {SystemID: "system_5", Name: "System 5", SecurityLevel: 0.2},
			"system_6": {SystemID: "system_6", Name: "System 6", SecurityLevel: 0.1},
		},
		Ports: map[string]*Port{},
		JumpConnections: []*JumpConnection{
			{FromSystemID: "system_1", ToSystemID: "system_2", FuelCost: 10},
			{FromSystemID: "system_2", ToSystemID: "system_3", FuelCost: 10},
			{FromSystemID: "system_1", ToSystemID: "system_4", FuelCost: 15},
			{FromSystemID: "system_4", ToSystemID: "system_5", FuelCost: 10},
			{FromSystemID: "system_3", ToSystemID: "system_6", FuelCost: 20},
			{FromSystemID: "system_5", ToSystemID: "system_6", FuelCost: 10},
		},
		systemsByRegion: make(map[string][]*System),
		portsBySystem:   make(map[string][]*Port),
	}

	// Build connection index (simulating what LoadWorld does)
	universe.connectionsByFrom = make(map[string][]*JumpConnection)
	for _, conn := range universe.JumpConnections {
		universe.connectionsByFrom[conn.FromSystemID] = append(
			universe.connectionsByFrom[conn.FromSystemID],
			conn,
		)
		// Add reverse connection for bidirectional navigation
		reverseConn := &JumpConnection{
			FromSystemID: conn.ToSystemID,
			ToSystemID:   conn.FromSystemID,
			FuelCost:     conn.FuelCost,
		}
		universe.connectionsByFrom[conn.ToSystemID] = append(
			universe.connectionsByFrom[conn.ToSystemID],
			reverseConn,
		)
	}

	// Validate - should pass because all systems are reachable
	err := wg.ValidateTopology(universe)
	assert.NoError(t, err, "All systems should be reachable in complex graph")
}

func TestValidateTopologyPartiallyConnectedGraph(t *testing.T) {
	logger := zerolog.Nop()
	wg := NewWorldGenerator(logger)

	// Create a graph with two disconnected components
	// Component 1: system_1 <-> system_2
	// Component 2: system_3 <-> system_4 (isolated)
	universe := &Universe{
		Systems: map[string]*System{
			"system_1": {SystemID: "system_1", Name: "System 1", SecurityLevel: 2.0},
			"system_2": {SystemID: "system_2", Name: "System 2", SecurityLevel: 0.8},
			"system_3": {SystemID: "system_3", Name: "System 3", SecurityLevel: 0.5},
			"system_4": {SystemID: "system_4", Name: "System 4", SecurityLevel: 0.3},
		},
		Ports: map[string]*Port{},
		JumpConnections: []*JumpConnection{
			{FromSystemID: "system_1", ToSystemID: "system_2", FuelCost: 10},
			{FromSystemID: "system_3", ToSystemID: "system_4", FuelCost: 10},
		},
		systemsByRegion: make(map[string][]*System),
		portsBySystem:   make(map[string][]*Port),
	}

	// Build connection index
	universe.connectionsByFrom = make(map[string][]*JumpConnection)
	for _, conn := range universe.JumpConnections {
		universe.connectionsByFrom[conn.FromSystemID] = append(
			universe.connectionsByFrom[conn.FromSystemID],
			conn,
		)
		reverseConn := &JumpConnection{
			FromSystemID: conn.ToSystemID,
			ToSystemID:   conn.FromSystemID,
			FuelCost:     conn.FuelCost,
		}
		universe.connectionsByFrom[conn.ToSystemID] = append(
			universe.connectionsByFrom[conn.ToSystemID],
			reverseConn,
		)
	}

	// Validate - should fail because systems are not all reachable
	err := wg.ValidateTopology(universe)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unreachable systems detected")
	// Should identify either system_3 or system_4 as unreachable (depending on start point)
}

func TestValidateTopologyEmptyUniverse(t *testing.T) {
	logger := zerolog.Nop()
	wg := NewWorldGenerator(logger)

	// Create empty universe
	universe := &Universe{
		Systems:           map[string]*System{},
		Ports:             map[string]*Port{},
		JumpConnections:   []*JumpConnection{},
		systemsByRegion:   make(map[string][]*System),
		portsBySystem:     make(map[string][]*Port),
		connectionsByFrom: make(map[string][]*JumpConnection),
	}

	err := wg.ValidateTopology(universe)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no systems defined")
}


func TestFindNearestPortFallback(t *testing.T) {
	// Test the fallback case where system has no ports and connected systems have no ports
	u := &Universe{
		Systems: map[string]*System{
			"system_1": {SystemID: "system_1", Name: "System 1"},
			"system_2": {SystemID: "system_2", Name: "System 2"},
			"system_3": {SystemID: "system_3", Name: "System 3"},
		},
		Ports: map[string]*Port{
			"port_3": {PortID: "port_3", Name: "Port 3", SystemID: "system_3"},
		},
		portsBySystem: map[string][]*Port{
			"system_1": nil, // No ports
			"system_2": nil, // No ports
			"system_3": {{PortID: "port_3", Name: "Port 3", SystemID: "system_3"}},
		},
		connectionsByFrom: map[string][]*JumpConnection{
			"system_1": {{FromSystemID: "system_1", ToSystemID: "system_2", FuelCost: 10}},
		},
	}

	// Should fall back to any port in the universe
	port := u.FindNearestPort("system_1")
	require.NotNil(t, port)
	assert.Equal(t, "port_3", port.PortID)
}

func TestFindNearestPortNoPorts(t *testing.T) {
	// Test case where universe has no ports at all
	u := &Universe{
		Systems: map[string]*System{
			"system_1": {SystemID: "system_1", Name: "System 1"},
		},
		Ports:             map[string]*Port{},
		portsBySystem:     map[string][]*Port{},
		connectionsByFrom: map[string][]*JumpConnection{},
	}

	port := u.FindNearestPort("system_1")
	assert.Nil(t, port)
}

func TestValidateTopologyInvalidJumpConnection(t *testing.T) {
	logger := zerolog.Nop()
	wg := NewWorldGenerator(logger)

	tests := []struct {
		name        string
		universe    *Universe
		expectedErr string
	}{
		{
			name: "invalid from_system",
			universe: &Universe{
				Systems: map[string]*System{
					"system_1": {SystemID: "system_1", Name: "System 1", SecurityLevel: 2.0},
				},
				Ports: map[string]*Port{},
				JumpConnections: []*JumpConnection{
					{FromSystemID: "nonexistent", ToSystemID: "system_1", FuelCost: 10},
				},
				systemsByRegion:   make(map[string][]*System),
				portsBySystem:     make(map[string][]*Port),
				connectionsByFrom: make(map[string][]*JumpConnection),
			},
			expectedErr: "non-existent from_system",
		},
		{
			name: "invalid to_system",
			universe: &Universe{
				Systems: map[string]*System{
					"system_1": {SystemID: "system_1", Name: "System 1", SecurityLevel: 2.0},
				},
				Ports: map[string]*Port{},
				JumpConnections: []*JumpConnection{
					{FromSystemID: "system_1", ToSystemID: "nonexistent", FuelCost: 10},
				},
				systemsByRegion:   make(map[string][]*System),
				portsBySystem:     make(map[string][]*Port),
				connectionsByFrom: make(map[string][]*JumpConnection),
			},
			expectedErr: "non-existent to_system",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := wg.ValidateTopology(tt.universe)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestValidateTopologyInvalidRegionReference(t *testing.T) {
	logger := zerolog.Nop()
	wg := NewWorldGenerator(logger)

	// Create universe with system referencing non-existent region
	universe := &Universe{
		Regions: map[string]*Region{
			"region_1": {RegionID: "region_1", Name: "Region 1"},
		},
		Systems: map[string]*System{
			"system_1": {SystemID: "system_1", Name: "System 1", RegionID: "nonexistent", SecurityLevel: 2.0},
		},
		Ports:             map[string]*Port{},
		JumpConnections:   []*JumpConnection{},
		systemsByRegion:   make(map[string][]*System),
		portsBySystem:     make(map[string][]*Port),
		connectionsByFrom: make(map[string][]*JumpConnection),
	}

	err := wg.ValidateTopology(universe)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "references non-existent region")
}

func TestLoadWorldWithRegions(t *testing.T) {
	logger := zerolog.Nop()
	wg := NewWorldGenerator(logger)

	// Create a temporary world config with regions
	tmpFile := t.TempDir() + "/test_world.json"
	worldJSON := `{
		"world": {
			"name": "Test World",
			"description": "Test",
			"seed": 42
		},
		"commodities": [],
		"regions": [
			{
				"region_id": "test_region",
				"name": "Test Region",
				"region_type": "core",
				"security_level": 2.0
			}
		],
		"systems": [
			{
				"system_id": "test_system",
				"name": "Test System",
				"region_id": "test_region",
				"security_rating": 2.0,
				"security_zone": "federated",
				"x": 0.0,
				"y": 0.0
			}
		],
		"ports": [
			{
				"port_id": "test_port",
				"system_id": "test_system",
				"name": "Test Port",
				"port_type": "trading"
			}
		],
		"jump_connections": []
	}`
	err := os.WriteFile(tmpFile, []byte(worldJSON), 0644)
	require.NoError(t, err)

	// Load the world
	universe, err := wg.LoadWorld(tmpFile)
	require.NoError(t, err)
	require.NotNil(t, universe)

	// Verify region was loaded
	region := universe.GetRegion("test_region")
	require.NotNil(t, region)
	assert.Equal(t, "Test Region", region.Name)
	assert.Equal(t, "core", region.RegionType)

	// Verify system is indexed by region
	systems := universe.GetSystemsByRegion("test_region")
	assert.Len(t, systems, 1)
	assert.Equal(t, "test_system", systems[0].SystemID)
}

func TestLoadWorldEmptyFile(t *testing.T) {
	logger := zerolog.Nop()
	wg := NewWorldGenerator(logger)

	// Create an empty JSON file
	tmpFile := t.TempDir() + "/empty.json"
	err := os.WriteFile(tmpFile, []byte("{}"), 0644)
	require.NoError(t, err)

	// Load should succeed but create empty universe
	universe, err := wg.LoadWorld(tmpFile)
	require.NoError(t, err)
	require.NotNil(t, universe)
	assert.Len(t, universe.Systems, 0)
	assert.Len(t, universe.Ports, 0)
	assert.Len(t, universe.JumpConnections, 0)
}

func TestGettersNonExistent(t *testing.T) {
	u := &Universe{
		Regions:           map[string]*Region{},
		Systems:           map[string]*System{},
		Ports:             map[string]*Port{},
		systemsByRegion:   map[string][]*System{},
		portsBySystem:     map[string][]*Port{},
		connectionsByFrom: map[string][]*JumpConnection{},
	}

	// Test getters with non-existent IDs
	assert.Nil(t, u.GetSystem("nonexistent"))
	assert.Nil(t, u.GetPort("nonexistent"))
	assert.Nil(t, u.GetRegion("nonexistent"))
	assert.Nil(t, u.GetSystemsByRegion("nonexistent"))
	assert.Nil(t, u.GetPortsBySystem("nonexistent"))
	assert.Nil(t, u.GetJumpConnections("nonexistent"))
}

func TestConcurrentAccess(t *testing.T) {
	// Create a universe with test data
	u := &Universe{
		Regions: map[string]*Region{
			"region_1": {RegionID: "region_1", Name: "Region 1"},
			"region_2": {RegionID: "region_2", Name: "Region 2"},
		},
		Systems: map[string]*System{
			"system_1": {SystemID: "system_1", Name: "System 1", RegionID: "region_1"},
			"system_2": {SystemID: "system_2", Name: "System 2", RegionID: "region_1"},
			"system_3": {SystemID: "system_3", Name: "System 3", RegionID: "region_2"},
		},
		Ports: map[string]*Port{
			"port_1": {PortID: "port_1", Name: "Port 1", SystemID: "system_1"},
			"port_2": {PortID: "port_2", Name: "Port 2", SystemID: "system_2"},
			"port_3": {PortID: "port_3", Name: "Port 3", SystemID: "system_3"},
		},
		systemsByRegion: map[string][]*System{
			"region_1": {
				{SystemID: "system_1", Name: "System 1", RegionID: "region_1"},
				{SystemID: "system_2", Name: "System 2", RegionID: "region_1"},
			},
			"region_2": {
				{SystemID: "system_3", Name: "System 3", RegionID: "region_2"},
			},
		},
		portsBySystem: map[string][]*Port{
			"system_1": {{PortID: "port_1", Name: "Port 1", SystemID: "system_1"}},
			"system_2": {{PortID: "port_2", Name: "Port 2", SystemID: "system_2"}},
			"system_3": {{PortID: "port_3", Name: "Port 3", SystemID: "system_3"}},
		},
		connectionsByFrom: map[string][]*JumpConnection{
			"system_1": {{FromSystemID: "system_1", ToSystemID: "system_2", FuelCost: 10}},
			"system_2": {{FromSystemID: "system_2", ToSystemID: "system_3", FuelCost: 15}},
		},
	}

	// Launch multiple goroutines that concurrently read from the universe
	const numGoroutines = 50
	const numIterations = 100
	
	done := make(chan bool, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numIterations; j++ {
				// Perform various read operations
				system := u.GetSystem("system_1")
				assert.NotNil(t, system)
				
				port := u.GetPort("port_1")
				assert.NotNil(t, port)
				
				region := u.GetRegion("region_1")
				assert.NotNil(t, region)
				
				systems := u.GetSystemsByRegion("region_1")
				assert.Len(t, systems, 2)
				
				ports := u.GetPortsBySystem("system_1")
				assert.Len(t, ports, 1)
				
				connections := u.GetJumpConnections("system_1")
				assert.Len(t, connections, 1)
				
				nearestPort := u.FindNearestPort("system_1")
				assert.NotNil(t, nearestPort)
			}
			done <- true
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}
