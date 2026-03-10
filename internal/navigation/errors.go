package navigation

import "errors"

// Navigation error types for game logic validation
var (
	// ErrShipDocked indicates the ship cannot jump while docked
	ErrShipDocked = errors.New("ship is docked")

	// ErrShipInCombat indicates the ship cannot jump while in combat
	ErrShipInCombat = errors.New("ship is in combat")

	// ErrShipDestroyed indicates the ship cannot jump while destroyed
	ErrShipDestroyed = errors.New("ship is destroyed")

	// ErrNoConnection indicates no jump connection exists between systems
	ErrNoConnection = errors.New("no jump connection exists")

	// ErrShipNotFound indicates the ship does not exist in the database
	ErrShipNotFound = errors.New("ship not found")

	// ErrInvalidSystemID indicates the target system does not exist
	ErrInvalidSystemID = errors.New("invalid system ID")

	// ErrAlreadyDocked indicates the ship is already docked
	ErrAlreadyDocked = errors.New("ship is already docked")

	// ErrPortNotFound indicates the port does not exist
	ErrPortNotFound = errors.New("port not found")

	// ErrPortNotInSystem indicates the port is not in the ship's current system
	ErrPortNotInSystem = errors.New("port is not in current system")

	// ErrNotDocked indicates the ship is not docked at a port
	ErrNotDocked = errors.New("ship is not docked")
)
