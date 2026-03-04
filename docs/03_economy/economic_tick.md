# Economic Tick Specification

## Version: 0.1

## Status: Draft

## Owner: Economy Systems

## Last Updated: 2026-03-03

---

# 1. Purpose

The Economic Tick subsystem governs how the universe economy evolves over time.

This system updates economic state during scheduled server ticks and drives:

* commodity production
* commodity consumption
* port inventory changes
* price adjustments
* economic scarcity events

The economic tick ensures that the universe economy remains dynamic and continuously generates new trading opportunities.

---

# 2. Scope

IN SCOPE:

* periodic economic updates
* commodity production
* commodity consumption
* port inventory balancing
* price recalculation triggers
* scarcity detection
* surplus detection

OUT OF SCOPE:

* player trading interactions (Port System)
* commodity definitions (Commodity Model)
* AI trader route selection (AI Trader System)
* resource extraction mechanics (Mining System)
* manufacturing systems

---

# 3. Design Principles

Non-negotiable constraints:

* server authoritative updates
* deterministic tick execution
* predictable update cadence
* scalable to thousands of ports
* minimal computational overhead

Economic updates must never depend on client input.

---

# 4. Core Concepts

Economic Tick
A scheduled server event that updates economic state.

Production
The addition of commodities to port inventory.

Consumption
The removal of commodities from port inventory.

Inventory Capacity
Maximum quantity a port can store for a commodity.

Scarcity
Condition where inventory falls significantly below normal levels.

Surplus
Condition where inventory exceeds expected levels.

Price Adjustment Trigger
Event where commodity prices are recalculated.

---

# 5. Data Model

## Entity: EconomicTickState (Transient)

Represents a single execution of the economic update cycle.

Fields:

* tick_id: int
* execution_time: timestamp
* ports_processed: int
* commodities_updated: int

---

## Entity: EconomicPortState (Persistent)

Tracks economic information associated with a port.

Fields:

* port_id: uint64
* last_update_tick: int
* total_trade_volume: int
* scarcity_flags: bitmask
* surplus_flags: bitmask

---

# 6. State Machine

Economic processing occurs as a deterministic sequence.

WAITING
→ Next tick scheduled.

PROCESSING
→ Economic tick executing.

COMPLETED
→ Updates finished.

ERROR
→ Tick aborted due to system failure.

Ticks must always return to WAITING after completion.

---

# 7. Core Mechanics

Each economic tick processes all active ports.

Processing sequence:

1. iterate through ports
2. apply commodity production
3. apply commodity consumption
4. clamp inventory to capacity limits
5. detect scarcity conditions
6. detect surplus conditions
7. flag price recalculation triggers

Production example:

```
inventory_current += production_rate
```

Consumption example:

```
inventory_current -= consumption_rate
```

Inventory clamp:

```
inventory_current = clamp(inventory_current, 0, inventory_capacity)
```

Scarcity detection:

```
if inventory_current < (inventory_capacity × 0.15)
    mark scarcity
```

Surplus detection:

```
if inventory_current > (inventory_capacity × 0.85)
    mark surplus
```

---

# 8. Mathematical Model

## Variables

ProductionRate
Commodity units generated per tick.

ConsumptionRate
Commodity units removed per tick.

InventoryLevel
Current commodity quantity.

InventoryCapacity
Maximum commodity quantity.

ScarcityThreshold
Inventory percentage below which scarcity is triggered.

SurplusThreshold
Inventory percentage above which surplus is triggered.

---

## Threshold Defaults

ScarcityThreshold = 0.15
SurplusThreshold = 0.85

ScarcityRatio =

InventoryLevel ÷ InventoryCapacity

Economic state determination:

Scarce if ScarcityRatio < ScarcityThreshold

Surplus if ScarcityRatio > SurplusThreshold

---

# 9. Tunable Parameters

EconomicTickInterval = 60 seconds

ScarcityThreshold = 0.15

SurplusThreshold = 0.85

DefaultProductionRateRange = 1–20 units per tick

DefaultConsumptionRateRange = 1–15 units per tick

---

# 10. Integration Points

Depends On:

* Commodity Model
* Port System
* Tick Engine

Exposes:

* updated port inventories
* scarcity events
* surplus events
* price recalculation triggers

Used By:

* AI Trader System
* Economic Analytics
* Mission Generation
* Trade Route Discovery

---

# 11. Failure & Edge Cases

Negative Inventory
Inventory automatically clamped to zero.

Inventory Overflow
Values clamped to capacity.

Tick Overlap
Second tick cannot start until previous tick completes.

Missing Port Data
Port skipped and error logged.

---

# 12. Performance Constraints

Economic tick must support:

* 10,000+ ports
* 50,000+ commodity updates per tick

Target runtime:

< 50 ms per tick cycle.

Processing must be parallelizable across worker threads.

---

# 13. Security Considerations

All economic updates occur server-side.

Clients cannot:

* trigger economic ticks
* alter production rates
* modify inventory values

Tick execution must validate all inventory state.

---

# 14. Telemetry & Logging

Log events:

* economic tick execution
* major scarcity events
* commodity market spikes
* abnormal inventory changes

Metrics tracked:

* port economic throughput
* commodity volatility
* trade route activity

---

# 15. Balancing Guidelines

The economic tick must create:

* recurring trade opportunities
* regional scarcity
* fluctuating commodity values
* incentives for long-distance trade

Economic stagnation must be avoided.

Scarcity events should periodically create profitable trading routes.

---

# 16. Non-Goals (v1)

Not included in the initial version:

* player-controlled production
* industrial manufacturing chains
* economic disasters
* faction economic warfare

---

# 17. Future Extensions

Possible expansions:

* economic shock events
* supply chain disruption
* dynamic market speculation
* faction-driven resource control
* galactic trade networks

---

# End of Document
