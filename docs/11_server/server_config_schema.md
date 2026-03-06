# Server Configuration Schema

## Version: 1.0
## Status: Draft
## Owner: Core Architecture
## Last Updated: 2026-03-05

---

# 1. Purpose

Documents every key in `config/server.json`. This is the canonical reference for all server configuration.

Default values are used if a key is omitted. The server validates the config on startup and exits with an error if required keys are missing or values are out of range.

The default config file is at `config/server.json`. See `docs/00_overview/TECH_STACK.md` for library references.

---

# 2. Full Schema

```json
{
  "server": {
    "ssh_port": 2222,
    "tick_interval_ms": 2000,
    "snapshot_interval_ticks": 100,
    "max_concurrent_players": 50,
    "db_path": "blacksector.db",
    "world_config_path": "config/world/alpha_sector.json",
    "missions_config_dir": "config/missions/"
  },
  "logging": {
    "level": "info",
    "log_file": "server.log",
    "debug_log_enabled": false,
    "debug_log_path": "debug.log",
    "debug_log_include_sql": false,
    "debug_log_include_tick_detail": true
  },
  "universe": {
    "universe_seed": 1701,
    "pirate_spawn_interval_ticks": 10,
    "pirate_activity_base": 0.30,
    "surrender_loss_percent": 40
  },
  "combat": {
    "courier_accuracy": 0.65,
    "shield_regen_per_tick": 5,
    "hull_repair_cost_per_point": 10,
    "flee_base_chance": 0.35,
    "flee_energy_bonus": 0.25,
    "missile_unit_price": 200
  },
  "economy": {
    "low_sec_price_multiplier": 1.18
  },
  "banking": {
    "interest_apply_interval_ticks": 500
  },
  "player": {
    "starting_credits": 1000,
    "starting_ship_class": "courier",
    "starting_system_id": "nexus_prime"
  },
  "irn": {
    "base_delay_ticks": 5,
    "max_delay_ticks": 20,
    "broadcast_rate_limit_per_hour": 2
  }
}
```

---

# 3. Key Reference

## 3.1 server

| Key                       | Type    | Default                            | Description                                                    |
| ------------------------- | ------- | ---------------------------------- | -------------------------------------------------------------- |
| `ssh_port`                | int     | 2222                               | TCP port for player SSH connections                            |
| `tick_interval_ms`        | int     | 2000                               | Milliseconds between server ticks                              |
| `snapshot_interval_ticks` | int     | 100                                | Save full world snapshot every N ticks                         |
| `max_concurrent_players`  | int     | 50                                 | Hard cap on simultaneous SSH sessions                          |
| `db_path`                 | string  | "blacksector.db"                   | Path to SQLite database file (relative to server binary)       |
| `world_config_path`       | string  | "config/world/alpha_sector.json"   | Path to world definition JSON                                  |
| `missions_config_dir`     | string  | "config/missions/"                 | Directory scanned for mission JSON files at startup            |

## 3.2 logging

| Key                           | Type    | Default       | Description                                                           |
| ----------------------------- | ------- | ------------- | --------------------------------------------------------------------- |
| `level`                       | string  | "info"        | Log level for server.log: "debug", "info", "warn", "error"           |
| `log_file`                    | string  | "server.log"  | Path for structured server log output                                 |
| `debug_log_enabled`           | bool    | false         | Enable verbose debug.log output                                       |
| `debug_log_path`              | string  | "debug.log"   | Path for debug log file                                               |
| `debug_log_include_sql`       | bool    | false         | Include every SQL statement in debug log (high volume)                |
| `debug_log_include_tick_detail` | bool  | true          | Log per-tick event detail when debug is enabled                       |

When `debug_log_enabled = true`, the debug log captures:
- Every tick start/end with duration (ms)
- Every player command received (player_id, command text, tick number)
- Every state machine transition (ship status, combat state)
- Every combat roll (hit check value, threshold, outcome; damage rolled, applied)
- Every mission event (accepted, progress, completed, abandoned)
- Every bank transaction
- Every DB write if `debug_log_include_sql = true`
- Every session connect and disconnect

This level of logging is designed to give an AI assistant (or human developer) full replay capability for debugging.

## 3.3 universe

| Key                          | Type   | Default | Description                                                         |
| ---------------------------- | ------ | ------- | ------------------------------------------------------------------- |
| `universe_seed`              | int64  | 1701    | Master PRNG seed. Controls all deterministic random outcomes        |
| `pirate_spawn_interval_ticks`| int    | 10      | Attempt pirate spawn every N ticks in eligible Low Sec systems      |
| `pirate_activity_base`       | float  | 0.30    | Base multiplier in pirate spawn formula                             |
| `surrender_loss_percent`     | int    | 40      | Percent of wallet credits lost on surrender                         |

Pirate spawn formula: `PirateSpawnChance = pirate_activity_base × (1 − system.security_rating)`

## 3.4 combat

| Key                          | Type   | Default | Description                                                 |
| ---------------------------- | ------ | ------- | ----------------------------------------------------------- |
| `courier_accuracy`           | float  | 0.65    | Base hit chance for courier class                           |
| `shield_regen_per_tick`      | int    | 5       | Shield points recovered per tick when not IN_COMBAT         |
| `hull_repair_cost_per_point` | int    | 10      | Credits per hull point repaired at port                     |
| `flee_base_chance`           | float  | 0.35    | Base flee success probability with no energy                |
| `flee_energy_bonus`          | float  | 0.25    | Max additional flee chance at full energy                   |
| `missile_unit_price`         | int    | 200     | Credits per missile at port                                 |

## 3.5 economy

| Key                       | Type   | Default | Description                                                        |
| ------------------------- | ------ | ------- | ------------------------------------------------------------------ |
| `low_sec_price_multiplier`| float  | 1.18    | Price multiplier applied to all commodities in Low Security space  |

## 3.6 banking

| Key                             | Type | Default | Description                                        |
| ------------------------------- | ---- | ------- | -------------------------------------------------- |
| `interest_apply_interval_ticks` | int  | 500     | Apply bank interest to all accounts every N ticks  |

500 ticks × 2s/tick = 1000 seconds ≈ ~17 minutes in-game "interest period."

## 3.7 player

| Key                    | Type   | Default       | Description                                          |
| ---------------------- | ------ | ------------- | ---------------------------------------------------- |
| `starting_credits`     | int    | 1000          | Credits given to new players on registration         |
| `starting_ship_class`  | string | "courier"     | Ship class assigned at registration                  |
| `starting_system_id`   | string | "nexus_prime" | System where new players spawn                       |

## 3.8 irn

| Key                            | Type | Default | Description                                                  |
| ------------------------------ | ---- | ------- | ------------------------------------------------------------ |
| `base_delay_ticks`             | int  | 5       | Minimum IRN message delivery delay                           |
| `max_delay_ticks`              | int  | 20      | Maximum IRN message delivery delay                           |
| `broadcast_rate_limit_per_hour`| int  | 2       | Max IRN broadcasts per player per in-game hour               |

---

# 4. Required Keys

The following keys have no safe default and are **required** in `server.json`. The server exits on startup if they are missing:

- `server.db_path`
- `server.world_config_path`
- `universe.universe_seed`

---

# End of Document
