package navigation

// Navigator defines the interface for ship navigation operations
type Navigator interface {
	// Jump attempts to move a ship to a connected system
	Jump(shipID string, targetSystemID int, currentTick int64) error

	// GetJumpConnections returns all valid jump destinations from a system
	GetJumpConnections(systemID int) ([]*JumpConnection, error)

	// CalculateFuelCost computes fuel cost for a jump
	CalculateFuelCost(fromSystemID, toSystemID int) (int, error)

	// ValidateJump checks if a jump is possible (connection exists, fuel available)
	ValidateJump(ship *Ship, targetSystemID int) error

	// Dock attempts to dock a ship at a port in the current system
	Dock(shipID string, portID int, currentTick int64) error

	// Undock attempts to undock a ship from its current port
	Undock(shipID string, currentTick int64) error
}
