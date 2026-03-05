# Event Log Schema Specification

## Version: 0.1

## Status: Draft

## Owner: Core Architecture

## Last Updated: 2026-03-05

---

# 1. Purpose

Defines the schema for the server event log.

The event log is an append-only record of significant game events. It serves:

* telemetry and operational monitoring
* debugging and post-incident analysis
* economic and gameplay analytics
* audit trail for player-affecting actions

The event log is separate from protocol messages and from SQLite game state. It is written to a structured log file or log table, not to the primary game database.

---

# 2. Design Principles

* Append-only — events are never modified or deleted
* Structured — all entries are JSON objects, one per line (NDJSON)
* Tick-indexed — all entries include the server tick at time of event
* Lightweight — event logging must not block tick execution
* Typed — each event type has a defined payload schema

---

# 3. Log Entry Envelope

All event log entries follow this envelope:

```json
{
  "tick": 4821,
  "timestamp": 1709612345,
  "event_type": "player_trade",
  "player_id": "550e8400-e29b-41d4-a716-446655440000",
  "system_id": 14,
  "payload": { ... }
}
```

## Envelope Fields

| Field        | Type    | Required | Description                              |
| ------------ | ------- | -------- | ---------------------------------------- |
| `tick`       | int     | yes      | Server tick number at time of event      |
| `timestamp`  | int     | yes      | Unix epoch seconds                       |
| `event_type` | string  | yes      | Identifies the event type                |
| `player_id`  | string  | no       | UUID of player involved (if applicable)  |
| `system_id`  | int     | no       | System ID where event occurred           |
| `payload`    | object  | yes      | Event-specific data (see below)          |

---

# 4. Event Type Catalog

## 4.1 Server Events

### server_start

```json
{
  "event_type": "server_start",
  "payload": {
    "protocol_version": "1.0",
    "tick_interval_ms": 2000,
    "config_files_loaded": ["commodities.json", "ship_classes.json"]
  }
}
```

---

### server_stop

```json
{
  "event_type": "server_stop",
  "payload": {
    "reason": "graceful_shutdown",
    "final_tick": 9043
  }
}
```

Reason values: `graceful_shutdown` | `crash` | `admin_halt`

---

### tick_slow

Emitted when a tick takes longer than the configured threshold.

```json
{
  "event_type": "tick_slow",
  "payload": {
    "tick": 4821,
    "duration_ms": 312,
    "threshold_ms": 200
  }
}
```

---

## 4.2 Player Events

### player_connect

```json
{
  "event_type": "player_connect",
  "player_id": "...",
  "payload": {
    "session_id": "...",
    "interface_mode": "TEXT",
    "remote_addr": "192.168.1.1"
  }
}
```

---

### player_disconnect

```json
{
  "event_type": "player_disconnect",
  "player_id": "...",
  "payload": {
    "session_id": "...",
    "reason": "client_closed",
    "session_duration_ticks": 1420
  }
}
```

Reason values: `client_closed` | `timeout` | `kicked` | `server_shutdown`

---

### player_auth_fail

```json
{
  "event_type": "player_auth_fail",
  "payload": {
    "remote_addr": "192.168.1.1",
    "reason": "invalid_token"
  }
}
```

---

## 4.3 Navigation Events

### player_jump

```json
{
  "event_type": "player_jump",
  "player_id": "...",
  "system_id": 42,
  "payload": {
    "from_system_id": 17,
    "to_system_id": 42,
    "fuel_consumed": 12
  }
}
```

---

## 4.4 Combat Events

### combat_start

```json
{
  "event_type": "combat_start",
  "player_id": "...",
  "system_id": 14,
  "payload": {
    "combat_id": "...",
    "aggressor_type": "player",
    "target_type": "npc"
  }
}
```

---

### combat_end

```json
{
  "event_type": "combat_end",
  "system_id": 14,
  "payload": {
    "combat_id": "...",
    "outcome": "player_victory",
    "duration_ticks": 38,
    "player_id": "..."
  }
}
```

Outcome values: `player_victory` | `player_destroyed` | `target_fled` | `player_fled`

---

### ship_destroyed

```json
{
  "event_type": "ship_destroyed",
  "player_id": "...",
  "system_id": 14,
  "payload": {
    "ship_id": "...",
    "destroyed_by": "npc_pirate",
    "combat_id": "..."
  }
}
```

---

## 4.5 Trade Events

### player_trade

