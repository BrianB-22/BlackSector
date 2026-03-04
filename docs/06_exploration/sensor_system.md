# Sensor System Specification
## Version: 0.1
## Status: Draft
## Owner: Game Systems Design
## Last Updated: 2026-03-04

---

# 1. Purpose

The Sensor System defines how ships detect signals, objects, and environmental phenomena in space.

Sensors are a foundational system used by multiple gameplay features including:

- exploration
- combat detection
- anomaly discovery
- navigation safety
- environmental awareness

The system models long-range detection similar to **submarine sonar concepts**, where objects emit signals that must be detected and resolved.

---

# 2. Scope

IN SCOPE:

- sensor range calculations
- signal detection mechanics
- signal resolution mechanics
- passive and active scanning
- sensor modifiers from ships and modules

OUT OF SCOPE:

- radar visualization UI
- targeting systems
- weapon locking mechanics
- mission scripting

---

# 3. Design Principles

The sensor system follows several core principles:

- detection should rarely be instantaneous
- larger objects are easier to detect
- stronger signals travel farther
- better sensors improve resolution speed
- distance and interference reduce detection quality

Sensors provide **probabilistic detection rather than guaranteed discovery**.

---

# 4. Core Concepts

### Signal

An emission produced by an object that can be detected by sensors.

Examples:

- mineral signatures
- thermal emissions
- electromagnetic radiation
- gravitational disturbances

---

### Passive Scan

Continuous sensor monitoring without emitting detectable signals.

Characteristics:

- lower detection range
- stealth operation
- no signal broadcast

---

### Active Scan

A directed sensor pulse used to detect signals more effectively.

Characteristics:

- higher detection range
- faster signal resolution
- increased detectability by other ships

---

### Signal Resolution

The process of improving knowledge about a detected signal until an object can be located.

---

# 5. Data Model

## Entity: SensorProfile

- sensor_id: UUID
- sensor_type: enum
- base_range: float
- signal_resolution_rate: float
- power_usage: float

---

## Entity: SignalSource

- signal_id: UUID
- source_object_id: UUID
- signal_type: enum
- signal_strength: float
- position: Vector3

---

## Entity: SensorContact

Transient

- contact_id: UUID
- player_id: UUID
- signal_type: enum
- estimated_direction: Vector3
- estimated_distance: float
- resolution_progress: float

---

# 6. State Machine (If Applicable)

Signal contacts progress through the following states:

UNDETECTED → DETECTED → RESOLVING → LOCKED

---

State transitions:

UNDETECTED → DETECTED

Occurs when signal strength exceeds detection threshold.

DETECTED → RESOLVING

Repeated scans increase signal resolution.

RESOLVING → LOCKED

Resolution reaches threshold and object location becomes known.

---

# 7. Core Mechanics

Sensor scans evaluate signals within range.

Typical detection workflow:

Ship sensors scan region  
↓  
Signal sources evaluated  
↓  
Detection probability calculated  
↓  
Signal contact created  
↓  
Repeated scans increase resolution  
↓  
Signal locked and object location revealed

Sensor performance is affected by:

- ship sensor strength
- signal strength
- distance
- environmental interference

---

# 8. Mathematical Model

Variables:

SensorStrength  
SignalStrength  
Distance  
SensorRange

---

DetectionProbability =

(SensorStrength × SignalStrength)  
÷ (Distance × DetectionDifficulty)

---

ResolutionGain =

BaseResolutionRate × SensorStrength

Resolution is capped at 1.0.

---

# 9. Tunable Parameters

BaseSensorRange = 10000 km

BaseResolutionRate = 0.05

MinimumDetectionChance = 0.02

ActiveScanRangeMultiplier = 2.0

ActiveScanDetectionPenalty = 1.5

---

# 10. Integration Points

Depends On:

- Exploration System
- Ship System
- Module System
- Region Partitioning

Provides data to:

- Exploration discovery mechanics
- Combat detection systems
- Navigation hazard detection

---

# 11. Failure & Edge Cases

If signal source becomes invalid:

Contact must be removed.

If player leaves sensor range:

Resolution progress decays.

Multiple ships detecting the same signal produce separate contacts.

---

# 12. Performance Constraints

Sensor scans must operate within region partitions.

Expected performance:

- <5ms processing per scan
- scalable to thousands of signals per region

Sensor queries must avoid global universe scans.

---

# 13. Security Considerations

All detection calculations occur server-side.

Clients cannot fabricate signal contacts.

Sensor results must be validated by server logic.

---

# 14. Telemetry & Logging

Tracked metrics:

- signals detected per hour
- average detection distance
- resolution time
- active scan usage

Telemetry supports balancing detection difficulty.

---

# 15. Balancing Guidelines

Sensors should reward specialized exploration ships.

Balancing targets:

- exploration ships detect signals faster
- combat ships have limited sensor capability
- stealth ships reduce detection probability

Sensor gameplay should emphasize **information warfare**.

---

# 16. Non-Goals (v1)

The sensor system will not include:

- deployable sensor probes
- automated scanning drones
- passive listening arrays
- planetary scanning

---

# 17. Future Extensions

Potential future features include:

- deep-space sensor arrays
- stealth field generators
- cooperative fleet scanning
- sensor jamming mechanics

---

# End of Document