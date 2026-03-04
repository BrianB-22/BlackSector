# AI Trader Model Specification

## Version: 0.1

## Status: Draft

## Owner: Economy Systems

## Last Updated: 2026-03-03

---

# 1. Purpose

The AI Trader Model defines autonomous economic agents responsible for transporting commodities between ports throughout the universe.

AI traders perform the following roles:

* move commodities between regions
* stabilize supply chains
* generate trade traffic
* create piracy opportunities
* populate the universe with economic activity

AI traders operate as **physical ships within the sector graph** and interact with all gameplay systems including navigation, combat, and economic simulation.

---

# 2. Scope

IN SCOPE:

* AI trader ship entities
* trade route selection
* cargo management
* risk-aware navigation
* trader personalities
* trader ship classes
* trader spawning and respawning
* AI trader naming system
* cargo drop on destruction

OUT OF SCOPE:

* player trading mechanics
* port inventory management
* commodity definitions
* combat resolution logic
* faction fleet behavior

---

# 3. Design Principles

Non-negotiable constraints:

* AI traders exist as physical ships
* server-authoritative decision making
* deterministic economic logic
* lightweight AI evaluation
* scalable to thousands of traders

AI traders must integrate seamlessly with navigation, economy, and combat systems.

---

# 4. Core Concepts

AI Trader
An autonomous ship that buys commodities at one port and sells them at another.

Trade Route
A sequence of sectors connecting a source port and destination port.

Cargo Hold
Container storing commodities transported by the trader.

Trader Personality
Behavioral profile affecting decision-making.

Trader Ship Class
Defines cargo capacity, warp performance, and durability.

Trader Population Controller
Subsystem responsible for maintaining the desired number of traders.

---

# 5. Data Model

## Entity: AITrader (Persistent)

* trader_id: uint64
* ship_name: string
* ship_class: enum
* personality: enum
* current_sector_id: uint64
* destination_sector_id: uint64
* cargo_manifest: map<string, int>
* cargo_capacity: int
* risk_tolerance: float
* last_trade_profit: float

---

## Entity: TradeRoute (Transient)

Represents a potential or active trading route.

Fields:

* origin_port_id: uint64
* destination_port_id: uint64
* commodity_id: string
* estimated_profit: float
* distance: int
* risk_score: float

---

## Entity: TraderPopulationState (Persistent)

Maintains global trader population.

Fields:

* active_traders: int
* target_population: int
* last_spawn_tick: int

---

# 6. State Machine

IDLE
→ Trader evaluating market opportunities.

SELECT_ROUTE
→ Trader calculating best trade route.

TRAVEL_TO_SOURCE
→ Trader navigating to source port.

BUY_COMMODITY
→ Trader purchasing cargo.

TRAVEL_TO_DESTINATION
→ Trader transporting cargo.

SELL_COMMODITY
→ Trader completing trade.

ESCAPE
→ Trader attempting to flee combat.

DESTROYED
→ Trader ship destroyed and removed from world.

RESPAWN_PENDING
→ Replacement trader scheduled.

---

# 7. Core Mechanics

AI traders evaluate potential trade routes periodically.

Evaluation sequence:

1. scan nearby ports
2. evaluate profitable commodity pairs
3. estimate route distance
4. estimate risk level
5. compute expected profit
6. select optimal route

Route profitability formula (simplified):

```text
expected_profit =
(price_sell - price_buy) × cargo_capacity
```

Adjusted by distance:

```text
profit_score =
expected_profit ÷ travel_distance
```

Adjusted by risk:

```text
risk_adjusted_profit =
profit_score × (1 - risk_score)
```

Trader selects the highest scoring route.

---

# 8. Mathematical Model

## Variables

BuyPrice
Commodity purchase price.

SellPrice
Commodity sale price.

CargoCapacity
Maximum cargo units.

TravelDistance
Number of sectors in route.

RiskScore
Probability of encountering danger.

---

## Profit Model

```
RawProfit =
(SellPrice - BuyPrice) × CargoCapacity
```

```
ProfitScore =
RawProfit ÷ TravelDistance
```

```
FinalRouteScore =
ProfitScore × (1 - RiskScore)
```

Routes with higher scores are preferred.

---

# 9. Tunable Parameters

TargetTraderPopulation = 500

TraderEvaluationInterval = 120 seconds

RiskAvoidanceWeight = 0.6

ProfitWeight = 1.0

MaximumRouteDistance = 40 sectors

TraderSpawnInterval = 30 seconds

CargoDropChanceOnDestruction = 0.8

---

# 10. Integration Points

Depends On:

* Commodity Model
* Port System
* Economic Tick
* Warp Mechanics
* Sector Model
* Combat System

Exposes:

* trader ship entities
* economic cargo transport
* trade route telemetry

Used By:

* piracy gameplay
* mission generation
* economic analytics
* faction intelligence systems

---

# 11. Failure & Edge Cases

No Profitable Routes
Trader enters idle state and waits for next evaluation cycle.

Destination Port Depleted
Trader re-evaluates route.

Cargo Capacity Exceeded
Transaction rejected.

Destroyed Trader
Cargo containers spawn in sector.

Navigation Failure
Trader recalculates route.

---

# 12. Performance Constraints

System must support:

* 500–5,000 AI traders
* real-time navigation updates
* route evaluation under 2ms per trader

Trader route evaluation should occur asynchronously when possible.

---

# 13. Security Considerations

AI trader decisions must be server-controlled.

Clients cannot:

* modify trader routes
* inject cargo
* influence trader spawning

Trader state updates must be validated during server ticks.

---

# 14. Telemetry & Logging

Track:

* trader route profitability
* trader destruction rate
* cargo transport volume
* piracy encounters

Log events:

* trader spawn
* trader destruction
* major cargo loss events
* economic anomalies

---

# 15. Balancing Guidelines

AI traders must maintain economic flow without dominating player activity.

Design targets:

* traders should populate major trade routes
* high-risk sectors should have fewer traders
* profitable routes should emerge organically
* piracy opportunities should occur regularly

Trader population must scale with universe size.

---

# 16. Non-Goals (v1)

Not included in initial release:

* escorted trade convoys
* trader diplomacy
* complex negotiation AI
* multi-stop trading routes

---

# 17. Future Extensions

Possible expansions include:

* faction trading fleets
* convoy systems
* smuggler traders
* dynamic economic speculation
* escort mission generation

---

# Trader Naming System

AI trader ships are assigned names from a predefined name bank.

Names are loaded from configuration files at server startup.

Example configuration location:

```
config/npc/trader_names.json
```

Each spawned trader receives a randomly selected name.

Names must be unique among active traders.

---

## Example Trader Name Configuration

```json
{
  "trader_names": [
    "Iron Horizon",
    "Stellar Venture",
    "Golden Nebula",
    "Outer Drift",
    "Solar Caravan",
    "Crimson Merchant",
    "Deep Star Runner",
    "Galactic Pioneer",
    "Silent Voyager",
    "Atlas Freighter",
    "Orion Trader",
    "Nova Wayfarer"
  ]
}
```

---

# End of Document
