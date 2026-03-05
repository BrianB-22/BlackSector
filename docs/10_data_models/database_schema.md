# Database Schema Specification

## Version: 0.4

## Status: Draft

## Owner: Core Architecture

## Last Updated: 2026-03-05

---

# 1. Purpose

Defines the persistent storage schema for the BlackSector server.

All game state that must survive a server restart is stored in SQLite. The schema is organized by domain and reflects the entity model defined in `entity_models.md`.

SQLite is chosen for its simplicity, zero-configuration deployment, and sufficient performance for a server supporting 50–100 concurrent players.

---

# 2. SQLite Runtime Configuration

The following PRAGMAs must be set on every database connection at open time, before any queries execute.

## 2.1 WAL Mode

```sql
PRAGMA journal_mode=WAL;
PRAGMA synchronous=NORMAL;
```

**Why WAL is required:**

The server runs a single-writer tick engine alongside multiple concurrent SSH sessions (readers). Without WAL, any active read transaction will block writes — the tick engine stalls waiting for readers to finish, and readers see `SQLITE_BUSY` errors during ticks.

WAL (Write-Ahead Log) allows:
- Concurrent readers without blocking the writer
- The tick engine to commit writes while SSH sessions are actively reading game state
- `synchronous=NORMAL` to reduce fsync overhead while remaining crash-safe

**Set on every connection open** — not once at startup. SQLite PRAGMAs are connection-scoped.

## 2.2 Other Required PRAGMAs

```sql
PRAGMA foreign_keys=ON;
PRAGMA busy_timeout=5000;
```

`foreign_keys=ON`: SQLite does not enforce FK constraints by default. Must be enabled per connection.

`busy_timeout=5000`: If a write lock is briefly held, retry for up to 5 seconds before returning `SQLITE_BUSY`. Prevents spurious errors during tick bursts.

---

# 3. Design Principles

* Schema reflects the entity model, not protocol messages
* All positions stored as two-dimensional (x, y) — no Z coordinate
* UUIDs used for player-facing identifiers; integer IDs for world entities
* All timestamps stored as Unix epoch integers
* All tick values stored as integers
* Booleans stored as INTEGER (0/1) for SQLite compatibility

---

# 3. Schema Conventions

```
INTEGER   — 64-bit signed integer
REAL      — 64-bit floating point
TEXT      — UTF-8 string
UUID      — TEXT in lowercase hyphenated format (e.g. "550e8400-e29b-41d4-a716-446655440000")
BOOL      — INTEGER (0 = false, 1 = true)
```

Primary keys are listed first. Foreign key relationships noted inline.

---

# 4. Player and Authentication Domain

## Table: players

```sql
CREATE TABLE players (
  player_id         TEXT PRIMARY KEY,
  player_name       TEXT NOT NULL UNIQUE,
  token_hash        TEXT NOT NULL,
  credits           INTEGER NOT NULL DEFAULT 0,
  created_at        INTEGER NOT NULL,
  last_login_at     INTEGER,
  is_banned         INTEGER NOT NULL DEFAULT 0
);
```

---

## Table: sessions

```sql
CREATE TABLE sessions (
  session_id        TEXT PRIMARY KEY,
  player_id         TEXT NOT NULL REFERENCES players(player_id),
  interface_mode    TEXT NOT NULL CHECK (interface_mode IN ('TEXT','GUI')),
  state             TEXT NOT NULL,
  connected_at      INTEGER NOT NULL,
  disconnected_at   INTEGER,
  linger_expiry_at  INTEGER,
  last_activity_at  INTEGER NOT NULL
);
```

State values: `CONNECTED` | `DISCONNECTED_LINGERING` | `DOCKED_OFFLINE` | `TERMINATED`

---

# 5. Ship Domain

## Table: ships

