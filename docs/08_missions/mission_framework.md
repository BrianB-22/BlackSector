# Mission Framework Specification

## Version: 0.1

## Status: Draft

## Owner: Mission Systems

## Last Updated: 2026-03-05

---

# 1. Purpose

The Mission Framework defines how missions are loaded, managed, and evaluated within the server.

Missions in Black Sector are **data-driven and externally defined**. All mission content is authored in JSON files outside the server binary. The server evaluates mission logic against game state during each tick cycle.

This architecture enables:

* community-authored mission content
* server admin curation of active missions
* mission sharing and distribution without server modification
* runtime hot-reload of mission content

---

# 2. Scope

IN SCOPE:

* mission file loading and validation
* mission lifecycle management
* objective evaluation during tick
* reward distribution
* player mission state tracking
* admin hot-reload support

OUT OF SCOPE:

* mission content authoring (see content_schema.md)
* admin control interface (see admin_controls.md)
* tick event injection (see event_injection.md)
* faction-specific mission chains
* player-created missions

---

# 3. Design Principles

Non-negotiable constraints:

* missions are defined entirely in external JSON files
* server binary contains no hardcoded mission content
* all mission evaluation is server-authoritative
* mission evaluation occurs inside the deterministic tick loop
* mission state is persistent and survives server restart
* admin controls which missions are active on their instance

---

# 4. Core Concepts

Mission Definition
An external JSON file describing a mission's objectives, triggers, rewards, and constraints.

Mission Instance
A server-side record of a player's active engagement with a mission.

Objective
A single step within a mission that the player must complete.

Objective Chain
A sequence of objectives that must be completed in order.

Reward
Credits, items, or upgrades granted upon mission completion.

Mission Registry
The in-memory collection of all loaded and active mission definitions.

---

# 5. Data Model

## Entity: MissionDefinition (Configuration)

Loaded from external JSON at startup or hot-reload. Immutable during runtime.

Fields:

* mission_id: string
* name: string
* description: string
* version: string
* author: string
* security_zones: list<string>
* min_security_rating: float or null
* objectives: list<ObjectiveDefinition>
* rewards: RewardDefinition
* expiry_ticks: int or null
* repeatable: bool
* enabled: bool

---

## Entity: MissionInstance (Persistent)

Tracks a player's active or completed engagement with a mission.

Fields:

* instance_id: uint64
* mission_id: string
* player_id: uint64
* current_objective_index: int
* state: enum
* accepted_tick: int
* completed_tick: int or null
* expired_tick: int or null

---

## Entity: ObjectiveProgress (Persistent)

Tracks progress toward a specific objective within a mission instance.

Fields:

* instance_id: uint64
* objective_id: string
* progress_current: int
* progress_required: int
* is_complete: bool

---

# 6. State Machine

## Mission Instance States

AVAILABLE
→ Mission is loaded and visible to eligible players.

ACCEPTED
→ Player has accepted the mission.

IN_PROGRESS
→ Player is actively working through objectives.

COMPLETED
→ All objectives satisfied; rewards distributed.

FAILED
→ Failure condition triggered (e.g., ship destroyed mid-escort).

EXPIRED
→ Mission timer elapsed before completion.

ABANDONED
→ Player voluntarily abandoned the mission.

Transitions:

```
AVAILABLE → ACCEPTED (player accepts)
ACCEPTED → IN_PROGRESS (first objective begins)
IN_PROGRESS → IN_PROGRESS (objective N completed, objective N+1 begins)
IN_PROGRESS → COMPLETED (final objective completed)
IN_PROGRESS → FAILED (failure condition met)
IN_PROGRESS → EXPIRED (expiry_ticks elapsed)
IN_PROGRESS → ABANDONED (player abandons)
```

---

# 7. Mission File Loading

Mission definitions are loaded from:

```
config/missions/
```

The server scans this directory at startup and loads all `.json` files found.

Loading sequence:

1. Scan `config/missions/` for `.json` files
2. Parse and validate each file against mission schema
3. Reject files that fail validation (log error, continue loading others)
4. Register valid missions into the Mission Registry
5. Apply enabled/disabled state per admin configuration

Each mission file may define one or multiple missions:

```json
{
  "missions": [ ... ]
}
```

Subdirectories within `config/missions/` are supported for organization:

```
config/missions/combat/
config/missions/trade/
config/missions/exploration/
```

---

# 8. Objective Evaluation

Mission objective progress is evaluated during the tick loop after all game actions are resolved.

Evaluation sequence (per tick):

1. Collect game events emitted this tick (kills, deliveries, arrivals, etc.)
2. For each active MissionInstance in IN_PROGRESS state:
   a. Retrieve current objective
   b. Check relevant events against objective conditions
   c. Update ObjectiveProgress
   d. If objective complete: advance to next objective or mark COMPLETED
   e. Check expiry_ticks
3. Distribute rewards for newly COMPLETED instances
4. Persist updated mission state

Objective evaluation must remain lightweight. No full-world scans per tick.

---

# 9. Reward Distribution

On mission completion:

* credits added to player account immediately
* items added to player cargo or station storage if cargo is full
* upgrades applied to player's current ship or queued for next dock

Reward distribution occurs inside the tick in which the final objective is completed.

---

# 10. Repeatable Missions

If `repeatable: true`, a completed mission instance returns to AVAILABLE state after a cooldown period defined in the mission definition.

A player may not hold two active instances of the same repeatable mission simultaneously.

---

# 11. Integration Points

Depends On:

* Tick Engine
* Event Injection System
* Player Account System
* Ship Cargo System
* Economy Engine (reward delivery)

Exposes:

* Mission Registry
* MissionInstance state
* ObjectiveProgress state
* MissionCompletedEvent

Used By:

* Admin Controls
* Player UI / Protocol Layer
* Event Injection System
* Telemetry

---

# 12. Failure & Edge Cases

Player Destroyed Mid-Mission
Mission moves to FAILED if the active objective has a ship_destroyed failure condition. Otherwise mission remains IN_PROGRESS.

Server Restart
All MissionInstance and ObjectiveProgress state is persistent. Missions resume on reconnect.

Mission File Removed After Accept
Instance continues to completion using the definition snapshot taken at accept time.

Invalid Reward Item
Reward skipped and error logged. Other rewards still distributed.

Duplicate Mission IDs
Second definition rejected at load time. Error logged.

---

# 13. Performance Constraints

Mission evaluation must support:

* up to 100 active mission instances concurrently
* objective evaluation under 5ms per tick cycle total

Mission evaluation runs after core game resolution and must not delay tick completion.

---

# 14. Security Considerations

All mission evaluation is server-side.

Clients cannot:

* claim objective completion directly
* modify reward values
* skip objectives
* alter mission state

Mission definitions are read-only at runtime. Files on disk may be updated by admin; changes apply on next hot-reload.

---

# 15. Telemetry & Logging

Log events:

* mission file loaded successfully
* mission file rejected (with reason)
* mission accepted by player
* mission completed
* mission failed or expired
* reward distribution

Metrics tracked:

* completion rate per mission
* average time to complete
* most popular missions
* failure cause distribution

---

# 16. Non-Goals (v1)

Not included in initial release:

* player-created missions
* faction-specific mission chains
* dynamic procedural mission generation
* mission marketplaces or trading
* multi-player cooperative mission instances

---

# 17. Future Extensions

Possible expansions:

* faction reputation rewards
* multi-player mission instances
* branching objective trees
* timed event missions (server-wide)
* mission prerequisites and unlock chains

---

# End of Document
