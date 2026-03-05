# Damage Model Specification

## Version: 0.2

## Status: Draft

## Owner: Combat Systems

## Last Updated: 2026-03-03

---

# 1. Purpose

The Damage Model defines how ships absorb, mitigate, and suffer damage during combat.

Ships are composed of layered defensive systems that degrade over time.

---

# 2. Core Damage Layers

Ships typically contain:

1. Shields
2. Armor
3. Hull
4. Internal Systems

Damage is applied sequentially through these layers.

---

# 3. Damage Types

Common damage types include:

* kinetic
* explosive
* energy
* electromagnetic

Each type interacts differently with defensive systems.

---

# 4. Core Mechanics

## 4.1 Damage Pipeline

```
incoming damage
→ shield absorption (Phase 1: shields absorb first)
→ overkill passthrough → hull damage
→ subsystem check (Phase 2)
```

**Phase 1 simplification:** Armor layer is not implemented in Phase 1. Damage passes directly from shields to hull.

## 4.2 Shield Absorption

Shields are a hit-point pool (not a percentage mitigation). All incoming damage is subtracted from `shield_points` first.

```
shield_damage = min(incoming_damage, shield_points)
passthrough   = incoming_damage - shield_damage
hull_damage   = passthrough
```

**Overkill:** If incoming damage exceeds remaining shield points, the remainder applies fully to hull. No damage is absorbed by a depleted shield.

Example: 30 damage incoming, shield at 20 → shield goes to 0, hull takes 10.

## 4.3 Shield Recharge

Shields do **not** recharge during active combat (while `ship.status = IN_COMBAT`).

Out of combat, shields recharge at a fixed rate per tick:

```
shield_regen_per_tick = 5   (configurable in server.json)
```

Shield recharge is capped at `max_shield_points`. Recharge begins on the first tick after combat ends.

Configured in `server.json`:

```json
"combat": {
  "shield_regen_per_tick": 5
}
```

## 4.4 Hull Damage

Hull has no regeneration. Hull can only be repaired by docking at a port with repair services.

Hull repair cost: `floor((max_hull - current_hull) × hull_repair_cost_per_point)`

Configured in `server.json`:

```json
"combat": {
  "hull_repair_cost_per_point": 10
}
```

## 4.5 Ship Destruction

When `hull_points` reaches 0:

- Ship status set to `DESTROYED`
- Combat ends immediately
- Death/respawn rules apply (see `docs/01_architecture/ship_system.md` Section 8)

## 4.6 Subsystem Failures (Phase 2)

Phase 2 may introduce subsystem damage:

* sensor damage (reduced detection range)
* propulsion failure (cannot jump until repaired)
* weapon disruption (weapon offline for N ticks)

Not implemented in Phase 1.

---

# 5. Tunable Parameters

| Parameter                  | Default | Location    |
| -------------------------- | ------- | ----------- |
| `shield_regen_per_tick`    | 5       | server.json |
| `hull_repair_cost_per_point` | 10    | server.json |

---

# 6. Balancing Guidelines

Small ships (Scout, Courier):

* Lower total HP (hull + shield)
* Shields matter more — less hull to fall back on
* Regen recovers them faster proportionally

Large ships (Fighter, Freighter):

* Higher total HP
* Fighter: high shield + hull — survives sustained fire
* Freighter: low shield, moderate hull — relies on fleeing

---

# End of Document
