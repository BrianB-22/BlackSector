# Phase 2: Expansion

## Version: 0.1

## Status: Draft

## Owner: Core Architecture

## Last Updated: 2026-03-05

---

# 1. Purpose

Defines the scope and success criteria for Phase 2: Expansion.

Phase 2 activates the full set of game systems designed in the specification. It builds on the validated vertical slice from Phase 1 and expands the world, economy, and player activities to deliver the complete BlackSector experience.

Phase 2 corresponds to Milestone 3 in `milestone_plan.md`.

---

# 2. Phase 2 Goals

* Full galaxy with multiple regions and security zones
* Dynamic economy driven by supply/demand and economic events
* AI traders maintaining economic baseline
* Mining as a distinct player activity
* Exploration as a distinct player activity
* Mission system with community-scriptable content
* Black Sector as the highest-risk, highest-reward zone

---

# 3. World Expansion

## 3.1 Full Universe

* 4–6 regions with distinct economic characters
* 500–1,000 star systems
* Jump network connecting all regions
* Region types: core, agricultural, industrial, frontier, black

## 3.2 Security Zones

| Zone           | Security Level | Description                              |
| -------------- | -------------- | ---------------------------------------- |
| High Security  | 0.7–1.0        | Patrolled, safe for new players          |
| Medium Security| 0.4–0.7        | Moderate risk, better rewards            |
| Low Security   | 0.0–0.4        | Dangerous, high profit potential         |
| Black Sector   | Special        | Lawless, contraband, extreme rewards     |

## 3.3 Black Sector

* Accessible only through specific jump points
* Contraband commodities legal here (alien_artifacts, stolen_goods, etc.)
* No law enforcement spawns
* Highest commodity prices and economic volatility
* Frequent economic events
* PvP risk is highest

---

# 4. Dynamic Economy

## 4.1 Supply and Demand

Commodity prices now fluctuate based on:

* Port inventory levels (low stock → higher prices)
* Player and AI trader activity
* Regional production/consumption patterns
* Active economic events

## 4.2 Economic Events

All events from `economic_events.md` activate:

* Food Shortage
* Industrial Boom
* Black Market Surge
* (and additional events defined in config)

Events occur on configurable intervals, create temporary opportunities, and expire naturally.

## 4.3 Full Commodity Set

Expansion of Phase 1 commodities to include:

* Medical Supplies (essential)
* Alien Artifacts (exotic, primarily in Black Sector)
* Stolen Goods (exotic, contraband)
* Industrial Equipment (industrial)
* Luxury Artifacts (luxury)
* Anomaly Resources (exotic, discovered via exploration)

---

# 5. AI Traders

* 50–80 AI traders active in the universe
* Traders select routes based on current price differentials
* Traders respond to economic events by shifting routes
* Traders target 60–70% of total trade volume
* Trader names loaded from `config/ai/trader_names.json`
* Traders can be destroyed by players (creates brief supply gap)

See `ai_trader_model.md` for full specification.

---

# 6. Mining System

* Asteroid fields in systems across all security zones
* Yield varies by field type and security zone
* Depletion model: fields become less productive with use, regenerate over time
* High Security fields: low yield, safe
* Black Sector fields: very high yield, dangerous
* Mining drones (future upgrade) increase yield and reduce hazard exposure
* Mineable commodities: raw_ore, rare_minerals, exotic_crystals

See `mining_system.md` and `mining_balancing_guidelines.md`.

---

# 7. Exploration System

* Sensor scans reveal hidden objects in a system
* Anomalies discoverable: derelict_ship, ancient_artifact, unstable_wormhole, energy_vortex, exotic_gas_cloud
* First-discovery bonus credits
* Player map system (discovered systems retained across sessions)
* Deep scan reveals greater detail (rarer discoveries)
* Frontier and Black Sector have higher anomaly density

See `exploration_system.md` and `mapping_data_model.md`.

---

# 8. Mission System

* Missions loaded from `config/missions/*.json`
* Community mission folder: `config/missions/community/`
* Admin hot-reload via `mission reload`
* Mission types: combat, delivery, acquisition, navigation, scan, dock, survival
* Chained objectives (sequential completion required)
* Rewards: credits, items, upgrades
* Failure conditions: ship destroyed, time expired

See `mission_framework.md` and `content_schema.md`.

---

# 9. Ship Upgrades

Basic upgrade set introduced in Phase 2:

| Upgrade              | Effect                                |
| -------------------- | ------------------------------------- |
| Cargo Expander       | +10 cargo capacity                    |
| Shield Booster       | +25 max shield points                 |
| Weapon Enhancer      | +20% weapon damage                    |
| Mining Drill         | +50% mining yield                     |
| Sensor Array         | Improved scan range and quality       |
| Fuel Optimizer       | -20% fuel cost per jump               |

Upgrades are purchased at ports and installed immediately.

---

# 10. Additional Ship Classes

Phase 2 introduces two additional ship classes:

| Class    | Role          | Hull | Shield | Energy | Cargo |
| -------- | ------------- | ---- | ------ | ------ | ----- |
| courier  | Starter       | 100  | 50     | 100    | 20    |
| freighter| Trade         | 150  | 40     | 80     | 60    |
| fighter  | Combat        | 120  | 80     | 120    | 10    |

Players can purchase new ship classes at specific ports.

---

# 11. Combat Expansion

* NPC pirate groups (2–3 ships) in Low Security and Black Sector
* Loot drop from destroyed NPCs (chance-based commodity drop)
* Combat in Black Sector can involve other players (PvP opt-in via region entry)
* Flee success probability (dependent on ship speed/upgrades)

---

# 12. Interface Expansion

* Port market listing now shows supply/demand indicators
* System map shows security zone color coding in TEXT mode
* Economic event notifications in player feed
* Mission tracker in HUD
* New player tutorial flow (first-run guidance text)

---

# 13. Success Criteria

Phase 2 is complete when:

* 25 concurrent players, tick duration under 100ms average
* Economy remains dynamic after 72 hours continuous play (no static equilibrium)
* AI traders account for 60–70% of trade volume (telemetry verified)
* All 3 player activities (trading, mining, exploration) are viable profit paths
* Missions complete and reward correctly for all 7 objective types
* Black Sector accessible, profitable, and demonstrably more dangerous
* Community mission JSON file loads correctly via hot-reload

---

# 14. Out of Scope (Phase 2)

| Feature              | Status         |
| -------------------- | -------------- |
| GUI client           | Phase 3+       |
| Faction systems      | Out of scope   |
| Fleet combat         | Out of scope   |
| Player-owned stations| Out of scope   |
| Manufacturing        | Out of scope   |
| Economic warfare     | Future         |
| Player relief missions| Future        |
| Faction embargoes    | Future         |

---

# End of Document
