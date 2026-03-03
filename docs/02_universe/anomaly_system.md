# Anomaly System Specification

## Version: 0.1
## Status: Draft
## Owner: Core Simulation

## Last Updated: 2026-03-02

---

# 1. Purpose

Defines the Anomaly System, which governs the generation, discovery, lifecycle, and impact of special space phenomena.

Anomalies are designed to:

- Drive exploration
- Create economic spikes
- Generate PvP hotspots
- Introduce temporary strategic shifts
- Provide rare discovery events

Anomalies are deterministic in origin but dynamic in interaction.

---

# 2. Scope

## IN SCOPE

- Anomaly generation rules
- Anomaly classification
- Discovery mechanics
- Lifecycle states
- Economic impact
- Risk scaling
- Server-wide event triggers

## OUT OF SCOPE

- Narrative story arcs
- Faction-driven dynamic wars
- Player-created anomalies
- Visual animation behavior

---

# 3. Design Principles

- Anomalies must be rare.
- Discovery must require effort.
- Anomalies must alter player behavior.
- Anomalies must introduce risk.
- Anomalies must not permanently destabilize economy.
- All anomaly behavior must be deterministic within tick model.

---

# 4. Anomaly Categories (v1)

## 4.1 Resource Anomalies

Examples:

- anomaly-rare-mineral-01
- anomaly-high-density-field-01

Effects:

- Increased yield multiplier
- Increased pirate activity
- Temporary economic impact

---

## 4.2 Environmental Anomalies

Examples:

- anomaly-gravitational-01
- anomaly-radiation-01
- anomaly-sensor-distortion-01

Effects:

- Reduced tracking efficiency
- Increased heat buildup
- Modified jump reliability

---

## 4.3 Pirate Activity Anomalies

Examples:

- anomaly-pirate-staging-01

Effects:

- Elevated NPC pirate spawns
- Increased PvP likelihood
- Higher loot drops

---

## 4.4 Exploration Data Anomalies

Examples:

- anomaly-ancient-structure-01
- anomaly-derelict-01

Effects:

- High-value tradeable data
- Temporary map reveal
- Rare economic information boost

---

# 5. Generation Model

Anomalies are generated per system using:

SystemSeed

\+

AnomalyFrequencyModifier

Anomaly spawn probability influenced by:

- SecurityRating (lower security = higher probability)
- Star type
- Region bias
- Distance from core

Anomalies must be:

- Deterministic from seed
- Activated when discovered
- Persisted when mutated

---

# 6. Discovery Mechanics

An anomaly may exist in one of two states:

HIDDEN  

DISCOVERED  

Discovery requires:

- Active scanning
- Drone-assisted scanning
- Exploration module usage
- Proximity threshold

Discovery probability increases with:

- Sensor strength
- Scan duration
- Exploration ship bonuses

Discovery emits:

anomaly\_discovered event

---

# 7. Anomaly Lifecycle

States:

HIDDEN  

→ DISCOVERED  

→ ACTIVE  

→ DEPLETING  

→ DORMANT  

→ (Optional Regeneration)

Lifecycle rules:

- Active anomalies may modify local system behavior.
- Resource anomalies deplete when mined.
- Environmental anomalies may decay over time.
- Pirate anomalies may dissipate after combat density drops.

Lifecycle driven entirely by tick engine.

---

# 8. Economic Impact Model

Resource anomalies may:

- Inject rare supply
- Trigger price volatility
- Alter trade routes
- Increase trader activity

Impact must be:

- Significant but temporary
- Regionally bounded
- Logged for telemetry

---

# 9. PvP Impact Model

Certain anomalies increase:

- Player traffic
- Risk density
- Detection exposure
- Pirate spawn multiplier

Anomalies must naturally generate conflict zones.

---

# 10. Server-Wide Event Trigger

If anomaly severity exceeds threshold:

Emit:

rare\_anomaly\_event

This may:

- Broadcast to region
- Increase travel interest
- Temporarily alter pirate aggression
- Shift market expectations

Server-wide events must be rare.

---

# 11. Persistence Model

Snapshot must store:

- Discovered anomalies
- Depletion level
- Active modifiers
- Remaining duration

Base anomaly presence may be regenerated from seed.

Mutated state must be persisted.

---

# 12. Performance Constraints

Anomaly system must:

- Avoid global scanning per tick
- Use O(1) per-system updates
- Only process active anomalies
- Avoid expensive recalculation

Inactive systems may remain dormant.

---

# 13. Determinism Requirements

Given:

- UniverseSeed
- SystemID
- TickNumber

Anomaly generation must be reproducible.

Discovery randomness must use deterministic PRNG.

---

# 14. Logging Requirements

Must log:

- Discovery events (INFO)
- Rare anomaly triggers (INFO)
- Anomaly depletion (DEBUG or INFO)
- Unexpected anomaly state transitions (WARN)

Logs must include:

- system\_id
- anomaly\_id
- tick

---

# 15. Testing Requirements

Tests must validate:

- Deterministic generation from seed
- Discovery probability behavior
- Lifecycle transitions
- Economic impact boundedness
- No infinite resource anomalies
- No permanent instability

Replay test must reproduce anomaly state exactly.

---

# 16. Non-Goals (v1)

- Dynamic anomaly migration
- Player-created anomalies
- Permanent galaxy-altering events
- Narrative-driven anomaly chains

---

# 17. Future Extensions

- Multi-system anomaly chains
- Time-limited galactic crises
- Procedural megastructures
- AI-driven anomaly escalation

All future evolution must preserve deterministic simulation core.

---

# End of Document
