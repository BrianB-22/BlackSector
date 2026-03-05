# Entity Models Specification

## Version: 0.1

## Status: Draft

## Owner: Core Architecture

## Last Updated: 2026-03-05

---

# 1. Purpose

Defines the runtime entity model for the BlackSector server.

This document describes entities as they exist in server memory during simulation — not as database rows or protocol messages.

Entity models are expressed as Go structs. Persistent fields are stored in SQLite via `database_schema.md`. Derived fields are computed at runtime and not persisted.

All positions are two-dimensional. No Z coordinate exists.

```go
type Vector2 struct {
    X float64
    Y float64
}
```

---

# 2. Player Domain

## Player

```go
type Player struct {
    PlayerID     string    // UUID
    PlayerName   string
    TokenHash    string    // not sent to clients
    CreatedAt    int64     // Unix epoch
    LastLoginAt  int64
    IsBanned     bool
}
```

---

## Session

```go
type Session struct {
    SessionID       string         // UUID — changes on each reconnect
    PlayerID        string
    InterfaceMode   InterfaceMode
    State           SessionState
    ConnectedAt     int64
    DisconnectedAt  *int64
    LingerExpiryAt  *int64
    LastActivityAt  int64
}

type InterfaceMode string
const (
    InterfaceModeText InterfaceMode = "TEXT"
    InterfaceModeGUI  InterfaceMode = "GUI"
)

type SessionState string
const (
    SessionStateConnected            SessionState = "CONNECTED"
    SessionStateDisconnectedLingering SessionState = "DISCONNECTED_LINGERING"
    SessionStateDockedOffline        SessionState = "DOCKED_OFFLINE"
    SessionStateTerminated           SessionState = "TERMINATED"
)
```

See `session_management.md` for lifecycle rules.

---

# 3. Ship Domain

## Ship

```go
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
    Cargo            []CargoSlot
    Upgrades         []ShipUpgrade
    CurrentSystemID  *int
    Position         Vector2
    Status           ShipStatus
    DockedAtPortID   *int
    LastUpdatedTick  int
}

type ShipStatus string
const (
    ShipStatusDocked    ShipStatus = "DOCKED"
    ShipStatusInSpace   ShipStatus = "IN_SPACE"
    ShipStatusInCombat  ShipStatus = "IN_COMBAT"
    ShipStatusDestroyed ShipStatus = "DESTROYED"
)
```

---

## CargoSlot

```go
type CargoSlot struct {
    SlotIndex   int
    CommodityID string
    Quantity    int
}
```

---

## ShipUpgrade

```go
type ShipUpgrade struct {
    UpgradeID        string
    InstalledAtTick  int
}
```

---

# 4. Universe Domain

## Region

```go
type Region struct {
    RegionID      int
    Name          string
    RegionType    RegionType
    SecurityLevel float64
}

type RegionType string
const (
    RegionTypeCore         RegionType = "core"
    RegionTypeAgricultural RegionType = "agricultural"
    RegionTypeIndustrial   RegionType = "industrial"
    RegionTypeFrontier     RegionType = "frontier"
    RegionTypeBlack        RegionType = "black"
)
```

Security level classification:
* `0.7–1.0` — High Security
* `0.4–0.7` — Medium Security
* `0.0–0.4` — Low Security
* `-1.0` — Black Sector (special flag)

---

## System

```go
type System struct {
    SystemID      int
    Name          string
    RegionID      int
    SecurityLevel float64
    Position      Vector2
    HazardLevel   float64
    // Loaded at startup, not persisted as fields:
    Ports       []*Port
    Connections []*JumpConnection
    HazardZones []*HazardZone
}
```

---

## JumpConnection

```go
type JumpConnection struct {
    ConnectionID       int
    FromSystemID       int
    ToSystemID         int
    Bidirectional      bool
    FuelCostModifier   float64
}
```

---

## HazardZone

