# Product Scope

## Version: 0.1
## Status: Draft
## Owner: Core Architecture

## Last Updated: 2026-03-02

---

# 1. Purpose

Defines what will be included in Version 1 (v1) of SpaceGame.

This document prevents scope creep and ensures focused delivery of a playable, stable simulation core.

---

# 2. Version 1 Definition

Version 1 is defined as:

A fully functional, persistent, multiplayer server supporting:

- Text-based SSH interface
- Procedurally generated universe
- Sensor-driven 1v1 combat
- Risk-tiered asteroid mining
- Exploration and anomaly discovery
- Background economic simulation
- Tradeable exploration data
- Command-based interaction
- Deterministic tick engine

v1 must be playable end-to-end without GUI.

---

# 3. Core Systems Included in v1

## 3.1 Server Core

- Headless server
- Deterministic tick engine
- Multi-user session handling
- Command queue
- Event emission system
- Snapshot persistence
- CLI admin tool

---

## 3.2 Protocol \& Interface

- Protocol v1 message envelope
- TEXT interface adapter (SSH primary)
- Interface negotiation framework
- Structured internal message model

GUI client NOT required for v1.

---

## 3.3 Combat

- 1v1 engagements
- Detection-only state
- Tracking buildup
- Scan-only interaction
- Silent shadowing
- Engagement lock
- Disengagement rules
- Projectile-based weapons
- Heat \& energy system
- Torpedoes
- Rail/kinetic weapons
- Mining lasers (dual use)
- EMP/disable weapon

No fleet combat.

---

## 3.4 Propulsion

- Engine classes
- Burn intensity control
- Signature tradeoffs
- Heat scaling
- Fuel usage
- Jump escape mechanics

No warp interdiction modules.

---

## 3.5 Mining

- Asteroid fields
- Density \& depletion
- Instability risk
- Hazard system
- Drone system
- Rare mineral chance
- Security-tier yield scaling

No automated mining.

---

## 3.6 Exploration

- System survey
- Anomaly generation
- Tradeable mapping data
- Data decay
- Rare anomaly event triggers
- Exploration-focused ships

No territory claiming.

---

## 3.7 Economy

- Commodity model
- Supply/demand tracking
- AI trader simulation
- Price volatility
- Rare discovery impact
- Persistent market shifts

No player-owned markets.

---

# 4. Explicitly Out of Scope (v1)

The following will NOT be included:

- Fleet combat
- Corporations
- Player-owned stations
- Territory control
- Complex faction politics
- Boarding mechanics
- Subsystem damage modeling
- Real-time physics simulation
- Browser client
- Binary protocol optimization
- Territory taxation
- Cooperative mining fleets

If not listed in Section 3, it is out of scope.

---

# 5. Playable Vertical Slice Definition

v1 must support this full loop:

1\. Player connects via SSH.

2\. Navigates to low-security system.

3\. Detects another player.

4\. Performs scan-only interaction.

5\. Either disengages or escalates.

6\. Combat resolves.

7\. Player mines asteroid field.

8\. Discovers rare mineral.

9\. Sells resource into fluctuating economy.

10\. Observes economic shift.

If this loop functions reliably, v1 is valid.

---

# 6. Stability Requirements

Before expanding beyond v1:

- Tick engine proven stable under 50+ concurrent users.
- No cross-system state corruption.
- Protocol v1 stable for 2 minor revisions.
- Heat \& energy model balanced.
- Economy not collapsing under mining saturation.
- No dominant combat build.

---

# 7. Performance Targets (v1)

- 50–200 concurrent players
- 500–1000 star systems
- <50ms tick resolution
- O(1) per-entity update cost
- No global recalculation spikes

---

# 8. Expansion Criteria

Expansion beyond v1 only allowed when:

- Core systems stable
- Performance validated
- Protocol frozen at 1.x
- No major architectural refactors pending

Future features must not destabilize:

- Server authority
- Tick determinism
- Risk-reward scaling
- Information asymmetry

---

# 9. Release Definition

v1 is considered complete when:

- Server runs persistently for 72 hours under load
- All core loops functional
- No critical exploit paths
- No infinite resource generation bugs
- No infinite invulnerability builds

v1 is simulation-stable, not content-complete.

---

# End of Document