```sql
CREATE TABLE ships (
  ship_id             TEXT PRIMARY KEY,
  player_id           TEXT NOT NULL REFERENCES players(player_id),
  ship_class          TEXT NOT NULL,
  hull_points         INTEGER NOT NULL,
  max_hull_points     INTEGER NOT NULL,
  shield_points       INTEGER NOT NULL,
  max_shield_points   INTEGER NOT NULL,
  energy_points       INTEGER NOT NULL,
  max_energy_points   INTEGER NOT NULL,
  cargo_capacity      INTEGER NOT NULL,
  current_system_id   INTEGER REFERENCES systems(system_id),
  position_x          REAL NOT NULL DEFAULT 0.0,
  position_y          REAL NOT NULL DEFAULT 0.0,
  status              TEXT NOT NULL,
  docked_at_port_id   INTEGER REFERENCES ports(port_id),
  last_updated_tick   INTEGER NOT NULL
);
```

Status values: `DOCKED` | `IN_SPACE` | `IN_COMBAT` | `DESTROYED`

---

## Table: ship_cargo

```sql
CREATE TABLE ship_cargo (
  ship_id        TEXT NOT NULL REFERENCES ships(ship_id),
  slot_index     INTEGER NOT NULL,
  commodity_id   TEXT NOT NULL REFERENCES commodities(commodity_id),
  quantity       INTEGER NOT NULL,
  PRIMARY KEY (ship_id, slot_index)
);
```

---

## Table: ship_drone_inventory

Tracks undeployed drones in a ship's drone bays. Drones are not commodities and are not stored in `ship_cargo`.

Bay capacity is defined per ship class in `config/ships/ship_classes.json` (`drone_bay_capacity`). Total bay usage = `SUM(ship_drone_inventory.quantity)` + active deployed drone count (`drones` table WHERE `owner_id` matches AND `status != 'destroyed'`). This sum must not exceed `drone_bay_capacity`. Enforced at application layer.

```sql
CREATE TABLE ship_drone_inventory (
  ship_id      TEXT NOT NULL REFERENCES ships(ship_id),
  drone_type   TEXT NOT NULL,   -- mapping | prospecting | decoy | relay
  quantity     INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (ship_id, drone_type)
);
```

---

## Table: ship_upgrades

```sql
CREATE TABLE ship_upgrades (
  ship_id              TEXT NOT NULL REFERENCES ships(ship_id),
  upgrade_id           TEXT NOT NULL,
  installed_at_tick    INTEGER NOT NULL,
  PRIMARY KEY (ship_id, upgrade_id)
);
```

---

# 6. Universe Domain

## Table: regions

```sql
CREATE TABLE regions (
  region_id      INTEGER PRIMARY KEY,
  name           TEXT NOT NULL,
  region_type    TEXT NOT NULL,
  security_level REAL NOT NULL
);
```

Region types: `core` | `agricultural` | `industrial` | `frontier` | `black`

Security level: `0.7–1.0` = High Security, `0.4–0.7` = Medium Security, `0.0–0.4` = Low Security, `-1.0` = Black Sector

---

## Table: systems

```sql
CREATE TABLE systems (
  system_id      INTEGER PRIMARY KEY,
  name           TEXT NOT NULL,
  region_id      INTEGER NOT NULL REFERENCES regions(region_id),
  security_level REAL NOT NULL,
  position_x     REAL NOT NULL,
  position_y     REAL NOT NULL,
  hazard_level   REAL NOT NULL DEFAULT 0.0
);
```

---

## Table: jump_connections

```sql
CREATE TABLE jump_connections (
  connection_id        INTEGER PRIMARY KEY,
  from_system_id       INTEGER NOT NULL REFERENCES systems(system_id),
  to_system_id         INTEGER NOT NULL REFERENCES systems(system_id),
  bidirectional        INTEGER NOT NULL DEFAULT 1,
  fuel_cost_modifier   REAL NOT NULL DEFAULT 1.0
);
```

---

## Table: hazard_zones

```sql
CREATE TABLE hazard_zones (
  hazard_id        INTEGER PRIMARY KEY,
  system_id        INTEGER NOT NULL REFERENCES systems(system_id),
  hazard_type      TEXT NOT NULL,
  position_x       REAL NOT NULL,
  position_y       REAL NOT NULL,
  radius           REAL NOT NULL,
  damage_per_tick  INTEGER NOT NULL DEFAULT 0,
  active           INTEGER NOT NULL DEFAULT 1
);
```

