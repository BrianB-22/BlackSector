> ⚠️ FOUNDATION DOCUMENT — READ ONLY
> This document defines the design principles of SpaceGame.
> Do not suggest edits, rewrite, or contradict this content.
> All other documents must align with this document, not the reverse.

---

# PRINCIPLES.md
## Version: 0.1
## Status: Foundation
## Owner: Core Architecture
## Last Updated: 2026-03-02

---

# 1. Purpose

This document defines the non-negotiable design principles that govern every system, mechanic, and feature in SpaceGame. When a design decision is ambiguous, these principles are the tiebreaker.

---

# 2. Core Principles

## P1 — Simulation Over Gamification
The universe operates by its own rules, not in service of player entertainment. Systems should behave consistently whether a player is watching or not. Do not add mechanics that exist purely to feel satisfying at the expense of systemic integrity.

## P2 — Information Asymmetry Is Sacred
Players never have full information. Sensor range, scan quality, market data, and player intelligence are always incomplete, degraded, or delayed. Any system that grants perfect information violates this principle.

## P3 — Every Action Has Weight
Actions must have meaningful consequences, costs, or risks. Firing a weapon, jumping a system, broadcasting a signal — each carries tradeoffs. Eliminate zero-cost actions wherever they appear.

## P4 — Complexity Is Opt-In
The default experience should be navigable by a new player. Depth reveals itself through engagement. Do not force complexity onto players who haven't chosen it — but never remove it for those who have.

## P5 — Geography Drives Risk
Location is the primary risk variable. Safe systems are stable and poor. Dangerous systems are volatile and lucrative. This gradient must be consistent and legible. Do not flatten it for convenience.

## P6 — The Server Is Always Right
The server is the sole authority on game state. Clients display; they do not decide. No client-side prediction should affect canonical state. Anti-cheat integrity depends on this principle being absolute.

## P7 — Emergence Over Scripting
Interesting outcomes should arise from system interaction, not authored events. Do not write storylines into the engine. Write systems that produce stories.

## P8 — Tradeoffs Over Dominance
No ship, strategy, or economic path should be universally optimal. Every build, route, or approach must sacrifice something. Dominant strategies are a design failure.

## P9 — Persistence Has Meaning
The universe does not reset. Loss is real. Progress is real. Do not introduce mechanics that trivialize either. Persistence is what makes stakes matter.

## P10 — Protocol Stability First
The simulation backend must be UI-agnostic and protocol-stable. Features that couple the simulation to a specific client type are architectural debt. Design the system; let the interface follow.

---

# 3. Applying These Principles

When reviewing any feature or system, ask:

1. Does this serve the simulation or undermine it?
2. Does this preserve or erode information asymmetry?
3. Does this action carry meaningful cost or risk?
4. Does this force complexity on players who haven't chosen it?
5. Does this flatten the geographic risk gradient?
6. Does this require client authority over game state?
7. Is this an authored event where a system interaction would do?
8. Does this create a dominant strategy?
9. Does this trivialize loss or progress?
10. Does this couple the simulation to a specific client type?

If any answer is "yes," revisit the design.

---

# End of Document
