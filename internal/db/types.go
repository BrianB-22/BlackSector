package db

// Player represents a registered player account
type Player struct {
	PlayerID     string
	PlayerName   string
	SSHUsername  string
	TokenHash    string
	PasswordHash string
	Credits      int64
	CreatedAt    int64
	LastLoginAt  *int64
	IsBanned     bool
}

// SessionState represents the state of a player session
type SessionState string

const (
	SessionConnected            SessionState = "CONNECTED"
	SessionDisconnectedLingering SessionState = "DISCONNECTED_LINGERING"
	SessionDockedOffline        SessionState = "DOCKED_OFFLINE"
	SessionTerminated           SessionState = "TERMINATED"
)

// Session represents an active player connection
type Session struct {
	SessionID      string
	PlayerID       string
	InterfaceMode  string // "TEXT" or "GUI"
	State          SessionState
	ConnectedAt    int64
	DisconnectedAt *int64
	LingerExpiryAt *int64
	LastActivityAt int64
}

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
	Status           string
	DockedAtPortID   *int
	LastUpdatedTick  int64
}

// CargoSlot represents one cargo hold entry
type CargoSlot struct {
	ShipID      string
	SlotIndex   int
	CommodityID string
	Quantity    int
}

// PortInventory represents commodity inventory at a port
type PortInventory struct {
	PortID       int
	CommodityID  string
	Quantity     int
	BuyPrice     int
	SellPrice    int
	UpdatedTick  int64
}

// Port represents a trading station or starbase
type Port struct {
	PortID            int
	SystemID          int
	Name              string
	PortType          string
	SecurityLevel     float64
	DockingFee        int
	HasBank           bool
	HasShipyard       bool
	HasUpgradeMarket  bool
	HasRepair         bool
	HasFuel           bool
}
