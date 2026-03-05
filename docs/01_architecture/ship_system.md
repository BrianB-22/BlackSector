# Ship System Specification

## Version: 0.1

## Status: Draft

## Owner: Core Architecture

## Last Updated: 2026-03-05

---

# 1. Purpose

Defines ship classes, ship upgrades, player credits, new player starting state, and ship destruction/respawn mechanics for BlackSector.

This spec covers the foundational player-facing game loop: you start with a ship and credits in Federated Space, earn money through trade/mining/missions, upgrade your ship, and manage the consequences of combat defeat.

---

# 2. Design Principles

* Ship classes are defined in config — not hardcoded. Stats can be tuned without a code deploy.
* Ships are persistent objects in the database. They outlive sessions.
* All player money is stored server-side. There is no "cash on ship" — credits are universal.
* Death consequences are server-configurable. The admin decides how harsh the game is.
* New players should be playable immediately — no mandatory purchase before first action.
* Upgrades occupy slots. Slot count varies by ship class. This bounds build complexity.

---

# 3. Federated Space

Federated Space is the starting zone of BlackSector — a cluster of government-controlled star systems at the center of the galaxy. It is the safest zone in the game and serves as the player onboarding area.

Properties:
* Highest security rating — NPC patrol coverage, no player-vs-player combat
* Contains the origin starbase where new players begin
* Full IRN relay coverage (100% reliability, ×1.0 delay)
* Drones are available for purchase at Federated starbases
* Players may always return to Federated Space regardless of location

Federated Space is a security zone above High Security. It is NOT a region in the normal sense — it is a fixed cluster of systems defined in world config.

```
Zone hierarchy (safest → most dangerous):
Federated Space → High Security → Medium Security → Low Security → Black Sector
```

The origin starbase system ID is set in `server.json` as `federated_origin_system_id`. This value is used for new player placement and standard-mode respawn.

> Note: `docs/02_universe/security_zones.md` should be updated to include Federated Space as a formal zone type.

---

# 4. Ship Classes

Ship classes are defined in `config/ships/ship_classes.json`. They are loaded at startup and hot-reloadable.

## 4.1 Phase 1 Classes

| Class     | Hull | Shield | Energy | Cargo | Upgrade Slots | Weapon Slots | Base Price | Speed  |
| --------- | ---- | ------ | ------ | ----- | ------------- | ------------ | ---------- | ------ |
| Courier   | 100  | 50     | 100    | 20    | 3             | 1            | 8,000 cr   | Medium |

Phase 1 ships only courier class. Additional classes are Phase 2.

## 4.2 Phase 2 Classes

| Class     | Hull | Shield | Energy | Cargo | Upgrade Slots | Weapon Slots | Base Price | Speed  |
| --------- | ---- | ------ | ------ | ----- | ------------- | ------------ | ---------- | ------ |
| Scout     | 60   | 40     | 130    | 8     | 4             | 1            | 12,000 cr  | Fast   |
| Freighter | 80   | 30     | 70     | 60    | 3             | 1            | 18,000 cr  | Slow   |
| Fighter   | 130  | 90     | 160    | 6     | 4             | 3            | 22,000 cr  | Fast   |

Class notes:
* **Scout**: Best sensors, suited for exploration and mapping. Light armament.
* **Freighter**: Maximum cargo, minimum combat viability. Trade specialist.
* **Fighter**: Full combat loadout. Multiple weapon slots. Minimal cargo.
* **Courier**: Balanced. The recommended starting class and general-purpose ship.

## 4.3 JSON Config Schema

`config/ships/ship_classes.json`:

