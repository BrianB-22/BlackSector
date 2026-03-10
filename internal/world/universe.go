package world

// GetSystem retrieves a system by ID (thread-safe)
func (u *Universe) GetSystem(systemID string) *System {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Systems[systemID]
}

// GetPort retrieves a port by ID (thread-safe)
func (u *Universe) GetPort(portID string) *Port {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Ports[portID]
}

// GetRegion retrieves a region by ID (thread-safe)
func (u *Universe) GetRegion(regionID string) *Region {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Regions[regionID]
}

// GetSystemsByRegion retrieves all systems in a region (thread-safe)
func (u *Universe) GetSystemsByRegion(regionID string) []*System {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.systemsByRegion[regionID]
}

// GetPortsBySystem retrieves all ports in a system (thread-safe)
func (u *Universe) GetPortsBySystem(systemID string) []*Port {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.portsBySystem[systemID]
}

// GetJumpConnections retrieves all jump connections from a system (thread-safe)
func (u *Universe) GetJumpConnections(systemID string) []*JumpConnection {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.connectionsByFrom[systemID]
}

// FindNearestPort finds the nearest port to a given system (simple implementation)
// In Phase 1, this returns the first port in the system, or the first port in a connected system
func (u *Universe) FindNearestPort(systemID string) *Port {
	u.mu.RLock()
	defer u.mu.RUnlock()

	// First, check if the system has any ports
	if ports := u.portsBySystem[systemID]; len(ports) > 0 {
		return ports[0]
	}

	// If no ports in current system, find first connected system with a port
	for _, conn := range u.connectionsByFrom[systemID] {
		destID := conn.ToSystemID
		if ports := u.portsBySystem[destID]; len(ports) > 0 {
			return ports[0]
		}
	}

	// Fallback: return any port (should not happen in valid topology)
	for _, port := range u.Ports {
		return port
	}

	return nil
}
