# Procedural Generation Specification

## Version: 0.1
## Status: Draft
## Owner: Core Simulation

## Last Updated: 2026-03-02

---

# 1. Purpose

Defines how the universe, star systems, resources, hazards, and anomalies are procedurally generated.

The procedural generation system must:

- Produce a large, explorable universe
- Be deterministic from a seed
- Support persistent simulation
- Allow regeneration from seed + tick
- Scale without manual content authoring

Procedural generation is foundational to exploration, economy, and risk distribution.

---

# 2. Scope

## IN SCOPE

- Universe seed model
- Star system generation
- Security tier distribution
- Resource field generation
- Hazard distribution
- Anomaly generation
- Deterministic regeneration rules

## OUT OF SCOPE

- Narrative content generation
- Hand-authored quest lines
- Visual asset generation
- Faction political systems (v1)

---

# 3. Design Principles

- Fully deterministic from seed
- No runtime nondeterministic randomness
- Generation separated from simulation logic
- Security gradient must create risk zones
- Resource distribution must shape economy
- Exploration must have meaningful unknowns

---

# 4. Universe Seed Model

The entire universe is derived from:

UniverseSeed (64-bit integer)

All generation functions must use:

- UniverseSeed
- SystemID
- LocalGenerationIndex

Randomness must be derived via deterministic PRNG.

No use of system time.

---

# 5. Universe Structure (v1)

Universe contains:

- 500–1000 star systems
- Each system assigned:

     - Unique ID

     - Spatial coordinates (abstract or grid-based)

     - SecurityRating

     - Resource richness modifier

     - Base pirate activity rating

Universe topology may be:

- Grid-based adjacency

OR

- Graph-based connectivity

Connectivity must be deterministic.

---

# 6. Security Tier Distribution

Each system assigned:

SecurityRating ∈ \[0.0 – 1.0]

OR

SecurityRating = NULL (Black Sector)

Zones:

- High Security (0.7 – 1.0)
- Medium Security (0.4 – 0.7)
- Low Security (0.0 – 0.4)
- Black Sector (NULL — unclassified, unmonitored)

Security influences:

- Mining yield multiplier
- Hazard probability
- Pirate spawn rate
- Rare mineral chance
- PvP density

Security distribution must form clusters, not pure random scatter.

Black Sector systems are rare. They must appear at the galaxy's outer edges or in isolated pockets unreachable by normal trade routes.

---

# 7. Star System Generation

For each system:

Generated attributes:

- SecurityRating
- BaseResourceDensity
- HazardDensity
- AnomalyFrequency
- MarketDemandProfile
- PirateActivityBase

Generation formula example:

SystemSeed = Hash(UniverseSeed + SystemID)

All attributes derived from SystemSeed.

---

# 8. Asteroid Field Generation

Each system generates:

- N asteroid fields
- Density ∈ \[0.2 – 1.0]
- InstabilityFactor ∈ \[0.0 – 0.8]
- MineralClass distribution
- HazardPresence probability

Density influenced by:

- BaseResourceDensity
- SecurityRating

Low security → higher density variance.

---

# 9. Hazard Distribution

HazardPresence probability:

(1 − SecurityRating) × HazardDensityMultiplier

Hazard types include:

- Minefields
- Radiation pockets
- Pirate ambush triggers

Hazards may be:

- Persistent per field
- Regenerated after depletion

---

# 10. Anomaly Generation

Anomalies represent exploration discoveries.

Each system has:

AnomalyFrequency

Anomalies include:

- Rare resource clusters
- Sensor distortion zones
- Pirate staging areas
- High-value exploration data nodes

Anomaly spawn must be:

- Deterministic per seed
- Activated when discovered
- Capable of triggering server-wide event

---

# 11. Dynamic Regeneration

Certain entities may regenerate over time:

- Asteroid field density
- Hazard reappearance
- Anomaly respawn (optional)

Regeneration logic must:

- Be tick-driven
- Be deterministic
- Avoid full system recalculation

---

# 12. Exploration Fog Model

Players initially know:

- System connectivity
- Core navigation data

Unknown until scanned:

- Field density
- Hazard details
- Rare anomaly presence
- Pirate activity intensity

Exploration data must be:

- Stored per player
- Tradeable
- Decay over time (optional)

---

# 13. Economic Impact Coupling

Procedural generation directly influences:

- Supply saturation
- Rare mineral availability
- Trade route profitability
- Pirate hotspots
- PvP clustering

Generation must avoid:

- Infinite rare material clusters
- Permanent economic stagnation
- Completely safe high-profit systems

---

# 14. Deterministic Regeneration Rules

Given:

- UniverseSeed
- SystemID
- TickNumber

System state must be reproducible.

If a system is regenerated without mutation:

It must produce identical baseline configuration.

Mutation events (mining depletion, discovery) stored in persistent state.

---

# 15. Performance Constraints

Generation must:

- Be O(1) per system lookup
- Avoid pre-generating entire universe at startup
- Support lazy system generation
- Cache active systems in memory

Inactive systems may be regenerated from seed.

---

# 16. Persistence Interaction

Snapshot must store:

- Mutated system state
- Depletion levels
- Active anomalies
- Market adjustments

Base procedural attributes may be regenerated from seed.

Only deltas must be persisted.

---

# 17. Non-Goals (v1)

- Hand-authored star systems
- Player-owned territory modification
- Dynamic universe expansion
- True orbital physics modeling

---

# 18. Future Extensions

- Sector-based political drift
- Dynamic security changes
- Procedural faction conflicts
- Regional resource migration
- Black hole / rare cosmic events

All extensions must preserve deterministic core.

---

# End of Document