```json
{
  "version": "1.0",
  "classes": [
    {
      "class_id": "courier",
      "name": "Courier",
      "description": "A balanced starter ship suited for trade runs and general use.",
      "role": "transport",
      "phase": 1,
      "base_stats": {
        "hull_points": 100,
        "shield_points": 50,
        "energy_points": 100,
        "cargo_capacity": 20
      },
      "upgrade_slots": 3,
      "weapon_slots": 1,
      "base_price": 8000,
      "speed_class": "medium",
      "available_in_federated_space": true
    },
    {
      "class_id": "scout",
      "name": "Scout",
      "description": "Fast and sensor-rich. Built for exploration before fighting.",
      "role": "exploration",
      "phase": 2,
      "base_stats": {
        "hull_points": 60,
        "shield_points": 40,
        "energy_points": 130,
        "cargo_capacity": 8
      },
      "upgrade_slots": 4,
      "weapon_slots": 1,
      "base_price": 12000,
      "speed_class": "fast",
      "available_in_federated_space": true
    },
    {
      "class_id": "freighter",
      "name": "Freighter",
      "description": "Bulk cargo hauler. Slow but carries more than anything else.",
      "role": "cargo",
      "phase": 2,
      "base_stats": {
        "hull_points": 80,
        "shield_points": 30,
        "energy_points": 70,
        "cargo_capacity": 60
      },
      "upgrade_slots": 3,
      "weapon_slots": 1,
      "base_price": 18000,
      "speed_class": "slow",
      "available_in_federated_space": true
    },
    {
      "class_id": "fighter",
      "name": "Fighter",
      "description": "Combat-optimized. High weapons capacity, minimal cargo.",
      "role": "combat",
      "phase": 2,
      "base_stats": {
        "hull_points": 130,
        "shield_points": 90,
        "energy_points": 160,
        "cargo_capacity": 6
      },
      "upgrade_slots": 4,
      "weapon_slots": 3,
      "base_price": 22000,
      "speed_class": "fast",
      "available_in_federated_space": true
    }
  ]
}
```

---

# 5. Ship Upgrades

Upgrades are purchased at ports and installed into ship upgrade slots. Each upgrade occupies one slot. Slot count is fixed per ship class.

Upgrades are defined in `config/ships/upgrades.json`. Hot-reloadable.

## 5.1 Upgrade Catalog

### Navigation / Systems

| Upgrade ID          | Name                  | Cost    | Effect                                              | Phase |
| ------------------- | --------------------- | ------- | --------------------------------------------------- | ----- |
| `enhanced_sensors`  | Enhanced Sensors      | 2,500   | Sensor range ×1.5; detect signals in adjacent systems | 1   |
| `fuel_optimizer`    | Fuel Optimizer        | 2,000   | Jump energy cost −20%                               | 1     |
| `autopilot_advanced`| Advanced Autopilot    | 1,800   | Multi-hop automated route planning                  | 2     |
| `emergency_jump`    | Emergency Jump Drive  | 3,500   | Jump at 40% energy (normal minimum: 60%)            | 2     |

### Combat

| Upgrade ID              | Name                   | Cost   | Effect                                            | Phase |
| ----------------------- | ---------------------- | ------ | ------------------------------------------------- | ----- |
| `shield_booster`        | Shield Booster         | 3,000  | Max shield +30%                                   | 1     |
| `hull_reinforcement`    | Hull Reinforcement     | 2,500  | Max hull +25%                                     | 1     |
| `targeting_computer`    | Targeting Computer     | 4,000  | Weapon accuracy +10%, critical hit chance +5%     | 2     |
| `missile_decoy_system`  | Missile Decoy System   | 3,500  | Fires countermeasures; 2 charges per combat       | 2     |
| `energy_capacitor`      | Energy Capacitor       | 2,200  | Max energy +30%                                   | 2     |

### Communications / Intelligence

| Upgrade ID         | Name               | Cost   | Effect                                                     | Phase |
| ------------------ | ------------------ | ------ | ---------------------------------------------------------- | ----- |
| `signal_intercept` | Signal Intercept   | 5,000  | Passive IRN intercept in Low Sec (40%) / Black Sector (70%)| 1     |
| `comms_jammer`     | Comms Jammer       | 6,500  | Active: blocks IRN transmissions in current system         | 1     |

### Mining

