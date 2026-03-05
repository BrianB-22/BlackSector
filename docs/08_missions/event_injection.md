# Mission Event Injection Specification

## Version: 0.1

## Status: Draft

## Owner: Mission Systems

## Last Updated: 2026-03-05

---

# 1. Purpose

This document defines how mission objective evaluation hooks into the server's deterministic tick engine.

The mission system does not poll game state. Instead, game subsystems **emit events** during tick execution. The mission evaluator consumes these events after the main game resolution phase and updates mission progress accordingly.

This keeps mission logic decoupled from game subsystems while remaining fully server-authoritative and deterministic.

---

# 2. Scope

IN SCOPE:

* game event types used for mission evaluation
* event emission points within the tick order
* mission evaluator position in tick sequence
* objective-to-event mapping
* event filtering and routing

OUT OF SCOPE:

* mission definition authoring (see content_schema.md)
* mission lifecycle management (see mission_framework.md)
* admin controls (see admin_controls.md)
* combat resolution logic
* economy tick logic

---

# 3. Design Principles

* game subsystems emit events; they have no knowledge of missions
* mission evaluator is a consumer of events, not a modifier of game state during resolution
* event injection occurs after all game resolution in a tick
* no subsystem calls the mission system directly
* mission evaluation is read-only with respect to game state; only mission state is mutated

---

# 4. Tick Execution Order

Mission evaluation occurs at the end of the tick, after all game state has been resolved:

```
1. Process player input commands
2. Resolve movement and navigation
3. Resolve combat
4. Resolve mining cycles
5. Resolve economic tick (production, consumption, prices)
6. Resolve exploration scans
7. --- Emit game events (all subsystems) ---
8. Evaluate mission objective progress  ← mission system runs here
9. Distribute mission rewards
10. Persist state changes
11. Broadcast updates to clients
```

Mission evaluation never modifies game state resolved in steps 1–6. It only updates mission instance state and queues rewards.

---

# 5. Game Event Types

The following events are emitted by game subsystems and consumed by the mission evaluator.

---

## EntityDestroyedEvent

Emitted by: Combat System

```
EntityDestroyedEvent {
    killer_player_id: uint64
    target_entity_id: uint64
    target_entity_type: enum   // pirate, ai_trader, any
    sector_id: uint64
    security_zone: string
    tick: int
}
```

Used by objective type: `kill`

---

## CommodityDeliveredEvent

Emitted by: Port System (on successful sell transaction)

```
CommodityDeliveredEvent {
    player_id: uint64
    port_id: uint64
    commodity_id: string
    quantity: int
    tick: int
}
```

Used by objective type: `deliver_commodity`

---

## CommodityAcquiredEvent

Emitted by: Port System (on buy transaction) and Mining System (on yield)

```
CommodityAcquiredEvent {
    player_id: uint64
    commodity_id: string
    quantity: int
    source: enum   // port, mining
    tick: int
}
```

Used by objective type: `acquire_commodity`

---

## PlayerEnteredSystemEvent

Emitted by: Navigation System (on warp arrival)

```
PlayerEnteredSystemEvent {
    player_id: uint64
    system_id: uint64
    sector_id: uint64
    security_zone: string
    tick: int
}
```

Used by objective type: `navigate_to`

---

## ObjectScannedEvent

Emitted by: Sensor System

```
ObjectScannedEvent {
    player_id: uint64
    object_id: uint64
    resolution_achieved: float
    tick: int
}
```

Used by objective type: `scan_object`

---

## PlayerDockedEvent

Emitted by: Port System (on successful dock)

```
PlayerDockedEvent {
    player_id: uint64
    port_id: uint64
    tick: int
}
```

Used by objective type: `dock_at`

---

## PlayerSurvivalTickEvent

Emitted by: Tick Engine (once per tick per active player)

```
PlayerSurvivalTickEvent {
    player_id: uint64
    sector_id: uint64
    security_zone: string
    tick: int
}
```

Used by objective type: `survive`

Survival progress increments only while the player remains in the required security zone.

---

## PlayerDestroyedEvent

Emitted by: Combat System (when player ship is destroyed)

```
PlayerDestroyedEvent {
    player_id: uint64
    sector_id: uint64
    tick: int
}
```

