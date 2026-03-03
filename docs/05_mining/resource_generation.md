# Resource Generation Specification
## Version: 0.1
## Status: Draft
## Owner: Economy Engine
## Last Updated: 2026-03-02

---

# 1. Purpose

Defines how asteroid resource extraction integrates into the global economy.

This subsystem governs:

- Rare mineral probability
- Resource yield classification
- Supply injection into system markets
- Economic ripple effects from extraction

---

# 2. Scope

IN SCOPE:
- Rare mineral probability modeling
- Resource classification
- Supply injection into economy
- Security-tier yield impact

OUT OF SCOPE:
- Price calculation logic (see price_engine.md)
- AI trader routing logic
- Market UI behavior

---

# 3. Design Principles

- Low-security space must produce higher rare yields.
- Rare resources must influence market dynamics.
- Resource output must scale with risk.
- Supply injection must be deterministic and server-authoritative.
- No client-side economic mutation.

---

# 4. Core Concepts

- RareMineralChance
- SupplyInjection
- SecurityYieldModifier
- ResourceClass
- SystemSupplyIndex

---

# 5. Data Model

## Entity: ResourceYield

- field_id: UUID
- system_id: UUID
- resource_type: enum
- quantity: float
- rarity_class: enum
- extraction_tick: int

Transient event object injected into economy engine.

---

# 6. Resource Classification

ResourceClass:

- Common
- Industrial
- Rare
- Volatile (future expansion)

Classification influenced by:

- Asteroid density
- Security rating
- Instability factor

---

# 7. Mathematical Model

## 7.1 Rare Mineral Probability

RareMineralChance =

Density  
× (1 − SecurityRating)  
× RareSpawnMultiplier  

Default RareSpawnMultiplier = 0.4

---

## 7.2 Security Yield Modifier

SecurityYieldModifier:

- High Security: 0.7
- Medium Security: 1.0
- Low Security: 1.3

Applied during mining yield calculation.

---

## 7.3 Supply Injection

On successful extraction:

SystemSupplyIndex(resource_type) += quantity

Supply update triggers price recalculation in Economy Engine.

---

# 8. Integration Points

Depends On:
- Mining System
- Economy Engine
- AI Trader Model
- Tick Engine

Exposes:
- ResourceInjectionEvent
- RareDiscoveryEvent (if threshold met)

---

# 9. Rare Discovery Threshold

If:

RareMineralChance ≥ 0.6  
AND  
Yield ≥ RareDiscoveryThreshold  

Then:

Emit RareDiscoveryEvent.

This may trigger:

- Temporary market volatility
- Pirate spawn multiplier
- Exploration interest spike

---

# 10. Tunable Parameters

- RareSpawnMultiplier
- SecurityYieldModifier values
- RareDiscoveryThreshold
- Density floor and ceiling
- ResourceClass weighting

---

# 11. Failure & Edge Cases

- Quantity clamped ≥ 0
- No negative supply injection
- RareDiscoveryEvent emitted once per field per threshold window

---

# 12. Performance Constraints

- O(1) supply update per extraction
- No full-market recalculation inside mining loop
- Market recalculation batched in economy tick phase

---

# 13. Security Considerations

- All supply injection server-authoritative
- No client-provided quantity accepted
- Rare discovery validation server-side only

---

# 14. Telemetry & Logging

Log:

- Rare mineral discovery frequency
- Yield distribution by security tier
- Market volatility spikes
- System supply saturation levels

---

# 15. Balancing Guidelines

- Low security must materially outproduce high security.
- Rare discoveries must feel meaningful but not destabilizing.
- Mining saturation must reduce profit margins over time.
- No permanent economic monopolies.

---

# 16. Non-Goals (v1)

- Player-controlled commodity manipulation
- Permanent resource depletion
- Player-owned resource claims
- Cross-server economic linking

---

# 17. Future Extensions

- Volatile mineral chain reactions
- Resource scarcity events
- Dynamic system resource shifts
- Faction-based resource bonuses

---

# End of Document