| Upgrade ID        | Name                | Cost   | Compatible Classes         | Effect                              | Phase |
| ----------------- | ------------------- | ------ | -------------------------- | ----------------------------------- | ----- |
| `mining_laser`    | Mining Laser        | 3,000  | Courier, Freighter, Scout  | Enables asteroid mining             | 1     |
| `mining_laser_mk2`| Mining Laser Mk II  | 5,500  | Courier, Freighter, Scout  | Mining yield +60% (requires Mk I)   | 2     |
| `cargo_expander`  | Cargo Expander      | 1,500  | Courier, Freighter         | +15 cargo capacity (max 3 per ship) | 1     |

### Economy / Trade

| Upgrade ID      | Name           | Cost   | Effect                                                      | Phase |
| --------------- | -------------- | ------ | ----------------------------------------------------------- | ----- |
| `trade_scanner` | Trade Scanner  | 4,000  | Shows commodity prices at ports within 2 jumps without dock | 2     |

## 5.2 Upgrade Rules

* Each upgrade occupies exactly 1 slot
* `max_per_ship` limits stacking (most upgrades: max 1; `cargo_expander`: max 3)
* `requires` enforces prerequisites (e.g., `mining_laser_mk2` requires `mining_laser`)
* `compatible_classes` is `["all"]` unless class-restricted
* Upgrades purchased at ports — not all ports stock all upgrades
* Upgrades are not refundable (sold for 0 credits, slot freed)
* Installed upgrades persist with the ship across sessions and restarts

## 5.3 JSON Config Schema

`config/ships/upgrades.json`:

```json
{
  "version": "1.0",
  "upgrades": [
    {
      "upgrade_id": "enhanced_sensors",
      "name": "Enhanced Sensors",
      "description": "Extends sensor range by 50% and enables detection of transmissions in adjacent systems.",
      "category": "navigation",
      "base_cost": 2500,
      "slot_cost": 1,
      "max_per_ship": 1,
      "compatible_classes": ["all"],
      "requires": null,
      "effects": {
        "sensor_range_multiplier": 1.5,
        "adjacent_transmission_detection": true
      },
      "phase": 1
    },
    {
      "upgrade_id": "mining_laser",
      "name": "Mining Laser",
      "description": "Enables manual asteroid mining. Required for all mining operations.",
      "category": "mining",
      "base_cost": 3000,
      "slot_cost": 1,
      "max_per_ship": 1,
      "compatible_classes": ["courier", "freighter", "scout"],
      "requires": null,
      "effects": {
        "enables_mining": true
      },
      "phase": 1
    },
    {
      "upgrade_id": "cargo_expander",
      "name": "Cargo Expander",
      "description": "Adds 15 cargo capacity. Can be installed up to 3 times.",
      "category": "cargo",
      "base_cost": 1500,
      "slot_cost": 1,
      "max_per_ship": 3,
      "compatible_classes": ["courier", "freighter"],
      "requires": null,
      "effects": {
        "cargo_capacity_bonus": 15
      },
      "phase": 1
    }
  ]
}
```

---

# 6. Player Credits

Credits are the universal currency of BlackSector. They are stored server-side — there is no physical cash on a ship.

## 6.1 Storage

Credits are stored as an integer column on the `players` table:

```sql
ALTER TABLE players ADD COLUMN credits INTEGER NOT NULL DEFAULT 0;
```

All credit transactions are atomic database writes committed immediately (not deferred to tick flush).

## 6.2 Credit Sources

* Trading: buy low, sell high at ports
* Mining: sell extracted resources at ports
* Missions: completion rewards
* Exploration: sell discovery data at ports with data markets
* Insurance payout: received on ship destruction (standard death mode)

## 6.3 Credit Sinks

* Purchasing ships at shipyards
* Purchasing and installing upgrades
* Docking fees at some ports
* Purchasing drones

## 6.4 Credit Display

Credits are displayed in the player HUD as `Cr` (e.g., `Cr 14,250`).

Player commands:
```
balance            — show current credits
```

---

# 7. New Player Starting State

## 7.1 Starting Location

New players begin in **Federated Space** at the origin starbase. The specific system is configured in `server.json`:

```json
"federated_origin_system_id": 1
```

The origin starbase is a large trading port with full services: shipyard, upgrade market, commodity market, drone market.