Hazard types: `asteroid_field` | `radiation_belt` | `debris_cloud` | `gravitational_anomaly`

---

# 7. Port and Economy Domain

## Table: ports

```sql
CREATE TABLE ports (
  port_id        INTEGER PRIMARY KEY,
  system_id      INTEGER NOT NULL REFERENCES systems(system_id),
  name           TEXT NOT NULL,
  port_type      TEXT NOT NULL,
  security_level REAL NOT NULL,
  docking_fee             INTEGER NOT NULL DEFAULT 0,
  has_bank                INTEGER NOT NULL DEFAULT 0,
  interest_rate_percent   REAL NOT NULL DEFAULT 0.0
);
```

Port types: `trading` | `mining` | `refueling` | `black_market`

---

## Table: commodities

```sql
CREATE TABLE commodities (
  commodity_id   TEXT PRIMARY KEY,
  name           TEXT NOT NULL,
  category       TEXT NOT NULL,
  base_price     INTEGER NOT NULL,
  volatility     REAL NOT NULL,
  is_contraband  INTEGER NOT NULL DEFAULT 0
);
```

Categories: `essential` | `industrial` | `luxury` | `exotic`

---

## Table: port_inventory

```sql
CREATE TABLE port_inventory (
  port_id          INTEGER NOT NULL REFERENCES ports(port_id),
  commodity_id     TEXT NOT NULL REFERENCES commodities(commodity_id),
  quantity         INTEGER NOT NULL DEFAULT 0,
  buy_price        INTEGER NOT NULL,
  sell_price       INTEGER NOT NULL,
  updated_tick     INTEGER NOT NULL,
  PRIMARY KEY (port_id, commodity_id)
);
```

---

## Table: market_price_history

```sql
CREATE TABLE market_price_history (
  id             INTEGER PRIMARY KEY AUTOINCREMENT,
  port_id        INTEGER NOT NULL REFERENCES ports(port_id),
  commodity_id   TEXT NOT NULL REFERENCES commodities(commodity_id),
  tick           INTEGER NOT NULL,
  buy_price      INTEGER NOT NULL,
  sell_price     INTEGER NOT NULL,
  buy_volume     INTEGER NOT NULL DEFAULT 0,
  sell_volume    INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_market_history ON market_price_history (port_id, commodity_id, tick DESC);
```

---

# 8. Economic Events Domain

## Table: active_economic_events

```sql
CREATE TABLE active_economic_events (
  event_instance_id    INTEGER PRIMARY KEY,
  event_id             TEXT NOT NULL,
  scope_type           TEXT NOT NULL,
  affected_region_id   INTEGER REFERENCES regions(region_id),
  affected_system_id   INTEGER REFERENCES systems(system_id),
  start_tick           INTEGER NOT NULL,
  end_tick             INTEGER NOT NULL,
  visibility           TEXT NOT NULL
);
```

Scope types: `system` | `region` | `galaxy`
Visibility: `public` | `hidden`

---

## Table: economic_event_cooldowns

```sql
CREATE TABLE economic_event_cooldowns (
  event_id        TEXT NOT NULL,
  scope_key       TEXT NOT NULL,
  last_ended_tick INTEGER NOT NULL,
  PRIMARY KEY (event_id, scope_key)
);
```

`scope_key` format: `"region:42"` | `"system:17"` | `"galaxy"`

---

# 9. AI Trader Domain

## Table: ai_traders

```sql
CREATE TABLE ai_traders (
  trader_id                  INTEGER PRIMARY KEY,
  name                       TEXT NOT NULL,
  ship_class                 TEXT NOT NULL,
  current_system_id          INTEGER REFERENCES systems(system_id),
  status                     TEXT NOT NULL,
  home_region_id             INTEGER REFERENCES regions(region_id),
  current_cargo_commodity    TEXT REFERENCES commodities(commodity_id),
  current_cargo_quantity     INTEGER NOT NULL DEFAULT 0,
  last_trade_tick            INTEGER
);
```

