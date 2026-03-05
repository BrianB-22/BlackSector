> ⚠️ FOUNDATION DOCUMENT — READ ONLY
> This document defines the high-level architecture of SpaceGame.
> Do not suggest edits, rewrite, or contradict this content.
> All other documents must align with this document, not the reverse.

---

# ARCHITECTURE.md
## Version: 0.1
## Status: Foundation
## Owner: Core Architecture
## Last Updated: 2026-03-02

---

# 1. Purpose

This document defines the structural skeleton of SpaceGame — how major systems relate to each other, what owns what, and where the boundaries between systems live. All module-level design docs must fit within this structure.

---

# 2. System Model

SpaceGame is a server-authoritative simulation. The server is the canonical source of all state. Clients are display and input terminals only.

```
[ Client Layer ]
    SSH / Telnet Terminal (v1)
    Future GUI Client (vNext)
         |
         | Protocol (structured text / future: binary)
         |
[ Server Layer ]
    Connection Manager
    Session Manager
         |
    Command Dispatcher
         |
    Simulation Engine  ←————————————————————┐
    ├── Tick Engine                          |
    ├── Universe Manager                     |
    │   ├── Sector / System Graph            |
    │   └── Procedural Generation            |
    ├── Entity Manager                       |
    │   ├── Ships                            |
    │   ├── Stations                         |
    │   ├── NPCs / AI Agents                 |
    │   └── Objects (debris, signals, etc.)  |
    ├── Combat Engine                        |
    │   ├── Sensor System                    |
    │   ├── Weapons & Targeting              |
    │   ├── Heat Management                  |
    │   └── Damage Resolution                |
    ├── Economic Engine                      |
    │   ├── Market System                    |
    │   ├── Resource Extraction              |
    │   ├── Supply/Demand Model              |
    │   └── AI Trader Agents                 |
    ├── Information System                   |
    │   ├── Scan & Detection                 |
    │   ├── Intelligence / Intel Market      |
    │   └── Signal Propagation               |
    └── Persistence Layer ——————————————————┘
        ├── State Serialization
        ├── Tick Snapshot
        └── Database Interface
```

---

# 3. Ownership Model

| Domain | Owned By | Notes |
|---|---|---|
| All game state | Server | Absolute. No exceptions. |
| Tick advancement | Tick Engine | All state changes at tick boundary |
| Universe topology | Universe Manager | Procedural, seed-reproducible |
| Entity lifecycle | Entity Manager | Ships, stations, NPCs, objects |
| Combat resolution | Combat Engine | Consumes sensor + entity state |
| Market state | Economic Engine | Driven by supply/demand + AI agents |
| Detection / intel | Information System | Feeds combat and economic decisions |
| Data persistence | Persistence Layer | Snapshots at tick; async writes |
| Client display | Client Layer | Read-only view of server state |

---

# 4. Key Relationships

- **Combat depends on Information.** Targeting, tracking, and engagement decisions are downstream of the sensor and scan system. Combat without information context is not valid.
- **Economy depends on Geography.** Market prices, pirate density, and opportunity are all functions of location in the risk gradient. The economic engine must be spatially aware.
- **Everything depends on the Tick Engine.** No system advances state outside of tick processing. The tick is the universal clock.
- **Persistence is downstream of simulation.** The persistence layer records state; it does not produce it. The simulation runs in memory; persistence is its journal.
- **Clients are downstream of everything.** The client layer renders state it receives. It has no upstream influence on simulation state.

---

# 5. Protocol Boundary

The protocol boundary sits between the Client Layer and the Server Layer. This boundary must remain stable and client-agnostic.

- v1 protocol: structured ANSI text over SSH/Telnet
- Future protocol: to be defined, but must not require changes to simulation internals
- The simulation must be fully operable with no connected clients

---

# 6. Expansion Model

New systems attach to the Simulation Engine as modules. They must:
- Respect tick boundaries
- Consume state through defined interfaces (not direct coupling)
- Not introduce client authority
- Not break the protocol boundary

---

# End of Document
