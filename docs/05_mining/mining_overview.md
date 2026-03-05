# Mining Subsystem Overview
## Version: 0.1
## Status: Draft
## Owner: Core Simulation
## Last Updated: 2026-03-02

---

# 1. Purpose

This document provides a high-level architectural overview of the Mining subsystem.

It explains how the following components interact:

- Mining System
- Hazard System
- Drone System
- Resource Generation System
- Economy Engine
- Combat Engine
- Tick Engine

This document contains no core math. Detailed mechanics are defined in subsystem-specific documents.

---

# 2. Subsystem Boundaries

The Mining subsystem is composed of four modular specifications:

- mining_system.md
- hazard_system.md
- resource_generation.md

Drone mechanics (prospecting drones) are defined in `docs/12_communications/communications_system.md` Section 14.

Each subsystem has a distinct responsibility:

| Subsystem | Responsibility |
|-----------|----------------|
| Mining System | Player extraction loop and depletion |
| Hazard System | Hidden environmental risks |
| Drone System | Deployable support mechanics |
| Resource Generation | Economic output modeling |

This separation prevents cross-domain coupling.

---

# 3. High-Level Mining Flow

The mining loop operates as follows:

1. Player enters asteroid field.
2. Player performs passive or active prospecting.
3. Player may deploy drone(s) for enhanced data.
4. Player initiates extraction cycle.
5. Tick engine resolves:
   - Yield calculation
   - Instability roll
   - Hazard trigger check
   - Heat and energy updates
6. Density reduced.
7. Resource injected into economic model.
8. PvP exposure modifiers applied.

---

# 4. Tick Engine Integration

Mining is fully tick-driven.

Each mining cycle occurs inside the authoritative tick loop:

Tick Order (relevant portions):

1. Process player commands
2. Resolve active mining cycles
3. Invoke hazard checks
4. Resolve drone activity
5. Update depletion
6. Inject resources into economy
7. Apply PvP exposure adjustments

Mining never executes outside the tick engine.

---

# 5. Security Tier Interaction

System SecurityRating influences:

- Yield multiplier
- Hazard probability
- Rare mineral chance
- Pirate spawn chance
- Detection exposure

High Security:
- Stable mining
- Low yield
- Low hazard
- Low PvP risk

Medium Security:
- Moderate yield
- Moderate risk

Low Security:
- High yield
- High hazard
- Increased pirate spawns
- Significant PvP exposure

This creates natural geographic risk zones.

---

# 6. Risk Layers

Mining risk is multi-layered:

1. Instability Risk  
   - Random yield loss
   - Heat spikes
   - Equipment strain

2. Environmental Hazard Risk  
   - Minefields
   - Radiation
   - Pirate ambush triggers

3. PvP Exposure Risk  
   - Increased signature while mining
   - Reduced mobility
   - Attracts pirates

4. Economic Risk  
   - Market price fluctuations
   - Over-mining supply shocks

Mining is never purely safe.

---

# 7. Drone Interaction Model

Drones enhance mining by:

- Improving hazard detection
- Increasing yield estimate accuracy
- Reducing uncertainty

However:

- Drones are consumable
- Drones can be destroyed
- Drone use increases exposure slightly
- Drones do not eliminate core risk

Drones reduce uncertainty but do not remove volatility.

---

# 8. Economic Injection Model

Mining outputs raw resources.

Each extraction:

- Generates commodity units
- Injects supply into system market
- Affects supply-demand ratio
- Influences AI trader routing

Rare mineral discovery:

- May trigger price shifts
- May increase pirate activity
- May generate mission opportunities

Mining directly impacts macro-economy.

---

# 9. PvP Integration

Mining increases vulnerability through:

- SignatureRadius × 1.25
- Velocity reduction (−30%)
- DetectionScore bonus to observers
- Increased pirate spawn chance

Mining hotspots naturally become PvP conflict zones.

Low-security mining is intentionally dangerous.

---

# 10. Performance Considerations

Mining must scale to:

- Dozens of simultaneous miners
- Hundreds of fields across systems

Constraints:

- Yield calculation <1ms per instance
- Hazard checks lightweight
- No full-system scans per tick
- Depletion updates constant-time

Mining cannot degrade tick performance.

---

# 11. Balancing Philosophy

Mining should:

- Reward calculated risk
- Encourage movement between systems
- Prevent permanent resource monopolies
- Generate economic dynamism
- Naturally create PvP tension

High security must feel stable.
Low security must feel lucrative but dangerous.

---

# 12. Non-Goals (v1)

- Fleet-based cooperative mining
- Automated mining bots
- Player-owned mining infrastructure
- Territory claiming

Mining is individual and risk-based in v1.

---

# 13. Future Extensions

- Mining specialization modules
- Equipment degradation
- Cooperative extraction bonuses
- Volatile mineral chain reactions
- Player-deployed mining arrays

Future features must preserve:
Risk, volatility, and economic impact.

---

# End of Document
