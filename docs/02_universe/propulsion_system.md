# Propulsion & Engine System Specification
## Version: 0.1  
## Status: Draft  
## Owner: Core Simulation  
## Last Updated: 2026-03-03  

---

# 1. Purpose

Defines the propulsion and engine mechanics governing:

- Ship movement between systems
- Energy consumption
- Fuel usage
- Heat generation
- Detection signature
- Strategic tradeoffs

Propulsion is not cosmetic movement.  
It is a strategic subsystem tightly coupled to combat, scanning, and survival.

---

# 2. Scope

## IN SCOPE

- Engine classes
- Burn intensity model
- Fuel consumption
- Heat generation
- Detection signature impact
- Jump mechanics
- Energy interaction
- Upgrade modifiers

## OUT OF SCOPE

- Real-time orbital physics
- Newtonian thrust modeling
- Atmospheric flight
- Micro-positioning inside system (v1 abstracted)

---

# 3. Design Principles

- Engines are tradeoff systems, not pure upgrades.
- Speed increases signature and heat.
- Escape requires planning.
- Fuel is finite.
- Jumping is strategic, not spam-based.
- Combat occurs at distance; propulsion affects engagement envelope.

---

# 4. Engine Classes (v1)

Each ship mounts one engine class.

## 4.1 Civilian Engine
- Moderate speed
- Low fuel efficiency
- Moderate heat
- Moderate signature

## 4.2 Industrial Engine
- Low speed
- High fuel efficiency
- Low heat
- Low signature

## 4.3 Military Engine
- High speed
- High fuel consumption
- High heat
- High signature

## 4.4 Stealth Engine
- Low-to-moderate speed
- Moderate fuel usage
- Low signature
- High heat under sustained burn

Engine class determines base modifiers.

---

# 5. Core Engine Attributes

Each engine has:

- MaxBurnLevel (0.0–1.0)
- BaseSpeed
- FuelRatePerTick
- HeatPerBurnUnit
- SignaturePerBurnUnit
- JumpChargeRate
- EnergyDrawPerTick

---

# 6. Burn Intensity Model

BurnLevel ∈ [0.0 – 1.0]

EffectiveSpeed =
BaseSpeed × BurnLevel

FuelConsumptionPerTick =
FuelRatePerTick × BurnLevel

HeatGeneratedPerTick =
HeatPerBurnUnit × BurnLevel²

SignatureGenerated =
SignaturePerBurnUnit × BurnLevel

Quadratic heat scaling discourages sustained max burn.

---

# 7. Detection Signature Impact

TotalSignature =

BaseShipSignature  
+ EngineSignature  
+ WeaponSignature  
+ MiningSignature  

High burn increases detection probability.

Stealth engines reduce SignaturePerBurnUnit.

---

# 8. Fuel System

Each ship has:

- FuelCapacity
- CurrentFuel

Fuel consumed each tick during burn.

If fuel reaches zero:

- BurnLevel forced to 0
- Jump unavailable
- Ship vulnerable

Refueling only possible at stations (v1).

---

# 9. Heat Interaction

Engine heat contributes to total ship heat.

If TotalHeat exceeds threshold:

- Tracking becomes easier
- Subsystems risk temporary shutdown (future extension)
- Jump charge efficiency reduced

Heat dissipates each tick based on cooling rating.

---

# 10. Jump Mechanics

Jump allows travel between connected systems.

Requirements:

- Valid jump route
- Minimum energy reserve
- Minimum fuel reserve
- JumpChargeMeter ≥ 100%

JumpChargeMeter increases per tick based on:

JumpChargeRate  
− HeatPenalty  
− DamagePenalty (future)

Jump may be interrupted if:

- EMP applied
- Energy drops below threshold
- Ship destroyed

---

# 11. Pre-Attack Jump Window

A ship may jump before combat escalation if:

- Not hard-locked
- Sufficient energy available
- Jump route valid

This reinforces long-range engagement philosophy.

---

# 12. Engine Upgrades

Upgrades may modify:

- BaseSpeed
- Fuel efficiency
- Heat scaling coefficient
- Signature scaling
- JumpChargeRate

Upgrades must not eliminate tradeoffs.

No engine may be strictly superior in all dimensions.

---

# 13. Security Zone Interaction

Low Security:

- Higher chance of interception during jump charge
- Higher pirate ambush probability during burn

High Security:

- Reduced hostile interception probability

Zone modifiers affect risk, not engine mechanics directly.

---

# 14. Determinism Requirements

Engine behavior must be:

- Tick-based
- Deterministic
- Independent of system clock
- Derived solely from ship state + tick number

No real-time physics allowed.

---

# 15. Performance Constraints

Engine update must be:

- O(1) per ship
- No global recalculation
- No pathfinding during tick
- No blocking operations

---

# 16. Logging Requirements

Log at:

DEBUG:
- Burn changes
- Jump charge progression

INFO:
- Jump initiated
- Jump completed

WARN:
- Fuel depletion
- Heat overload threshold reached

---

# 17. Testing Requirements

Tests must validate:

- Fuel consumption correctness
- Heat scaling curve
- Signature scaling
- Jump eligibility conditions
- Deterministic burn calculations
- No negative resource values

Replay test must confirm identical propulsion behavior across runs.

---

# 18. Non-Goals (v1)

- True acceleration curves
- Inter-system warp interdiction
- Dynamic gravitational modeling
- Mid-jump combat

---

# 19. Future Extensions

- Interdiction modules
- Afterburner spikes
- Engine damage states
- Emergency jump overload
- Fuel type specialization

All future additions must preserve tradeoff-driven design.

---

# End of Document