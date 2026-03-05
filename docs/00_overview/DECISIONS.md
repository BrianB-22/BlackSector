> ⚠️ FOUNDATION DOCUMENT — READ ONLY
> This document is an append-only log of key design decisions.
> Do not remove or alter existing entries.
> New entries may be appended by the project owner only.
> Understanding the "why" behind decisions prevents drift and revisionism.

---

# DECISIONS.md
## Version: 0.1
## Status: Foundation (Living Document — append only)
## Owner: Core Architecture
## Last Updated: 2026-03-02

---

# 1. Purpose

This document records significant design decisions: what was decided, why, and what alternatives were rejected. It exists to prevent two failure modes:

1. **Drift** — A later module quietly reverses a decision without acknowledging it.
2. **Revisionism** — A future designer asks "why did we do it this way?" and gets no answer.

Every entry must include: the decision, the reasoning, and the alternatives considered.

---

# 2. Decision Log

---

## D-001 — Text-First Interface (SSH/Telnet)
**Date:** 2026-03-02
**Status:** Accepted

**Decision:** The primary client interface for v1 is terminal-based over SSH or Telnet, using structured ANSI text.

**Reasoning:**
- Establishes a clean protocol boundary between simulation and display.
- Forces UI-agnostic simulation design from day one.
- Eliminates visual asset dependencies in early development.
- Terminal UI is a deliberate aesthetic choice, not a technical limitation.
- Future GUI clients consume the same server protocol without requiring simulation changes.

**Alternatives Considered:**
- Web client (rejected: couples frontend framework to early-stage backend)
- Native GUI client (rejected: asset and rendering overhead too early)
- Browser-based terminal emulator (deferred: possible future client, not v1)

---

## D-002 — Server-Authoritative Architecture
**Date:** 2026-03-02
**Status:** Accepted

**Decision:** The server is the sole authority on all game state. Clients display only; they do not influence canonical state.

**Reasoning:**
- Multiplayer integrity requires a single source of truth.
- Anti-cheat and fairness depend on this being absolute.
- Enables multiple client types without state conflicts.
- Simplifies reasoning about game state during development.

**Alternatives Considered:**
- Client-side prediction with server reconciliation (rejected: introduces trust surface and complexity incompatible with early-stage development)
- Peer-to-peer (rejected: incompatible with persistent simulation and anti-cheat requirements)

---

## D-003 — Tick-Based Determinism
**Date:** 2026-03-02
**Status:** Accepted

**Decision:** All simulation state advances at discrete tick boundaries. No continuous real-time state updates.

**Reasoning:**
- Deterministic ticks make state reproducible and debuggable.
- Simplifies persistence (snapshot at tick boundary).
- Enables fair simultaneous resolution of actions.
- Reduces server load compared to continuous simulation.

**Alternatives Considered:**
- Real-time continuous simulation (rejected: adds complexity, debugging difficulty, and load without meaningful player benefit given text-first interface)
- Event-driven asynchronous updates (deferred: may be used within tick processing, not as replacement)

---

## D-004 — Information Asymmetry as Core Mechanic
**Date:** 2026-03-02
**Status:** Accepted

**Decision:** Players never have full information. Sensor range, scan quality, and market data are always incomplete, degraded, or delayed. There is no mechanic that grants complete information.

**Reasoning:**
- Information asymmetry is the primary driver of strategic depth.
- Creates meaningful decisions around detection, evasion, and intelligence gathering.
- Makes exploration and intelligence economically valuable.
- Prevents the game from becoming a solved optimization problem.

**Alternatives Considered:**
- Full local awareness (rejected: removes strategic fog and makes combat deterministic)
- Purchasable full-system maps (rejected: trivializes exploration and intel economy)

---

## D-005 — No Safe Zones (v1) — Amended for Federated Space
**Date:** 2026-03-02
**Status:** Accepted (amended 2026-03-05)

**Decision:** High-security systems significantly reduce risk but do not eliminate it. With one explicit exception, no region provides mechanical immunity from attack or loss.

**Exception — Federated Space:** PvP combat between players is server-side disabled in Federated Space. This is a deliberate new-player onboarding protection. NPC pirates are also absent. Federated Space is not a loophole for farming — mining yields are intentionally very low and it is not a viable long-term play zone.

**Reasoning:**
- Persistent stakes require that loss always be possible — everywhere except Federated Space.
- Safe zones create gameplay gravity that pulls players away from risk-reward content. Federated Space is the exception that proves the rule: it exists to get players started, not to shelter them indefinitely.
- The risk gradient only has meaning if both ends are real. Federated Space is the floor of the gradient, not a bypass.

**Alternatives Considered:**
- No safe zone at all (rejected: new players with no ship and no knowledge need a protected start)
- Station docking as safe state (partial: docked ships are harder targets, but not immune outside Federated Space)
- Mechanical immunity in all High Security systems (rejected: dilutes the risk gradient)

---

## D-006 — Procedural Universe Generation
**Date:** 2026-03-02
**Status:** Accepted

**Decision:** The universe is procedurally generated from a reproducible seed. Hand-crafted systems are rare exceptions, not the baseline.

**Reasoning:**
- Scale requires procedural generation — hand-crafting a vast universe is not feasible.
- Seed reproducibility enables debugging and support.
- Procedural generation reinforces the simulation-not-theme-park identity.

**Alternatives Considered:**
- Fully hand-crafted universe (rejected: does not scale; contradicts simulation identity)
- Hybrid (accepted as exception case only: key lore locations may be hand-placed, but the generation system cannot depend on this)

---

# 3. How to Add an Entry

Copy this template and append to Section 2:

```
## D-XXX — [Short Title]
**Date:** YYYY-MM-DD
**Status:** Accepted | Deferred | Rejected

**Decision:** [One sentence statement of what was decided.]

**Reasoning:**
- [Bullet points explaining why]

**Alternatives Considered:**
- [What else was considered and why it was rejected or deferred]
```

---

# End of Document
