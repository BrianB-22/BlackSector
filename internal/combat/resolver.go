package combat

// CombatResolver defines the interface for combat operations
type CombatResolver interface {
	// SpawnPirate creates a pirate encounter in Low Security systems
	SpawnPirate(systemID int, targetShipID string, tick int64) (*CombatInstance, error)

	// ProcessAttack resolves a player attack action
	ProcessAttack(combatID string, attackerID string, tick int64) (*CombatResult, error)

	// ProcessFlee attempts to disengage from combat
	ProcessFlee(combatID string, playerID string, tick int64) (*FleeResult, error)

	// ProcessSurrender ends combat with credit penalty
	ProcessSurrender(combatID string, playerID string, tick int64) error

	// ResolveCombatTick processes all active combat instances (pirate counter-attacks)
	ResolveCombatTick(tick int64) ([]*CombatEvent, error)

	// CheckPirateSpawns evaluates spawn probability for ships in Low Security systems
	CheckPirateSpawns(tick int64) ([]*CombatEvent, error)

	// GetActiveCombat retrieves the active combat instance for a ship
	GetActiveCombat(shipID string) (*CombatInstance, error)

	// GetPirateShip retrieves pirate ship data
	GetPirateShip(pirateShipID string) (*PirateShip, error)
}
