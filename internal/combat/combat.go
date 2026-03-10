package combat

import (
	"database/sql"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Database interface defines required database operations for combat system
type Database interface {
	// Ship operations
	GetShipByID(shipID string) (*Ship, error)
	UpdateShipStatus(shipID string, status string, tick int64) error
	UpdateShipDamage(shipID string, hull, shield int, tick int64) error
	ClearShipCargo(shipID string) error
	RespawnShip(shipID string, systemID int, portID int, tick int64) error

	// Player operations
	GetPlayerByShipID(shipID string) (string, error)
	GetPlayerCredits(playerID string) (int64, error)
	UpdatePlayerCredits(playerID string, credits int) error

	// Combat operations
	CreateCombatInstance(combat *CombatInstance) error
	GetCombatInstance(combatID string) (*CombatInstance, error)
	GetActiveCombatByShip(shipID string) (*CombatInstance, error)
	UpdateCombatStatus(combatID string, status CombatStatus, tick int64) error
	UpdateCombatTurn(combatID string, turnNumber int) error
	DeleteCombatInstance(combatID string) error

	// Pirate operations (ephemeral - in-memory only)
	CreatePirateShip(pirate *PirateShip) error
	GetPirateShip(pirateShipID string) (*PirateShip, error)
	UpdatePirateShip(pirate *PirateShip) error
	DeletePirateShip(pirateShipID string) error

	// System queries
	GetSystemSecurityLevel(systemID int) (float64, error)
	GetShipsInSpace() ([]*Ship, error)
	FindNearestPort(systemID int) (int, error)

	// Transaction management
	BeginTx() (*sql.Tx, error)
	CommitTx(tx *sql.Tx) error
	RollbackTx(tx *sql.Tx) error
}

// Config holds combat system configuration
type Config struct {
	PirateActivityBase    float64 // Base spawn probability (default 0.10)
	SpawnCheckInterval    int     // Ticks between spawn checks (default 5)
	SurrenderLossPercent  int     // Percentage of credits lost on surrender (default 40)
	InsurancePayout       int     // Credits granted on ship destruction (default 5000)
	RaiderSpawnWeight     float64 // Spawn probability for raiders (default 0.70)
	MarauderSpawnWeight   float64 // Spawn probability for marauders (default 0.30)
}

// DefaultConfig returns the default combat configuration
func DefaultConfig() *Config {
	return &Config{
		PirateActivityBase:   0.10,
		SpawnCheckInterval:   5,
		SurrenderLossPercent: 40,
		InsurancePayout:      5000,
		RaiderSpawnWeight:    0.70,
		MarauderSpawnWeight:  0.30,
	}
}

// CombatSystem is the concrete implementation of the CombatResolver interface
type CombatSystem struct {
	cfg          *Config
	db           Database
	pirateTiers  map[string]*PirateTierConfig
	logger       zerolog.Logger
}

// NewCombatSystem creates a new combat system instance
func NewCombatSystem(cfg *Config, db Database, logger zerolog.Logger) *CombatSystem {
	cs := &CombatSystem{
		cfg:         cfg,
		db:          db,
		pirateTiers: make(map[string]*PirateTierConfig),
		logger:      logger,
	}

	// Initialize pirate tier configurations
	cs.initializePirateTiers()

	return cs
}

// initializePirateTiers sets up the pirate tier definitions
func (cs *CombatSystem) initializePirateTiers() {
	// Raider: easier tier (70% spawn rate)
	cs.pirateTiers["raider"] = &PirateTierConfig{
		Tier:          "raider",
		Hull:          60,
		Shield:        20,
		DamageMin:     12,
		DamageMax:     18,
		Accuracy:      0.60,
		FleeThreshold: 0.15,
		SpawnWeight:   cs.cfg.RaiderSpawnWeight,
	}

	// Marauder: harder tier (30% spawn rate)
	cs.pirateTiers["marauder"] = &PirateTierConfig{
		Tier:          "marauder",
		Hull:          90,
		Shield:        40,
		DamageMin:     18,
		DamageMax:     28,
		Accuracy:      0.65,
		FleeThreshold: 0.10,
		SpawnWeight:   cs.cfg.MarauderSpawnWeight,
	}
}

// SpawnPirate creates a pirate encounter in Low Security systems
func (cs *CombatSystem) SpawnPirate(systemID int, targetShipID string, tick int64) (*CombatInstance, error) {
	// Get target ship
	ship, err := cs.db.GetShipByID(targetShipID)
	if err != nil {
		return nil, fmt.Errorf("spawn pirate: %w", err)
	}
	if ship == nil {
		return nil, fmt.Errorf("spawn pirate: ship not found")
	}

	// Validate ship is in space
	if ship.Status != "IN_SPACE" {
		return nil, fmt.Errorf("spawn pirate: %w", ErrShipNotInSpace)
	}

	// Check if ship already in combat
	existingCombat, err := cs.db.GetActiveCombatByShip(targetShipID)
	if err != nil {
		return nil, fmt.Errorf("spawn pirate: %w", err)
	}
	if existingCombat != nil {
		return nil, fmt.Errorf("spawn pirate: %w", ErrShipAlreadyInCombat)
	}

	// Select pirate tier (70% raider, 30% marauder)
	tier := cs.selectPirateTier()

	// Create pirate ship
	pirate := cs.createPirateShip(tier)
	if err := cs.db.CreatePirateShip(pirate); err != nil {
		return nil, fmt.Errorf("spawn pirate: %w", err)
	}

	// Create combat instance
	combat := &CombatInstance{
		CombatID:     uuid.New().String(),
		PlayerShipID: targetShipID,
		PirateShipID: pirate.ShipID,
		SystemID:     systemID,
		StartTick:    tick,
		Status:       CombatActive,
		TurnNumber:   0,
	}

	if err := cs.db.CreateCombatInstance(combat); err != nil {
		// Clean up pirate ship
		cs.db.DeletePirateShip(pirate.ShipID)
		return nil, fmt.Errorf("spawn pirate: %w", err)
	}

	// Update ship status to IN_COMBAT
	if err := cs.db.UpdateShipStatus(targetShipID, "IN_COMBAT", tick); err != nil {
		// Clean up combat and pirate
		cs.db.DeleteCombatInstance(combat.CombatID)
		cs.db.DeletePirateShip(pirate.ShipID)
		return nil, fmt.Errorf("spawn pirate: %w", err)
	}

	cs.logger.Info().
		Str("combat_id", combat.CombatID).
		Str("player_ship_id", targetShipID).
		Str("pirate_tier", tier).
		Int("system_id", systemID).
		Int64("tick", tick).
		Msg("pirate encounter spawned")

	return combat, nil
}

// selectPirateTier randomly selects a pirate tier based on spawn weights
func (cs *CombatSystem) selectPirateTier() string {
	roll := rand.Float64()
	if roll < cs.cfg.RaiderSpawnWeight {
		return "raider"
	}
	return "marauder"
}

// createPirateShip instantiates a new pirate ship with tier-specific stats
func (cs *CombatSystem) createPirateShip(tier string) *PirateShip {
	config := cs.pirateTiers[tier]
	return &PirateShip{
		ShipID:          uuid.New().String(),
		Tier:            tier,
		HullPoints:      config.Hull,
		MaxHull:         config.Hull,
		ShieldPoints:    config.Shield,
		MaxShield:       config.Shield,
		WeaponDamageMin: config.DamageMin,
		WeaponDamageMax: config.DamageMax,
		Accuracy:        config.Accuracy,
		FleeThreshold:   config.FleeThreshold,
	}
}

// ProcessAttack resolves a player attack action
func (cs *CombatSystem) ProcessAttack(combatID string, attackerID string, tick int64) (*CombatResult, error) {
	// Get combat instance
	combat, err := cs.db.GetCombatInstance(combatID)
	if err != nil {
		return nil, fmt.Errorf("process attack: %w", err)
	}
	if combat == nil {
		return nil, fmt.Errorf("process attack: %w", ErrCombatNotFound)
	}

	// Validate combat is active
	if combat.Status != CombatActive {
		return nil, fmt.Errorf("process attack: %w", ErrCombatNotActive)
	}

	// Validate attacker is the player in this combat
	if combat.PlayerShipID != attackerID {
		return nil, fmt.Errorf("process attack: invalid attacker")
	}

	// Get player ship
	playerShip, err := cs.db.GetShipByID(combat.PlayerShipID)
	if err != nil {
		return nil, fmt.Errorf("process attack: %w", err)
	}
	if playerShip == nil {
		return nil, fmt.Errorf("process attack: player ship not found")
	}

	// Check if player ship is destroyed
	if playerShip.HullPoints <= 0 {
		return nil, fmt.Errorf("process attack: %w", ErrShipDestroyed)
	}

	// Get pirate ship
	pirate, err := cs.db.GetPirateShip(combat.PirateShipID)
	if err != nil {
		return nil, fmt.Errorf("process attack: %w", err)
	}
	if pirate == nil {
		return nil, fmt.Errorf("process attack: pirate ship not found")
	}

	// Resolve player attack on pirate
	result := cs.resolveDamage(playerShip.WeaponDamage, pirate, 1.0) // Player always hits

	// Update pirate ship
	if err := cs.db.UpdatePirateShip(pirate); err != nil {
		return nil, fmt.Errorf("process attack: %w", err)
	}

	// Check if pirate destroyed
	if pirate.HullPoints <= 0 {
		result.TargetDestroyed = true
		if err := cs.endCombat(combat, "destroyed", tick); err != nil {
			return nil, fmt.Errorf("process attack: %w", err)
		}
		return result, nil
	}

	// Check if pirate flees
	hullPercent := float64(pirate.HullPoints) / float64(pirate.MaxHull)
	if hullPercent <= pirate.FleeThreshold {
		result.TargetFled = true
		if err := cs.endCombat(combat, "fled", tick); err != nil {
			return nil, fmt.Errorf("process attack: %w", err)
		}
		return result, nil
	}

	// Pirate counter-attacks
	counterResult := cs.resolveDamage(
		rand.Intn(pirate.WeaponDamageMax-pirate.WeaponDamageMin+1)+pirate.WeaponDamageMin,
		playerShip,
		pirate.Accuracy,
	)

	// Update player ship
	if err := cs.db.UpdateShipDamage(playerShip.ShipID, playerShip.HullPoints, playerShip.ShieldPoints, tick); err != nil {
		return nil, fmt.Errorf("process attack: %w", err)
	}

	// Check if player destroyed
	if playerShip.HullPoints <= 0 {
		if err := cs.handleShipDestruction(playerShip, combat, tick); err != nil {
			return nil, fmt.Errorf("process attack: %w", err)
		}
	}

	// Increment turn number
	combat.TurnNumber++
	if err := cs.db.UpdateCombatTurn(combatID, combat.TurnNumber); err != nil {
		return nil, fmt.Errorf("process attack: %w", err)
	}

	cs.logger.Debug().
		Str("combat_id", combatID).
		Bool("player_hit", result.Hit).
		Int("player_damage", result.Damage).
		Bool("pirate_hit", counterResult.Hit).
		Int("pirate_damage", counterResult.Damage).
		Int("turn", combat.TurnNumber).
		Msg("combat turn resolved")

	return result, nil
}

// resolveDamage applies damage to a target (ship or pirate)
func (cs *CombatSystem) resolveDamage(damage int, target interface{}, accuracy float64) *CombatResult {
	// Roll for hit
	hitRoll := rand.Float64()
	if hitRoll > accuracy {
		return &CombatResult{
			Hit:    false,
			Damage: 0,
		}
	}

	result := &CombatResult{
		Hit:    true,
		Damage: damage,
	}

	// Apply damage based on target type
	switch t := target.(type) {
	case *PirateShip:
		// Apply to shields first
		shieldDamage := min(damage, t.ShieldPoints)
		t.ShieldPoints -= shieldDamage
		remainingDamage := damage - shieldDamage

		// Apply remaining to hull
		hullDamage := min(remainingDamage, t.HullPoints)
		t.HullPoints -= hullDamage

		result.ShieldDamage = shieldDamage
		result.HullDamage = hullDamage
		result.TargetShield = t.ShieldPoints
		result.TargetHull = t.HullPoints
		result.TargetDestroyed = t.HullPoints <= 0

	case *Ship:
		// Apply to shields first
		shieldDamage := min(damage, t.ShieldPoints)
		t.ShieldPoints -= shieldDamage
		remainingDamage := damage - shieldDamage

		// Apply remaining to hull
		hullDamage := min(remainingDamage, t.HullPoints)
		t.HullPoints -= hullDamage

		result.ShieldDamage = shieldDamage
		result.HullDamage = hullDamage
		result.TargetShield = t.ShieldPoints
		result.TargetHull = t.HullPoints
		result.TargetDestroyed = t.HullPoints <= 0
	}

	return result
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ProcessFlee attempts to disengage from combat
func (cs *CombatSystem) ProcessFlee(combatID string, playerID string, tick int64) (*FleeResult, error) {
	// Get combat instance
	combat, err := cs.db.GetCombatInstance(combatID)
	if err != nil {
		return nil, fmt.Errorf("process flee: %w", err)
	}
	if combat == nil {
		return nil, fmt.Errorf("process flee: %w", ErrCombatNotFound)
	}

	// Validate combat is active
	if combat.Status != CombatActive {
		return nil, fmt.Errorf("process flee: %w", ErrCombatNotActive)
	}

	// End combat with fled status
	if err := cs.endCombat(combat, "fled", tick); err != nil {
		return nil, fmt.Errorf("process flee: %w", err)
	}

	cs.logger.Info().
		Str("combat_id", combatID).
		Str("player_id", playerID).
		Int64("tick", tick).
		Msg("player fled from combat")

	return &FleeResult{
		Success: true,
		Reason:  "Successfully disengaged from combat",
	}, nil
}

// ProcessSurrender ends combat with credit penalty
func (cs *CombatSystem) ProcessSurrender(combatID string, playerID string, tick int64) error {
	// Get combat instance
	combat, err := cs.db.GetCombatInstance(combatID)
	if err != nil {
		return fmt.Errorf("process surrender: %w", err)
	}
	if combat == nil {
		return fmt.Errorf("process surrender: %w", ErrCombatNotFound)
	}

	// Validate combat is active
	if combat.Status != CombatActive {
		return fmt.Errorf("process surrender: %w", ErrCombatNotActive)
	}

	// Get player credits
	credits, err := cs.db.GetPlayerCredits(playerID)
	if err != nil {
		return fmt.Errorf("process surrender: %w", err)
	}

	// Calculate credit loss (40% of wallet)
	lossAmount := int(float64(credits) * float64(cs.cfg.SurrenderLossPercent) / 100.0)
	newCredits := int(credits) - lossAmount

	// Update player credits
	if err := cs.db.UpdatePlayerCredits(playerID, newCredits); err != nil {
		return fmt.Errorf("process surrender: %w", err)
	}

	// End combat
	if err := cs.endCombat(combat, "ended", tick); err != nil {
		return fmt.Errorf("process surrender: %w", err)
	}

	cs.logger.Info().
		Str("combat_id", combatID).
		Str("player_id", playerID).
		Int("credits_lost", lossAmount).
		Int64("tick", tick).
		Msg("player surrendered")

	return nil
}

// endCombat terminates a combat instance and cleans up
func (cs *CombatSystem) endCombat(combat *CombatInstance, reason string, tick int64) error {
	// Update combat status
	var status CombatStatus
	switch reason {
	case "destroyed":
		status = CombatEnded
	case "fled":
		status = CombatFled
	default:
		status = CombatEnded
	}

	if err := cs.db.UpdateCombatStatus(combat.CombatID, status, tick); err != nil {
		return fmt.Errorf("end combat: %w", err)
	}

	// Update player ship status back to IN_SPACE
	if err := cs.db.UpdateShipStatus(combat.PlayerShipID, "IN_SPACE", tick); err != nil {
		return fmt.Errorf("end combat: %w", err)
	}

	// Delete ephemeral pirate ship
	if err := cs.db.DeletePirateShip(combat.PirateShipID); err != nil {
		cs.logger.Warn().
			Err(err).
			Str("pirate_ship_id", combat.PirateShipID).
			Msg("failed to delete pirate ship")
	}

	// Delete combat instance
	if err := cs.db.DeleteCombatInstance(combat.CombatID); err != nil {
		cs.logger.Warn().
			Err(err).
			Str("combat_id", combat.CombatID).
			Msg("failed to delete combat instance")
	}

	return nil
}

// handleShipDestruction processes player ship destruction
func (cs *CombatSystem) handleShipDestruction(ship *Ship, combat *CombatInstance, tick int64) error {
	cs.logger.Info().
		Str("ship_id", ship.ShipID).
		Str("player_id", ship.PlayerID).
		Str("combat_id", combat.CombatID).
		Msg("player ship destroyed")

	// Clear cargo
	if err := cs.db.ClearShipCargo(ship.ShipID); err != nil {
		return fmt.Errorf("handle ship destruction: %w", err)
	}

	// Find nearest port for respawn
	nearestPort, err := cs.db.FindNearestPort(ship.CurrentSystemID)
	if err != nil {
		return fmt.Errorf("handle ship destruction: %w", err)
	}

	// Respawn ship at nearest port
	if err := cs.db.RespawnShip(ship.ShipID, ship.CurrentSystemID, nearestPort, tick); err != nil {
		return fmt.Errorf("handle ship destruction: %w", err)
	}

	// Grant insurance payout
	credits, err := cs.db.GetPlayerCredits(ship.PlayerID)
	if err != nil {
		return fmt.Errorf("handle ship destruction: %w", err)
	}

	newCredits := int(credits) + cs.cfg.InsurancePayout
	if err := cs.db.UpdatePlayerCredits(ship.PlayerID, newCredits); err != nil {
		return fmt.Errorf("handle ship destruction: %w", err)
	}

	// End combat
	if err := cs.endCombat(combat, "ended", tick); err != nil {
		return fmt.Errorf("handle ship destruction: %w", err)
	}

	cs.logger.Info().
		Str("player_id", ship.PlayerID).
		Int("insurance_payout", cs.cfg.InsurancePayout).
		Int("respawn_port", nearestPort).
		Msg("ship respawned with insurance payout")

	return nil
}

// ResolveCombatTick processes all active combat instances (currently no-op as counter-attacks happen in ProcessAttack)
func (cs *CombatSystem) ResolveCombatTick(tick int64) ([]*CombatEvent, error) {
	// In Phase 1, pirate counter-attacks happen immediately after player attacks
	// This method is reserved for future tick-based combat mechanics
	return nil, nil
}

// CheckPirateSpawns evaluates spawn probability for ships in Low Security systems
func (cs *CombatSystem) CheckPirateSpawns(tick int64) ([]*CombatEvent, error) {
	// Only check every N ticks
	if tick%int64(cs.cfg.SpawnCheckInterval) != 0 {
		return nil, nil
	}

	// Get all ships in space
	ships, err := cs.db.GetShipsInSpace()
	if err != nil {
		return nil, fmt.Errorf("check pirate spawns: %w", err)
	}

	events := make([]*CombatEvent, 0)

	for _, ship := range ships {
		// Get system security level
		securityLevel, err := cs.db.GetSystemSecurityLevel(ship.CurrentSystemID)
		if err != nil {
			cs.logger.Warn().
				Err(err).
				Int("system_id", ship.CurrentSystemID).
				Msg("failed to get system security level")
			continue
		}

		// Skip if not Low Security (< 0.4)
		if securityLevel >= 0.4 {
			continue
		}

		// Calculate spawn chance
		spawnChance := cs.cfg.PirateActivityBase * (1.0 - securityLevel)
		roll := rand.Float64()

		if roll < spawnChance {
			// Spawn pirate
			combat, err := cs.SpawnPirate(ship.CurrentSystemID, ship.ShipID, tick)
			if err != nil {
				cs.logger.Warn().
					Err(err).
					Str("ship_id", ship.ShipID).
					Int("system_id", ship.CurrentSystemID).
					Msg("failed to spawn pirate")
				continue
			}

			events = append(events, &CombatEvent{
				Type:     "spawn",
				CombatID: combat.CombatID,
				PlayerID: ship.PlayerID,
				SystemID: ship.CurrentSystemID,
				Tick:     tick,
				Details: map[string]interface{}{
					"pirate_ship_id": combat.PirateShipID,
				},
			})
		}
	}

	return events, nil
}

// GetActiveCombat retrieves the active combat instance for a ship
func (cs *CombatSystem) GetActiveCombat(shipID string) (*CombatInstance, error) {
	combat, err := cs.db.GetActiveCombatByShip(shipID)
	if err != nil {
		return nil, fmt.Errorf("get active combat: %w", err)
	}
	return combat, nil
}

// GetPirateShip retrieves pirate ship data
func (cs *CombatSystem) GetPirateShip(pirateShipID string) (*PirateShip, error) {
	pirate, err := cs.db.GetPirateShip(pirateShipID)
	if err != nil {
		return nil, fmt.Errorf("get pirate ship: %w", err)
	}
	return pirate, nil
}
