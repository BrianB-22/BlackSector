# Supply Demand Model Specification

## Version: 0.1

## Status: Draft

## Owner: Economy Systems

## Last Updated: 2026-03-03

---

# 1. Purpose

The Supply Demand Model provides the quantitative framework used to measure market conditions within the galactic economy.

This subsystem evaluates:

* commodity supply levels
* commodity demand pressure
* scarcity and surplus states
* market imbalance signals

These signals are used by other economic systems including:

* Price Engine
* AI Trader Model
* Economic Events
* Mission Generation

The supply–demand model does **not calculate prices**.
Instead, it exposes normalized economic indicators used by pricing and AI systems.

---

# 2. Scope

IN SCOPE:

* commodity supply metrics
* demand pressure calculations
* scarcity classification
* surplus classification
* supply/demand ratios
* market signal generation

OUT OF SCOPE:

* price calculation (Price Engine)
* commodity definitions (Commodity Model)
* port trading mechanics (Port System)
* economic event generation (Economic Events)
* AI trader decision logic (AI Trader Model)

---

# 3. Design Principles

Non-negotiable constraints:

* deterministic calculations
* server authoritative state
* low computational cost
* normalized market signals
* scalable across thousands of ports

The supply-demand system must provide stable signals that prevent extreme volatility.

---

# 4. Core Concepts

Supply
The quantity of a commodity available within a port or region.

Demand
The rate at which a commodity is consumed or purchased.

Supply Ratio
A normalized value representing how full the inventory is.

Demand Pressure
A measure of consumption pressure relative to production.

Scarcity
A condition where supply falls below acceptable levels.

Surplus
A condition where supply significantly exceeds expected demand.

Market Signal
Normalized indicators used by pricing and AI decision systems.

---

# 5. Data Model

## Entity: CommoditySupplyState (Persistent)

Represents current supply information for a commodity within a port.

Fields:

* port_id: uint64
* commodity_id: string
* inventory_current: int
* inventory_capacity: int
* production_rate: int
* consumption_rate: int

---

## Entity: MarketSignal (Transient)

Represents computed supply-demand indicators.

Fields:

* commodity_id: string
* supply_ratio: float
* demand_pressure: float
* scarcity_score: float
* surplus_score: float

MarketSignal objects are generated during economic evaluation cycles.

---

# 6. State Machine

Market state is derived from calculated metrics rather than stored state.

NORMAL
→ Balanced supply and demand.

SCARCE
→ Supply falling below scarcity threshold.

SURPLUS
→ Supply exceeding surplus threshold.

CRITICAL_SHORTAGE
→ Supply approaching depletion.

These classifications help other systems adjust behavior.

---

# 7. Core Mechanics

Supply-demand metrics are evaluated during the economic update cycle.

Evaluation sequence:

1. retrieve commodity inventory
2. calculate supply ratio
3. calculate demand pressure
4. evaluate scarcity conditions
5. generate market signals

Example calculation:

```text
SupplyRatio = InventoryCurrent / InventoryCapacity
```

Demand pressure calculation:

```text
DemandPressure = ConsumptionRate / max(ProductionRate, 1)
```

Scarcity detection:

```text
if SupplyRatio < ScarcityThreshold
    mark scarce
```

Surplus detection:

```text
if SupplyRatio > SurplusThreshold
    mark surplus
```

---

# 8. Mathematical Model

## Variables

InventoryCurrent
Current quantity stored at port.

InventoryCapacity
Maximum storage capacity.

ProductionRate
Units generated per economic tick.

ConsumptionRate
Units consumed per economic tick.

SupplyRatio
Normalized supply value.

DemandPressure
Normalized consumption pressure.

---

## Supply Ratio

```
SupplyRatio = InventoryCurrent / InventoryCapacity
```

Range:

```
0.0 → empty
1.0 → full
```

---

## Demand Pressure

```
DemandPressure = ConsumptionRate / max(ProductionRate, 1)
```

Interpretation:

| Value | Meaning                   |
| ----- | ------------------------- |
| <1    | production exceeds demand |
| 1     | balanced                  |

> 1 | demand exceeds production |

---

## Scarcity Score

```
ScarcityScore = max(0, ScarcityThreshold - SupplyRatio)
```

---

## Surplus Score

```
SurplusScore = max(0, SupplyRatio - SurplusThreshold)
```

---

# 9. Tunable Parameters

ScarcityThreshold = 0.20

CriticalShortageThreshold = 0.05

SurplusThreshold = 0.80

MaximumDemandPressure = 5.0

SignalUpdateInterval = 60 seconds

---

# 10. Integration Points

Depends On:

* Commodity Model
* Port System
* Economic Tick

Exposes:

* supply_ratio
* demand_pressure
* scarcity_score
* surplus_score

Used By:

* Price Engine
* AI Trader Model
* Economic Events
* Mission Generation

---

# 11. Failure & Edge Cases

Zero Capacity
Supply ratio defaults to zero.

Zero Production
Demand pressure treated as high demand.

Negative Inventory
Inventory automatically clamped to zero.

Extreme Demand
Demand pressure capped at configured maximum.

---

# 12. Performance Constraints

System must support:

* up to 1,000 ports
* 10–50 commodities
* near-real-time updates

Evaluation cost must remain below 10 ms per cycle.

---

# 13. Security Considerations

All supply-demand calculations occur server-side.

Clients cannot:

* modify inventory values
* alter supply signals
* influence demand pressure directly

Market signals must only be generated during server economic updates.

---

# 14. Telemetry & Logging

Track metrics including:

* commodity scarcity frequency
* regional supply imbalance
* average demand pressure
* port inventory distribution

Log events:

* critical shortages
* supply collapse
* abnormal demand spikes

---

# 15. Balancing Guidelines

Supply-demand signals must encourage trade movement.

Design goals:

* shortages should appear regularly
* surplus regions should emerge naturally
* markets should rebalance over time

Scarcity events should encourage exploration of new regions.

---

# 16. Non-Goals (v1)

Not included in the initial version:

* inter-port supply chain modeling
* predictive demand modeling
* advanced economic forecasting

---

# 17. Future Extensions

Potential future enhancements:

* regional supply aggregation
* multi-port demand analysis
* economic forecasting models
* adaptive commodity production

---

# End of Document