Used by: failure condition `ship_destroyed`

---

# 6. Event Routing

Events are not broadcast to all missions. The evaluator routes events only to mission instances where they are relevant.

The mission evaluator maintains an in-memory index:

```
player_id → list of active MissionInstances
```

For each incoming event:

1. Look up player_id in index
2. For each active instance, retrieve current objective
3. Check if event type matches objective type
4. If match: evaluate parameters and update progress

This keeps evaluation O(active_missions_per_player) per event, not O(all_missions).

---

# 7. Objective Evaluation Logic

Each objective type maps to a specific event and parameter check:

| Objective Type    | Event                    | Progress Condition                                     |
| ----------------- | ------------------------ | ------------------------------------------------------ |
| kill              | EntityDestroyedEvent     | target_type matches AND security_zone matches (if set) |
| deliver_commodity | CommodityDeliveredEvent  | port_id matches AND commodity_id matches               |
| acquire_commodity | CommodityAcquiredEvent   | commodity_id matches                                   |
| navigate_to       | PlayerEnteredSystemEvent | system_id matches                                      |
| scan_object       | ObjectScannedEvent       | object_id matches AND resolution_achieved >= required  |
| dock_at           | PlayerDockedEvent        | port_id matches                                        |
| survive           | PlayerSurvivalTickEvent  | security_zone matches; progress = ticks in zone        |

For quantitative objectives (kill, acquire_commodity, deliver_commodity):
progress_current increments per qualifying event until progress_current >= progress_required.

For boolean objectives (navigate_to, dock_at, scan_object):
objective completes immediately on first qualifying event.

---

# 8. Failure Condition Evaluation

`ship_destroyed` failure:

On receiving `PlayerDestroyedEvent` for a player with an active mission instance where the current objective has `failure_condition.type = "ship_destroyed"`:

→ Mission instance transitions to FAILED state immediately.

`time_elapsed` failure:

On each tick, the evaluator checks elapsed ticks for active instances with time-limited objectives.

If `ticks_elapsed_on_objective >= failure_condition.ticks`:

→ Mission instance transitions to FAILED state.

---

# 9. Reward Injection

Reward distribution occurs in step 9 of the tick order, immediately after objective evaluation.

On COMPLETED transition:

1. Query RewardDefinition from MissionDefinition
2. Add credits to player account
3. Add items to player cargo or station storage queue if cargo is full
4. Queue upgrade grants for application at next dock

Reward injection completes within the same tick as the completion event.

---

# 10. Event Emission Contract

Game subsystems that emit mission-relevant events must follow this contract:

* events are emitted after the subsystem resolves its own state changes
* events carry only the data fields defined in Section 5
* subsystems do not check mission state before emitting
* subsystems do not conditionally suppress events based on mission context

The mission system is a passive consumer. Game subsystems have no awareness of it.

---

# 11. Failure & Edge Cases

Player Logs Off Mid-Mission
Mission instance persists. On reconnect, progress resumes. Survival-type objectives pause during disconnect.

Duplicate Events in One Tick
Each event is evaluated once. Duplicate entity IDs in the same tick are deduplicated by the evaluator.

Event Emitted for Player With No Active Mission
Discarded immediately.

Objective Complete but Reward Fails
Mission marks COMPLETED. Reward error is logged. Admin may manually grant via console.

---

# 12. Performance Constraints

Event evaluation must complete under 5ms total per tick across all active instances.

The player → instance index must be maintained in memory. No database queries during event routing.

---

# 13. Security Considerations

Events are emitted only by server subsystems.

Clients cannot:

* inject synthetic game events
* trigger objective completion directly
* modify progress counters

All event data originates from server-resolved game state and is never sourced from client input.

---

# 14. Non-Goals (v1)

Not included in initial release:

* complex conditional event chains (if event A then require event B within N ticks)
* multi-player shared objective progress
* event sourcing or full event log replay
* client-side mission progress prediction

---

# 15. Future Extensions

Possible expansions:

* composite event conditions (AND / OR across multiple event types)
* timed event windows (objective must be completed between tick X and tick Y)
* global server event triggers (server-wide mission activations)
* cooperative objective sharing across player groups

---

# End of Document
