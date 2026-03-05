# Snapshot Format Specification

## Version: 0.1

## Status: Draft

## Owner: Core Architecture

## Last Updated: 2026-03-05

---

# 1. Purpose

Defines the format used to serialize and restore full server state.

Snapshots are used for:

* crash recovery — restore the last known good state after an unclean shutdown
* scheduled persistence — periodic full-state flush to disk beyond incremental SQLite writes
* server migration — move a running game to a new host

Snapshots complement the SQLite database. The database is the primary persistence mechanism during normal operation. Snapshots provide a point-in-time consistent image of the full simulation state.

---

# 2. Design Principles

* A snapshot must be sufficient to fully reconstruct server state without any other input
* Snapshot format is JSON for portability and debuggability
* Configuration data (commodities, ship classes, missions) is NOT included — it is reloaded from config files
* Snapshots are versioned to allow future migration
* Writing a snapshot must not block the tick loop; snapshots are written asynchronously

---

# 3. File Location and Naming

Snapshots are written to a configurable directory.

Default path: `snapshots/`

Filename format:

```
snapshot_{tick}_{timestamp}.json
```

Example:

```
snapshot_004821_1709612345.json
```

The server always loads the most recent valid snapshot on startup if present.

A `snapshot_latest.json` symlink (or copy on systems without symlink support) is maintained pointing to the current snapshot.

---

# 4. Snapshot Envelope

```json
{
  "snapshot_version": "1.0",
  "tick": 4821,
  "timestamp": 1709612345,
  "server_name": "Black Sector",
  "protocol_version": "1.0",
  "state": { ... }
}
```

## Envelope Fields

| Field              | Type   | Description                                    |
| ------------------ | ------ | ---------------------------------------------- |
| `snapshot_version` | string | Format version for migration compatibility     |
| `tick`             | int    | Server tick at time of snapshot                |
| `timestamp`        | int    | Unix epoch seconds                             |
| `server_name`      | string | Server name from configuration                 |
| `protocol_version` | string | Protocol version in use                        |
| `state`            | object | Full game state (see below)                    |

---

# 5. State Object Structure

```json
{
  "state": {
    "players": [ ... ],
    "sessions": [ ... ],
    "ships": [ ... ],
    "ai_traders": [ ... ],
    "port_inventories": [ ... ],
    "active_economic_events": [ ... ],
    "economic_event_cooldowns": [ ... ],
    "mission_instances": [ ... ],
    "anomalies": [ ... ],
    "asteroid_fields": [ ... ],
    "player_map_data": [ ... ],
    "player_waypoints": [ ... ],
    "market_price_history_recent": [ ... ],
    "economic_event_state": { ... }
  }
}
```

World structure (regions, systems, ports, jump connections, hazard zones, commodities) is NOT included in the snapshot. These are static and reloaded from the database at startup.

---

# 6. Included State Sections

## 6.1 players

Array of player records.

```json
{
  "player_id": "550e8400-e29b-41d4-a716-446655440000",
  "player_name": "Orion",
  "token_hash": "...",
  "created_at": 1709600000,
  "last_login_at": 1709612000,
  "is_banned": false
}
```

---

## 6.2 sessions

Array of non-terminated session records.

```json
{
  "session_id": "...",
  "player_id": "...",
  "interface_mode": "TEXT",
  "state": "DISCONNECTED_LINGERING",
  "connected_at": 1709610000,
  "disconnected_at": 1709612000,
  "linger_expiry_at": 1709612300,
  "last_activity_at": 1709612000
}
```

Only sessions in state `CONNECTED`, `DISCONNECTED_LINGERING`, or `DOCKED_OFFLINE` are included. `TERMINATED` sessions are excluded.

---

## 6.3 ships

Array of all ships, including destroyed ships (for persistence of destruction state).

```json
{
  "ship_id": "...",
  "player_id": "...",
  "ship_class": "courier",
  "hull_points": 80,
  "max_hull_points": 100,
  "shield_points": 40,
  "max_shield_points": 50,
  "energy_points": 75,
  "max_energy_points": 100,
  "cargo_capacity": 20,
  "cargo": [
    { "slot_index": 0, "commodity_id": "food_supplies", "quantity": 10 }
  ],
  "upgrades": [],
  "current_system_id": 14,
  "position": { "x": 12.5, "y": -8.3 },
  "status": "IN_SPACE",
  "docked_at_port_id": null,
  "last_updated_tick": 4820
}
```

---

## 6.4 ai_traders

Array of AI trader state records.

```json
{
  "trader_id": 42,
  "name": "Wandering Khal",
  "ship_class": "freighter",
  "current_system_id": 7,
  "status": "TRAVELING",
  "home_region_id": 2,
  "current_cargo_commodity": "refined_ore",
  "current_cargo_quantity": 30,
  "last_trade_tick": 4790
}
```

