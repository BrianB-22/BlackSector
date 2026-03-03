# System Generation Specification

## Version: 0.1
## Status: Draft
## Owner: Core Simulation

## Last Updated: 2026-03-02

---

# 1. Purpose

Defines how individual star systems are procedurally generated.

System generation determines:

- Star type
- Planets
- Stations
- Asteroid fields
- Anomalies
- Security rating
- Pirate baseline
- Economic bias

System generation must be deterministic, lightweight, and compatible with persistent mutation.

---

# 2. Scope

## IN SCOPE

- Deterministic system attribute generation
- Planet generation
- Station placement
- Asteroid field generation hooks
- Anomaly seeding
- Security assignment
- Region integration

## OUT OF SCOPE

- Galaxy-wide topology (see galaxy\_structure.md)
- Region clustering (see region\_partitioning.md)
- Runtime anomaly lifecycle
- Economic simulation logic

---

# 3. Design Principles

- Fully deterministic from seed.
- Lightweight generation (O(1) per system).
- Lazy-load capable.
- Mutations stored separately.
- No runtime randomness outside PRNG.
- Visual identifiers stable.

---

# 4. Generation Inputs

Each system generation uses:

- UniverseSeed
- SystemID
- RegionID
- RegionModifiers

Derived:

SystemSeed = Hash(UniverseSeed + SystemID)

All procedural attributes derived from SystemSeed.

---

# 5. Star Generation

StarType determined from SystemSeed.

Possible types (v1):

- star-yellow-dwarf-01
- star-red-dwarf-01
- star-blue-giant-01
- star-white-dwarf-01
- star-neutron-01

Star type influences:

- AnomalyFrequencyModifier
- HazardBias
- ExplorationValueWeight

Star type must be deterministic.

---

# 6. Security Assignment

SecurityRating derived from:

- Region average security
- Distance from galactic core
- SystemSeed noise modifier

Final SecurityRating must fall within region band constraints.

SecurityRating immutable in v1.

---

# 7. Planet Generation

Each system contains:

- 1–8 planets (configurable)

Planet count derived from:

SystemSeed

Planet attributes:

- PlanetID
- PlanetType (visual reference)
- ResourceAffinityModifier
- HabitableFlag (boolean, v1 cosmetic)
- OrbitIndex (abstract)

PlanetType examples:

- planet-desert-01
- planet-ice-01
- planet-gas-giant-01
- planet-rocky-01

Planet visuals derived deterministically from SystemSeed + PlanetIndex.

---

# 8. Station Generation

Each system may contain:

- 0–2 stations (configurable)

Station presence influenced by:

- SecurityRating
- Region type
- Economic bias

Station types:

- station-industrial-01
- station-trade-hub-01
- station-military-01
- station-frontier-01
- station-mining-01

High Security systems more likely to host trade hubs.

Low Security systems more likely to host frontier or mining stations.

---

# 9. Asteroid Field Seeding

Asteroid field generation includes:

- FieldCount (1–5 typical)
- BaseDensity
- InstabilityFactor
- HazardPresenceFlag

FieldCount influenced by:

- Region resource bias
- SecurityRating

Low Security → higher density variance.

Actual depletion state stored in persistent mutation layer.

---

# 10. Anomaly Seeding

Each system seeded with:

- Potential anomaly slots
- Base anomaly probability

Actual anomaly activation determined by:

- Exploration actions
- Tick lifecycle logic

SystemSeed defines baseline anomaly potential.

---

# 11. Pirate Baseline Assignment

Each system assigned:

PirateActivityBase

Derived from:

- Region pirate modifier
- SecurityRating

Low Security → higher baseline.

Pirate spawning handled during runtime.

---

# 12. Economic Bias Assignment

Each system assigned:

- CommodityDemandBias
- CommoditySupplyBias
- VolatilityWeight

Derived from:

- Region economic bias
- Star type
- SecurityRating

Used by Economy Engine during initialization.

---

# 13. Deterministic PRNG Requirements

System generation must:

- Use seeded PRNG
- Never use system time
- Never depend on runtime ordering
- Produce identical output for identical seed

PRNG state not shared across systems.

---

# 14. Lazy Generation Model

System generation must:

- Occur on first access
- Cache in memory if active
- Regenerate from seed if unloaded
- Apply mutation overlay from persistence

Inactive systems need not remain in memory.

---

# 15. Mutation Overlay Model

Base system attributes derived from seed.

Mutations stored separately:

- Depleted asteroid density
- Discovered anomalies
- Market shifts
- Active combat remnants

On load:

1\. Generate base system

2\. Apply mutation overlay

3\. Register in active simulation

---

# 16. Performance Constraints

System generation must:

- Complete in <1ms per system
- Avoid heavy allocation
- Avoid cross-system scanning
- Be constant-time complexity

System lookup must be O(1).

---

# 17. Testing Requirements

Tests must verify:

- Deterministic generation
- Security band correctness
- Station distribution validity
- Planet count bounds
- Star type reproducibility
- No invalid visual ID references

Replay test must confirm identical generation across runs.

---

# 18. Non-Goals (v1)

- Realistic orbital simulation
- Dynamic planet migration
- Player-built stations
- System destruction events
- Dynamic system creation

---

# 19. Future Extensions

- Deep-space systems beyond frontier
- Rare system archetypes
- System instability cycles
- Procedural megastructures
- Region-driven system mutation

All extensions must preserve deterministic core.

---

# End of Document