## 7.2 Starting Resources

Starting values are configured in `server.json`:

```json
"new_player": {
  "starting_credits": 10000,
  "starter_ship_class": "courier",
  "starter_ship_free": true
}
```

On registration:
1. Player account is created
2. A Courier-class ship is provisioned at the origin starbase (free — no credits deducted if `starter_ship_free: true`)
3. Starting credits are deposited
4. Player enters the game docked at the origin starbase

With 10,000 starting credits and a free Courier, a new player can immediately:
* Buy supplies and begin trade runs
* Purchase a mining laser and start prospecting
* Explore nearby systems

## 7.3 Starter Ship Provisioning

The starter ship is created with base stats from the ship class config. No upgrades pre-installed.

```
Ship:     Courier (free)
Hull:     100/100
Shield:   50/50
Energy:   100/100
Cargo:    0/20
Upgrades: (none)
Status:   DOCKED — Federated Station Alpha (System 1)
```

---

# 8. Ship Destruction and Respawn

## 8.1 Death Mode Configuration

Death behavior is set globally in `server.json`:

```json
"death": {
  "mode": "standard",
  "standard": {
    "respawn_system_id": 1,
    "cargo_loss": true,
    "ship_loss": true,
    "insurance_payout_percent": 30,
    "minimum_credits_floor": 3000
  },
  "permadeath": {
    "account_action": "reset"
  }
}
```

## 8.2 Standard Death Mode

When a ship is destroyed in standard mode:

**Immediate effects (tick of destruction):**
1. Ship status set to `DESTROYED` in database
2. Cargo dropped as a lootable debris field at the destruction position (other players can salvage it for `cargo_salvage_window_ticks` ticks, then it expires)
3. Ship upgrades are lost (not salvageable)
4. Destruction logged as a game event

**Credit settlement (same tick):**
1. Insurance payout calculated: `floor(ship_class.base_price × insurance_payout_percent / 100)`
2. Payout credited to player's account
3. If `credits + payout < minimum_credits_floor`, credits are topped up to the floor value
4. Ensures player can always afford a new starter ship

**Respawn:**
1. Player is teleported to the respawn system (default: Federated Space origin)
2. A new Courier-class ship is provisioned at the respawn starbase (player pays for it using their retained credits)
3. Player receives a notification on next tick or login:

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  YOUR SHIP HAS BEEN DESTROYED.

  Location:  System 14 (Vega Prime)
  Cause:     Weapons fire (attacker: ghost)

  Insurance payout: 2,400 Cr (30% of Courier base value)
  Credits retained: 14,250 Cr
  You have been transferred to Federated Station Alpha.

  Visit the shipyard to purchase a new vessel.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**Summary: you lose the ship, you lose the cargo, you keep the money.**

## 8.3 Permadeath Mode

When a ship is destroyed in permadeath mode:

1. Ship destroyed (same as standard)
2. Cargo lost (same as standard, lootable)
3. Credits lost
4. Account status set to `permadead`
5. On next login attempt, player receives:

```
Your pilot did not survive.
Account: nova — Status: PERMADEAD

You may create a new character by reconnecting with a new SSH username.
```

6. The `player_name` is freed after 30 days (configurable), allowing it to be claimed by a new registration

Note: Admins can override permadeath via `bsctl player revive <name>` to restore an account in exceptional cases.

## 8.4 PvP Death Restrictions

Ship destruction by another player (PvP) is only possible in:
* Low Security systems (opt-in — both parties must have engaged)
* Black Sector (always enabled)
* Federated Space: PvP disabled — NPC enforcers intervene

NPC pirates can destroy ships in all zones except Federated Space.

## 8.5 Ship Purchase Flow

After death (or any time the player wants a new ship), they dock at a port with a shipyard:

```
shipyard                     — list available ships at current port
shipyard buy <class>         — purchase a ship (replaces current if DESTROYED, or adds if allowed)
shipyard info <class>        — view stats and price for a ship class
```

Only ports with `has_shipyard: true` (set in world config) sell ships. The Federated Space origin starbase always has a shipyard.

