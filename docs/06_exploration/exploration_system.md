# Exploration System Specification
## Version: 0.1
## Status: Draft
## Owner: Game Systems Design
## Last Updated: 2026-03-04

---

# 1. Purpose

The exploration subsystem governs how players discover hidden objects, signals, and anomalies within the universe.

The system supports the gameplay pillar of **discovery and intelligence gathering**.

Exploration provides players with information advantages that can influence:

- mining
- navigation
- trade routes
- mission opportunities
- strategic positioning

All exploration targets are **pre-generated during universe creation**.  
Exploration mechanics reveal these objects progressively through scanning and signal analysis.

Exploration **does not generate new world objects**.

---

# 2. Scope

IN SCOPE:

- Signal detection
- Progressive discovery
- Exploration scanning
- Discovery state tracking
- Mapping updates
- Exploration interaction with anomalies
- Exploration interaction with mining targets
- Player discovery records

OUT OF SCOPE:

- Universe generation
- Procedural anomaly creation
- Combat detection mechanics
- Navigation pathfinding
- Mission generation

---

# 3. Design Principles

The exploration system follows these constraints:

- Server authoritative
- Deterministic discovery results
- No client-side object creation
- Tick-driven updates
- Signal-based discovery instead of direct object visibility
- Progressive information reveal

Exploration should feel like **long-range submarine detection in space**.

Objects are rarely instantly identified.

---

# 4. Core Concepts

### Exploration Object

A hidden world object that can be discovered through exploration.

Examples:

- asteroid clusters
- derelict ships
- anomalies
- hidden stations
- hazardous regions

---

### Signal

A detectable emission produced by an exploration object.

Signals provide incomplete information.

Signals include:

- signal type
- signal strength
- approximate direction
- approximate distance

---

### Discovery State

Represents the player's knowledge about an exploration object.

States:

UNDISCOVERED  
DETECTED  
LOCATED  
SURVEYED

---

### Survey

A detailed scan that reveals the full attributes of a discovered object.

---

# 5. Data Model

## Entity: ExplorationObject

Persistent

- object_id: UUID
- object_type: enum
- position: Vector2
- signal_type: enum
- signal_strength: float
- region_id: UUID
- is_hidden: bool

Generated during universe creation.

---

## Entity: SignalContact

Transient

- signal_id: UUID
- player_id: UUID
- signal_type: enum
- estimated_direction: Vector2
- estimated_distance: float
- signal_strength: float

Computed during scans.

---

## Entity: PlayerDiscovery

Persistent

- player_id: UUID
- object_id: UUID
- discovery_state: enum
- discovery_timestamp: datetime
- survey_progress: float

Tracks player knowledge.

---

# 6. State Machine (If Applicable)

Exploration discovery follows a defined state progression.

UNDISCOVERED → DETECTED → LOCATED → SURVEYED

Transitions:

UNDISCOVERED → DETECTED  
Occurs when signal is detected during scanning.

DETECTED → LOCATED  
Occurs when signal resolution reaches sufficient accuracy.

LOCATED → SURVEYED  
Occurs when a detailed scan completes.

Failure state:

If player leaves sensor range, resolution progress may decay.

---

# 7. Core Mechanics

Exploration operates through scanning actions.

Typical workflow:

Player performs scan  
↓  
Nearby signals evaluated  
↓  
Signals returned as contacts  
↓  
Repeated scans increase signal resolution  
↓  
Object coordinates determined  
↓  
Survey reveals full object data

Signal detection probability depends on:

- sensor strength
- signal strength
- distance
- environmental interference

---

# 8. Mathematical Model

## Variables

SignalStrength  
Range: 0 – 100

SensorStrength  
Range: 0 – 100

Distance  
Units: kilometers

ResolutionProgress  
Range: 0 – 1

---

## Formulas

DetectionProbability =

(SensorStrength / 100)  
× (SignalStrength / 100)  
÷ DistanceFactor

Where:

DistanceFactor = max(1, Distance / SensorRange)

---

ResolutionGain =

BaseResolutionRate  
× (SensorStrength / 100)

Resolution is capped at 1.0.

---

# 9. Tunable Parameters

BaseResolutionRate = 0.05  
MinimumDetectionChance = 0.02  
MaxDetectionDistance = 20000 km  
SignalDecayRate = 0.01 per tick  
SurveyCompletionThreshold = 1.0

These parameters are adjustable without code modification.

---

# 10. Integration Points

Depends On:

- Universe Generation System
- Region Partitioning
- Tick Engine
- Data Persistence Layer

Exposes:

- Discovery events
- Mapping updates
- Exploration telemetry
- Mission triggers

---

# 11. Failure & Edge Cases

Invalid states:

If object_id does not exist → scan ignored.

Exploit prevention:

Players cannot force discovery without scanning.

Race condition:

If multiple players discover simultaneously, discovery records remain independent.

Timeout behavior:

Signal resolution progress decays if player exits scan range.

---

# 12. Performance Constraints

Exploration scanning must not query the entire universe dataset.

Scan operations operate on **region partitions only**.

Expected constraints:

- <5ms scan processing per player action
- Must support thousands of exploration objects per region cluster

---

# 13. Security Considerations

Client cannot reveal hidden objects.

All signal calculations occur server-side.

Client requests only trigger scans.

Discovery states are server validated.

---

# 14. Telemetry & Logging

Events logged:

- signal detection events
- object discovery events
- survey completion events
- anomaly discoveries

Telemetry supports balancing exploration density and discovery pacing.

---

# 15. Balancing Guidelines

Exploration should reward risk and patience.

Design goals:

- discoveries should rarely be instantaneous
- rare anomalies should require extensive scanning
- high-risk regions should produce stronger signals
- exploration ships should outperform combat ships in scanning

---

# 16. Non-Goals (v1)

The exploration system will not include:

- procedural anomaly spawning
- player-deployed sensor probes
- cooperative scanning mechanics
- automated mapping drones

---

# 17. Future Extensions

Potential future features:

- faction intelligence networks
- exploration contracts
- deployable sensor arrays
- cooperative fleet surveying
- anomaly research systems

---

# End of Document