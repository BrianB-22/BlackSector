# Heat \& Energy Model Specification

## Version: 0.1
## Status: Draft
## Owner: Core Simulation

## Last Updated: 2026-03-02

---

# 1. Purpose

Defines the unified Heat and Energy systems governing ship performance during combat and high-intensity operations.

This subsystem controls:

- Action gating
- Performance degradation
- Tactical pacing
- Risk escalation
- Resource tradeoffs

Heat and Energy are core balancing levers for combat.

---

# 2. Scope

IN SCOPE:

- Energy capacity and regeneration
- Heat accumulation and dissipation
- Action energy costs
- Heat performance penalties
- Overheat behavior
- Integration with propulsion and combat

OUT OF SCOPE:

- Fuel economy (handled in propulsion)
- Reactor upgrade tree (future expansion)
- Subsystem damage modeling (future expansion)

---

# 3. Design Principles

- Energy limits immediate action frequency.
- Heat limits sustained aggression.
- High output increases vulnerability.
- Overheating must meaningfully degrade performance.
- No infinite action loops.
- All changes resolved in tick engine.

---

# 4. Core Concepts

- MaxEnergy
- EnergyLevel
- EnergyRegen
- HeatLevel
- HeatPenalty
- HeatDecay
- OverheatThreshold
- ActionEnergyCost
- ActionHeatCost

---

# 5. Data Model

## Entity: ShipPowerState

- max\_energy: float
- energy\_level: float
- base\_energy\_regen: float
- heat\_level: float
- heat\_decay\_rate: float
- overheat\_threshold: float
- reactor\_efficiency: float

Transient combat state.

---

# 6. State Model

NORMAL  

→ HIGH\_HEAT  

→ OVERHEATED  

→ RECOVERY  

→ NORMAL

---

# 7. Energy System

## 7.1 Energy Regeneration

EnergyRegenPerTick =

BaseEnergyRegen

× ReactorEfficiency

× (1 − HeatPenalty)

EnergyLevel(t+1) =

min(MaxEnergy,

EnergyLevel(t) + EnergyRegenPerTick − EnergySpentThisTick)

---

## 7.2 Energy Costs (Baseline)

| Action            | Energy Cost |

|------------------|------------|

| Missile Fire      | 20         |

| Railgun Fire      | 12         |

| Hard Evasion      | 18         |

| Engine Burn       | 15         |

| Active Scan       | 15         |

| Jam               | 14         |

| Jump Initiation   | Ship-based |

If EnergyLevel < RequiredCost:

Action fails.

---

# 8. Heat System

## 8.1 Heat Accumulation

HeatIncreasePerAction:

| Action            | Heat Cost |

|------------------|----------|

| Missile Fire      | 12       |

| Railgun Fire      | 6        |

| Hard Evasion      | 10       |

| Engine Burn       | 8        |

| Active Scan       | 8        |

| Jam               | 9        |

| Jump Initiation   | 15       |

HeatLevel(t+1) =

max(0,

HeatLevel(t)

\+ HeatGeneratedThisTick

− HeatDecayRate)

---

## 8.2 Heat Decay

HeatDecayRate = 5 per tick (default)

Decay applies after action resolution.

---

# 9. Heat Penalty Model

HeatPenalty = (HeatLevel / 100)²

HeatPenalty affects:

- TrackingGain
- WeaponAccuracy
- ShieldRecharge
- EnergyRegen
- JumpChargeRate

HeatPenalty ∈ \[0–1]

Non-linear scaling ensures:

- Low heat = minor penalty
- High heat = severe degradation

---

# 10. Overheat Behavior

OverheatThreshold = 90 (default)

If HeatLevel ≥ OverheatThreshold:

- BurnIntensity capped at 0.5
- WeaponAccuracy −20%
- ShieldRecharge −30%
- EnergyRegen −25%
- Jump disabled

Overheat ends when HeatLevel < 70.

---

# 11. Heat \& Energy Interaction

High heat reduces:

EnergyRegenPerTick.

Aggressive play increases:

- Immediate output
- Long-term vulnerability

Energy gating prevents burst spamming.

Heat gating prevents sustained dominance.

---

# 12. Integration with Combat

Heat affects:

- TrackingConfidence growth
- Firing Solution Quality
- Hit probability indirectly
- Shield sustainability

Energy affects:

- Action availability
- Disengagement viability
- Jump escape timing

---

# 13. Integration with Propulsion

Engine Burn increases:

- Heat accumulation
- Energy consumption
- Signature

High-thrust engines multiply heat generation.

---

# 14. Tunable Parameters

- BaseEnergyRegen
- HeatDecayRate
- HeatPenalty exponent (default 2)
- OverheatThreshold
- Action energy costs
- Action heat costs

---

# 15. Failure \& Edge Cases

If HeatLevel > 100:

Clamp at 100.

If EnergyLevel < 0:

Clamp at 0.

If ship overheats mid-jump:

Jump fails.

Simultaneous energy and heat caps resolved deterministically.

---

# 16. Performance Constraints

- O(1) update per ship per tick
- No cross-ship heat dependencies
- No global recalculation loops

---

# 17. Security Considerations

- All energy and heat mutation server-side only
- Action validation before execution
- No client-side prediction authority

---

# 18. Telemetry \& Logging

Log:

- Average heat per engagement
- Overheat frequency
- Energy starvation events
- Action failure due to low energy

Used for balancing.

---

# 19. Balancing Guidelines

- Heat should meaningfully punish reckless aggression.
- Energy starvation should occur in sustained fights.
- Overheat must feel dangerous.
- Conservative players should outperform reckless ones long-term.
- Heat decay must allow recovery if player disengages.

---

# 20. Non-Goals (v1)

- Reactor meltdown explosions
- Subsystem damage
- Heat damage to hull
- Thermal stealth mechanics

---

# 21. Future Extensions

- Reactor class differentiation
- Heat vent modules
- Thermal signature spoofing
- Heat-based critical failure events

---

# End of Document
