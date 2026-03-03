# Security Zones Specification

## Version: 0.1
## Status: Draft
## Owner: Core Simulation

## Last Updated: 2026-03-02

---

# 1. Purpose

Defines the Security Zone system, which governs:

- Risk gradients
- PvP legality
- Pirate activity
- Mining yield scaling
- Hazard probability
- Economic stability

Security Zones are the primary macro-level risk control mechanism in the galaxy.

---

# 2. Scope

## IN SCOPE

- Security rating model
- Zone classification
- Gameplay modifiers by zone
- PvP behavior rules
- NPC enforcement (v1 minimal)
- Economic impact scaling

## OUT OF SCOPE

- Faction politics
- Territory control
- Dynamic sovereignty wars
- Complex law enforcement AI

---

# 3. Design Principles

- Higher reward requires higher risk.
- Security must be geographically coherent.
- Zones must materially change behavior.
- Safety must not be absolute.
- Frontier must feel dangerous.
- Zone rules must be deterministic.

---

# 4. Security Rating Model

Each star system has:

SecurityRating ∈ \[0.0 – 1.0]

SecurityRating influences:

- Mining yield multiplier
- Hazard probability
- Pirate spawn rate
- Rare anomaly probability
- PvP exposure
- Economic volatility

SecurityRating is assigned during procedural generation and is immutable in v1.

---

# 5. Zone Classification

Security zones are grouped as:

High Security  

Medium Security  

Low Security  

Thresholds (configurable):

High Security:

0.7 – 1.0

Medium Security:

0.4 – 0.7

Low Security:

0.0 – 0.4

Zone classification derived from SecurityRating.

---

# 6. High Security Zone Rules

Characteristics:

- Stable economy
- Lower mining yield multiplier
- Reduced hazard probability
- Lower pirate spawn rate
- PvP discouraged but not impossible

Gameplay effects:

- Lower RareMineralChance
- Reduced PvP incentives
- Slower economic volatility
- Safer trade routes

High Security must feel stable, not profitable.

---

# 7. Medium Security Zone Rules

Characteristics:

- Balanced risk and reward
- Moderate mining yield
- Moderate pirate presence
- Occasional anomalies
- PvP viable

Gameplay effects:

- Standard yield multiplier
- Moderate anomaly frequency
- Balanced trade profitability

Medium Security is transitional space.

---

# 8. Low Security Zone Rules

Characteristics:

- High mining yield
- High anomaly frequency
- High pirate spawn rate
- Elevated hazard probability
- PvP common

Gameplay effects:

- Increased RareMineralChance
- Increased economic volatility
- Elevated detection exposure
- Increased resource density variance

Low Security must feel lucrative but unstable.

---

# 9. PvP Behavior Model by Zone

High Security:

- PvP allowed but reputational consequences (future)
- No instant law enforcement in v1
- Player choice drives behavior

Medium Security:

- PvP neutral
- No structural penalties in v1

Low Security:

- PvP expected
- No penalties
- Increased pirate overlap

Zone must shape behavior indirectly, not hard-disable PvP.

---

# 10. Mining \& Resource Scaling

Security influences:

SecurityYieldModifier:

High:

0.7

Medium:

1.0

Low:

1.3 (configurable)

RareMineralChance scaled by:

(1 − SecurityRating)

Lower security increases rare discovery probability.

---

# 11. Hazard Scaling

HazardProbability:

(1 − SecurityRating) × HazardBaseMultiplier

High Security:

Rare hazards

Low Security:

Frequent hazards

Hazard intensity may scale with zone.

---

# 12. Pirate Activity Scaling

Pirate spawn probability scaled by:

PirateActivityBase × (1 − SecurityRating)

Low Security:

- Frequent pirate encounters
- Elevated ambush chance

High Security:

- Rare pirate presence

---

# 13. Economic Volatility Model

Security influences:

- Price fluctuation amplitude
- Trade spread width
- Supply shock impact

Low Security:

- Higher volatility
- Larger price swings

High Security:

- Stabilized market behavior

---

# 14. Detection \& Exposure Impact

Security zones may influence:

- Passive detection sensitivity
- Mining exposure multiplier
- Scan clarity degradation (optional)

Low Security systems may:

- Increase detection range due to reduced monitoring
- Increase PvP encounter frequency

---

# 15. Persistence Interaction

SecurityRating is:

- Deterministic
- Seed-derived
- Immutable in v1

Dynamic modification not supported in v1.

---

# 16. Determinism Requirements

Given:

- UniverseSeed
- SystemID

SecurityRating must be reproducible.

No runtime randomness allowed in assignment.

---

# 17. Performance Constraints

Security zone effects must:

- Be constant-time lookups
- Avoid global recalculation
- Be cached per system

Zone logic applied during subsystem resolution.

---

# 18. Testing Requirements

Tests must verify:

- Proper classification by thresholds
- Yield scaling correctness
- Hazard scaling correctness
- Pirate scaling correctness
- Rare mineral scaling correctness
- Deterministic assignment

Replay tests must confirm identical behavior across runs.

---

# 19. Non-Goals (v1)

- Dynamic security shifts
- Law enforcement AI fleets
- Reputation-based lockouts
- Faction-controlled space
- Territory taxation

---

# 20. Future Extensions

- Dynamic security drift
- Regional instability waves
- Faction influence systems
- Enforcement NPC patrols
- Temporary lockdown events

All expansions must preserve deterministic core.

---

# End of Document