Status values: `IDLE` | `TRAVELING` | `TRADING` | `DOCKED` | `DESTROYED`

---

# 10. Mission Domain

## Table: mission_instances

```sql
CREATE TABLE mission_instances (
  instance_id      TEXT PRIMARY KEY,
  mission_id       TEXT NOT NULL,
  player_id        TEXT NOT NULL REFERENCES players(player_id),
  status           TEXT NOT NULL,
  accepted_tick    INTEGER,
  started_tick     INTEGER,
  completed_tick   INTEGER,
  failed_reason    TEXT,
  expires_at_tick  INTEGER
);

CREATE INDEX idx_missions_player ON mission_instances (player_id, status);
```

Status values: `AVAILABLE` | `ACCEPTED` | `IN_PROGRESS` | `COMPLETED` | `FAILED` | `EXPIRED` | `ABANDONED`

---

## Table: objective_progress

```sql
CREATE TABLE objective_progress (
  instance_id      TEXT NOT NULL REFERENCES mission_instances(instance_id),
  objective_index  INTEGER NOT NULL,
  status           TEXT NOT NULL,
  current_value    INTEGER NOT NULL DEFAULT 0,
  required_value   INTEGER NOT NULL,
  PRIMARY KEY (instance_id, objective_index)
);
```

Status values: `PENDING` | `ACTIVE` | `COMPLETED` | `FAILED`

---

# 11. Exploration Domain

## Table: player_map_data

```sql
CREATE TABLE player_map_data (
  player_id          TEXT NOT NULL REFERENCES players(player_id),
  system_id          INTEGER NOT NULL REFERENCES systems(system_id),
  discovered_at_tick INTEGER NOT NULL,
  scan_level         INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (player_id, system_id)
);
```

Scan levels: `0` = visited, `1` = basic scan, `2` = deep scan

---

## Table: anomalies

```sql
CREATE TABLE anomalies (
  anomaly_id                INTEGER PRIMARY KEY,
  system_id                 INTEGER NOT NULL REFERENCES systems(system_id),
  anomaly_type              TEXT NOT NULL,
  position_x                REAL NOT NULL,
  position_y                REAL NOT NULL,
  is_discovered             INTEGER NOT NULL DEFAULT 0,
  discovered_by_player_id   TEXT REFERENCES players(player_id),
  resource_quantity         INTEGER NOT NULL DEFAULT 0,
  depleted                  INTEGER NOT NULL DEFAULT 0
);
```

---

# 12. Mining Domain

## Table: asteroid_fields

```sql
CREATE TABLE asteroid_fields (
  field_id          INTEGER PRIMARY KEY,
  system_id         INTEGER NOT NULL REFERENCES systems(system_id),
  field_type        TEXT NOT NULL,
  position_x        REAL NOT NULL,
  position_y        REAL NOT NULL,
  depletion_level   REAL NOT NULL DEFAULT 0.0,
  last_mined_tick   INTEGER
);
```

Field types: `common` | `rich` | `rare` | `depleted`

---

## Table: asteroid_field_resources

```sql
CREATE TABLE asteroid_field_resources (
  field_id            INTEGER NOT NULL REFERENCES asteroid_fields(field_id),
  commodity_id        TEXT NOT NULL REFERENCES commodities(commodity_id),
  base_yield          INTEGER NOT NULL,
  current_multiplier  REAL NOT NULL DEFAULT 1.0,
  PRIMARY KEY (field_id, commodity_id)
);
```

---

# 13. Navigation Domain

## Table: player_waypoints

```sql
CREATE TABLE player_waypoints (
  waypoint_id      INTEGER PRIMARY KEY,
  player_id        TEXT NOT NULL REFERENCES players(player_id),
  system_id        INTEGER NOT NULL REFERENCES systems(system_id),
  name             TEXT NOT NULL,
  position_x       REAL NOT NULL,
  position_y       REAL NOT NULL,
  created_at_tick  INTEGER NOT NULL
);
```

