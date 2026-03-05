# Rare Anomaly Events Specification
## Version: 0.1
## Status: Draft
## Owner: Game Systems Design
## Last Updated: 2026-03-04

---

# 1. Purpose

The Rare Anomaly Events system defines extremely uncommon exploration discoveries that provide unique gameplay opportunities.

Rare anomalies represent the most valuable and mysterious discoveries in the universe. They provide:

- rare resources
- unique mission chains
- strategic locations
- narrative lore opportunities

Rare anomalies are **pre-generated during universe creation** and hidden within the exploration system.

They are intentionally scarce to preserve the excitement of discovery.

---

# 2. Scope

IN SCOPE:

- rare anomaly types
- anomaly discovery mechanics
- anomaly rarity distribution
- anomaly interaction with exploration
- anomaly event triggers

OUT OF SCOPE:

- procedural anomaly generation
- anomaly combat encounters
- story mission scripting
- anomaly destruction mechanics

---

# 3. Design Principles

Rare anomalies should follow these principles:

- extremely rare discoveries
- difficult to detect
- valuable rewards
- encourage exploration in dangerous regions
- produce memorable gameplay moments

Rare anomalies should not appear frequently enough to become routine discoveries.

---

# 4. Core Concepts

### Rare Anomaly

A unique or extremely rare exploration object with special gameplay properties.

Examples include:

- alien megastructures
- spatial distortions
- ancient derelicts
- exotic energy sources
- unstable wormholes

---

### Anomaly Signature

Rare anomalies emit unusual signal signatures that distinguish them from common exploration objects.

These signatures may include:

- gravitational disturbances
- exotic energy emissions
- irregular sensor patterns

---

### Anomaly Event

An event triggered when a rare anomaly is discovered or surveyed.

Events may include:

- mission generation
- discovery announcements
- research opportunities

---

# 5. Data Model

## Entity: RareAnomaly

Persistent

- anomaly_id: UUID
- anomaly_type: enum
- position: Vector2
- region_id: UUID
- signal_type: enum
- signal_strength: float
- rarity_tier: enum
- is_discovered: bool

---

## Entity: AnomalyDiscovery

Persistent

- discovery_id: UUID
- anomaly_id: UUID
- player_id: UUID
- discovery_timestamp: datetime
- survey_completed: bool

---

## Entity: AnomalyEvent

Transient

- event_id: UUID
- anomaly_id: UUID
- event_type: enum
- trigger_timestamp: datetime

---

# 6. State Machine (If Applicable)

Rare anomalies follow the standard exploration discovery progression.

UNDISCOVERED → DETECTED → LOCATED → SURVEYED

---

Additional event states:

SURVEYED → EVENT_TRIGGERED

Anomaly events may trigger when survey is completed.

---

# 7. Core Mechanics

Rare anomalies are hidden exploration objects.

Typical workflow:

Player scans region  
↓  
Weak anomaly signal detected  
↓  
Repeated scans increase resolution  
↓  
Signal resolved to anomaly location  
↓  
Player performs detailed survey  
↓  
Anomaly fully revealed  
↓  
Event triggered

Rare anomalies require stronger sensors or multiple scans to resolve.

---

# 8. Mathematical Model

Variables:

SignalStrength  
SensorStrength  
DetectionDifficulty  
Distance

---

DetectionProbability =

(SensorStrength × SignalStrength)  
÷ (Distance × DetectionDifficulty)

---

DetectionDifficulty is higher for rare anomalies.

Typical anomaly difficulty multiplier:

3.0 – 10.0

---

# 9. Tunable Parameters

CommonAnomalyDifficulty = 2.0

RareAnomalyDifficulty = 5.0

LegendaryAnomalyDifficulty = 10.0

AnomalySignalStrengthRange = 5 – 25

AnomalySurveyTimeMultiplier = 2.0

---

# 10. Integration Points

Depends On:

- Exploration System
- Mapping Data Model
- Mission System
- Universe Generation System

Provides triggers for:

- exploration missions
- research activities
- narrative content
- rare economic opportunities

---

# 11. Failure & Edge Cases

If anomaly removed by admin utility:

Discovery records remain but object marked inactive.

Multiple players discovering the anomaly simultaneously produce separate discovery records.

If player leaves scan range during survey:

Survey progress pauses.

---

# 12. Performance Constraints

Rare anomalies should not significantly increase scan processing cost.

Anomaly objects should be indexed by region.

Scan queries must remain region-scoped.

---

# 13. Security Considerations

Clients cannot directly detect anomalies.

All detection logic occurs server-side.

Anomaly discovery must originate from valid exploration mechanics.

---

# 14. Telemetry & Logging

Tracked metrics:

- anomaly discovery rate
- time to discovery
- survey completion frequency
- anomaly region distribution

Telemetry supports balancing rarity and discovery pacing.

---

# 15. Balancing Guidelines

Rare anomalies must remain rare enough to feel significant.

Balancing goals:

- anomalies should not cluster excessively
- legendary anomalies should be extremely scarce
- discovery should require meaningful exploration effort

Players discovering rare anomalies should feel a strong sense of achievement.

---

# 16. Non-Goals (v1)

The rare anomaly system will not include:

- dynamic anomaly spawning
- anomaly combat bosses
- destructible anomalies
- player-created anomalies

---

# 17. Future Extensions

Potential future features include:

- anomaly research systems
- anomaly-based technology unlocks
- faction conflict over anomalies
- exploration guild discovery rankings

---

# End of Document