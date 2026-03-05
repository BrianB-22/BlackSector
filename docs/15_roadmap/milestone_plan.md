# Milestone Plan

## Version: 0.1

## Status: Draft

## Owner: Core Architecture

## Last Updated: 2026-03-05

---

# 1. Purpose

Defines the development milestone structure for BlackSector.

Milestones represent stable, testable states of the server. Each milestone builds on the previous and delivers a defined set of playable functionality.

---

# 2. Milestone Summary

| Milestone | Name                   | Description                                              |
| --------- | ---------------------- | -------------------------------------------------------- |
| M1        | Foundation             | Server boots, players can connect via SSH                |
| M2        | Vertical Slice         | Core playable loop: navigate, trade, fight               |
| M3        | World Expansion        | Economy events, mining, exploration, missions            |
| M4        | Depth and Balance      | AI traders, full economic model, balancing               |
| M5        | GUI Client             | Optional graphical frontend on TCP port 2223             |

---

# 3. Milestone 1 — Foundation

Goal: A running headless server that accepts SSH connections and responds to basic protocol messages.

### Deliverables

* Server binary compiles and runs
* SQLite database initialized with schema
* Configuration loading from JSON files
* SSH listener on port 2222
* Handshake protocol (TEXT mode)
* Session management (connect, disconnect, linger)
* Admin CLI via stdin
* Tick loop running at 2-second interval
* Server log output (structured JSON)
* Snapshot save and load

### Success Criteria

* Player can SSH in and receive `handshake_ack`
* Server survives restart and restores state from snapshot
* Admin can run `server status` and receive output

---

# 4. Milestone 2 — Vertical Slice

Goal: A minimal playable game loop. A player can create a ship, navigate between systems, trade commodities, and fight NPC pirates.

See `phase_1_vertical_slice.md` for full detail.

### Deliverables

* Player registration and ship creation
* Universe: 1 region, 10–20 systems, static commodity prices
* Navigation: jump between connected systems
* Port docking and undocking
* Basic commodity trading (buy/sell at ports)
* NPC pirates that attack players in Low Security space
* Turn-based combat (attack, flee)
* Ship destruction and respawn
* Player credits tracking
* TEXT mode terminal interface (ANSI rendered)

### Success Criteria

* 5 concurrent testers can play independently without interference
* Player can complete a trade run profitably
* Player can win and lose combat against NPC pirates
* No data loss on server restart

---

# 5. Milestone 3 — World Expansion

Goal: Full game systems active. Players have a rich variety of activities, the economy is dynamic, and the world is large enough to explore.

See `phase_2_expansion.md` for full detail.

### Deliverables

* Full universe: multiple regions, 500+ systems
* Black Sector region with contraband mechanics
* Economic events system (food shortage, industrial boom, etc.)
* Mining system with asteroid fields
* Exploration system with sensor scans and anomalies
* Mission system with external JSON scripts
* AI traders (60–70% of trade volume)
* Dynamic commodity pricing (supply and demand)
* Community mission folder support

### Success Criteria

* 25 concurrent testers, stable tick performance
* No single trade route dominates — routes shift over time
* Mining and exploration feel distinct from trading
* Missions complete correctly, rewards distributed
* AI traders maintain economic baseline

---

# 6. Milestone 4 — Depth and Balance

Goal: The game is balanced for long-term play. Edge cases handled, exploitation paths closed, telemetry in place.

### Deliverables

* Economic balancing pass across all commodity types
* Frontier and Black Sector reward/risk calibration
* Ship class balance pass
* Mining yield and depletion tuning
* Event frequency and impact tuning
* Full telemetry and operational monitoring
* Rate limiting and security hardening
* Performance validation at 100 concurrent players
* Snapshot reliability testing

### Success Criteria

* 100 concurrent players, tick duration under 100ms average
* No single dominant player strategy
* Economy remains dynamic after 7-day continuous uptime
* Zero critical errors in 48-hour stability test

---

# 7. Milestone 5 — GUI Client

Goal: Optional graphical frontend for players who prefer it. SSH interface remains fully functional and unchanged.

### Deliverables

* TLS/TCP listener on port 2223 (hardened for production)
* GUI client application consuming structured protocol messages
* System map visualization
* Ship status dashboard
* Market interface
* Combat display
* Navigation interface
* All gameplay features accessible via GUI

### Success Criteria

* GUI player can play a full game session without SSH
* SSH and GUI players can coexist on the same server simultaneously
* Protocol remains unchanged from Milestone 4

---

# 8. Scope Boundaries

The following are explicitly out of scope for all milestones in this plan:

* Faction systems and faction warfare
* Fleet combat
* Player-owned stations and infrastructure
* Manufacturing chains
* WebSocket protocol support
* Binary protocol compression
* Distributed / multi-server architecture

---

# 9. Technical Debt Tracking

Known technical debt items are tracked in `technical_debt_log.md`.

Debt items should be addressed before or during Milestone 4 where they affect performance or reliability.

---

# End of Document
