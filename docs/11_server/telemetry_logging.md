# Telemetry and Logging Specification

## Version: 0.1

## Status: Draft

## Owner: Core Architecture

## Last Updated: 2026-03-05

---

# 1. Purpose

Defines the logging and telemetry model for the BlackSector server.

This document covers:

* log levels and output channels
* structured server log format
* event log integration
* performance metrics
* operational health signals

The event log schema (game events such as trades, combat, missions) is defined in `event_log_schema.md`. This document covers server operational logging.

---

# 2. Design Principles

* Structured logs — all entries are machine-readable JSON
* Non-blocking — log writes never stall the tick loop
* Separated concerns — server operational log vs. game event log
* Appropriate verbosity — production runs at INFO level by default

---

# 3. Log Channels

The server writes to two separate log channels:

| Channel      | File                | Format  | Contents                              |
| ------------ | ------------------- | ------- | ------------------------------------- |
| Server log   | `logs/server.log`   | JSON    | Server operational events             |
| Event log    | `logs/events.log`   | NDJSON  | Game events (trades, combat, etc.)    |

Both files use one JSON object per line.

When running under systemd, `logs/server.log` is also captured by `journald`.

---

# 4. Log Levels

| Level    | When to Use                                         |
| -------- | --------------------------------------------------- |
| DEBUG    | Detailed trace data (development only, not default) |
| INFO     | Normal operational events                           |
| WARN     | Unusual but recoverable conditions                  |
| ERROR    | Failures that affect one player or operation        |
| CRITICAL | Failures that affect the server or all players      |

Default production level: `INFO`

Log level is configurable in `config/server.json`:

```json
{
  "log_level": "INFO"
}
```

---

# 5. Server Log Entry Format

All server log entries follow this structure:

```json
{
  "level": "INFO",
  "tick": 4821,
  "timestamp": 1709612345,
  "msg": "player connected",
  "player_id": "550e8400-e29b-41d4-a716-446655440000",
  "session_id": "abc123...",
  "interface_mode": "TEXT",
  "remote_addr": "192.168.1.1"
}
```

## Required Fields

| Field       | Type   | Description                               |
| ----------- | ------ | ----------------------------------------- |
| `level`     | string | Log level                                 |
| `tick`      | int    | Current server tick (0 before tick loop)  |
| `timestamp` | int    | Unix epoch seconds                        |
| `msg`       | string | Human-readable description                |

Additional context fields are added as needed per entry.

---

# 6. Standard Server Log Events

## 6.1 Startup and Shutdown

| Event              | Level | Description                            |
| ------------------ | ----- | -------------------------------------- |
| server_starting    | INFO  | Server process starting                |
| config_loaded      | INFO  | Configuration loaded successfully      |
| db_opened          | INFO  | SQLite database opened                 |
| world_loaded       | INFO  | World data loaded (N systems, N ports) |
| missions_loaded    | INFO  | Mission definitions loaded             |
| snapshot_loaded    | INFO  | Snapshot loaded (or fresh start)       |
| ssh_listener_ready | INFO  | SSH listener bound on port 2222        |
| tcp_listener_ready | INFO  | TLS/TCP listener bound on port 2223    |
| server_ready       | INFO  | Server ready, tick loop starting       |
| server_stopping    | INFO  | Shutdown initiated                     |
| server_stopped     | INFO  | Server stopped cleanly                 |

---

## 6.2 Session Events

| Event               | Level | Description                             |
| ------------------- | ----- | --------------------------------------- |
| session_connected   | INFO  | New player session established          |
| session_disconnected| INFO  | Player session disconnected             |
| session_auth_failed | WARN  | Authentication failure                  |
| session_kicked      | INFO  | Admin kicked a player                   |
| session_expired     | INFO  | Linger session expired                  |
| handshake_timeout   | WARN  | Client did not complete handshake       |
| rate_limited        | WARN  | Session exceeded command rate limit     |
| session_terminated  | INFO  | Session terminated due to violations    |

---

## 6.3 Tick Performance