```go
type HazardZone struct {
    HazardID      int
    SystemID      int
    HazardType    HazardType
    Position      Vector2
    Radius        float64
    DamagePerTick int
    Active        bool
}

type HazardType string
const (
    HazardTypeAsteroidField      HazardType = "asteroid_field"
    HazardTypeRadiationBelt      HazardType = "radiation_belt"
    HazardTypeDebrisCloud        HazardType = "debris_cloud"
    HazardTypeGravitationalAnomaly HazardType = "gravitational_anomaly"
)
```

---

# 5. Port and Economy Domain

## Port

```go
type Port struct {
    PortID        int
    SystemID      int
    Name          string
    PortType      PortType
    SecurityLevel float64
    DockingFee    int
    Inventory     map[string]*PortInventoryEntry  // keyed by commodity_id
}

type PortType string
const (
    PortTypeTrading     PortType = "trading"
    PortTypeMining      PortType = "mining"
    PortTypeRefueling   PortType = "refueling"
    PortTypeBlackMarket PortType = "black_market"
)
```

---

## PortInventoryEntry

```go
type PortInventoryEntry struct {
    CommodityID  string
    Quantity     int
    BuyPrice     int
    SellPrice    int
    UpdatedTick  int
}
```

---

## Commodity

```go
type Commodity struct {
    CommodityID  string
    Name         string
    Category     CommodityCategory
    BasePrice    int
    Volatility   float64
    IsContraband bool
}

type CommodityCategory string
const (
    CommodityCategoryEssential  CommodityCategory = "essential"
    CommodityCategoryIndustrial CommodityCategory = "industrial"
    CommodityCategoryLuxury     CommodityCategory = "luxury"
    CommodityCategoryExotic     CommodityCategory = "exotic"
)
```

---

## ActiveEconomicEvent

```go
type ActiveEconomicEvent struct {
    EventInstanceID  int
    EventID          string
    ScopeType        EventScopeType
    AffectedRegionID *int
    AffectedSystemID *int
    StartTick        int
    EndTick          int
    Visibility       EventVisibility
}

type EventScopeType string
const (
    EventScopeSystem  EventScopeType = "system"
    EventScopeRegion  EventScopeType = "region"
    EventScopeGalaxy  EventScopeType = "galaxy"
)

type EventVisibility string
const (
    EventVisibilityPublic EventVisibility = "public"
    EventVisibilityHidden EventVisibility = "hidden"
)
```

---

# 6. AI Trader Domain

## AITrader

```go
type AITrader struct {
    TraderID                int
    Name                    string
    ShipClass               string
    CurrentSystemID         *int
    Status                  TraderStatus
    HomeRegionID            *int
    CurrentCargoCommodity   *string
    CurrentCargoQuantity    int
    LastTradeTick           *int
}

type TraderStatus string
const (
    TraderStatusIdle      TraderStatus = "IDLE"
    TraderStatusTraveling TraderStatus = "TRAVELING"
    TraderStatusTrading   TraderStatus = "TRADING"
    TraderStatusDocked    TraderStatus = "DOCKED"
    TraderStatusDestroyed TraderStatus = "DESTROYED"
)
```

---

# 7. Mission Domain

## MissionInstance

```go
type MissionInstance struct {
    InstanceID     string
    MissionID      string       // references JSON config
    PlayerID       string
    Status         MissionStatus
    AcceptedTick   *int
    StartedTick    *int
    CompletedTick  *int
    FailedReason   *string
    ExpiresAtTick  *int
    Objectives     []ObjectiveProgress
}

type MissionStatus string
const (
    MissionStatusAvailable  MissionStatus = "AVAILABLE"
    MissionStatusAccepted   MissionStatus = "ACCEPTED"
    MissionStatusInProgress MissionStatus = "IN_PROGRESS"
    MissionStatusCompleted  MissionStatus = "COMPLETED"
    MissionStatusFailed     MissionStatus = "FAILED"
    MissionStatusExpired    MissionStatus = "EXPIRED"
    MissionStatusAbandoned  MissionStatus = "ABANDONED"
)
```

