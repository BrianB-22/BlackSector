package navigation

import (
	"fmt"
	"strconv"

	"github.com/BrianB-22/BlackSector/internal/world"
	"github.com/rs/zerolog"
)

// ShipRepository defines the database operations needed for navigation
type ShipRepository interface {
	GetShipByID(shipID string) (*Ship, error)
	UpdateShipPosition(shipID string, systemID int, tick int64) error
	UpdateShipDockStatus(shipID string, status ShipStatus, dockedAtPortID *int, tick int64) error
}

// NavigationSystem is the concrete implementation of the Navigator interface
type NavigationSystem struct {
	logger   zerolog.Logger
	universe *world.Universe
	db       ShipRepository
}

// NewNavigationSystem creates a new NavigationSystem instance
func NewNavigationSystem(universe *world.Universe, db ShipRepository, logger zerolog.Logger) *NavigationSystem {
	return &NavigationSystem{
		logger:   logger,
		universe: universe,
		db:       db,
	}
}

// Jump attempts to move a ship to a connected system
func (n *NavigationSystem) Jump(shipID string, targetSystemID int, currentTick int64) error {
	n.logger.Debug().
		Str("ship_id", shipID).
		Int("target_system", targetSystemID).
		Int64("tick", currentTick).
		Msg("jump requested")

	// Retrieve ship from database
	ship, err := n.db.GetShipByID(shipID)
	if err != nil {
		return fmt.Errorf("jump: %w", err)
	}
	if ship == nil {
		return fmt.Errorf("jump: %w", ErrShipNotFound)
	}

	// Validate ship status - must be IN_SPACE
	switch ShipStatus(ship.Status) {
	case StatusDocked:
		return fmt.Errorf("jump: %w", ErrShipDocked)
	case StatusInCombat:
		return fmt.Errorf("jump: %w", ErrShipInCombat)
	case StatusDestroyed:
		return fmt.Errorf("jump: %w", ErrShipDestroyed)
	case StatusInSpace:
		// Valid status, continue
	default:
		return fmt.Errorf("jump: invalid ship status: %s", ship.Status)
	}

	// Validate target system exists
	targetSystemIDStr := strconv.Itoa(targetSystemID)
	targetSystem := n.universe.GetSystem(targetSystemIDStr)
	if targetSystem == nil {
		return fmt.Errorf("jump: %w: %d", ErrInvalidSystemID, targetSystemID)
	}

	// Validate jump connection exists
	currentSystemIDStr := strconv.Itoa(ship.CurrentSystemID)
	connections := n.universe.GetJumpConnections(currentSystemIDStr)
	
	connectionExists := false
	for _, conn := range connections {
		if conn.ToSystemID == targetSystemIDStr {
			connectionExists = true
			break
		}
	}

	if !connectionExists {
		return fmt.Errorf("jump from system %d to %d: %w", ship.CurrentSystemID, targetSystemID, ErrNoConnection)
	}

	// Update ship position in database
	if err := n.db.UpdateShipPosition(shipID, targetSystemID, currentTick); err != nil {
		return fmt.Errorf("jump: %w", err)
	}

	n.logger.Info().
		Str("ship_id", shipID).
		Str("player_id", ship.PlayerID).
		Int("from_system", ship.CurrentSystemID).
		Int("to_system", targetSystemID).
		Int64("tick", currentTick).
		Msg("jump completed")

	return nil
}