| Event           | Level | Description                                   |
| --------------- | ----- | --------------------------------------------- |
| tick_complete   | DEBUG | Normal tick completed (debug level to avoid log spam) |
| tick_slow       | WARN  | Tick exceeded slow threshold                  |
| tick_very_slow  | ERROR | Tick exceeded 2x slow threshold               |

Example `tick_slow` entry:

```json
{
  "level": "WARN",
  "tick": 4821,
  "timestamp": 1709612345,
  "msg": "tick_slow",
  "duration_ms": 612,
  "threshold_ms": 500,
  "command_count": 22,
  "player_count": 14
}
```

---

## 6.4 Persistence

| Event              | Level | Description                              |
| ------------------ | ----- | ---------------------------------------- |
| snapshot_written   | INFO  | Snapshot saved successfully              |
| snapshot_failed    | ERROR | Snapshot write failed                    |
| db_flush_slow      | WARN  | SQLite flush exceeded threshold          |
| db_error           | ERROR | Database write failure                   |

---

## 6.5 Configuration

| Event              | Level | Description                              |
| ------------------ | ----- | ---------------------------------------- |
| config_reload      | INFO  | Hot-reload triggered                     |
| mission_load_error | ERROR | Mission file failed to load              |
| event_load_error   | ERROR | Economic event definition invalid        |

---

## 6.6 Critical Errors

| Event                    | Level    | Description                              |
| ------------------------ | -------- | ---------------------------------------- |
| db_connection_lost       | CRITICAL | SQLite connection lost                   |
| tick_loop_panic          | CRITICAL | Panic inside tick loop (recovered)       |
| subsystem_init_failed    | CRITICAL | Subsystem failed to initialize           |
| snapshot_load_failed     | CRITICAL | Could not load snapshot on startup       |

CRITICAL entries always trigger a server shutdown attempt.

---

# 7. Performance Metrics

The server tracks running metrics that are periodically logged and available via the `server status` admin command.

## Tick Metrics (logged every 100 ticks)

```json
{
  "level": "INFO",
  "tick": 5000,
  "timestamp": 1709612345,
  "msg": "tick_metrics",
  "period_ticks": 100,
  "avg_tick_ms": 48,
  "max_tick_ms": 112,
  "min_tick_ms": 31,
  "slow_ticks": 0
}
```

---

## Session Metrics (logged every 100 ticks)

```json
{
  "level": "INFO",
  "tick": 5000,
  "timestamp": 1709612345,
  "msg": "session_metrics",
  "connected_players": 14,
  "peak_players": 18,
  "total_commands_processed": 1842,
  "total_commands_rejected": 7
}
```

---

## Economic Metrics (logged every 500 ticks)

```json
{
  "level": "INFO",
  "tick": 5000,
  "timestamp": 1709612345,
  "msg": "economy_metrics",
  "active_economic_events": 2,
  "ai_traders_active": 61,
  "total_trades_period": 283,
  "total_trade_volume_credits": 1420500
}
```

---

# 8. Log Rotation

Logs are rotated based on either time or size:

* `events.log`: rotate daily or at 100MB
* `server.log`: rotate daily or at 50MB

Retain 30 days of rotated logs (configurable).

Configuration:

```json
{
  "log_rotation_max_size_mb": 100,
  "log_rotation_retain_days": 30
}
```

Log rotation is handled by the OS logrotate facility (see `deployment_model.md`). The server uses `copytruncate` mode — no server restart required.

---

# 9. Telemetry Summary

Key signals for operational monitoring:

| Signal                          | Normal Indicator           |
| ------------------------------- | -------------------------- |
| Tick duration                   | < 100ms average            |
| Slow tick frequency             | < 1 per 1000 ticks         |
| Session auth failures           | < 10 per minute            |
| Rate-limited events             | < 20 per minute            |
| DB flush duration               | < 50ms per tick            |
| Snapshot write success          | 100%                       |
| CRITICAL errors                 | 0                          |

Alert if any of the above thresholds are consistently exceeded.

---

# 10. Non-Goals (v1)

* Prometheus metrics endpoint
* Distributed tracing
* Log aggregation pipeline (ELK, Loki, etc.)
* Real-time telemetry dashboard
* Alerting integration

These may be added in a future operations tooling phase.

---

# End of Document
