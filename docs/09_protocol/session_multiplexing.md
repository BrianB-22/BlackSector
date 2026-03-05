# Session Multiplexing Specification

## Version: 0.1

## Status: Draft

## Owner: Core Architecture

## Last Updated: 2026-03-05

---

# 1. Purpose

Defines the relationship between players, sessions, and connections in Black Sector.

This document clarifies the terminology used across protocol documents and establishes the rules governing session lifecycle, conflict handling, and reconnection behavior.

---

# 2. Terminology

**Player**
A registered account identity. Identified by `player_id`. Persistent across sessions.

**Session**
A single active gameplay context for a player. One player may have at most one active session at a time. Identified by `session_id`. Created on successful handshake, destroyed on disconnect or timeout.

**Connection**
The underlying network link between a client and the server. A connection may be SSH (TEXT mode) or TCP (GUI mode), determined by which port the client connects to. A session is bound to exactly one connection.

**Interface Mode**
Either `TEXT` (SSH port) or `GUI` (TCP port). Determined by connection port. Not negotiable.

---

# 3. One Session Per Player

A player may have only one active session at a time.

This applies regardless of interface mode. A player cannot hold a TEXT session and a GUI session simultaneously.

If a player connects while an existing session is active, the server rejects the new connection with `session_already_active`. The player must terminate the existing session before a new one can be established.

This rule simplifies server state management and prevents conflicting game actions from two simultaneous clients.

---

# 4. Session Lifecycle

```
[Player connects]
        ↓
[Handshake completes successfully]
        ↓
Session created → state: ACTIVE
        ↓
[Player sends commands, receives events]
        ↓
[Player disconnects cleanly OR connection drops]
        ↓
Session state → DISCONNECTED
        ↓
[Linger period per session_management.md]
        ↓
Session destroyed
```

Session IDs are not reused. Each new connection generates a new session_id.

---

# 5. Session State

| State           | Description                                              |
| --------------- | -------------------------------------------------------- |
| ACTIVE          | Player connected and interacting                         |
| DISCONNECTED    | Connection dropped; linger period active (ship in space) |
| TERMINATED      | Session cleanly ended; player safely docked              |

For ship persistence behavior during DISCONNECTED state, see `session_management.md`.

---

# 6. Reconnection

When a player reconnects after a disconnect:

1. New connection established (same or different port)
2. Handshake completes with same `player_id`
3. Server checks for existing session in DISCONNECTED state
4. If found and linger period has not expired: new session resumes game state
5. New `session_id` is issued; old session_id is invalidated
6. Server sends current state snapshot to new client

The player receives fresh state on reconnect regardless of interface mode. A player may reconnect via TEXT after having used GUI, or vice versa.

---

# 7. Switching Interface Mode

A player may switch between TEXT (SSH) and GUI (TCP) across separate sessions — not within the same session.

Workflow:

1. Player disconnects from SSH session (TEXT mode)
2. Player's ship enters linger state per session_management.md
3. Player connects via TCP port (GUI mode)
4. New handshake completes; new session_id issued
5. Session resumes with current game state

There is no in-session mode switching. The interface mode is fixed for the lifetime of a connection.

---

# 8. Rate Limiting Scope

Rate limits are applied per connection. Since one player = one active connection at a time, rate limits are effectively per player.

See `rate_limiting.md` for specific limits.

---

# 9. Event Broadcasting

All game events for a player are delivered to their single active session.

There is no fanout — events are sent to exactly one connection. If the connection is dropped, events are discarded until the player reconnects and requests state sync.

See `state_sync.md` for reconnect state recovery.

---

# 10. Security Considerations

* session_id is server-generated and must not be predictable
* clients cannot supply their own session_id
* a valid player_token in handshake_response is required to create a session
* expired or revoked tokens must be rejected at handshake

---

# 11. Non-Goals (v1)

* simultaneous TEXT + GUI sessions for the same player
* session transfer without disconnect
* session sharing between players
* spectator sessions

---

# 12. Future Extensions

* session transfer between ports without disconnect (reconnect without ship linger)
* spectator session type (read-only, no commands)
* admin observer sessions

---

# End of Document