---

## 6.5 port_inventories

Array of port inventory entries for all ports.

```json
{
  "port_id": 22,
  "commodity_id": "food_supplies",
  "quantity": 500,
  "buy_price": 120,
  "sell_price": 100,
  "updated_tick": 4800
}
```

---

## 6.6 active_economic_events

Array of currently active economic event instances.

```json
{
  "event_instance_id": 412,
  "event_id": "food_shortage",
  "scope_type": "region",
  "affected_region_id": 3,
  "affected_system_id": null,
  "start_tick": 4500,
  "end_tick": 5200,
  "visibility": "public"
}
```

---

## 6.7 economic_event_cooldowns

Array of active cooldown records.

```json
{
  "event_id": "food_shortage",
  "scope_key": "region:3",
  "last_ended_tick": 4820
}
```

---

## 6.8 mission_instances

Array of non-terminal mission instances.

```json
{
  "instance_id": "...",
  "mission_id": "pirate_hunt",
  "player_id": "...",
  "status": "IN_PROGRESS",
  "accepted_tick": 4700,
  "started_tick": 4701,
  "completed_tick": null,
  "failed_reason": null,
  "expires_at_tick": 5200,
  "objectives": [
    {
      "objective_index": 0,
      "status": "COMPLETED",
      "current_value": 3,
      "required_value": 3
    },
    {
      "objective_index": 1,
      "status": "ACTIVE",
      "current_value": 0,
      "required_value": 1
    }
  ]
}
```

---

## 6.9 anomalies

Full array of anomaly records.

```json
{
  "anomaly_id": 7,
  "system_id": 31,
  "anomaly_type": "derelict_ship",
  "position": { "x": 55.1, "y": -22.8 },
  "is_discovered": true,
  "discovered_by_player_id": "...",
  "resource_quantity": 200,
  "depleted": false
}
```

---

## 6.10 asteroid_fields

Full array of asteroid field state records.

```json
{
  "field_id": 88,
  "system_id": 9,
  "field_type": "rich",
  "position": { "x": 3.2, "y": 14.7 },
  "depletion_level": 0.34,
  "last_mined_tick": 4810,
  "resources": [
    {
      "commodity_id": "raw_ore",
      "base_yield": 20,
      "current_multiplier": 0.85
    }
  ]
}
```

---

## 6.11 player_map_data

Array of all player map discovery records.

```json
{
  "player_id": "...",
  "system_id": 31,
  "discovered_at_tick": 4200,
  "scan_level": 2
}
```

---

## 6.12 player_waypoints

Array of all player waypoints.

```json
{
  "waypoint_id": 1,
  "player_id": "...",
  "name": "Good Mining Spot",
  "system_id": 9,
  "position": { "x": 3.2, "y": 14.7 },
  "created_at_tick": 3900
}
```

---

## 6.13 market_price_history_recent

Recent market history entries (last 500 ticks). Full history remains in the database.

```json
{
  "port_id": 22,
  "commodity_id": "food_supplies",
  "tick": 4820,
  "buy_price": 130,
  "sell_price": 110,
  "buy_volume": 200,
  "sell_volume": 150
}
```

---

## 6.14 economic_event_state

Scalar state values for the economic event subsystem.

```json
{
  "last_event_spawn_tick": 4700,
  "event_spawn_counter": 42
}
```

---

# 7. Snapshot Lifecycle

## Writing

1. Snapshot is triggered by the tick engine after tick resolution
2. Current state is serialized to a temporary file
3. Temporary file is renamed atomically to the snapshot filename
4. `snapshot_latest.json` link is updated
5. Old snapshots beyond the retention window are deleted

Snapshots are written asynchronously. The tick loop does not wait for the write to complete.

---

## Reading on Startup

1. Server checks for `snapshot_latest.json` (or most recent timestamped snapshot)
2. If found, validates `snapshot_version` and `protocol_version`
3. Applies state from snapshot to in-memory structures
4. Continues ticking from `tick + 1`
5. If no snapshot found, starts from tick 0 with empty world state

---

## Validation

Before applying a snapshot, the server validates:

* `snapshot_version` is supported
* `protocol_version` matches or is compatible
* All required sections are present
* Referenced IDs (system_id, port_id, etc.) exist in the loaded world data

Invalid snapshots are rejected and the server halts with a descriptive error.

---

# 8. Snapshot Interval

Default: every 100 ticks (approximately every 3.3 minutes at 2-second tick interval).

Configurable via `server.snapshot_interval_ticks` in server configuration.

---

# 9. Retention

Default: keep last 10 snapshots.

Older snapshots are deleted automatically after a successful write.

---

# 10. Non-Goals (v1)

* Incremental / delta snapshots
* Compressed snapshot format
* Remote snapshot storage
* Snapshot diff and merge
* Cross-version migration tooling

---

# End of Document