// GetJumpConnections returns all valid jump destinations from a system
func (n *NavigationSystem) GetJumpConnections(systemID int) ([]*JumpConnection, error) {
	// Convert int ID to string for world lookup
	systemIDStr := strconv.Itoa(systemID)
	
	n.logger.Debug().
		Int("system_id", systemID).
		Str("system_id_str", systemIDStr).
		Msg("get jump connections requested")
	
	// Get connections from world Universe
	worldConnections := n.universe.GetJumpConnections(systemIDStr)
	
	// Convert world.JumpConnection to navigation.JumpConnection
	navConnections := make([]*JumpConnection, 0, len(worldConnections))
	for _, wc := range worldConnections {
		// Parse string IDs to int
		fromID, err := strconv.Atoi(wc.FromSystemID)
		if err != nil {
			return nil, fmt.Errorf("parse from_system_id %s: %w", wc.FromSystemID, err)
		}
		
		toID, err := strconv.Atoi(wc.ToSystemID)
		if err != nil {
			return nil, fmt.Errorf("parse to_system_id %s: %w", wc.ToSystemID, err)
		}
		
		// All connections are bidirectional in Phase 1
		navConnections = append(navConnections, &JumpConnection{
			ConnectionID:     len(navConnections) + 1, // Generate sequential ID
			FromSystemID:     fromID,
			ToSystemID:       toID,
			Bidirectional:    true,
			FuelCostModifier: float64(wc.FuelCost),
		})
	}
	
	n.logger.Debug().
		Int("system_id", systemID).
		Int("connection_count", len(navConnections)).
		Msg("jump connections retrieved")
	
	return navConnections, nil
}

// CalculateFuelCost computes fuel cost for a jump
func (n *NavigationSystem) CalculateFuelCost(fromSystemID, toSystemID int) (int, error) {
	// Implementation will be added in subsequent tasks
	n.logger.Debug().
		Int("from_system", fromSystemID).
		Int("to_system", toSystemID).
		Msg("calculate fuel cost requested")
	return 0, nil
}

// ValidateJump checks if a jump is possible (connection exists, fuel available)
func (n *NavigationSystem) ValidateJump(ship *Ship, targetSystemID int) error {
	// Implementation will be added in subsequent tasks
	n.logger.Debug().
		Str("ship_id", ship.ShipID).
		Int("target_system", targetSystemID).
		Msg("validate jump requested")
	return nil
}

// Dock attempts to dock a ship at a port in the current system
func (n *NavigationSystem) Dock(shipID string, portID int, currentTick int64) error {
	n.logger.Debug().
		Str("ship_id", shipID).
		Int("port_id", portID).
		Int64("tick", currentTick).
		Msg("dock requested")

	// Retrieve ship from database
	ship, err := n.db.GetShipByID(shipID)
	if err != nil {
		return fmt.Errorf("dock: %w", err)
	}
	if ship == nil {
		return fmt.Errorf("dock: %w", ErrShipNotFound)
	}

	// Validate ship status - must be IN_SPACE
	switch ShipStatus(ship.Status) {
	case StatusDocked:
		return fmt.Errorf("dock: %w", ErrAlreadyDocked)
	case StatusInCombat:
		return fmt.Errorf("dock: %w", ErrShipInCombat)
	case StatusDestroyed:
		return fmt.Errorf("dock: %w", ErrShipDestroyed)
	case StatusInSpace:
		// Valid status, continue
	default:
		return fmt.Errorf("dock: invalid ship status: %s", ship.Status)
	}

	// Validate port exists
	portIDStr := strconv.Itoa(portID)
	port := n.universe.GetPort(portIDStr)
	if port == nil {
		return fmt.Errorf("dock: %w: %d", ErrPortNotFound, portID)
	}

	// Validate port is in ship's current system
	currentSystemIDStr := strconv.Itoa(ship.CurrentSystemID)
	if port.SystemID != currentSystemIDStr {
		return fmt.Errorf("dock: %w: port %d is in system %s, ship is in system %d", 
			ErrPortNotInSystem, portID, port.SystemID, ship.CurrentSystemID)
	}

	// Update ship status to DOCKED and set DockedAtPortID
	if err := n.db.UpdateShipDockStatus(shipID, StatusDocked, &portID, currentTick); err != nil {
		return fmt.Errorf("dock: %w", err)
	}

	n.logger.Info().
		Str("ship_id", shipID).
		Str("player_id", ship.PlayerID).
		Int("port_id", portID).
		Str("port_name", port.Name).
		Int("system_id", ship.CurrentSystemID).
		Int64("tick", currentTick).
		Msg("dock completed")

	return nil
}

