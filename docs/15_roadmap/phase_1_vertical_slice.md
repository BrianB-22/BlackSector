# Phase 1: Vertical Slice

## Version: 0.1

## Status: Draft

## Owner: Core Architecture

## Last Updated: 2026-03-05

---

# 1. Purpose

Defines the scope and success criteria for the Phase 1 Vertical Slice.

The vertical slice is the minimum playable version of BlackSector. It delivers one complete slice through the game's core systems: connection, navigation, trading, and combat.

Phase 1 corresponds to Milestone 2 in `milestone_plan.md`.

---

# 2. Design Philosophy

The vertical slice is intentionally narrow. It proves that all core systems work together correctly, that the server architecture is sound, and that the fundamental game loop is enjoyable.

Features not included in Phase 1 are deferred to Phase 2, not because they are unimportant, but because they are not needed to validate the core loop.

---

# 3. In Scope

## 3.1 World

* 1 region (core/industrial type)
* 15–20 star systems connected by jump points
* 1–3 ports per system
* High Security and Low Security systems present
* Static world generation (hand-crafted or seed-based, not procedural)

## 3.2 Player

* Player registration via SSH (player name + token issuance)
* One ship per player (courier class as starting ship)
* Player credits (starting balance configurable)
* Ship cargo (static cargo capacity)

## 3.3 Navigation

* Jump between connected systems (fuel cost model)
* System map display (TEXT mode: ASCII system list with connection indicators)
* Docking and undocking at ports

## 3.4 Economy

* 6–8 commodities (2 essential, 3 industrial, 1 luxury)
* Static base prices (no dynamic pricing in Phase 1)
* Port buy/sell interface
* Cargo manifest display
* Credits tracking and display

## 3.5 Combat

* NPC pirates spawn in Low Security systems
* Turn-based combat (attack, flee)
* Weapon damage and shield/hull mechanics
* Ship destruction (player loses cargo and credits penalty; respawns at nearest port)
* NPC destruction (no loot in Phase 1)

## 3.6 Interface

* Full TEXT mode terminal interface
* ANSI color rendering for system maps, ship status, market listings
* Command-line input with tab completion (if feasible)
* Status bar showing: system, credits, hull/shields/energy
* Help system (`help` command lists available commands)

## 3.7 Server

* SSH listener on port 2222
* Handshake and session management
* Tick loop at 2-second interval
* SQLite persistence
* Snapshot save/load (every 100 ticks)
* Admin CLI (server status, player kick, server shutdown)
* Server log output

---

# 4. Out of Scope (Phase 1)

| Feature                    | Deferred To   |
| -------------------------- | ------------- |
| Dynamic pricing            | Phase 2       |
| Economic events            | Phase 2       |
| AI traders                 | Phase 2       |
| Mining                     | Phase 2       |
| Exploration and scanning   | Phase 2       |
| Missions                   | Phase 2       |
| Multiple regions           | Phase 2       |
| Black Sector               | Phase 2       |
| Medium Security space      | Phase 2       |
| Ship upgrades              | Phase 2       |
| Combat loot                | Phase 2       |
| GUI client (port 2223)     | Phase 2+      |
| Faction systems            | Out of scope  |
| Fleet combat               | Out of scope  |

---

# 5. Commodity Set (Phase 1)

| ID             | Name           | Category   | Base Price |
| -------------- | -------------- | ---------- | ---------- |
| food_supplies  | Food Supplies  | essential  | 100        |
| fuel_cells     | Fuel Cells     | essential  | 150        |
| raw_ore        | Raw Ore        | industrial | 80         |
| refined_ore    | Refined Ore    | industrial | 240        |
| machinery      | Machinery      | industrial | 600        |
| electronics    | Electronics    | industrial | 800        |
| luxury_goods   | Luxury Goods   | luxury     | 1500       |

Prices are static in Phase 1. Each port carries a subset of these commodities.

---

# 6. Ship Class (Phase 1)

One starting ship class:

| Field          | Value         |
| -------------- | ------------- |
| Class          | courier       |
| Hull Points    | 100           |
| Shield Points  | 50            |
| Energy Points  | 100           |
| Cargo Capacity | 20 units      |
| Weapon Damage  | 15            |

---

# 7. Success Criteria

Phase 1 is complete when:

* A new player can SSH in and have a functional character within 2 minutes
* A player can complete a profitable trade run between two ports
* A player can survive or die to an NPC pirate and continue playing after death
* 5 concurrent players can play simultaneously without server errors
* Server survives restart and restores state from snapshot with no data loss
* Tick duration remains under 100ms with 5 concurrent players

---

# 8. Known Limitations (Accepted for Phase 1)

* Static commodity prices — no economic simulation
* No AI-driven traders filling the market
* Small world (15–20 systems) — limited exploration
* No ship variety — all players start with the same ship
* No combat loot — NPC pirates drop nothing
* Limited TEXT interface polish — functional, not beautiful

These limitations are intentional and documented. They will be addressed in Phase 2.

---

# 9. Testing Plan

* Unit testing of subsystem logic (combat math, navigation pathfinding)
* Integration test: server boots, player connects, executes trade, survives combat
* Load test: 5 concurrent players running automated trade loops, measure tick duration
* Persistence test: server restart mid-session, verify state recovery

---

# End of Document
