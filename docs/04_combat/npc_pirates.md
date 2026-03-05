# NPC Pirates Specification

## Version: 0.1
## Status: Draft
## Owner: Combat / Core Simulation
## Last Updated: 2026-03-05

---

# 1. Purpose

Defines NPC pirate behavior for Phase 1: spawn rules, stats, engagement logic, and outcomes.

NPC pirates are the primary source of combat threat in Phase 1. They make Low Security space dangerous without requiring other players.

---

# 2. Scope

## IN SCOPE

- Spawn rules and frequency
- Basic engagement AI (attack, pursue, disengage)
- Pirate stats (hull, shield, weapons)
- Combat outcome handling
- Integration with security zone spawn rates

## OUT OF SCOPE (Phase 1)

- Pirate loot drops (Phase 2)
- Named pirate factions
- Pirate base structures
- Coordinated pirate fleets
- Pirate bounty / wanted system
- Pirate dialogue or negotiation

---

# 3. Design Principles

- Pirates exist to make Low Security feel dangerous.
- Phase 1 AI is intentionally simple — attack on sight, flee when near death.
- Pirate difficulty scales with security zone.
- No loot in Phase 1 — death penalty is cargo/ship loss, not kill reward.
- Pirates are ephemeral — they spawn per-encounter, not as persistent world entities.

---

# 4. Spawn Rules

## 4.1 Zones

Pirates spawn only in:

| Zone            | Spawn Enabled | Notes                          |
| --------------- | ------------- | ------------------------------ |
| Federated Space | No            | NPC patrols enforce no pirates |
| High Security   | Rare          | Low probability                |
| Medium Security | Moderate      | Phase 2 only                   |
| Low Security    | Frequent      | Phase 1 active zone            |
| Black Sector    | Maximum       | Phase 2 only                   |

## 4.2 Spawn Trigger

Pirate encounters are checked once per player tick-action while the player is in an eligible system and not docked.

Spawn check formula (from `security_zones.md`):

```
PirateSpawnChance = PirateActivityBase × (1 − SecurityRating)
```

`PirateActivityBase` is configured in `server.json`.

Default values:

```json
"npc_pirates": {
  "activity_base": 0.10,
  "spawn_check_interval_ticks": 5
}
```

This gives Low Security (SecurityRating ~0.2) a base spawn chance of ~8% per check interval.

## 4.3 Encounter Initiation

When a spawn check succeeds:

1. A pirate NPC is instantiated in the player's current system
2. Combat engagement begins immediately on the next tick
3. The player receives a notification:

```
⚠  PIRATE INTERCEPT — A hostile vessel has locked on to your position.
```

---

# 5. Pirate Stats

Pirate stats scale with the security zone. Phase 1 (Low Security only):

| Tier        | Hull | Shield | Weapon Damage | Accuracy | Flee Threshold |
| ----------- | ---- | ------ | ------------- | -------- | -------------- |
| Raider      | 60   | 20     | 12–18         | 60%      | 15% hull       |
| Marauder    | 90   | 40     | 18–28         | 65%      | 10% hull       |

Tier is selected randomly at spawn with weighted probability:

```
Raider:   70%
Marauder: 30%
```

Pirate stats are configurable in `config/npcs/pirate_tiers.json`.

---

# 6. Combat AI Behavior

Pirate AI follows a fixed state machine:

```
SPAWNED → ENGAGING → (FLEEING | DESTROYED)
```

## 6.1 ENGAGING

- Every combat tick: attempts weapon fire against player
- Uses standard weapon damage resolution (see `combat_math.md`)
- No shield/hull repair during combat
- Does not use missiles or special weapons in Phase 1

## 6.2 Flee Condition

When pirate hull drops to or below `flee_threshold`:

- Pirate transitions to FLEEING state
- Combat ends — player may not re-engage
- Pirate is removed from the world after fleeing

```
PIRATE RETREATS — The hostile vessel has broken off and fled.
```

## 6.3 DESTROYED

When pirate hull reaches 0:

- Pirate is removed from world
- No loot dropped (Phase 1)
- Player receives kill notification:

```
HOSTILE DESTROYED — You have neutralized the threat.
```

- Event logged to `game_events` table

---

# 7. Player Options During Pirate Encounter

Standard combat commands apply (see `engagement_flow.md`). Additionally:

```
surrender          — end combat, pirate takes a portion of wallet credits
```

**Surrender outcome:**
- Combat ends without ship destruction
- Pirate takes `floor(wallet_credits × surrender_loss_percent / 100)` from wallet
- Default `surrender_loss_percent`: 40 (configurable)
- Bank accounts are not touched

Configured in `server.json`:

```json
"npc_pirates": {
  "surrender_loss_percent": 40
}
```

---

# 8. Data Model

NPC pirates are **ephemeral** — they are not persisted to the database. A pirate exists only for the duration of a single combat encounter and is discarded when combat ends (flee or destroy).

Pirate encounters are logged to the game event log:

```
event_type: "pirate_encounter"
event_type: "pirate_destroyed"
event_type: "pirate_fled"
event_type: "player_surrendered_to_pirate"
```

---

# 9. Integration Points

- **Security Zones** (`docs/02_universe/security_zones.md`): spawn rate formula and zone eligibility
- **Combat Math** (`docs/04_combat/combat_math.md`): damage resolution, hit probability
- **Engagement Flow** (`docs/04_combat/engagement_flow.md`): turn structure, command processing
- **Ship System** (`docs/01_architecture/ship_system.md`): ship destruction on player death

---

# 10. Configuration Reference

Full `server.json` NPC pirate block:

```json
"npc_pirates": {
  "activity_base": 0.10,
  "spawn_check_interval_ticks": 5,
  "surrender_loss_percent": 40
}
```

Tier stats in `config/npcs/pirate_tiers.json`:

```json
{
  "version": "1.0",
  "tiers": [
    {
      "tier_id": "raider",
      "hull_points": 60,
      "shield_points": 20,
      "weapon_damage_min": 12,
      "weapon_damage_max": 18,
      "accuracy_percent": 60,
      "flee_threshold_percent": 15,
      "spawn_weight": 70
    },
    {
      "tier_id": "marauder",
      "hull_points": 90,
      "shield_points": 40,
      "weapon_damage_min": 18,
      "weapon_damage_max": 28,
      "accuracy_percent": 65,
      "flee_threshold_percent": 10,
      "spawn_weight": 30
    }
  ]
}
```

---

# 11. Non-Goals (Phase 1)

- Pirate loot / cargo drops
- Named factions or lore-attached pirates
- Pirate bases or home systems
- Pirate bounty system
- Multiple pirates per encounter
- Pirate negotiation or dialogue

---

# 12. Future Extensions (Phase 2+)

- Loot drops (cargo, credits, contraband)
- Pirate faction affiliations with territory
- Coordinated multi-ship ambush encounters
- Wanted / bounty system
- Black Sector elite pirate variants
- Pirate patrol routes in Low Security

---

# End of Document
