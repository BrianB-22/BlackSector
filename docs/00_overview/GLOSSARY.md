> ⚠️ FOUNDATION DOCUMENT — READ ONLY
> This document defines canonical terminology for SpaceGame.
> Do not suggest edits, rewrite, or contradict this content.
> All other documents must use these definitions exactly.
> When in doubt about a term, this document is authoritative.

---

# GLOSSARY.md
## Version: 0.1
## Status: Foundation
## Owner: Core Architecture
## Last Updated: 2026-03-02

---

# 1. Purpose

This document defines canonical terms used across all SpaceGame design documents. Consistent terminology prevents semantic drift — where the same word means different things in different modules.

When writing or reviewing any design doc, use these definitions exactly. If a term you need is missing, add it here before using it elsewhere.

---

# 2. Universe & Geography

**Universe** — The full procedurally generated space in which the simulation runs. Seeded at generation; persistent thereafter.

**Sector** — A named region of space containing one or more systems. The top-level geographic division.

**System** — A discrete navigable location within a sector, typically centered on a star. The primary unit of geography. Has a security rating.

**Security Rating** — A numeric or tiered classification of a system describing its relative safety. High security = stable, lower profit. Low security = dangerous, higher profit. The primary risk-reward geographic lever.

**Jump Point** — A navigable connection between two systems. Travel between systems requires transit through a jump point.

**Region** — An informal grouping of sectors. Not a hard simulation boundary; used for narrative and map legibility.

---

# 3. Entities

**Entity** — Any discrete object tracked by the Entity Manager. Includes ships, stations, NPCs, debris, signals, and environmental objects.

**Ship** — A player- or NPC-controlled mobile entity capable of navigation, combat, and trade.

**Station** — A fixed or semi-fixed entity serving as a hub for trade, repair, docking, and information exchange.

**NPC** — A non-player-controlled entity governed by AI agent logic. Includes traders, pirates, faction patrols, and others.

**AI Trader** — A specific NPC type that participates in the economic simulation, moving goods between systems in response to market conditions.

**Object** — A non-entity entity: debris fields, signal sources, derelicts, natural phenomena. Interactable but not agent-driven.

---

# 4. Combat & Sensors

**Tick** — The fundamental unit of simulation time. All state changes occur at tick boundaries. The tick is the universal clock.

**Sensor Range** — The maximum distance at which a ship's sensor suite can detect another entity. Affected by ship loadout, environmental conditions, and target signature.

**Signature** — The detectable output of a ship: heat, EM emissions, mass. Higher signature = easier to detect. Affected by ship activity (engines, weapons, scanning).

**Scan** — An active sensor action that attempts to gather detailed information about a detected entity. Costs time, energy, and increases the scanning ship's own signature.

**Track** — A maintained sensor lock on a detected entity. Required for weapons targeting. Degrades if the target leaves sensor range or suppresses signature.

**Targeting** — The process of establishing a weapons solution against a tracked entity. Requires an active track and sufficient time to resolve.

**Engagement** — A combat interaction between two or more entities. Begins at detection; escalates through tracking, targeting, and weapons fire.

**Heat** — A resource that accumulates from ship activity (engines, weapons, scanning). Exceeding heat limits forces cooldown or causes damage. A core tradeoff resource.

**Energy** — A ship resource consumed by active systems. Must be managed across weapons, sensors, engines, and shields.

---

# 5. Economy

**Market** — A system-level exchange for goods and commodities. Prices are driven by supply, demand, and local conditions. Not player-set in v1.

**Commodity** — A tradable good with supply and demand properties. Prices fluctuate based on economic activity.

**Supply/Demand Model** — The underlying simulation that drives market prices. Affected by trade activity, mining, piracy, and AI trader behavior.

**Risk-Reward Gradient** — The spatial relationship between geographic danger and economic opportunity. Low security systems offer higher potential profit but greater physical risk.

**Intel** — Information about another player's position, loadout, activity, or intent. A tradable asset in the information economy.

---

# 6. Information Systems

**Information Asymmetry** — The design condition in which no player has complete knowledge of the universe state. A core principle, not a bug.

**Detection** — The act of a sensor system identifying that an entity exists within sensor range. Does not imply identity or detail.

**Identification** — The determination of an entity's type, faction, or identity. Requires scan time and proximity beyond basic detection.

**Signal** — An ambient or directed emission that can be detected by sensor systems. May indicate presence, activity, or distress.

**Intel Market** — A player-driven or NPC-mediated system for buying and selling information about other entities, locations, or events.

---

# 7. Technical

**Server-Authoritative** — The architectural principle that the server owns all canonical game state. Clients are display terminals only.

**Tick Engine** — The simulation component responsible for advancing game state at discrete tick intervals.

**Persistence Layer** — The system responsible for serializing and storing simulation state. Downstream of the simulation; does not drive it.

**Protocol** — The defined communication format between client and server. v1: structured ANSI text over SSH/Telnet.

**Session** — A single connected client interaction with the server. A player may have one active session at a time.

---

# End of Document
