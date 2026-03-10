package navigation

// ShipStatus represents the current state of a ship
type ShipStatus string

const (
	StatusDocked    ShipStatus = "DOCKED"
	StatusInSpace   ShipStatus = "IN_SPACE"
	StatusInCombat  ShipStatus = "IN_COMBAT"
	StatusDestroyed ShipStatus = "DESTROYED"
)

// Ship represents a player's vessel
type Ship struct {
	ShipID           string
	PlayerID         string
	ShipClass        string
	HullPoints       int
	MaxHullPoints    int
	ShieldPoints     int
	MaxShieldPoints  int
	EnergyPoints     int
	MaxEnergyPoints  int
	CargoCapacity    int
	MissilesCurrent  int
	CurrentSystemID  int
	PositionX        float64
	PositionY        float64
	Status           ShipStatus
	DockedAtPortID   *int
	LastUpdatedTick  int64
}

// JumpConnection represents a navigable route between systems
type JumpConnection struct {
	ConnectionID     int
	FromSystemID     int
	ToSystemID       int
	Bidirectional    bool
	FuelCostModifier float64
}
// SystemMapData contains information about a system and its connections for display
type SystemMapData struct {
	CurrentSystem   *SystemInfo
	JumpConnections []*JumpConnectionInfo
	Ports           []*PortInfo
}

// SystemInfo contains display information about a system
type SystemInfo struct {
	SystemID      int
	Name          string
	SecurityLevel float64
	SecurityZone  string
}

// JumpConnectionInfo contains display information about a jump connection
type JumpConnectionInfo struct {
	DestinationSystemID   int
	DestinationSystemName string
	SecurityLevel         float64
	SecurityZone          string
	FuelCost              int
}

// PortInfo contains display information about a port
type PortInfo struct {
	PortID   int
	Name     string
	PortType string
}

