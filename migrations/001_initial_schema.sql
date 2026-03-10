-- BlackSector Initial Database Schema
-- Migration: 001
-- Description: Creates all tables for the BlackSector game server
-- Version: 0.4

-- ============================================================================
-- 1. Player and Authentication Domain
-- ============================================================================

CREATE TABLE players (
  player_id         TEXT PRIMARY KEY,
  player_name       TEXT NOT NULL UNIQUE,
  token_hash        TEXT NOT NULL,
  credits           INTEGER NOT NULL DEFAULT 0,
  created_at        INTEGER NOT NULL,
  last_login_at     INTEGER,
  is_banned         INTEGER NOT NULL DEFAULT 0
);

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

-- ============================================================================
-- 2. Ship Domain
-- ============================================================================

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
  missiles_current    INTEGER NOT NULL DEFAULT 0,
  current_system_id   INTEGER REFERENCES systems(system_id),
  position_x          REAL NOT NULL DEFAULT 0.0,
  position_y          REAL NOT NULL DEFAULT 0.0,
  status              TEXT NOT NULL,
  docked_at_port_id   INTEGER REFERENCES ports(port_id),
  last_updated_tick   INTEGER NOT NULL
);

CREATE TABLE ship_cargo (
  ship_id        TEXT NOT NULL REFERENCES ships(ship_id),
  slot_index     INTEGER NOT NULL,
  commodity_id   TEXT NOT NULL REFERENCES commodities(commodity_id),
  quantity       INTEGER NOT NULL,
  PRIMARY KEY (ship_id, slot_index)
);

CREATE TABLE ship_drone_inventory (
  ship_id      TEXT NOT NULL REFERENCES ships(ship_id),
  drone_type   TEXT NOT NULL,
  quantity     INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (ship_id, drone_type)
);

CREATE TABLE ship_upgrades (
  ship_id              TEXT NOT NULL REFERENCES ships(ship_id),
  upgrade_id           TEXT NOT NULL,
  installed_at_tick    INTEGER NOT NULL,
  PRIMARY KEY (ship_id, upgrade_id)
);

-- ============================================================================
-- 3. Universe Domain
-- ============================================================================

CREATE TABLE regions (
  region_id      INTEGER PRIMARY KEY,
  name           TEXT NOT NULL,
  region_type    TEXT NOT NULL,
  security_level REAL NOT NULL
);

CREATE TABLE systems (
  system_id      INTEGER PRIMARY KEY,
  name           TEXT NOT NULL,
  region_id      INTEGER NOT NULL REFERENCES regions(region_id),
  security_level REAL NOT NULL,
  position_x     REAL NOT NULL,
  position_y     REAL NOT NULL,
  hazard_level   REAL NOT NULL DEFAULT 0.0
);

CREATE TABLE jump_connections (
  connection_id        INTEGER PRIMARY KEY,
  from_system_id       INTEGER NOT NULL REFERENCES systems(system_id),
  to_system_id         INTEGER NOT NULL REFERENCES systems(system_id),
  bidirectional        INTEGER NOT NULL DEFAULT 1,
  fuel_cost_modifier   REAL NOT NULL DEFAULT 1.0
);

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

-- ============================================================================
-- 4. Port and Economy Domain
-- ============================================================================

