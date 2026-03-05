# Security Zones Specification

## Version: 0.2
## Status: Draft
## Owner: Core Simulation

## Last Updated: 2026-03-05

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

OR

SecurityRating = NULL (Black Sector — unclassified, unmonitored space)

OR

SecurityRating = 2.0 (Federated Space — special fixed zone, above the normal scale)

SecurityRating influences:

- Mining yield multiplier
- Hazard probability
- Pirate spawn rate
- Rare anomaly probability
- PvP exposure
- Economic volatility

SecurityRating is assigned during procedural generation and is immutable in v1.

Black Sector systems are rare and treated as a distinct classification outside the normal rating scale.

---

# 5. Zone Classification

Security zones are grouped as:

Federated Space *(new player starting area, center of galaxy)*

High Security

Medium Security

Low Security

Black Sector

Thresholds (configurable):

Federated Space:

SecurityRating = 2.0 (special fixed value — not procedurally generated)

High Security:

0.7 – 1.0

Medium Security:

0.4 – 0.7

Low Security:

0.0 – 0.4

Black Sector:

SecurityRating = NULL (not on the 0.0–1.0 scale)

Zone classification derived from SecurityRating.

Black Sector systems are procedurally seeded but rare. They represent the game's namesake extreme — space that exists outside any known authority or mapping.

Federated Space systems are hand-placed at the center of the galaxy and defined in world config, not generated procedurally. They are the safe starting zone for new players.

---

# 6. Federated Space Zone Rules

Federated Space is a fixed cluster of government-controlled systems at the center of the galaxy. It is the starting zone for all new players and the safest area in the game.

Characteristics:

- No pirate spawns — NPC patrol coverage enforced
- PvP fully disabled — combat between players is blocked server-side
- Full IRN relay coverage (100% reliability, ×1.0 delay)
- Economy is highly stable, low volatility
- Mining yields are very low (Federated Space is not a resource zone)
- Contains the origin starbase (new player spawn, standard-mode respawn)
- Ships can always jump back to Federated Space from anywhere in the galaxy

Gameplay effects:

- SecurityYieldModifier: 0.3 (very low — farming Federated Space is intentionally unproductive)
- PirateActivityBase: 0 (no pirates)
- PvP: disabled
- IRN Reliability: 100%
- Drone availability: all types stocked at origin starbase

Federated Space must feel safe, stable, and welcoming — but not a place to get rich.

See `docs/01_architecture/ship_system.md` for starting state and respawn details.

---

# 7. High Security Zone Rules

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

# 8. Medium Security Zone Rules

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

# 9. Low Security Zone Rules

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

# 10. Black Sector Zone Rules

Characteristics:

- No security rating
- Contraband markets active
- Extreme mining yield variance
- Maximum hazard probability
- Maximum pirate spawn rate
- Rare and exotic commodity presence
- No law enforcement of any kind
- PvP unrestricted and expected

Gameplay effects:

- Highest RareMineralChance in the game
- Extreme economic volatility
- Contraband goods tradeable without restriction
- Survival is not guaranteed

Black Sector must feel lawless, ancient, and dangerous.

These are the systems the game is named after.

---

# 11. PvP Behavior Model by Zone

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

Black Sector:

- PvP constant
- No rules
- Entering is a deliberate risk decision

Zone must shape behavior indirectly, not hard-disable PvP.

---

# 12. Mining \& Resource Scaling

Security influences:

SecurityYieldModifier:

High:

0.7

Medium:

1.0

Low:

1.3 (configurable)

Black Sector:

2.0+ (extreme variance, configurable)

RareMineralChance scaled by:

(1 − SecurityRating)

Lower security increases rare discovery probability.

Black Sector uses maximum RareMineralChance with high variance — yields may be extraordinary or nearly nothing.

---

# 13. Hazard Scaling

HazardProbability:

(1 − SecurityRating) × HazardBaseMultiplier

High Security:

Rare hazards

Low Security:

Frequent hazards

Black Sector:

Constant hazards — assumed hostile environment

Hazard intensity may scale with zone.

---

# 14. Pirate Activity Scaling

Pirate spawn probability scaled by:

PirateActivityBase × (1 − SecurityRating)

Low Security:

- Frequent pirate encounters
- Elevated ambush chance

High Security:

- Rare pirate presence

Black Sector:

- Maximum pirate density
- Ambush assumed, not exceptional

---

# 15. Economic Volatility Model

Security influences:

- Price fluctuation amplitude
- Trade spread width
- Supply shock impact

Low Security:

- Higher volatility
- Larger price swings

High Security:

- Stabilized market behavior

Black Sector:

- Extreme volatility
- Contraband goods tradeable
- No price floor enforcement

---

# 16. Detection \& Exposure Impact

Security zones may influence:

- Passive detection sensitivity
- Mining exposure multiplier
- Scan clarity degradation (optional)

Low Security systems may:

- Increase detection range due to reduced monitoring
- Increase PvP encounter frequency

---

# 17. Persistence Interaction

SecurityRating is:

- Deterministic
- Seed-derived
- Immutable in v1

Dynamic modification not supported in v1.

---

# 18. Determinism Requirements

Given:

- UniverseSeed
- SystemID

SecurityRating must be reproducible.

No runtime randomness allowed in assignment.

---

# 19. Performance Constraints

Security zone effects must:

- Be constant-time lookups
- Avoid global recalculation
- Be cached per system

Zone logic applied during subsystem resolution.

---

# 20. Testing Requirements

Tests must verify:

- Proper classification by thresholds
- Yield scaling correctness
- Hazard scaling correctness
- Pirate scaling correctness
- Rare mineral scaling correctness
- Deterministic assignment
- Black Sector correctly identified by NULL SecurityRating
- Black Sector modifiers applied independently of rating scale

Replay tests must confirm identical behavior across runs.

---

# 21. Non-Goals (v1)

- Dynamic security shifts
- Law enforcement AI fleets
- Reputation-based lockouts
- Faction-controlled space
- Territory taxation

---

# 22. Future Extensions

- Dynamic security drift
- Regional instability waves
- Faction influence systems
- Enforcement NPC patrols
- Temporary lockdown events

All expansions must preserve deterministic core.

---

# End of Document