// Undock attempts to undock a ship from its current port
func (n *NavigationSystem) Undock(shipID string, currentTick int64) error {
	n.logger.Debug().
		Str("ship_id", shipID).
		Int64("tick", currentTick).
		Msg("undock requested")

	// Retrieve ship from database
	ship, err := n.db.GetShipByID(shipID)
	if err != nil {
		return fmt.Errorf("undock: %w", err)
	}
	if ship == nil {
		return fmt.Errorf("undock: %w", ErrShipNotFound)
	}

	// Validate ship status - must be DOCKED
	switch ShipStatus(ship.Status) {
	case StatusInSpace:
		return fmt.Errorf("undock: %w", ErrNotDocked)
	case StatusInCombat:
		return fmt.Errorf("undock: %w", ErrShipInCombat)
	case StatusDestroyed:
		return fmt.Errorf("undock: %w", ErrShipDestroyed)
	case StatusDocked:
		// Valid status, continue
	default:
		return fmt.Errorf("undock: invalid ship status: %s", ship.Status)
	}

	// Update ship status to IN_SPACE and clear DockedAtPortID
	if err := n.db.UpdateShipDockStatus(shipID, StatusInSpace, nil, currentTick); err != nil {
		return fmt.Errorf("undock: %w", err)
	}

	n.logger.Info().
		Str("ship_id", shipID).
		Str("player_id", ship.PlayerID).
		Int("system_id", ship.CurrentSystemID).
		Int64("tick", currentTick).
		Msg("undock completed")

	return nil
}
// GetSystemMap returns formatted system map data for display
func (n *NavigationSystem) GetSystemMap(systemID int) (*SystemMapData, error) {
	n.logger.Debug().
		Int("system_id", systemID).
		Msg("get system map requested")

	// Convert int ID to string for world lookup
	systemIDStr := strconv.Itoa(systemID)

	// Get current system information
	system := n.universe.GetSystem(systemIDStr)
	if system == nil {
		return nil, fmt.Errorf("get system map: %w: %d", ErrInvalidSystemID, systemID)
	}

	// Parse system ID back to int for display
	sysID, err := strconv.Atoi(system.SystemID)
	if err != nil {
		return nil, fmt.Errorf("parse system_id %s: %w", system.SystemID, err)
	}

	// Build current system info
	currentSystemInfo := &SystemInfo{
		SystemID:      sysID,
		Name:          system.Name,
		SecurityLevel: system.SecurityLevel,
		SecurityZone:  system.SecurityZone,
	}

	// Get jump connections from this system
	worldConnections := n.universe.GetJumpConnections(systemIDStr)
	jumpConnections := make([]*JumpConnectionInfo, 0, len(worldConnections))

	for _, wc := range worldConnections {
		// Get destination system details
		destSystem := n.universe.GetSystem(wc.ToSystemID)
		if destSystem == nil {
			n.logger.Warn().
				Str("to_system_id", wc.ToSystemID).
				Msg("destination system not found, skipping connection")
			continue
		}

		// Parse destination system ID
		destID, err := strconv.Atoi(wc.ToSystemID)
		if err != nil {
			n.logger.Warn().
				Str("to_system_id", wc.ToSystemID).
				Err(err).
				Msg("failed to parse destination system ID, skipping")
			continue
		}

		jumpConnections = append(jumpConnections, &JumpConnectionInfo{
			DestinationSystemID:   destID,
			DestinationSystemName: destSystem.Name,
			SecurityLevel:         destSystem.SecurityLevel,
			SecurityZone:          destSystem.SecurityZone,
			FuelCost:              wc.FuelCost,
		})
	}

	// Get ports in this system
	worldPorts := n.universe.GetPortsBySystem(systemIDStr)
	ports := make([]*PortInfo, 0, len(worldPorts))

	for _, wp := range worldPorts {
		// Parse port ID
		pID, err := strconv.Atoi(wp.PortID)
		if err != nil {
			n.logger.Warn().
				Str("port_id", wp.PortID).
				Err(err).
				Msg("failed to parse port ID, skipping")
			continue
		}

		ports = append(ports, &PortInfo{
			PortID:   pID,
			Name:     wp.Name,
			PortType: wp.PortType,
		})
	}

	mapData := &SystemMapData{
		CurrentSystem:   currentSystemInfo,
		JumpConnections: jumpConnections,
		Ports:           ports,
	}

	n.logger.Debug().
		Int("system_id", systemID).
		Int("connection_count", len(jumpConnections)).
		Int("port_count", len(ports)).
		Msg("system map data retrieved")

	return mapData, nil
}