```json
{
  "event_type": "player_trade",
  "player_id": "...",
  "system_id": 7,
  "payload": {
    "port_id": 22,
    "commodity_id": "refined_ore",
    "quantity": 50,
    "price_per_unit": 340,
    "trade_type": "sell",
    "total_credits": 17000
  }
}
```

Trade types: `buy` | `sell`

---

## 4.6 Mining Events

### mining_yield

```json
{
  "event_type": "mining_yield",
  "player_id": "...",
  "system_id": 9,
  "payload": {
    "field_id": 88,
    "commodity_id": "raw_ore",
    "quantity_extracted": 14,
    "depletion_after": 0.34
  }
}
```

---

## 4.7 Exploration Events

### anomaly_discovered

```json
{
  "event_type": "anomaly_discovered",
  "player_id": "...",
  "system_id": 31,
  "payload": {
    "anomaly_id": 7,
    "anomaly_type": "derelict_ship",
    "first_discovery": true
  }
}
```

---

### system_discovered

```json
{
  "event_type": "system_discovered",
  "player_id": "...",
  "system_id": 31,
  "payload": {
    "first_discovery": true
  }
}
```

---

## 4.8 Economic Events

### economic_event_start

```json
{
  "event_type": "economic_event_start",
  "payload": {
    "event_instance_id": 412,
    "event_id": "food_shortage",
    "scope_type": "region",
    "affected_region_id": 3,
    "end_tick": 5200,
    "visibility": "public"
  }
}
```

---

### economic_event_end

```json
{
  "event_type": "economic_event_end",
  "payload": {
    "event_instance_id": 412,
    "event_id": "food_shortage",
    "actual_duration_ticks": 380
  }
}
```

---

## 4.9 Mission Events

### mission_accepted

```json
{
  "event_type": "mission_accepted",
  "player_id": "...",
  "payload": {
    "instance_id": "...",
    "mission_id": "escort_convoy"
  }
}
```

---

### mission_completed

```json
{
  "event_type": "mission_completed",
  "player_id": "...",
  "payload": {
    "instance_id": "...",
    "mission_id": "escort_convoy",
    "duration_ticks": 412,
    "rewards": {
      "credits": 5000,
      "items": [],
      "upgrades": []
    }
  }
}
```

---

### mission_failed

```json
{
  "event_type": "mission_failed",
  "player_id": "...",
  "payload": {
    "instance_id": "...",
    "mission_id": "escort_convoy",
    "reason": "ship_destroyed"
  }
}
```

---

## 4.10 Admin Events

### admin_command

```json
{
  "event_type": "admin_command",
  "payload": {
    "command": "mission reload",
    "result": "success",
    "details": "12 missions loaded"
  }
}
```

---

### config_reload

```json
{
  "event_type": "config_reload",
  "payload": {
    "config_type": "missions",
    "files_loaded": 12,
    "files_failed": 0
  }
}
```

---

# 5. Storage

## Log File

Event log entries are written to a rotating NDJSON log file.

Default path: `logs/events.log`

Each entry is one line of JSON followed by a newline character.

Log rotation: daily or when file exceeds 100MB, configurable.

---

## Optional Database Table

For servers requiring queryable event history, events may also be written to a SQLite table:

```sql
CREATE TABLE event_log (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  tick         INTEGER NOT NULL,
  timestamp    INTEGER NOT NULL,
  event_type   TEXT NOT NULL,
  player_id    TEXT,
  system_id    INTEGER,
  payload_json TEXT NOT NULL
);

CREATE INDEX idx_event_log_tick        ON event_log (tick);
CREATE INDEX idx_event_log_player      ON event_log (player_id, tick);
CREATE INDEX idx_event_log_event_type  ON event_log (event_type, tick);
```

Use of the database table is optional and configurable. The file log is always written.

---

# 6. Retention

Log files are retained for a configurable duration (default: 30 days).

The event_log database table, if used, may be pruned after a configurable number of ticks (default: 7 days equivalent).

---

# 7. Performance Constraints

* Event log writes must not block tick execution
* Log writes are asynchronous via a buffered channel
* Log buffer overflow (if consumer falls behind) drops the oldest entries and emits a warning
* Target: under 0.5ms overhead per tick for log flushing

---

# 8. Non-Goals (v1)

* Real-time log streaming to external systems
* Structured log aggregation (e.g. ELK stack)
* Replay from event log (snapshot-based recovery is used instead)
* Distributed log coordination

---

# End of Document
