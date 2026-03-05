# Economic Events Specification

## Version: 0.1

## Status: Draft

## Owner: Economy Systems

## Last Updated: 2026-03-03

---

# 1. Purpose

The Economic Events subsystem introduces temporary disruptions and opportunities within the galactic economy.

Economic events dynamically modify economic behavior by altering:

* commodity production
* commodity consumption
* commodity pricing
* legality rules
* regional demand

These events prevent economic stagnation and create emergent trading opportunities for players and AI traders.

Economic events operate on top of the existing economic simulation defined by:

* Commodity Model
* Port System
* Economic Tick
* AI Trader Model

---

# 2. Scope

IN SCOPE:

* economic event definitions
* event spawning logic
* event duration management
* event cooldown rules
* economic modifier application
* event visibility (public vs hidden)
* event scope (system, region, galaxy)
* economic trigger conditions

OUT OF SCOPE:

* mission systems
* faction warfare events
* combat events
* anomaly events
* narrative story events

---

# 3. Design Principles

Non-negotiable constraints:

* server authoritative event generation
* deterministic event execution
* lightweight event evaluation
* events must create economic imbalance
* event definitions loaded from configuration

Events must enhance economic dynamism without destabilizing the entire economy.

---

# 4. Core Concepts

Economic Event
A temporary condition affecting economic behavior within a defined scope.

Event Scope
Defines the geographic area affected by the event.

Event Visibility
Determines whether the event is publicly announced or hidden.

Event Duration
The time period during which an event remains active.

Event Cooldown
Minimum time before the same event type may occur again.

Economic Modifier
Adjustment applied to economic variables such as production or price.

---

# 5. Data Model

## Entity: EconomicEventDefinition (Configuration)

Loaded from JSON configuration.

Fields:

* event_id: string
* name: string
* scope: enum
* visibility: enum
* duration_min_minutes: int
* duration_max_minutes: int
* cooldown_minutes: int
* economic_modifiers: object
* trigger_conditions: optional<object>
* region_bias: optional<string>

---

## Entity: ActiveEconomicEvent (Persistent)

Represents an event currently affecting the universe.

Fields:

* event_instance_id: uint64
* event_id: string
* scope_type: enum
* affected_region_id: optional<uint32>
* affected_system_id: optional<uint32>
* start_tick: int
* end_tick: int
* visibility: enum

---

## Entity: EconomicEventState (Persistent)

Tracks event system state.

Fields:

* active_events: list<ActiveEconomicEvent>
* last_event_spawn_tick: int
* event_spawn_counter: int

---

# 6. State Machine

INACTIVE
→ No active event.

SPAWN_PENDING
→ Event selected and preparing activation.

ACTIVE
→ Event modifiers applied.

EXPIRING
→ Event nearing completion.

ENDED
→ Modifiers removed and cooldown initiated.

Cooldown prevents immediate reoccurrence of the same event.

---

# 7. Core Mechanics

Economic events are evaluated periodically during the economic update cycle.

Event generation sequence:

1. event timer reached
2. eligible event definitions evaluated
3. trigger conditions checked
4. target region or system selected
5. event instance created
6. economic modifiers applied

Event expiration sequence:

1. event duration reached
2. modifiers removed
3. cooldown timer started
4. event instance archived

Events may affect:

* port production rates
* port consumption rates
* commodity price multipliers
* legality rules
* AI trader priorities

---

# 8. Mathematical Model

## Variables

BasePrice
Standard commodity value.

PriceMultiplier
Modifier applied during event.

ProductionMultiplier
Modifier applied to commodity production.

ConsumptionMultiplier
Modifier applied to commodity consumption.

---

## Modifier Application

```text
AdjustedPrice = BasePrice × PriceMultiplier
```

```text
AdjustedProduction = ProductionRate × ProductionMultiplier
```

```text
AdjustedConsumption = ConsumptionRate × ConsumptionMultiplier
```

---

## Stacking Rules

Multiple events may affect the same commodity.

Combined multiplier:

```text
CombinedMultiplier =
product(all_event_multipliers)
```

Clamp limits:

MinimumMultiplier = 0.25
MaximumMultiplier = 5.0

---

# 9. Tunable Parameters

MinorEventInterval = 20 minutes

RegionalEventInterval = 120 minutes

GlobalEventInterval = 720 minutes

MaxConcurrentEvents = 6

MaxPriceMultiplier = 5.0

MinimumEventDuration = 30 minutes

MaximumEventDuration = 360 minutes

---

# 10. Integration Points

Depends On:

* Commodity Model
* Port System
* Economic Tick
* Region Partitioning

Exposes:

* active economic events
* commodity modifier signals
* economic news notifications

Used By:

* AI Trader System
* Mission Generation
* Player Economic Interface
* Galactic News System

---

# 11. Failure & Edge Cases

Invalid Event Configuration
Event rejected during configuration load.

Overlapping Event Scope Conflict
Events allowed but modifiers capped.

Missing Target Region
Event spawn cancelled.

Event Duration Overflow
Duration clamped to maximum allowed value.

---

# 12. Performance Constraints

System must support:

* evaluation across up to 1,000 ports
* multiple concurrent events
* near-zero impact on economic tick performance

Event evaluation should execute in under 5ms.

---

# 13. Security Considerations

All event generation must occur server-side.

Clients cannot:

* trigger economic events
* modify economic modifiers
* alter event durations

Event configuration files must be validated before loading.

---

# 14. Telemetry & Logging

Track metrics including:

* event frequency
* commodity price spikes
* regional economic volatility
* trade volume changes during events

Log events:

* event creation
* event expiration
* modifier anomalies

---

# 15. Balancing Guidelines

Economic events must:

* introduce temporary scarcity
* encourage trade route shifts
* create profit opportunities
* avoid complete economic collapse

Event intensity should scale with universe size.

Events should favor frontier regions slightly more than core regions.

---

# 16. Non-Goals (v1)

Not included in the initial version:

* player-triggered economic events
* faction-driven market manipulation
* permanent economic shifts
* narrative economic campaigns

---

# 17. Future Extensions

Possible expansions include:

* player relief missions
* faction trade embargoes
* economic warfare mechanics
* industrial disasters
* galactic market speculation

---

# Event Configuration

Event definitions are stored in JSON configuration files.

Example location:

```
config/economy/economic_events.json
```

Events are loaded during server startup.

---

## Example Event Configuration

```json
{
  "events": [
    {
      "event_id": "food_shortage",
      "name": "Food Shortage",
      "scope": "region",
      "visibility": "public",
      "duration_min_minutes": 60,
      "duration_max_minutes": 180,
      "cooldown_minutes": 360,
      "economic_modifiers": {
        "food_supplies": {
          "demand_multiplier": 2.0,
          "price_multiplier": 1.8
        }
      },
      "region_bias": "agricultural"
    },
    {
      "event_id": "industrial_boom",
      "name": "Industrial Boom",
      "scope": "system",
      "visibility": "public",
      "duration_min_minutes": 90,
      "duration_max_minutes": 240,
      "cooldown_minutes": 480,
      "economic_modifiers": {
        "refined_ore": {
          "consumption_multiplier": 2.0
        }
      },
      "region_bias": "industrial"
    },
    {
      "event_id": "black_market_surge",
      "name": "Black Market Surge",
      "scope": "region",
      "visibility": "hidden",
      "duration_min_minutes": 120,
      "duration_max_minutes": 240,
      "cooldown_minutes": 600,
      "economic_modifiers": {
        "alien_artifacts": {
          "price_multiplier": 2.5
        }
      }
    }
  ]
}
```

---

# End of Document
