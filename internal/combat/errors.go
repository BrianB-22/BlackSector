package combat

import "errors"

// Combat error types for game logic validation
var (
	// ErrShipNotInSpace indicates the ship must be in space for combat
	ErrShipNotInSpace = errors.New("ship must be in space for combat")

	// ErrShipAlreadyInCombat indicates the ship is already engaged in combat
	ErrShipAlreadyInCombat = errors.New("ship is already in combat")

	// ErrCombatNotFound indicates the combat instance does not exist
	ErrCombatNotFound = errors.New("combat instance not found")

	// ErrCombatNotActive indicates the combat is not in active state
	ErrCombatNotActive = errors.New("combat is not active")

	// ErrNotPlayerTurn indicates it's not the player's turn
	ErrNotPlayerTurn = errors.New("not player's turn")

	// ErrShipDestroyed indicates the ship has been destroyed
	ErrShipDestroyed = errors.New("ship has been destroyed")

	// ErrInvalidPirateTier indicates an unknown pirate tier was specified
	ErrInvalidPirateTier = errors.New("invalid pirate tier")

	// ErrInsufficientCredits indicates player doesn't have enough credits to surrender
	ErrInsufficientCredits = errors.New("insufficient credits for surrender")
)
