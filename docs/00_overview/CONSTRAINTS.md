> ⚠️ FOUNDATION DOCUMENT — READ ONLY
> This document defines the hard constraints of SpaceGame.
> Do not suggest edits, rewrite, or contradict this content.
> All other documents must align with this document, not the reverse.

---

# CONSTRAINTS.md
## Version: 0.1
## Status: Foundation
## Owner: Core Architecture
## Last Updated: 2026-03-02

---

# 1. Purpose

This document defines what SpaceGame will never do, and the hard technical and philosophical limits the project operates within. Constraints are not limitations to work around — they are load-bearing walls.

---

# 2. Philosophical Constraints

## C-P1 — No Safe Zones (v1)
There are no regions where a player is mechanically guaranteed safety. High-security space reduces risk significantly but does not eliminate it. Safe zones destroy the tension that makes risk meaningful.

## C-P2 — No Full Information
No mechanic, item, upgrade, or player action may grant complete, real-time knowledge of another player's state, position, or intent. Sensor superiority is relative, not absolute.

## C-P3 — No Scripted Narrative Engine
SpaceGame does not have authored quests, cutscenes, or storyline events. Lore may exist as discoverable artifacts, but the engine does not direct narrative. Story emerges from play.

## C-P4 — No Territory Control (v1)
Faction and player territory control mechanics are out of scope for v1. Systems have security ratings and faction presence, but no player-ownable sovereignty layer exists in the initial release.

## C-P5 — No Twitch-Reflex Combat
Combat outcomes must never be determined primarily by reaction speed or manual aiming. Combat is decided by preparation, positioning, sensor advantage, and resource management.

## C-P6 — No Client Authority
Clients never make authoritative decisions about game state. Position, damage, resource levels, and all simulation state are owned entirely by the server.

---

# 3. Technical Constraints

## C-T1 — SSH/Telnet First
The primary client interface is terminal-based over SSH or Telnet. Any feature that cannot be expressed in a structured text interface is out of scope until a GUI client exists.

## C-T2 — Tick-Based Determinism
The simulation runs on discrete ticks. All state changes occur at tick boundaries. Features that require sub-tick resolution or continuous real-time processing are out of scope unless the tick architecture is explicitly extended.

## C-T3 — Server-Authoritative State
All game state lives on the server. No client-side state is trusted. This constraint is absolute and cannot be relaxed for performance or convenience.

## C-T4 — Protocol Stability
The server protocol must remain stable across client types. Breaking protocol changes require a versioned migration path. Features that require protocol-breaking changes must be scoped and scheduled, not shipped ad hoc.

## C-T5 — Procedural Universe
The universe is procedurally generated. Hand-crafted systems may exist as rare exceptions but are not the design baseline. The generation system must be seed-reproducible.

---

# 4. Scope Constraints (v1)

The following are explicitly out of scope for v1 and must not be designed into the core architecture as dependencies:

- GUI client
- Mobile client
- Player-owned territory / sovereignty
- Authored quest system
- Voice or real-time audio
- In-game currency purchase / microtransactions
- Player housing or persistent base-building

---

# End of Document