---

## ObjectiveProgress

```go
type ObjectiveProgress struct {
    ObjectiveIndex  int
    Status          ObjectiveStatus
    CurrentValue    int
    RequiredValue   int
}

type ObjectiveStatus string
const (
    ObjectiveStatusPending   ObjectiveStatus = "PENDING"
    ObjectiveStatusActive    ObjectiveStatus = "ACTIVE"
    ObjectiveStatusCompleted ObjectiveStatus = "COMPLETED"
    ObjectiveStatusFailed    ObjectiveStatus = "FAILED"
)
```

---

# 8. Exploration Domain

## PlayerMapEntry

```go
type PlayerMapEntry struct {
    PlayerID         string
    SystemID         int
    DiscoveredAtTick int
    ScanLevel        int  // 0=visited, 1=basic, 2=deep
}
```

---

## Anomaly

```go
type Anomaly struct {
    AnomalyID               int
    SystemID                int
    AnomalyType             AnomalyType
    Position                Vector2
    IsDiscovered            bool
    DiscoveredByPlayerID    *string
    ResourceQuantity        int
    Depleted                bool
}

type AnomalyType string
const (
    AnomalyTypeDerelictShip      AnomalyType = "derelict_ship"
    AnomalyTypeAncientArtifact   AnomalyType = "ancient_artifact"
    AnomalyTypeExoticGasCloud    AnomalyType = "exotic_gas_cloud"
    AnomalyTypeUnstableWormhole  AnomalyType = "unstable_wormhole"
    AnomalyTypeEnergyVortex      AnomalyType = "energy_vortex"
)
```

---

# 9. Mining Domain

## AsteroidField

```go
type AsteroidField struct {
    FieldID         int
    SystemID        int
    FieldType       FieldType
    Position        Vector2
    DepletionLevel  float64  // 0.0 = full, 1.0 = empty
    LastMinedTick   *int
    Resources       []FieldResource
}

type FieldType string
const (
    FieldTypeCommon   FieldType = "common"
    FieldTypeRich     FieldType = "rich"
    FieldTypeRare     FieldType = "rare"
    FieldTypeDepleted FieldType = "depleted"
)
```

---

## FieldResource

```go
type FieldResource struct {
    CommodityID       string
    BaseYield         int
    CurrentMultiplier float64
}
```

---

# 10. Navigation Domain

## Waypoint

```go
type Waypoint struct {
    WaypointID     int
    PlayerID       string
    Name           string
    SystemID       int
    Position       Vector2
    CreatedAtTick  int
}
```

---

# 11. Transient (Runtime-Only) Entities

These entities exist only in server memory and are never written to the database.

## CombatInstance

```go
type CombatInstance struct {
    CombatID       string   // UUID generated at combat start
    SystemID       int
    StartTick      int
    Status         CombatStatus
    Participants   []*CombatParticipant
    PendingActions []CombatAction
}

type CombatStatus string
const (
    CombatStatusActive    CombatStatus = "ACTIVE"
    CombatStatusResolving CombatStatus = "RESOLVING"
    CombatStatusEnded     CombatStatus = "ENDED"
)
```

---

## CombatParticipant

```go
type CombatParticipant struct {
    ParticipantID   string
    ParticipantType ParticipantType
    IsAlive         bool
}

type ParticipantType string
const (
    ParticipantTypePlayer ParticipantType = "player"
    ParticipantTypeNPC    ParticipantType = "npc"
)
```

---

## CommandEnvelope

```go
type CommandEnvelope struct {
    CorrelationID string
    SessionID     string
    PlayerID      string
    Command       string
    Parameters    map[string]any
    QueuedAtTick  int
}
```

---

# 12. Non-Goals (v1)

The following entity types are out of scope for v1:

* Faction entities and membership
* Fleet entities and compositions
* Player-owned stations or infrastructure
* Manufacturing chains
* Persistent NPC economies

---

# End of Document