---

# 14. Communications Domain

## Table: messages

```sql
CREATE TABLE messages (
  message_id        TEXT PRIMARY KEY,
  message_type      TEXT NOT NULL,     -- proximity | system | irn_direct | irn_broadcast | distress | dead_drop
  sender_id         TEXT NOT NULL REFERENCES players(player_id),
  sender_name       TEXT NOT NULL,     -- denormalized for display after player deletion
  recipient_id      TEXT REFERENCES players(player_id),
  origin_system_id  INTEGER NOT NULL REFERENCES systems(system_id),
  origin_position_x REAL,
  origin_position_y REAL,
  content           TEXT NOT NULL,     -- max 500 characters
  sent_tick         INTEGER NOT NULL,
  deliver_at_tick   INTEGER NOT NULL,
  delivered         INTEGER NOT NULL DEFAULT 0,
  read_at_tick      INTEGER,
  expires_at_tick   INTEGER NOT NULL,
  dead_drop_port_id INTEGER REFERENCES ports(port_id),
  intercepted_by    TEXT REFERENCES players(player_id)
);

CREATE INDEX idx_messages_recipient ON messages (recipient_id, delivered, read_at_tick);
CREATE INDEX idx_messages_deliver   ON messages (deliver_at_tick, delivered);
CREATE INDEX idx_messages_dead_drop ON messages (dead_drop_port_id, recipient_id);
```

---

## Table: drones

```sql
CREATE TABLE drones (
  drone_id          TEXT PRIMARY KEY,
  drone_name        TEXT NOT NULL,
  drone_type        TEXT NOT NULL,     -- mapping | prospecting | decoy | relay
  owner_id          TEXT NOT NULL REFERENCES players(player_id),
  current_system_id INTEGER NOT NULL REFERENCES systems(system_id),
  position_x        REAL NOT NULL,
  position_y        REAL NOT NULL,
  hull_points       INTEGER NOT NULL,
  max_hull_points   INTEGER NOT NULL,
  energy_points     INTEGER NOT NULL,
  max_energy_points INTEGER NOT NULL,
  cargo_current     INTEGER NOT NULL DEFAULT 0,
  cargo_capacity    INTEGER NOT NULL DEFAULT 0,
  status            TEXT NOT NULL,     -- active | standby | returning | destroyed
  deployed_at_tick  INTEGER NOT NULL,
  last_command_tick INTEGER,
  last_report_tick  INTEGER
);

CREATE INDEX idx_drones_owner  ON drones (owner_id, status);
CREATE INDEX idx_drones_system ON drones (current_system_id);
```

---

## Table: drone_commands

```sql
CREATE TABLE drone_commands (
  command_id        TEXT PRIMARY KEY,
  drone_id          TEXT NOT NULL REFERENCES drones(drone_id),
  owner_id          TEXT NOT NULL REFERENCES players(player_id),
  command_type      TEXT NOT NULL,     -- move | scan | survey | prospect | sample | mimic | pattern | activate | deactivate | report | return | standby | recall
  parameters        TEXT,             -- JSON (e.g. target coordinates, pattern args)
  issued_tick       INTEGER NOT NULL,
  deliver_at_tick   INTEGER NOT NULL,
  delivered         INTEGER NOT NULL DEFAULT 0,
  executed          INTEGER NOT NULL DEFAULT 0,
  origin_system_id  INTEGER NOT NULL REFERENCES systems(system_id),
  exposure_applied  INTEGER NOT NULL DEFAULT 0,
  intercepted_by    TEXT REFERENCES players(player_id)
);

CREATE INDEX idx_drone_commands_deliver ON drone_commands (deliver_at_tick, delivered);
CREATE INDEX idx_drone_commands_drone   ON drone_commands (drone_id, executed);
```

---

## Table: drone_telemetry

