# Exploration Ships Specification
## Version: 0.1
## Status: Draft
## Owner: Game Systems Design
## Last Updated: 2026-03-04

---

# 1. Purpose

The exploration ships subsystem defines ship classes and capabilities optimized for exploration activities.

Exploration vessels prioritize sensor performance, signal resolution, and operational range over combat capability.

These ships enable players to efficiently locate hidden signals, resolve exploration targets, and survey anomalies.

---

# 2. Scope

IN SCOPE:

- exploration ship classes
- exploration ship capabilities
- sensor performance modifiers
- stealth characteristics related to exploration
- ship-based exploration bonuses

OUT OF SCOPE:

- combat ship balance
- ship construction mechanics
- sensor mechanics implementation
- ship manufacturing systems

---

# 3. Design Principles

Exploration ships should:

- outperform combat ships in sensor capability
- have limited combat strength
- provide long operational range
- allow safer operation in hazardous regions
- encourage solo exploration gameplay

Exploration vessels emphasize **information advantage rather than combat power**.

---

# 4. Core Concepts

### Exploration Ship

A ship class designed to maximize detection and survey efficiency.

Characteristics include:

- enhanced sensors
- improved signal resolution
- reduced detection signature
- extended travel range

---

### Sensor Efficiency

Sensor efficiency determines how quickly signals can be detected and resolved.

Exploration ships receive bonuses to sensor efficiency.

---

### Survey Capability

Survey capability affects how quickly exploration objects can be fully analyzed once located.

---

# 5. Data Model

## Entity: ShipClass

- class_id: UUID
- class_name: string
- ship_role: enum
- sensor_modifier: float
- stealth_modifier: float
- survey_speed_modifier: float
- cargo_capacity: float
- combat_rating: float

---

## Entity: ShipModuleSlot

- slot_id: UUID
- ship_class_id: UUID
- module_type: enum
- slot_size: enum

Exploration ships typically include additional sensor module slots.

---

# 6. State Machine (If Applicable)

Exploration ships do not require a state machine.

However, ship activity modes may influence sensor performance.

Operational modes include:

TRAVEL  
SCAN  
SURVEY

Mode transitions occur based on player actions.

---

# 7. Core Mechanics

Exploration ships influence exploration mechanics through stat modifiers.

Modifiers include:

SensorRangeModifier  
SignalResolutionModifier  
SurveySpeedModifier

These modifiers affect exploration calculations.

Example workflow:

Player enters scan mode  
↓  
Ship sensors activate  
↓  
Nearby signals evaluated  
↓  
Signal resolution rate modified by ship stats  
↓  
Object discovered

---

# 8. Mathematical Model

Variables:

BaseSensorRange  
ShipSensorModifier  
BaseResolutionRate

---

EffectiveSensorRange =

BaseSensorRange × ShipSensorModifier

---

EffectiveResolutionRate =

BaseResolutionRate × ShipSensorModifier

---

SurveyCompletionTime =

BaseSurveyTime ÷ SurveySpeedModifier

---

# 9. Tunable Parameters

ScoutSensorModifier = 1.25  
SurveyShipSensorModifier = 1.5  
ExplorerSensorModifier = 1.75

ScoutSurveyModifier = 1.0  
SurveyShipSurveyModifier = 1.5  
ExplorerSurveyModifier = 1.75

CombatRatingPenalty = -40%

---

# 10. Integration Points

Depends On:

- Exploration System
- Sensor System
- Ship System
- Module System

Provides modifiers to:

- signal detection
- signal resolution
- survey operations

---

# 11. Failure & Edge Cases

Sensor modules damaged → reduced sensor effectiveness.

Ship destroyed → exploration progress lost if not persisted.

Incorrect module installation → exploration bonuses disabled.

---

# 12. Performance Constraints

Ship modifiers must be applied during sensor calculations without adding significant overhead.

Modifier evaluation should remain constant-time per scan event.

---

# 13. Security Considerations

All ship modifiers validated server-side.

Client cannot modify exploration bonuses.

Ship stats retrieved from authoritative server data.

---

# 14. Telemetry & Logging

Tracked events:

- exploration ship usage
- signals detected by ship class
- anomaly discoveries per ship class
- survey completion time

Telemetry used to balance exploration efficiency.

---

# 15. Balancing Guidelines

Exploration ships should:

- detect signals faster than combat ships
- survey objects significantly faster
- remain vulnerable in combat scenarios

Exploration ships should encourage risk-reward gameplay.

---

# 16. Non-Goals (v1)

The exploration ships system will not include:

- automated exploration drones
- AI exploration fleets
- remote scanning satellites

---

# 17. Future Extensions

Potential future enhancements:

- deployable sensor probes
- fleet-based cooperative scanning
- long-range exploration cruisers
- automated mapping modules

---

# End of Document