CREATE TABLE ports (
  port_id        INTEGER PRIMARY KEY,
  system_id      INTEGER NOT NULL REFERENCES systems(system_id),
  name           TEXT NOT NULL,
  port_type      TEXT NOT NULL,
  security_level REAL NOT NULL,
  docking_fee             INTEGER NOT NULL DEFAULT 0,
  has_bank                INTEGER NOT NULL DEFAULT 0,
  interest_rate_percent   REAL NOT NULL DEFAULT 0.0,
  has_shipyard            INTEGER NOT NULL DEFAULT 0,
  has_upgrade_market      INTEGER NOT NULL DEFAULT 0,
  has_drone_market        INTEGER NOT NULL DEFAULT 0,
  has_missile_supply      INTEGER NOT NULL DEFAULT 0,
  has_repair              INTEGER NOT NULL DEFAULT 1,
  has_fuel                INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE commodities (
  commodity_id   TEXT PRIMARY KEY,
  name           TEXT NOT NULL,
  category       TEXT NOT NULL,
  base_price     INTEGER NOT NULL,
  volatility     REAL NOT NULL,
  is_contraband  INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE port_inventory (
  port_id          INTEGER NOT NULL REFERENCES ports(port_id),
  commodity_id     TEXT NOT NULL REFERENCES commodities(commodity_id),
  quantity         INTEGER NOT NULL DEFAULT 0,
  buy_price        INTEGER NOT NULL,
  sell_price       INTEGER NOT NULL,
  updated_tick     INTEGER NOT NULL,
  PRIMARY KEY (port_id, commodity_id)
);

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

-- ============================================================================
-- 5. Economic Events Domain
-- ============================================================================

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

CREATE TABLE economic_event_cooldowns (
  event_id        TEXT NOT NULL,
  scope_key       TEXT NOT NULL,
  last_ended_tick INTEGER NOT NULL,
  PRIMARY KEY (event_id, scope_key)
);

-- ============================================================================
-- 6. AI Trader Domain
-- ============================================================================

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

-- ============================================================================
-- 7. Mission Domain
-- ============================================================================

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

CREATE TABLE objective_progress (
  instance_id      TEXT NOT NULL REFERENCES mission_instances(instance_id),
  objective_index  INTEGER NOT NULL,
  status           TEXT NOT NULL,
  current_value    INTEGER NOT NULL DEFAULT 0,
  required_value   INTEGER NOT NULL,
  PRIMARY KEY (instance_id, objective_index)
);

-- ============================================================================
-- 8. Exploration Domain
-- ============================================================================

CREATE TABLE player_map_data (
  player_id          TEXT NOT NULL REFERENCES players(player_id),
  system_id          INTEGER NOT NULL REFERENCES systems(system_id),
  discovered_at_tick INTEGER NOT NULL,
  scan_level         INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (player_id, system_id)
);

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

-- ============================================================================
-- 9. Mining Domain
-- ============================================================================

CREATE TABLE asteroid_fields (
  field_id          INTEGER PRIMARY KEY,
  system_id         INTEGER NOT NULL REFERENCES systems(system_id),
  field_type        TEXT NOT NULL,
  position_x        REAL NOT NULL,
  position_y        REAL NOT NULL,
  depletion_level   REAL NOT NULL DEFAULT 0.0,
  last_mined_tick   INTEGER
);

CREATE TABLE asteroid_field_resources (
  field_id            INTEGER NOT NULL REFERENCES asteroid_fields(field_id),
  commodity_id        TEXT NOT NULL REFERENCES commodities(commodity_id),
  base_yield          INTEGER NOT NULL,
  current_multiplier  REAL NOT NULL DEFAULT 1.0,
  PRIMARY KEY (field_id, commodity_id)
);

-- ============================================================================
-- 10. Navigation Domain
-- ============================================================================

CREATE TABLE player_waypoints (
  waypoint_id      INTEGER PRIMARY KEY,
  player_id        TEXT NOT NULL REFERENCES players(player_id),
  system_id        INTEGER NOT NULL REFERENCES systems(system_id),
  name             TEXT NOT NULL,
  position_x       REAL NOT NULL,
  position_y       REAL NOT NULL,
  created_at_tick  INTEGER NOT NULL
);

-- ============================================================================
-- 11. Communications Domain
-- ============================================================================

CREATE TABLE messages (
  message_id        TEXT PRIMARY KEY,
  message_type      TEXT NOT NULL,
  sender_id         TEXT NOT NULL REFERENCES players(player_id),
  sender_name       TEXT NOT NULL,
  recipient_id      TEXT REFERENCES players(player_id),
  origin_system_id  INTEGER NOT NULL REFERENCES systems(system_id),
  origin_position_x REAL,
  origin_position_y REAL,
  content           TEXT NOT NULL,
  sent_tick         INTEGER NOT NULL,
  deliver_at_tick   INTEGER NOT NULL,
  delivered         INTEGER NOT NULL DEFAULT 0,
  read_at_tick      INTEGER,
  expires_at_tick   INTEGER NOT NULL,
  dead_drop_port_id INTEGER REFERENCES ports(port_id),
  intercepted_by    TEXT REFERENCES players(player_id)
);

CREATE TABLE drones (
  drone_id          TEXT PRIMARY KEY,
  drone_name        TEXT NOT NULL,
  drone_type        TEXT NOT NULL,
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
  status            TEXT NOT NULL,
  deployed_at_tick  INTEGER NOT NULL,
  last_command_tick INTEGER,
  last_report_tick  INTEGER
);

CREATE TABLE drone_commands (
  command_id        TEXT PRIMARY KEY,
  drone_id          TEXT NOT NULL REFERENCES drones(drone_id),
  owner_id          TEXT NOT NULL REFERENCES players(player_id),
  command_type      TEXT NOT NULL,
  parameters        TEXT,
  issued_tick       INTEGER NOT NULL,
  deliver_at_tick   INTEGER NOT NULL,
  delivered         INTEGER NOT NULL DEFAULT 0,
  executed          INTEGER NOT NULL DEFAULT 0,
  origin_system_id  INTEGER NOT NULL REFERENCES systems(system_id),
  exposure_applied  INTEGER NOT NULL DEFAULT 0,
  intercepted_by    TEXT REFERENCES players(player_id)
);

CREATE TABLE drone_telemetry (
  report_id         TEXT PRIMARY KEY,
  drone_id          TEXT NOT NULL REFERENCES drones(drone_id),
  owner_id          TEXT NOT NULL REFERENCES players(player_id),
  report_type       TEXT NOT NULL,
  payload           TEXT NOT NULL,
  generated_tick    INTEGER NOT NULL,
  deliver_at_tick   INTEGER NOT NULL,
  delivered         INTEGER NOT NULL DEFAULT 0,
  origin_system_id  INTEGER NOT NULL REFERENCES systems(system_id)
);

-- ============================================================================
-- 12. Banking Domain
-- ============================================================================

CREATE TABLE player_bank_accounts (
  account_id          TEXT PRIMARY KEY,
  player_id           TEXT NOT NULL REFERENCES players(player_id),
  port_id             INTEGER NOT NULL REFERENCES ports(port_id),
  balance             INTEGER NOT NULL DEFAULT 0,
  opened_at_tick      INTEGER NOT NULL,
  last_interest_tick  INTEGER NOT NULL DEFAULT 0,
  UNIQUE (player_id, port_id)
);

CREATE TABLE bank_transactions (
  transaction_id        TEXT PRIMARY KEY,
  player_id             TEXT NOT NULL REFERENCES players(player_id),
  port_id               INTEGER REFERENCES ports(port_id),
  transaction_type      TEXT NOT NULL,
  amount                INTEGER NOT NULL,
  balance_after         INTEGER NOT NULL,
  counterparty_id       TEXT REFERENCES players(player_id),
  counterparty_port_id  INTEGER REFERENCES ports(port_id),
  tick                  INTEGER NOT NULL
);

-- ============================================================================
-- 13. Performance Indexes
-- ============================================================================

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
CREATE INDEX idx_missions_player    ON mission_instances (player_id, status);
CREATE INDEX idx_market_history     ON market_price_history (port_id, commodity_id, tick DESC);
CREATE INDEX idx_messages_recipient ON messages (recipient_id, delivered, read_at_tick);
CREATE INDEX idx_messages_deliver   ON messages (deliver_at_tick, delivered);
CREATE INDEX idx_messages_dead_drop ON messages (dead_drop_port_id, recipient_id);
CREATE INDEX idx_drones_owner       ON drones (owner_id, status);
CREATE INDEX idx_drones_system      ON drones (current_system_id);
CREATE INDEX idx_drone_commands_deliver ON drone_commands (deliver_at_tick, delivered);
CREATE INDEX idx_drone_commands_drone   ON drone_commands (drone_id, executed);
CREATE INDEX idx_drone_telemetry_deliver ON drone_telemetry (deliver_at_tick, delivered);
CREATE INDEX idx_drone_telemetry_owner   ON drone_telemetry (owner_id, delivered);
