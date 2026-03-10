package missions

// MissionStatus represents the current state of a mission instance
type MissionStatus string

const (
	MissionAvailable  MissionStatus = "AVAILABLE"
	MissionAccepted   MissionStatus = "ACCEPTED"
	MissionInProgress MissionStatus = "IN_PROGRESS"
	MissionCompleted  MissionStatus = "COMPLETED"
	MissionFailed     MissionStatus = "FAILED"
	MissionExpired    MissionStatus = "EXPIRED"
	MissionAbandoned  MissionStatus = "ABANDONED"
)

// ObjectiveStatus represents the current state of a mission objective
type ObjectiveStatus string

const (
	ObjectivePending   ObjectiveStatus = "PENDING"
	ObjectiveActive    ObjectiveStatus = "ACTIVE"
	ObjectiveCompleted ObjectiveStatus = "COMPLETED"
)

// MissionDefinition represents a mission template loaded from JSON
type MissionDefinition struct {
	MissionID           string                 `json:"mission_id"`
	Name                string                 `json:"name"`
	Description         string                 `json:"description"`
	Version             string                 `json:"version"`
	Author              string                 `json:"author"`
	Enabled             bool                   `json:"enabled"`
	Repeatable          bool                   `json:"repeatable"`
	RepeatCooldownTicks int                    `json:"repeat_cooldown_ticks"`
	SecurityZones       []string               `json:"security_zones"`
	AvailableAtPorts    []int                  `json:"available_at_ports,omitempty"`
	ExpiryTicks         *int                   `json:"expiry_ticks,omitempty"`
	Objectives          []*ObjectiveDefinition `json:"objectives"`
	Rewards             *RewardDefinition      `json:"rewards"`
}

// ObjectiveDefinition defines a mission objective
type ObjectiveDefinition struct {
	ObjectiveID string                 `json:"objective_id"`
	Type        string                 `json:"type"` // "deliver_commodity", "navigate_to", "kill"
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// RewardDefinition defines mission rewards
type RewardDefinition struct {
	Credits int           `json:"credits"`
	Items   []*ItemReward `json:"items,omitempty"`
}

// ItemReward represents an item reward
type ItemReward struct {
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
}

// MissionInstance represents an active mission for a player
type MissionInstance struct {
	InstanceID            string
	MissionID             string
	PlayerID              string
	Status                MissionStatus
	AcceptedTick          int64
	StartedTick           *int64
	CompletedTick         *int64
	FailedReason          *string
	ExpiresAtTick         *int64
}

// ObjectiveProgress tracks progress on a mission objective
type ObjectiveProgress struct {
	InstanceID     string
	ObjectiveIndex int
	Status         string
	CurrentValue   int
	RequiredValue  int
}

// MissionEvent represents a mission state change event
type MissionEvent struct {
	Type       string // "completed", "failed", "expired", "accepted", "abandoned"
	InstanceID string
	PlayerID   string
	MissionID  string
	Tick       int64
	Details    map[string]interface{}
}

// MissionListing represents a mission available to a player
type MissionListing struct {
	MissionID       string
	Name            string
	Description     string
	Objectives      []*ObjectiveDefinition
	Rewards         *RewardDefinition
	ExpiryTicks     *int
	Repeatable      bool
	SecurityZones   []string
}

// MissionFile represents the JSON structure of a mission configuration file
type MissionFile struct {
	Missions []*MissionDefinition `json:"missions"`
}
