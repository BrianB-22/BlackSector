package world

import (
	"fmt"
)

// ValidateTopology ensures all systems are reachable and jump connections are valid
func (wg *WorldGenerator) ValidateTopology(u *Universe) error {
	wg.logger.Info().Msg("validating world topology")

	// Validate that all systems exist
	if len(u.Systems) == 0 {
		return fmt.Errorf("no systems defined in world configuration")
	}

	// Validate that all ports reference valid systems
	for portID, port := range u.Ports {
		if _, exists := u.Systems[port.SystemID]; !exists {
			return fmt.Errorf("port %s references non-existent system %s", portID, port.SystemID)
		}
	}

	// Validate that all systems reference valid regions (if regions are used)
	if len(u.Regions) > 0 {
		for systemID, system := range u.Systems {
			if system.RegionID != "" {
				if _, exists := u.Regions[system.RegionID]; !exists {
					return fmt.Errorf("system %s references non-existent region %s", systemID, system.RegionID)
				}
			}
		}
	}

	// Validate jump connections reference valid systems
	for _, conn := range u.JumpConnections {
		if _, exists := u.Systems[conn.FromSystemID]; !exists {
			return fmt.Errorf("jump connection references non-existent from_system %s", conn.FromSystemID)
		}
		if _, exists := u.Systems[conn.ToSystemID]; !exists {
			return fmt.Errorf("jump connection references non-existent to_system %s", conn.ToSystemID)
		}
	}

	// Validate that all systems are reachable (connectivity check)
	if err := wg.validateConnectivity(u); err != nil {
		return fmt.Errorf("connectivity validation failed: %w", err)
	}

	// Validate security levels are within valid ranges
	for systemID, system := range u.Systems {
		if system.SecurityLevel < 0.0 || system.SecurityLevel > 2.0 {
			return fmt.Errorf("system %s has invalid security level %.2f (must be 0.0-2.0)", systemID, system.SecurityLevel)
		}
	}

	// Validate that at least one Federated Space system exists (SecurityLevel = 2.0)
	hasFedSpace := false
	for _, system := range u.Systems {
		if system.SecurityLevel == 2.0 {
			hasFedSpace = true
			break
		}
	}
	if !hasFedSpace {
		return fmt.Errorf("no Federated Space system found (SecurityLevel = 2.0)")
	}

	wg.logger.Info().Msg("world topology validation successful")
	return nil
}


// validateConnectivity performs a breadth-first search to ensure all systems are reachable
func (wg *WorldGenerator) validateConnectivity(u *Universe) error {
	if len(u.Systems) == 0 {
		return nil
	}

	// Start from any system (use the first one we find)
	var startSystemID string
	for id := range u.Systems {
		startSystemID = id
		break
	}

	// BFS to find all reachable systems
	visited := make(map[string]bool)
	queue := []string{startSystemID}
	visited[startSystemID] = true

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		// Check all connections from this system
		for _, conn := range u.connectionsByFrom[currentID] {
			destID := conn.ToSystemID

			// If not visited, add to queue
			if !visited[destID] {
				visited[destID] = true
				queue = append(queue, destID)
			}
		}
	}

	// Check if all systems were reached
	if len(visited) != len(u.Systems) {
		unreachable := []string{}
		for systemID := range u.Systems {
			if !visited[systemID] {
				unreachable = append(unreachable, systemID)
			}
		}
		return fmt.Errorf("unreachable systems detected: %v", unreachable)
	}

	return nil
}
