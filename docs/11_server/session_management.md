# Session Management Specification

## Version: 0.1
## Status: Draft
## Owner: Core Architecture

## Last Updated: 2026-03-05

---

# 1. Purpose

Defines how player sessions are managed across connection, disconnection, and logout events.

This specification governs:

- Session lifecycle
- Disconnect behavior
- Ship persistence on logout
- Safe logout requirements
- Reconnection handling
- Forced disconnection policy

---

# 2. Scope

## IN SCOPE

- Session state model
- Disconnect handling rules
- Ship state on logout
- Dock-required logout mechanic
- Lingering ship behavior
- Reconnection flow

## OUT OF SCOPE

- Authentication flow (see identity_and_registration.md)
- Command queue mechanics (see command_queue.md)
- Transport layer handling (see system_architecture.md)

---

# 3. Design Principles

- The universe does not pause when a player disconnects.
- Safety on logout is not guaranteed.
- Players who plan ahead log off safely.
- Players who do not accept the consequences.
- Ship state must always be consistent with simulation state.
- Disconnection must never corrupt world state.

---

# 4. Session States

A session may be in one of the following states:

CONNECTED
→ Active authenticated session with command submission enabled.

DISCONNECTED_LINGERING
→ Player has disconnected. Ship remains in space for a configurable duration.

DOCKED_OFFLINE
→ Player logged off while docked. Ship is stored safely at the port or station.

DESTROYED
→ Ship was destroyed while player was offline or during lingering phase.

---

# 5. Logout Behavior

## 5.1 Docked Logout (Safe)

If a player disconnects while their ship is docked at a port or station:

- Ship is placed in DOCKED_OFFLINE state.
- Ship is removed from the active sector entity list.
- Ship state is persisted in the station's storage record.
- No exposure to combat, scanning, or hazards while offline.

On reconnect:

- Ship restored to the station where it was docked.
- Full state preserved.

---

## 5.2 Undocked Disconnect (Unsafe)

If a player disconnects while their ship is in open space:

- Ship enters DISCONNECTED_LINGERING state.
- Ship remains in the sector entity list for LingerDurationTicks.
- Ship is detectable, scannable, and attackable during this window.
- Ship does not move, fire, or take active actions.
- Heat and energy continue to decay normally each tick.

After LingerDurationTicks expires:

- If ship is not under active engagement: ship phases out and is stored in the last visited station.
- If ship is under active engagement: ship remains until engagement resolves.

On reconnect before linger expires:

- Session resumes with ship in its current state.

On reconnect after linger expires:

- Ship is retrieved from storage at the last visited station.

---

## 5.3 Reconnect After Destruction

If a ship is destroyed while the player is offline:

- Player is notified on next login.
- Ship loss is permanent.
- Player must acquire a new ship.

---

# 6. Safe Logout Command

Players may issue an explicit logout command:

LOGOUT

If docked:

- Immediate safe logout.

If undocked:

- Command rejected.
- Player informed they must dock before logging out safely.
- Player may still close connection (triggering undocked disconnect rules).

Players cannot force a safe logout from open space.

---

# 7. Lingering Ship Rules

While in DISCONNECTED_LINGERING state:

- Ship appears in sector scan results.
- Ship signature reflects its last active state.
- Ship may be targeted and attacked.
- Ship does not respond to attacks (no active defense).
- Ship does not issue commands.

LingerDurationTicks is configurable.

Default: 10 ticks (approximately 20 seconds at default tick rate).

---

# 8. Forced Disconnection

If a session is terminated by the server (timeout, admin action, error):

- Same rules apply as undocked disconnect.
- No distinction between voluntary and forced disconnection.

---

# 9. Port and Station Storage

Ships docked offline are stored in the station's ship storage.

Storage record includes:

- ship_id
- account_id
- station_id
- docked_at_tick
- ship_state snapshot

Storage capacity per station is configurable.

No storage fees in v1.

---

# 10. Tunable Parameters

LingerDurationTicks = 10

MaxStoredShipsPerStation = configurable (no hard limit in v1)

---

# 11. Integration Points

Depends On:

- Identity & Registration (authentication state)
- Sector Model (entity list management)
- Persistence Model (ship state snapshot)
- Tick Engine (linger countdown, engagement resolution)

Exposes:

- PlayerDisconnected event
- ShipLingeringExpired event
- PlayerReconnected event

---

# 12. Logging Requirements

Must log:

- Player disconnect (INFO) — include docked/undocked state
- Linger phase start (INFO)
- Linger phase expiry (INFO)
- Ship attacked while lingering (WARN)
- Ship destroyed while offline (WARN)
- Player reconnect (INFO)

---

# 13. Testing Requirements

Tests must validate:

- Docked logout removes ship from sector entity list
- Undocked disconnect places ship in linger state
- Linger countdown decrements correctly each tick
- Ship phases out correctly after linger expires
- Reconnect before linger expiry restores correct state
- Reconnect after linger expiry places ship at last station
- Destruction while offline correctly flags ship as lost

---

# 14. Non-Goals (v1)

- Offline protection timers (safe windows after login)
- Ship insurance or loss compensation
- Offline passive income
- Crew management during offline periods

---

# 15. Future Extensions

- Reputation consequences for attacking offline ships
- Rental fees for long-term station storage
- Emergency beacon on linger phase entry
- Configurable linger behavior per security zone

---

# End of Document
