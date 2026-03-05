# Message Types Specification
## Version: 0.3
## Status: Draft
## Owner: Core Architecture
## Last Updated: 2026-03-04

---

# 1. Purpose

Defines the canonical list of message categories used by the protocol.

Message types determine how payload schemas are interpreted by clients.

All messages must include the `type` field defined in the message envelope.

---

# 2. Design Principles

Message types must be:

- deterministic
- domain-scoped
- forward compatible
- semantically stable

Message names must follow:

<domain>_<action>

---

# 3. System Messages

Used for session and protocol operations.

Examples:

handshake_init  
handshake_ack  
handshake_reject  
tick_update  
server_shutdown  

---

# 4. Command Messages

Represent client-submitted actions.

Examples:

command_submit  
command_accept  
command_reject  

Commands represent player intent but are not guaranteed execution.

---

# 5. Combat Messages

Represent combat state events.

Examples:

combat_start  
combat_update  
weapon_fired  
projectile_spawn  
ship_destroyed  

---

# 6. Exploration Messages

Represent discovery and scanning.

Examples:

scan_result  
signal_detected  
anomaly_discovered  
navigation_corridor_discovered  

---

# 7. Mining Messages

Examples:

mining_start  
mining_yield  
mining_interrupted  

---

# 8. Economy Messages

Examples:

market_update  
trade_executed  
data_market_listing  

---

# 9. Navigation Messages

Examples:

navigation_route_update  
jump_initiated  
jump_completed  
hazard_warning  

---

# 10. Communications Messages

Represent player-to-player and player-to-drone messaging.

Examples:

message_received
irn_outage

---

# 11. Drone Messages

Represent drone control and telemetry.

Examples:

drone_telemetry

Note: Drone commands are submitted via `command_submit` with `command: "drone_command"`. Drone telemetry responses use the `drone_telemetry` type.

---

# 12. Banking Messages

Represent bank account operations and notifications.

Examples:

bank_interest_credited
bank_payment_received

Note: All banking commands are submitted via `command_submit`. Banking messages are server-initiated notifications only (interest accrual, incoming payments).

---

# 13. Error Messages

Examples:

error_invalid_command
error_rate_limited
error_unauthorized

---

# End of Document