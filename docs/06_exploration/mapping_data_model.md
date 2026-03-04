# Mapping Data Model Specification
## Version: 0.1
## Status: Draft
## Owner: Game Systems Design
## Last Updated: 2026-03-04

---

# 1. Purpose

The Mapping Data Model defines how exploration knowledge about the universe is stored, updated, and shared.

Mapping represents the **player’s accumulated understanding of space**. It records discoveries, resolved signals, surveyed objects, and environmental data gathered through exploration activities.

Mapping data enables other systems to consume exploration knowledge, including:

- navigation
- mining
- missions
- economic intelligence
- hazard awareness

The mapping system does **not generate exploration objects**. It records player discovery results.

---

# 2. Scope

IN SCOPE:

- storage of exploration discoveries
- player knowledge tracking
- mapping resolution levels
- exploration data persistence
- discovery ownership
- map update propagation

OUT OF SCOPE:

- universe generation
- anomaly generation
- sensor detection calculations
- navigation pathfinding
- mission generation logic

---

# 3. Design Principles

Mapping data must follow these rules:

- Server authoritative
- Deterministic persistence
- Player-specific knowledge
- Efficient spatial lookup
- Scalable for large universes

Mapping should represent **information, not terrain generation**.

Players gradually build their own map knowledge through exploration.

---

# 4. Core Concepts

### Region Map

A region map represents a spatial area containing exploration data.

Regions correspond to the **region partitioning system** used by the universe.

---

### Map Entry

A map entry represents a known exploration object.

Map entries are created when objects reach the **LOCATED discovery state**.

---

### Discovery Record

Tracks which players have discovered which objects.

Each discovery record links a player to an exploration object.

---

### Mapping Accuracy

Mapping accuracy reflects the level of information known about an area.

Levels include:

UNKNOWN  
PARTIAL  
MAPPED  
SURVEYED

---

# 5. Data Model

## Entity: RegionMap

Persistent

- region_id: UUID
- asteroid_density: float
- hazard_level: float
- navigation_difficulty: float
- mapping_accuracy: enum

---

## Entity: MapEntry

Persistent

- entry_id: UUID
- object_id: UUID
- region_id: UUID
- object_type: enum
- coordinates: Vector3
- discovery_level: enum
- discovered_by: UUID
- discovery_timestamp: datetime

---

## Entity: PlayerMapKnowledge

Persistent

- player_id: UUID
- region_id: UUID
- mapping_accuracy: enum
- last_updated: datetime

---

## Entity: DiscoveryRecord

Persistent

- discovery_id: UUID
- player_id: UUID
- object_id: UUID
- discovery_state: enum
- discovery_timestamp: datetime
- survey_progress: float

---

# 6. State Machine (If Applicable)

Map entries follow the same discovery progression as exploration objects.

UNKNOWN → DETECTED → LOCATED → SURVEYED

Transitions occur through exploration scanning and survey actions.

---

Transition definitions:

DETECTED

Signal contact recorded but no coordinates available.

LOCATED

Object coordinates resolved.

SURVEYED

Full object attributes revealed.

---

# 7. Core Mechanics

Mapping updates occur when exploration discovery states change.

Typical workflow:

Player scans region  
↓  
Signal detected  
↓  
Signal resolved to object location  
↓  
Map entry created  
↓  
Object surveyed  
↓  
Map entry updated with full attributes

Mapping knowledge is stored independently per player.

---

# 8. Mathematical Model

Mapping itself does not use complex calculations.

However discovery thresholds affect map creation.

Variables:

ResolutionProgress  
SurveyProgress

---

MapEntryCreationThreshold =

ResolutionProgress ≥ 1.0

---

SurveyCompletionThreshold =

SurveyProgress ≥ 1.0

---

# 9. Tunable Parameters

RegionMapRefreshInterval = 60 seconds

MapEntryCreationThreshold = 1.0

SurveyCompletionThreshold = 1.0

MapKnowledgeDecayRate = 0 (disabled in v1)

---

# 10. Integration Points

Depends On:

- Exploration System
- Region Partitioning System
- Data Persistence Layer
- Tick Engine

Provides information to:

- Navigation System
- Mining System
- Mission System
- Economic Data Market

---

# 11. Failure & Edge Cases

Invalid object references must be rejected.

If a discovered object is removed (admin utility):

Map entry must be marked **inactive**.

Concurrent discovery by multiple players produces independent discovery records.

---

# 12. Performance Constraints

Mapping queries must operate using region partitions.

Expected constraints:

- <3ms region map lookup
- scalable to millions of map entries

Mapping writes should occur asynchronously to avoid gameplay delays.

---

# 13. Security Considerations

Client cannot directly modify mapping data.

Discovery records validated server-side.

Map entries only created through valid exploration mechanics.

---

# 14. Telemetry & Logging

Tracked metrics:

- discoveries per region
- anomaly discovery frequency
- map coverage per player
- average time to survey objects

Telemetry used to tune exploration density.

---

# 15. Balancing Guidelines

Mapping progression should feel gradual.

Design targets:

- new players discover frequent minor objects
- high-value discoveries remain rare
- fully mapping a region should require significant effort

Mapping knowledge should provide **strategic advantage** without trivializing exploration.

---
## Entity: NavigationCorridor

Persistent

Navigation corridors represent discovered safe or optimal travel paths through hazardous or inefficient regions of space.

Corridors are pre-generated during universe creation and may also be inserted by administrative utilities.

Corridors become available to players only after they are discovered through exploration.

Fields:

corridor_id: UUID  
region_id: UUID  
start_position: Vector3  
end_position: Vector3  
corridor_width: float  
hazard_reduction: float  
discovered_by: UUID  
discovery_timestamp: datetime  

---

## Corridor Map Integration

Navigation corridors are stored as map features within a player's exploration knowledge.

Corridors allow the navigation system to calculate safer or shorter travel routes through hazardous regions.

Corridors may significantly reduce route hazard cost during pathfinding calculations.

---

## Corridor Discovery

Corridors are discovered through exploration scanning.

Discovery workflow:

Player scans region  
↓  
Weak structural signal detected  
↓  
Signal resolution identifies corridor path  
↓  
Corridor mapped  
↓  
Corridor stored in player mapping data

# 16. Non-Goals (v1)

The mapping system will not include:

- shared faction maps
- global discovery broadcasts
- dynamic map decay
- automated exploration probes

---

# 17. Future Extensions

Potential future features include:

- player-sold map data
- faction intelligence sharing
- dynamic map updates
- exploration contracts
- shared fleet mapping

---

# End of Document