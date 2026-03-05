# Mission Content Schema Specification

## Version: 0.1

## Status: Draft

## Owner: Mission Systems

## Last Updated: 2026-03-05

---

# 1. Purpose

This document defines the JSON schema used to author mission files for Black Sector.

Mission files are external JSON documents loaded by the server at startup or hot-reload. Anyone may author and share mission files. Server admins decide which files to activate on their instance.

---

# 2. File Structure

Each mission file is a JSON document containing a top-level `missions` array:

```json
{
  "missions": [
    { ... },
    { ... }
  ]
}
```

A single file may contain one or many missions.

Files are placed in `config/missions/` or any subdirectory within it.

---

# 3. Mission Definition Schema

```json
{
  "mission_id": "string (unique, snake_case)",
  "name": "string (display name)",
  "description": "string (player-facing summary)",
  "version": "string (semver, e.g. '1.0.0')",
  "author": "string (optional, for attribution)",
  "enabled": true,
  "repeatable": false,
  "repeat_cooldown_ticks": 0,
  "security_zones": ["high", "medium", "low", "black"],
  "expiry_ticks": null,
  "objectives": [ ... ],
  "rewards": { ... }
}
```

## Fields

mission_id
Unique identifier. Snake_case. Must be unique across all loaded mission files. Duplicate IDs cause the second definition to be rejected on load.

name
Short display name shown to players.

description
Player-facing description explaining the mission premise.

version
Semver string. Used for tracking when mission files are updated.

author
Optional. Attribution for community-authored missions.

enabled
Boolean. If false, mission is loaded but not offered to players. Allows admin to load a file but selectively disable missions within it.

repeatable
Boolean. If true, players may complete this mission multiple times.

repeat_cooldown_ticks
Integer. Ticks before a completed repeatable mission becomes available again. Ignored if repeatable is false.

security_zones
Array of zone names where this mission is offered. Valid values: `"high"`, `"medium"`, `"low"`, `"black"`. If empty, mission is available in all zones.

expiry_ticks
Integer or null. Number of ticks after acceptance before the mission expires. Null means no expiry.

---

# 4. Objective Schema

Objectives are evaluated in array order. Each objective must be completed before the next begins.

```json
{
  "objective_id": "string (unique within mission)",
  "type": "string (objective type)",
  "description": "string (player-facing step description)",
  "parameters": { ... },
  "failure_condition": { ... } or null
}
```

## Objective Types

### kill

Player must destroy a specified number of targets.

```json
{
  "objective_id": "obj_1",
  "type": "kill",
  "description": "Destroy 3 pirate ships in low security space.",
  "parameters": {
    "target_type": "pirate",
    "quantity": 3,
    "security_zone": "low"
  }
}
```

target_type values: `"pirate"`, `"ai_trader"`, `"any"`

security_zone: optional filter — if omitted, kills count in any zone.

---

### deliver_commodity

Player must deliver a quantity of a commodity to a specific port.

```json
{
  "objective_id": "obj_2",
  "type": "deliver_commodity",
  "description": "Deliver 20 units of food supplies to Port Helios.",
  "parameters": {
    "commodity_id": "food",
    "quantity": 20,
    "port_id": "port_helios"
  }
}
```

The player must source the commodity themselves. The mission does not provide it.

---

### acquire_commodity

Player must have a quantity of a commodity in their cargo hold.

```json
{
  "objective_id": "obj_1",
  "type": "acquire_commodity",
  "description": "Load 10 units of ore into your cargo hold.",
  "parameters": {
    "commodity_id": "ore",
    "quantity": 10
  }
}
```

Typically used as a precursor step before a deliver_commodity objective.

---

### navigate_to

Player must reach a specified system.

```json
{
  "objective_id": "obj_3",
  "type": "navigate_to",
  "description": "Travel to the Outer Drift system.",
  "parameters": {
    "system_id": "outer_drift"
  }
}
```

Objective completes when player's current sector belongs to the target system.

---

### scan_object

Player must scan a specific exploration object or anomaly.

```json
{
  "objective_id": "obj_1",
  "type": "scan_object",
  "description": "Locate and scan the derelict station in the Vance system.",
  "parameters": {
    "object_id": "derelict_vance_7",
    "required_resolution": 0.8
  }
}
```

required_resolution: float 0.0–1.0. Sensor resolution required for the objective to count.

---

### dock_at

Player must dock at a specific port.

```json
{
  "objective_id": "obj_4",
  "type": "dock_at",
  "description": "Return to Port Helios to claim your reward.",
  "parameters": {
    "port_id": "port_helios"
  }
}
```

---

### survive

Player must remain alive for a number of ticks while in a specified security zone.

```json
{
  "objective_id": "obj_2",
  "type": "survive",
  "description": "Survive in low security space for 60 ticks.",
  "parameters": {
    "duration_ticks": 60,
    "security_zone": "low"
  }
}
```

