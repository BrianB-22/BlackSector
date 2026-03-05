# Command Queue Specification

## Version: 0.1

## Status: Draft

## Owner: Core Architecture

## Last Updated: 2026-03-05

---

# 1. Purpose

Defines the command queue model used to accept and process player commands within the tick-based simulation.

Player commands arrive asynchronously from network sessions. The command queue bridges asynchronous input from player connections to synchronous execution inside the tick loop.

---

# 2. Design Principles

* Command enqueue is non-blocking — session goroutines never wait on tick logic
* Command execution is synchronous within the tick loop
* Execution order within a tick is deterministic
* Commands that exceed per-session limits are rejected immediately
* Malformed commands are rejected without entering the queue

---

# 3. Queue Architecture

```
Session Goroutine (SSH / TCP)
        │
        │  command_submit message received
        │
        ▼
  ┌─────────────────────┐
  │   Per-Session Queue  │  (goroutine-safe, bounded)
  └─────────┬───────────┘
            │
            │  dequeued at tick start
            ▼
  ┌──────────────────────┐
  │   Tick Command Batch │  (all commands for this tick)
  └──────────┬───────────┘
             │
             ▼
       Tick Execution
```

Each active session has its own bounded command queue. At the start of each tick, the tick engine drains all per-session queues.

---

# 4. Command Envelope

Commands arrive wrapped in the standard protocol envelope:

```json
{
  "type": "command_submit",
  "timestamp": 1709612345,
  "correlation_id": "<uuid>",
  "payload": {
    "command": "navigate",
    "parameters": {
      "destination_system_id": 42
    }
  }
}
```

## Validation Before Enqueue

Before a command is accepted into the queue, the session layer validates:

* message envelope is well-formed JSON
* `type` is `command_submit`
* `correlation_id` is present and is a valid UUID
* `payload.command` is a non-empty string
* `payload.parameters` is an object (may be empty)

Validation failures result in an immediate `error_invalid_command` response. The command is not enqueued.

---

# 5. Per-Session Queue Limits

| Limit                          | Default | Configurable |
| ------------------------------ | ------- | ------------ |
| Max commands per tick          | 3       | Yes          |
| Max queue depth (pending)      | 6       | Yes          |
| Max command payload size       | 4096 B  | Yes          |

## Per-Tick Limit

If a session submits more than `max_commands_per_tick` commands in one tick window, excess commands are rejected with `error_rate_limited`. Commands already in the queue for the current tick are not affected.

## Queue Depth Limit

If the queue contains `max_queue_depth` commands already waiting (e.g. submitted faster than ticks can process them), new submissions are rejected with `error_rate_limited`. This prevents queue buildup during slow ticks.

---

# 6. Execution Order

Within a tick, commands are executed in the following order:

1. Admin commands (from admin interface)
2. Player commands, ordered by session age (oldest session first)
3. Within a session, commands are executed in FIFO order

This ordering is deterministic given the same connection state.

---

# 7. Command Execution

During tick step 1, the tick engine:

1. Drains all per-session queues into the tick command batch
2. Sorts by execution order rules above
3. Executes each command via the command dispatcher

The command dispatcher routes commands to the appropriate subsystem handler:

| Command Category | Handler      |
| ---------------- | ------------ |
| `navigate`       | Navigation   |
| `jump`           | Navigation   |
| `attack`         | Combat       |
| `flee`           | Combat       |
| `mine`           | Mining       |
| `scan`           | Exploration  |
| `dock`           | Port/Economy |
| `buy`            | Economy      |
| `sell`           | Economy      |
| `accept_mission` | Missions     |
| `abandon_mission`| Missions     |

Unknown commands are rejected with `error_invalid_command`.

---

# 8. Command Response

Each command receives a response sent back to the originating session.

## Accept

Command was valid and queued for execution this tick:

```json
{
  "type": "command_accept",
  "timestamp": 4821,
  "correlation_id": "<uuid>",
  "payload": { "queued": true }
}
```

## Reject

Command was invalid, rate-limited, or the player cannot perform the action:

```json
{
  "type": "command_reject",
  "timestamp": 4821,
  "correlation_id": "<uuid>",
  "payload": { "reason": "Insufficient energy" }
}
```

Responses are sent by the end of the tick in which the command was processed.

---

# 9. Admin Commands

Admin commands submitted via stdin bypass the per-session queue and are executed with priority at the start of the tick.

Admin commands do not have a `correlation_id` — responses are written to the admin stdout.

Admin commands are never rate-limited. However, only one admin command per tick is processed to avoid tick stall.

---

# 10. Command Cancellation

Commands already in the queue cannot be cancelled by the client.

If a player disconnects before their queued commands are executed:

* Commands in the queue are still executed that tick
* Results are discarded (no session to receive them)

---

# 11. Security

* Commands are only accepted from authenticated sessions
* Sessions must complete the full handshake before any `command_submit` is accepted
* Command parameters are validated by the relevant subsystem handler before execution
* Parameters cannot affect server state other than through the defined command interface

---

# 12. Performance Constraints

* Enqueue operation: O(1), non-blocking
* Tick dequeue: O(sessions × max_commands_per_tick)
* Total command processing per tick: < 50ms for 100 concurrent players

---

# 13. Non-Goals (v1)

* Command prioritization by player
* Command cancellation
* Deferred or scheduled commands
* Cross-player command dependencies

---

# End of Document
