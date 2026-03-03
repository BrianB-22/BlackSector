# Mining System Specification
## Version: 0.1
## Status: Draft
## Owner: Core Simulation
## Last Updated: 2026-03-02

---

# 1. Purpose

Defines the player-facing asteroid mining loop, including:

- Prospecting
- Extraction
- Depletion
- PvP exposure during mining
- Integration with hazard and drone subsystems

Mining is a probabilistic, energy-constrained economic activity.

---

# 2. Scope

IN SCOPE:
- Mining cycle execution
- Yield calculation
- Depletion behavior
- PvP exposure modifiers

OUT OF SCOPE:
- Hazard logic (see hazard_system.md)
- Drone mechanics (see drone_system.md)
- Rare anomaly events
- Economic price modeling

---

# 3. Design Principles

- Server-authoritative resolution
- Tick-based execution
- Risk scales with security tier
- Yield is probabilistic
- Mining increases vulnerability

---

# 4. Core Concepts

- Density: Resource concentration (0.2–1.0)
- InstabilityFactor: Extraction risk multiplier (0–0.8)
- DepletionLevel: Field exhaustion state (0–1)
- SecurityYieldModifier: Zone-based yield multiplier

---

# 5. Data Model

## Entity: AsteroidField

- field_id: UUID
- system_id: UUID
- density: float
- instability_factor: float
- depletion_level: float
- hazard_presence: bool
- mineral_class: enum

Persistent entity.

---

# 6. State Machine

AVAILABLE → ACTIVE_MINING → DEPLETING → LOW_YIELD → REGENERATING

Transitions driven by depletion and tick regeneration.

---

# 7. Core Mechanics

Mining is executed in 1-tick cycles.

Each cycle:
- EnergyCost = 20
- HeatIncrease = +6
- Yield calculated
- Density reduced

Mining reduces ship velocity by 30%.

---

# 8. Mathematical Model

## Variables

- BaseYield = 50
- RandomFactor ∈ [0.6 – 1.4]
- SecurityYieldModifier:
  - High Sec: 0.7
  - Medium Sec: 1.0
  - Low Sec: 1.3

## Yield Formula

Yield =

(BaseYield × Density)
× RandomFactor
× (1 − InstabilityFactor)
× SecurityYieldModifier

---

# 9. Tunable Parameters

- BaseYield
- Density depletion rate (default 0.02)
- Regeneration rate (0.005 per 100 ticks)
- SecurityYieldModifier values

---

# 10. Integration Points

Depends On:
- Hazard System
- Drone System
- Economy Engine
- Tick Engine

Exposes:
- MiningStartEvent
- MiningYieldEvent

---

# 11. Failure & Edge Cases

If Energy < 20:
Mining cycle fails.

If hazard undetected:
Hazard system invoked.

Density floor = 0.1

---

# 12. Performance Constraints

Must process <1ms per active mining instance.

---

# 13. Security Considerations

- No client-side yield calculation
- No bypass of energy check
- No density underflow

---

# 14. Telemetry & Logging

Log:
- Yield distribution
- Instability failure rate
- Mining per security tier

---

# 15. Balancing Guidelines

- Low security must significantly out-reward high security
- Instability must create occasional zero-yield cycles
- Mining risk must scale naturally with PvP density

---

# 16. Non-Goals (v1)

- Cooperative fleet mining
- Automated mining bots
- Player-owned mining stations

---

# 17. Future Extensions

- Subsystem damage
- Equipment degradation
- Mining specialization modules

---

# End of Document