Progress resets if the player leaves the specified zone.

---

## Failure Conditions

An optional `failure_condition` block may be added to any objective. If the condition is met, the mission moves to FAILED state immediately.

```json
{
  "failure_condition": {
    "type": "ship_destroyed"
  }
}
```

Supported failure condition types:

`ship_destroyed`
Mission fails if player ship is destroyed while this objective is active.

`time_elapsed`
Mission fails if this objective is not completed within the specified ticks.

```json
{
  "failure_condition": {
    "type": "time_elapsed",
    "ticks": 300
  }
}
```

---

# 5. Reward Schema

```json
{
  "rewards": {
    "credits": 5000,
    "items": [
      {
        "item_id": "repair_kit",
        "quantity": 2
      }
    ],
    "upgrades": [
      {
        "upgrade_id": "cargo_expansion_t1"
      }
    ]
  }
}
```

## Fields

credits
Integer. Credits added to player account on completion. Use 0 for no credit reward.

items
Array of item grants. Each entry specifies an `item_id` (matching a server-known item definition) and `quantity`. Items are placed in cargo or held at the last visited station if cargo is full.

upgrades
Array of upgrade grants. Each entry specifies an `upgrade_id` (matching a server-known upgrade definition). Upgrades are queued and applied at next dock.

Any reward field may be omitted or set to an empty array. At least one reward is recommended.

---

# 6. Complete Example

A three-step mission chaining combat, acquisition, and delivery:

```json
{
  "missions": [
    {
      "mission_id": "emergency_ore_run",
      "name": "Emergency Ore Delivery",
      "description": "A mining station is desperate for refined ore. Clear a pirate blockade and deliver the goods.",
      "version": "1.0.0",
      "author": "community_author",
      "enabled": true,
      "repeatable": true,
      "repeat_cooldown_ticks": 500,
      "security_zones": ["low"],
      "expiry_ticks": 2000,
      "objectives": [
        {
          "objective_id": "step_1",
          "type": "kill",
          "description": "Destroy 2 pirate ships blocking the trade lane.",
          "parameters": {
            "target_type": "pirate",
            "quantity": 2,
            "security_zone": "low"
          },
          "failure_condition": null
        },
        {
          "objective_id": "step_2",
          "type": "acquire_commodity",
          "description": "Load 15 units of refined ore.",
          "parameters": {
            "commodity_id": "ore",
            "quantity": 15
          },
          "failure_condition": null
        },
        {
          "objective_id": "step_3",
          "type": "deliver_commodity",
          "description": "Deliver the ore to Station Ironhold.",
          "parameters": {
            "commodity_id": "ore",
            "quantity": 15,
            "port_id": "station_ironhold"
          },
          "failure_condition": {
            "type": "ship_destroyed"
          }
        }
      ],
      "rewards": {
        "credits": 8000,
        "items": [
          {
            "item_id": "repair_kit",
            "quantity": 2
          }
        ],
        "upgrades": []
      }
    }
  ]
}
```

---

# 7. Validation Rules

The server validates all mission files on load. A file that fails any rule is rejected entirely.

Required field validation:

* `mission_id` must be present, non-empty, and unique
* `objectives` must contain at least one entry
* Each objective must have `objective_id`, `type`, and `parameters`
* `rewards` must be present (may have zero values)

Type validation:

* `enabled`, `repeatable` must be boolean
* `expiry_ticks`, `repeat_cooldown_ticks` must be integer or null
* `security_zones` must contain only valid zone names

Reference validation:

* `commodity_id` values must match a loaded commodity definition
* `port_id` values must match a known port (warning only — port may be procedurally generated)
* `item_id` and `upgrade_id` values must match server-known definitions

On validation failure:

* Error is written to server log with file name and reason
* File is skipped
* Other valid files continue to load normally

---

# 8. Authoring Guidelines

For community mission authors:

* Use descriptive `mission_id` values prefixed with your handle: `authorname_missionname`
* Always set `version` and `author` fields
* Test missions in a local server instance before sharing
* Avoid `port_id` references to ports that may not exist in all universe seeds
* Use `security_zones` to scope missions appropriately — Black Sector missions should expect extreme danger
* Chain objectives to tell a story, not just stack grind steps

---

# 9. Non-Goals (v1)

Not supported in initial schema:

* conditional branching (if objective A fails, take path B)
* dynamic NPC dialogue trees
* player-choice reward selection
* faction reputation as a reward type

---

# 10. Future Extensions

Possible schema expansions:

* `prerequisites` field — require completion of another mission first
* `npc_dialogue` field — text shown at objective start/completion
* `faction_reputation` reward type
* `branch` objective type for conditional chains
* `world_event` trigger type for server-wide missions

---

# End of Document