```sql
CREATE TABLE drone_telemetry (
  report_id         TEXT PRIMARY KEY,
  drone_id          TEXT NOT NULL REFERENCES drones(drone_id),
  owner_id          TEXT NOT NULL REFERENCES players(player_id),
  report_type       TEXT NOT NULL,     -- status | scan_result | survey_update | prospect_result | mining_yield | alert | destroyed
  payload           TEXT NOT NULL,     -- JSON report data
  generated_tick    INTEGER NOT NULL,
  deliver_at_tick   INTEGER NOT NULL,
  delivered         INTEGER NOT NULL DEFAULT 0,
  origin_system_id  INTEGER NOT NULL REFERENCES systems(system_id)
);

CREATE INDEX idx_drone_telemetry_deliver ON drone_telemetry (deliver_at_tick, delivered);
CREATE INDEX idx_drone_telemetry_owner   ON drone_telemetry (owner_id, delivered);
```

---

# 15. Banking Domain

## Table: player_bank_accounts

```sql
CREATE TABLE player_bank_accounts (
  account_id          TEXT PRIMARY KEY,
  player_id           TEXT NOT NULL REFERENCES players(player_id),
  port_id             INTEGER NOT NULL REFERENCES ports(port_id),
  balance             INTEGER NOT NULL DEFAULT 0,
  opened_at_tick      INTEGER NOT NULL,
  last_interest_tick  INTEGER NOT NULL DEFAULT 0,
  UNIQUE (player_id, port_id)
);
```

---

## Table: bank_transactions

Audit log for all banking activity.

```sql
CREATE TABLE bank_transactions (
  transaction_id        TEXT PRIMARY KEY,
  player_id             TEXT NOT NULL REFERENCES players(player_id),
  port_id               INTEGER REFERENCES ports(port_id),
  transaction_type      TEXT NOT NULL,  -- deposit | withdraw | transfer_out | transfer_in | interest | send | receive
  amount                INTEGER NOT NULL,
  balance_after         INTEGER NOT NULL,
  counterparty_id       TEXT REFERENCES players(player_id),
  counterparty_port_id  INTEGER REFERENCES ports(port_id),
  tick                  INTEGER NOT NULL
);
```

---

# 16. Performance Indexes

```sql
CREATE INDEX idx_sessions_player    ON sessions (player_id);
CREATE INDEX idx_ships_player       ON ships (player_id);
CREATE INDEX idx_ships_system       ON ships (current_system_id);
CREATE INDEX idx_ports_system       ON ports (system_id);
CREATE INDEX idx_systems_region     ON systems (region_id);
CREATE INDEX idx_jumps_from         ON jump_connections (from_system_id);
CREATE INDEX idx_traders_system     ON ai_traders (current_system_id);
CREATE INDEX idx_map_player         ON player_map_data (player_id);
CREATE INDEX idx_events_end_tick    ON active_economic_events (end_tick);
CREATE INDEX idx_anomalies_system   ON anomalies (system_id);
CREATE INDEX idx_fields_system      ON asteroid_fields (system_id);
CREATE INDEX idx_waypoints_player   ON player_waypoints (player_id);
CREATE INDEX idx_bank_accounts_player ON player_bank_accounts (player_id);
CREATE INDEX idx_bank_tx_player     ON bank_transactions (player_id);
CREATE INDEX idx_bank_tx_tick       ON bank_transactions (tick);
```

---

# 17. Configuration Data (Not Stored in Database)

The following are loaded from JSON config files at startup and are NOT in the database:

| Config File                              | Contents                         |
| ---------------------------------------- | -------------------------------- |
| `config/economy/commodities.json`        | Commodity definitions            |
| `config/ships/ship_classes.json`         | Ship class definitions           |
| `config/ships/upgrades.json`             | Upgrade definitions              |
| `config/economy/economic_events.json`    | Economic event definitions       |
| `config/missions/*.json`                 | Mission scripts                  |
| `config/ai/trader_names.json`            | AI trader name lists             |
| `config/drones/drone_types.json`         | Drone class definitions (hull, energy, sensors, cost) |

These may be hot-reloaded via the admin CLI without a server restart.

---

# 18. Non-Goals (v1)

* Full-text search
* Replication or read replicas
* Sharding or partitioning
* Time-series analytics storage

---

# End of Document
