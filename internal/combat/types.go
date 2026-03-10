package combat

// CombatStatus represents the state of a combat encounter
type CombatStatus string

const (
	CombatActive CombatStatus = "ACTIVE"
	CombatEnded  CombatStatus = "ENDED"
	CombatFled   CombatStatus = "FLED"
)

// CombatInstance represents an active combat encounter between player and pirate
type CombatInstance struct {
	CombatID     string
	PlayerShipID string
	PirateShipID string
	SystemID     int
	StartTick    int64
	Status       CombatStatus
	TurnNumber   int
}

// PirateShip represents an ephemeral NPC enemy
type PirateShip struct {
	ShipID          string
	Tier            string // "raider" or "marauder"
	HullPoints      int
	MaxHull         int
	ShieldPoints    int
	MaxShield       int
	WeaponDamageMin int
	WeaponDamageMax int
	Accuracy        float64 // 0.60 for raider, 0.65 for marauder
	FleeThreshold   float64 // 0.15 for raider, 0.10 for marauder
}

// DamageRange represents min/max damage values
type DamageRange struct {
	Min int
	Max int
}

// CombatResult contains the outcome of an attack action
type CombatResult struct {
	Hit             bool
	Damage          int
	ShieldDamage    int
	HullDamage      int
	TargetShield    int
	TargetHull      int
	TargetDestroyed bool
	TargetFled      bool
}

// FleeResult contains the outcome of a flee attempt
type FleeResult struct {
	Success bool
	Reason  string
}

// CombatEvent represents a significant combat occurrence for logging/notification
type CombatEvent struct {
	Type       string // "spawn", "destroyed", "fled", "player_destroyed"
	CombatID   string
	PlayerID   string
	SystemID   int
	Tick       int64
	Details    map[string]interface{}
}

// PirateTierConfig defines stats for a pirate tier
type PirateTierConfig struct {
	Tier            string
	Hull            int
	Shield          int
	DamageMin       int
	DamageMax       int
	Accuracy        float64
	FleeThreshold   float64
	SpawnWeight     float64 // Probability weight (raider=0.70, marauder=0.30)
}

// Ship represents a player's vessel (minimal interface for combat)
type Ship struct {
	ShipID          string
	PlayerID        string
	ShipClass       string
	HullPoints      int
	MaxHullPoints   int
	ShieldPoints    int
	MaxShieldPoints int
	WeaponDamage    int
	Status          string
	CurrentSystemID int
}
