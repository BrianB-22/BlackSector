# Missile System Specification

## Version: 0.1

## Status: Draft

## Owner: Combat Systems

## Last Updated: 2026-03-03

---

# 1. Purpose

The Missile System governs long-range guided weapons.

Missiles represent a primary offensive weapon type in long-range combat.

---

# 2. Core Concepts

Missile Lock
Target must be detected and tracked.

Missile Flight
Missile travels toward target over time.

Missile Interception
Missiles may be destroyed by countermeasures.

---

# 3. Core Mechanics

Missile lifecycle:

```
target lock
→ missile launch
→ flight tracking
→ interception checks
→ impact resolution
```

---

# 4. Variables

* missile_speed
* missile_range
* guidance_accuracy
* warhead_damage
* missile_capacity (per ship — see `docs/01_architecture/ship_system.md` Section 4)

Missile capacity is the maximum number of missiles a ship can carry. It is defined per ship class in `config/ships/ship_classes.json` as `missile_capacity`. Current values:

| Class     | Missile Capacity |
| --------- | ---------------- |
| Courier   | 4                |
| Scout     | 4                |
| Freighter | 2                |
| Fighter   | 12               |

Missiles are purchased at ports and consume a slot from the ship's magazine. Fired missiles are expended and must be restocked at port.

---

# 5. Counterplay

Missiles may be defeated through:

* point-defense systems
* decoys
* electronic countermeasures

---

# End of Document
