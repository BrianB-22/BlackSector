# Handshake Protocol Specification

## Version: 0.2

## Status: Draft

## Owner: Core Architecture

## Last Updated: 2026-03-05

---

# 1. Purpose

Defines the full connection negotiation process between a client and the server.

The handshake establishes:

* player identity and authentication
* session creation
* interface mode (determined by connection port, not negotiated)
* tick interval acknowledgement

This document supersedes handshake_flow.md.

---

# 2. Dual-Port Architecture

The server listens on two distinct ports:

| Port | Default | Interface Mode | Transport  | Client Type           |
| ---- | ------- | -------------- | ---------- | --------------------- |
| SSH  | 2222    | TEXT           | SSH        | Terminal / any SSH client |
| TCP  | 2223    | GUI            | Raw TCP    | GUI application       |

**Interface mode is determined entirely by which port the client connects to.**

There is no interface_mode negotiation during the handshake. The server knows the mode the moment a connection arrives.

Port assignments are configurable in server configuration. The defaults above are illustrative.

---

# 3. Design Principles

* interface mode is implicit from port — not negotiated
* one active session per player at a time
* handshake must complete before any game commands are accepted
* handshake failure closes the connection immediately
* server is authoritative on session_id, tick_interval, and all timestamps

---

# 4. Handshake Flow — SSH (TEXT Mode)

```
[Client connects to SSH port 2222]
        ↓
[SSH authentication completes]
        ↓
Server  →  handshake_init     (TEXT mode assumed)
Client  →  handshake_response
Server  →  handshake_ack  OR  handshake_reject
        ↓
[Session ACTIVE — terminal interface begins]
```

### 4.1 handshake_init (Server → Client)

Sent immediately after SSH authentication succeeds.

```json
{
  "type": "handshake_init",
  "timestamp": 1709600000,
  "protocol_version": "1.0",
  "interface_mode": "TEXT",
  "server_name": "Black Sector",
  "motd": "Welcome. Watch your back out there.",
  "payload": {}
}
```

Fields:

* `interface_mode` — always `"TEXT"` on SSH port. Informational only.
* `motd` — message of the day, displayed to player on connect.

---

### 4.2 handshake_response (Client → Server)

```json
{
  "type": "handshake_response",
  "timestamp": 1709600001,
  "protocol_version": "1.0",
  "correlation_id": "<uuid>",
  "payload": {
    "player_token": "<auth_token>"
  }
}
```

Fields:

* `player_token` — authentication credential issued at account creation or prior login.
* `protocol_version` — must match or be compatible with server version.

---

### 4.3 handshake_ack (Server → Client)

Sent on successful authentication and session creation.

```json
{
  "type": "handshake_ack",
  "timestamp": 1709600002,
  "correlation_id": "<uuid>",
  "payload": {
    "session_id": "<uuid>",
    "player_id": "<uuid>",
    "tick_interval_ms": 2000,
    "interface_mode": "TEXT"
  }
}
```

Session is ACTIVE after client receives handshake_ack.

---

# 5. Handshake Flow — TCP (GUI Mode)

```
[Client connects to TCP port 2223]
        ↓
[TLS handshake completes — connection encrypted]
        ↓
Server  →  handshake_init     (GUI mode assumed)
Client  →  handshake_response
Server  →  handshake_ack  OR  handshake_reject
        ↓
[Session ACTIVE — structured JSON stream begins]
```

TLS must complete before any protocol messages are exchanged. Plaintext connections on port 2223 are rejected immediately. All subsequent traffic — including the player_token — is encrypted in transit.

### 5.1 handshake_init (Server → Client)

```json
{
  "type": "handshake_init",
  "timestamp": 1709600000,
  "protocol_version": "1.0",
  "interface_mode": "GUI",
  "server_name": "Black Sector",
  "motd": "Welcome. Watch your back out there.",
  "payload": {}
}
```

Fields:

* `interface_mode` — always `"GUI"` on TCP port. Informational only.

---

### 5.2 handshake_response (Client → Server)

Identical structure to TEXT mode:

```json
{
  "type": "handshake_response",
  "timestamp": 1709600001,
  "protocol_version": "1.0",
  "correlation_id": "<uuid>",
  "payload": {
    "player_token": "<auth_token>"
  }
}
```

---

### 5.3 handshake_ack (Server → Client)

```json
{
  "type": "handshake_ack",
  "timestamp": 1709600002,
  "correlation_id": "<uuid>",
  "payload": {
    "session_id": "<uuid>",
    "player_id": "<uuid>",
    "tick_interval_ms": 2000,
    "interface_mode": "GUI"
  }
}
```

---

# 6. Rejection Flow

If authentication fails, version is incompatible, or an existing session conflict is not resolved:

```json
{
  "type": "handshake_reject",
  "timestamp": 1709600002,
  "correlation_id": "<uuid>",
  "payload": {
    "reason": "string"
  }
}
```

Connection is closed immediately after handshake_reject is sent.

Common rejection reasons:

| Reason                        | Description                                    |
| ----------------------------- | ---------------------------------------------- |
| `invalid_token`               | Authentication token not recognized            |
| `version_mismatch`            | Protocol version incompatible                  |
| `session_already_active`      | Player already has an active session           |
| `server_full`                 | Concurrent player limit reached                |
| `handshake_timeout`           | Client did not respond within timeout window   |

---

# 7. Session Conflict Handling

One active session per player is permitted at a time.

If a player attempts to connect while an existing session is active:

* Server sends `handshake_reject` with reason `session_already_active`
* The existing session is NOT terminated automatically
* The player must disconnect their existing session before reconnecting

This applies regardless of interface mode. A player cannot hold a TEXT and GUI session simultaneously.

See `session_multiplexing.md` for full session lifecycle details.

---

# 8. Handshake Timeout

If the client connects but does not complete the handshake within the configured timeout:

* Server sends `handshake_reject` with reason `handshake_timeout`
* Connection is closed

Default timeout: 30 seconds.

---

# 9. Protocol Version Compatibility

Version format: `MAJOR.MINOR`

Compatibility rules:

* Major version mismatch → `handshake_reject` with reason `version_mismatch`
* Minor version mismatch → allowed if server considers it backward-compatible
* Server defines minimum supported protocol version in configuration

---

# 10. Security Considerations

* SSH port authentication is handled by the SSH transport layer before the application handshake begins
* TCP port must validate `player_token` in `handshake_response` — no transport-level auth
* All tokens are validated server-side; clients cannot forge session_id or player_id
* Handshake messages must be processed before any game commands are accepted

---

# 11. Non-Goals (v1)

* interface mode negotiation (mode is implicit from port)
* simultaneous TEXT + GUI sessions for the same player
* guest or anonymous sessions
* OAuth or third-party authentication

---

# 12. Future Extensions

* token refresh during active session
* multi-factor authentication on TCP port
* session transfer between ports (switch from SSH to GUI without disconnect)

---

# End of Document
