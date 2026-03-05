# Rate Limiting Specification

## Version: 0.2

## Status: Draft

## Owner: Core Architecture

## Last Updated: 2026-03-05

---

# 1. Purpose

Protects the server from abuse, command flooding, and malformed message attacks.

Rate limits apply per connection. Since one player may have only one active connection at a time, this is effectively per-player limiting. See `session_multiplexing.md`.

---

# 2. Limits

| Limit                          | Default Value     | Scope         |
| ------------------------------ | ----------------- | ------------- |
| Max commands per tick          | 3                 | Per connection |
| Max malformed messages per minute | 10             | Per connection |
| Max handshake attempts per minute | 5              | Per IP address |
| Max bytes per message          | 4096 bytes        | Per message    |

These values are configurable in server configuration.

---

# 3. Command Rate Limit

A maximum of 3 commands per tick will be accepted from a single connection.

Commands submitted beyond this limit within a single tick are rejected immediately with:

```json
{
  "type": "error_rate_limited",
  "timestamp": <tick>,
  "correlation_id": "<uuid>",
  "payload": {
    "reason": "Command limit exceeded for this tick",
    "retry_after_tick": <next_tick>
  }
}
```

Queued commands that arrive early in a tick window are processed in submission order.

---

# 4. Malformed Message Limit

If a connection sends more than 10 malformed or unparseable messages within 60 seconds:

* Session is terminated
* Connection is closed
* IP address is temporarily blocked for a configurable cooldown period (default: 60 seconds)

A malformed message is defined as any message that:

* cannot be parsed as valid JSON
* is missing required envelope fields (`type`, `timestamp`, `payload`)
* exceeds the maximum message size

---

# 5. Handshake Flood Protection

More than 5 connection attempts from the same IP address within 60 seconds triggers a temporary block.

This applies to both SSH and TCP ports independently.

---

# 6. Enforcement

Rate limit enforcement occurs at the connection layer, before commands enter the tick queue.

Enforcement sequence:

1. Message received
2. Check message size — reject if exceeds limit
3. Parse JSON — increment malformed counter if invalid
4. Check command rate for current tick — reject if exceeded
5. Accept message into queue

---

# 7. Escalating Response

| Violation                         | Response                              |
| --------------------------------- | ------------------------------------- |
| Command limit exceeded this tick  | error_rate_limited, retry next tick   |
| Malformed message (under limit)   | error_invalid_command, session continues |
| Malformed message (limit reached) | Session terminated, IP cooldown       |
| Handshake flood (IP)              | IP temporarily blocked                |

Repeated session terminations from the same IP may result in longer block durations.

---

# 8. Non-Goals (v1)

* per-command-type rate limits
* per-player-reputation rate limit adjustments
* global server-wide command throttling

---

# End of Document
