package world

// NewTestUniverse creates a Universe for testing with the given jump connections.
// It automatically builds the bidirectional connection cache.
func NewTestUniverse(connections []*JumpConnection) *Universe {
	universe := &Universe{
		Systems:         make(map[string]*System),
		Ports:           make(map[string]*Port),
		Regions:         make(map[string]*Region),
		JumpConnections: connections,
		systemsByRegion: make(map[string][]*System),
		portsBySystem:   make(map[string][]*Port),
	}

	// Build connection index (simulating what LoadWorld does)
	universe.connectionsByFrom = make(map[string][]*JumpConnection)
	for _, conn := range connections {
		// Forward connection
		universe.connectionsByFrom[conn.FromSystemID] = append(
			universe.connectionsByFrom[conn.FromSystemID],
			conn,
		)
		
		// Reverse connection (bidirectional in Phase 1)
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

	return universe
}

// NewTestUniverseComplete creates a complete Universe for testing with systems, ports, and connections.
// It builds all necessary caches including port-by-system and connection-by-from indexes.
func NewTestUniverseComplete(systems map[string]*System, ports map[string]*Port, connections []*JumpConnection) *Universe {
	universe := &Universe{
		Systems:         systems,
		Ports:           ports,
		Regions:         make(map[string]*Region),
		JumpConnections: connections,
		systemsByRegion: make(map[string][]*System),
		portsBySystem:   make(map[string][]*Port),
		connectionsByFrom: make(map[string][]*JumpConnection),
	}

	// Build port index
	for _, port := range ports {
		universe.portsBySystem[port.SystemID] = append(
			universe.portsBySystem[port.SystemID],
			port,
		)
	}

	// Build connection index (bidirectional)
	for _, conn := range connections {
		// Forward connection
		universe.connectionsByFrom[conn.FromSystemID] = append(
			universe.connectionsByFrom[conn.FromSystemID],
			conn,
		)
		
		// Reverse connection (bidirectional in Phase 1)
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

	// Build system by region index
	for _, system := range systems {
		if system.RegionID != "" {
			universe.systemsByRegion[system.RegionID] = append(
				universe.systemsByRegion[system.RegionID],
				system,
			)
		}
	}

	return universe
}
