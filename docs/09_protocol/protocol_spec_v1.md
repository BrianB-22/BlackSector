# Protocol Specification v1

## Version: 0.1
## Status: Draft
## Owner: Core Architecture

## Last Updated: 2026-03-02

---

# 1. Purpose

Defines the communication contract between clients and the headless server.

This protocol governs:

- Connection negotiation
- Interface mode selection (TEXT | GUI)
- Message envelope structure
- Command submission
- Event broadcasting
- Tick-based state updates
- Version compatibility

The protocol is transport-agnostic.

---

# 2. Scope

IN SCOPE:

- Handshake negotiation
- Message envelope format
- Command structure
- Event categories
- Error handling
- Rate limiting model

OUT OF SCOPE:

- Transport encryption (handled by SSH/TLS layer)
- UI rendering rules (see interface\_model.md)
- Internal simulation logic
- Database schema

---

# 3. Design Principles

- Server-authoritative simulation.
- Structured messages only.
- No UI formatting in protocol payload.
- Versioned and forward-compatible.
- Deterministic message processing.
- Transport-independent.

---

# 4. Transport Model

The server listens on two ports with distinct purposes:

| Port | Default | Transport | Interface Mode | Client         |
| ---- | ------- | --------- | -------------- | -------------- |
| SSH  | 2222    | SSH       | TEXT           | Any SSH client |
| GUI  | 2223    | TLS/TCP   | GUI            | GUI application |

**SSH port (2222) is the primary interface in v1.** Black Sector is a rich text-based game accessed via SSH. No dedicated client is required.

**TCP port (2223) is reserved for future GUI clients.** The architecture supports it from day one; the GUI client itself is future scope.

Interface mode is determined by which port the client connects to. It is not negotiated during the handshake.

Transport layer must not alter protocol semantics.

Protocol messages are line-delimited JSON objects in v1.

---

# 5. Versioning Model

Protocol version format:

MAJOR.MINOR

Example:

1.0

Compatibility rules:

- Major version mismatch → connection rejected.
- Minor version mismatch → allowed if backward compatible.
- Server defines minimum supported version.

---

# 6. Connection Handshake

## 6.1 Server → Client (Greeting)

On connect, server sends (interface_mode reflects the port connected to):

{

     "type": "handshake\_init",

     "protocol\_version": "1.0",

     "interface\_mode": "TEXT",

     "server\_name": "Black Sector",

     "motd": "..."

}

---

## 6.2 Client → Server (Response)

Client responds with:

{

     "type": "handshake\_response",

     "protocol\_version": "1.0",

     "correlation\_id": "\<uuid\>",

     "payload": {

       "player\_token": "\<auth\_token\>"

     }

}

---

## 6.3 Handshake Resolution

If accepted:

{

     "type": "handshake\_ack",

     "session\_id": "...",

     "tick\_interval\_ms": 2000

}

If rejected:

{

     "type": "handshake\_reject",

     "reason": "Protocol version mismatch"

}

Connection closed after rejection.

---

# 7. Message Envelope Format

All protocol messages must follow this structure:

{

     "type": "<message\_type>",

     "timestamp": <server\_tick\_or\_epoch>,

     "correlation\_id": "<uuid\_optional>",

     "payload": { ... }

}

Fields:

- type: required
- timestamp: required (server authoritative)
- correlation\_id: required for command/response pairs
- payload: required (object)

No raw text messages permitted at protocol layer.

---

# 8. Message Categories

Message types are grouped logically.

## 8.1 System

- handshake\_init
- handshake\_ack
- handshake\_reject
- server\_shutdown
- tick\_update

## 8.2 Command

- command\_submit
- command\_accept
- command\_reject

## 8.3 Combat

- combat\_start
- combat\_update
- projectile\_spawn
- ship\_destroyed

## 8.4 Mining

- mining\_start
- mining\_yield
- hazard\_trigger

## 8.5 Exploration

- scan\_result
- anomaly\_discovered

## 8.6 Economy

- market\_update
- rare\_discovery\_event

## 8.7 Error

- error\_invalid\_command
- error\_rate\_limited
- error\_unauthorized

---

# 9. Command Submission Model

Client submits:

{

     "type": "command\_submit",

     "timestamp": <client\_timestamp>,

     "correlation\_id": "<uuid>",

     "payload": {

       "command": "fire\_weapon",

       "parameters": { ... }

     }

}

Server validates and responds:

ACCEPT:

{

     "type": "command\_accept",

     "correlation\_id": "<uuid>",

     "payload": { "queued": true }

}

REJECT:

{

     "type": "command\_reject",

     "correlation\_id": "<uuid>",

     "payload": { "reason": "Insufficient energy" }

}

Command execution occurs inside tick engine.

---

# 10. Tick Update Model

Server emits periodic tick update:

{

     "type": "tick\_update",

     "timestamp": <tick\_number>,

     "payload": {

       "tick": <int>,

       "server\_load": <float\_optional>

     }

}

Subsystem updates are sent as discrete event messages.

No full-state dump per tick in v1.

---

# 11. TEXT Mode Behavior

In TEXT mode:

- Server produces structured messages internally.
- Text adapter converts messages to ANSI-rendered output.
- Client receives formatted terminal content.
- Protocol envelope still governs underlying communication.

TEXT mode is an adapter layer, not a different protocol.

---

# 12. GUI Mode Behavior

In GUI mode:

- Server sends structured JSON messages only.
- No ANSI codes.
- No formatting instructions.
- Client responsible for presentation.

GUI must respect protocol message envelope strictly.

---

# 13. Rate Limiting \& Security

Rate limits are enforced per connection. Since one player may have only one active session at a time, this is effectively per-player.

Server enforces:

- Max commands per tick per connection (default: 3).
- Max malformed messages per minute (default: 10).
- Max handshake attempts per IP per minute (default: 5).
- Disconnect on repeated protocol violations.

Invalid message format → immediate rejection.

Authentication (player\_token) must be validated before gameplay commands are accepted.

See `rate\_limiting.md` and `session\_multiplexing.md`.

---

# 14. Error Handling

Errors follow envelope format:

{

     "type": "error\_invalid\_command",

     "timestamp": <tick>,

     "correlation\_id": "<uuid>",

     "payload": {

       "reason": "Unknown command"

     }

}

Errors must not crash session.

---

# 15. Backward Compatibility Strategy

- Additive changes allowed in MINOR version.
- Field removal requires MAJOR increment.
- Deprecated fields must remain for one minor cycle.
- Clients must ignore unknown fields.

---

# 16. Performance Constraints

- O(1) envelope parse per message.
- No per-message blocking I/O inside tick loop.
- Command enqueue must be non-blocking.

---

# 17. Non-Goals (v1)

- Binary compression
- WebSocket support
- Streaming delta encoding
- Peer-to-peer communication
- Client-side simulation authority

---

# 18. Future Extensions

- Binary protocol variant
- Message batching
- Event compression
- Real-time streaming GUI mode
- Replay synchronization

---

# End of Document
