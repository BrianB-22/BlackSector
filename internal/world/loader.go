package world

import (
	"encoding/json"
	"fmt"
	"os"
)

// WorldConfig represents the JSON structure of the world configuration file
type WorldConfig struct {
	World           *WorldMetadata    `json:"world"`
	Commodities     []*Commodity      `json:"commodities"`
	Regions         []*Region         `json:"regions,omitempty"`
	Systems         []*System         `json:"systems"`
	Ports           []*Port           `json:"ports"`
	JumpConnections []*JumpConnection `json:"jump_connections"`
	PortInventories []*PortCommodity  `json:"port_inventories,omitempty"`
}

// WorldMetadata contains metadata about the world
type WorldMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Seed        int    `json:"seed"`
}

// LoadWorld loads the static world configuration from the specified JSON file
func (wg *WorldGenerator) LoadWorld(configPath string) (*Universe, error) {
	wg.logger.Info().Str("path", configPath).Msg("loading world configuration")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read world config: %w", err)
	}

	var config WorldConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse world config JSON: %w", err)
	}

	universe := &Universe{
		Regions:           make(map[string]*Region),
		Systems:           make(map[string]*System),
		Ports:             make(map[string]*Port),
		JumpConnections:   config.JumpConnections,
		systemsByRegion:   make(map[string][]*System),
		portsBySystem:     make(map[string][]*Port),
		connectionsByFrom: make(map[string][]*JumpConnection),
	}

	// Acquire write lock while populating the universe cache
	universe.mu.Lock()
	defer universe.mu.Unlock()

	// Load regions (if present)
	for _, region := range config.Regions {
		universe.Regions[region.RegionID] = region
	}

	// Load systems and build region index
	for _, system := range config.Systems {
		universe.Systems[system.SystemID] = system
		if system.RegionID != "" {
			universe.systemsByRegion[system.RegionID] = append(
				universe.systemsByRegion[system.RegionID],
				system,
			)
		}
	}

	// Load ports and build system index
	for _, port := range config.Ports {
		universe.Ports[port.PortID] = port
		universe.portsBySystem[port.SystemID] = append(
			universe.portsBySystem[port.SystemID],
			port,
		)
	}

	// Build jump connection index (all connections are bidirectional in Phase 1)
	for _, conn := range config.JumpConnections {
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

	wg.logger.Info().
		Int("regions", len(universe.Regions)).
		Int("systems", len(universe.Systems)).
		Int("ports", len(universe.Ports)).
		Int("connections", len(universe.JumpConnections)).
		Int("commodities", len(config.Commodities)).
		Msg("world configuration loaded successfully")

	return universe, nil
}