---

# 9. Data Model Changes

## 9.1 players table

Add `credits` column:

```sql
ALTER TABLE players ADD COLUMN credits INTEGER NOT NULL DEFAULT 0;
```

## 9.2 ship_drone_inventory table

Drones stored as undeployed cargo on a ship are tracked in a separate table (not in `ship_cargo`, which is for commodities only):

```sql
CREATE TABLE ship_drone_inventory (
  ship_id      TEXT NOT NULL REFERENCES ships(ship_id),
  drone_type   TEXT NOT NULL,   -- mapping | prospecting | decoy | relay
  quantity     INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (ship_id, drone_type)
);
```

When a drone is deployed, a row is removed from this table and a `drones` record is created. When recalled (drone docks at port), the reverse occurs.

---

# 10. Server Configuration Reference

All new fields added to `server.json`:

```json
{
  "federated_origin_system_id": 1,

  "new_player": {
    "starting_credits": 10000,
    "starter_ship_class": "courier",
    "starter_ship_free": true
  },

  "death": {
    "mode": "standard",
    "standard": {
      "respawn_system_id": 1,
      "cargo_loss": true,
      "ship_loss": true,
      "insurance_payout_percent": 30,
      "minimum_credits_floor": 3000,
      "cargo_salvage_window_ticks": 50
    },
    "permadeath": {
      "account_action": "reset",
      "name_release_days": 30
    }
  }
}
```

---

# 11. Protocol Messages

## Ship status update (Server → Client)

```json
{
  "type": "ship_status_update",
  "timestamp": 8400,
  "correlation_id": null,
  "payload": {
    "ship_id": "...",
    "hull_points": 0,
    "max_hull_points": 100,
    "status": "DESTROYED",
    "cause": "weapons_fire",
    "attacker_name": "ghost",
    "insurance_payout": 2400,
    "credits_after": 14250,
    "respawn_system_id": 1
  }
}
```

## Shipyard purchase command (Client → Server)

```json
{
  "type": "command_submit",
  "timestamp": 8500,
  "correlation_id": "...",
  "payload": {
    "command": "shipyard_buy",
    "parameters": {
      "class_id": "courier"
    }
  }
}
```

---

# 12. Player Commands

```
balance                      — show current credit balance

shipyard                     — list ships available at current port (must be docked)
shipyard info <class>        — view stats and price
shipyard buy <class>         — purchase a ship

ship                         — show current ship stats
ship upgrades                — list installed upgrades and empty slots
ship upgrade buy <upgrade_id>— purchase and install an upgrade (must be docked at port with upgrade market)
ship upgrade remove <upgrade_id> — remove upgrade (slot freed, credits not refunded)
ship cargo                   — show cargo manifest
```

---

# 13. Balancing Guidelines

* Starting credits (10,000) should feel like enough to get started but not enough to skip progression
* Courier base price (8,000) ensures new players can replace their starter ship after one moderate trade run
* Insurance at 30% means death is a setback, not a total wipe — keeps players in the game
* The credits floor (3,000) ensures a player can always buy a new starter ship no matter how poor they are after death
* Upgrade costs should scale with impact — comms upgrades (signal intercept, jammer) are expensive because they shift information asymmetry
* Cargo expander max-3 prevents infinite cargo accumulation — a freighter with 3 expanders maxes at 105 units, maintaining class identity

---

# 14. Non-Goals (v1)

* Ship insurance as a purchasable policy (flat payout only)
* Ship loans or credit debt
* Crew/officer system
* Ship customization (paint, name beyond player choice)
* Multiple ships simultaneously active per player
* Ship trade between players (ships are not transferable)

---

# 15. Future Extensions

* Ship naming (player assigns name, displayed in combat log)
* Faction-specific ship classes (unlocked by reputation)
* Ship crafting / modification at specialized ports
* Black market ship acquisition in Low Sec / Black Sector
* Ship insurance policies (purchasable, higher payout tiers)
* Ship storage at ports (park a second ship while flying another)
* Bounty system on ship destruction (PvP kills yield a credit reward)

---

# End of Document
