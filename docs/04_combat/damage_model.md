# Damage Model Specification

## Version: 0.1

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

Damage pipeline:

```
incoming damage
→ shield mitigation
→ armor mitigation
→ hull damage
→ subsystem check
```

Subsystem failures may include:

* sensor damage
* propulsion failure
* weapon disruption

---

# 5. Tunable Parameters

* ShieldCapacity
* ArmorStrength
* HullIntegrity
* SubsystemFailureChance

---

# 6. Balancing Guidelines

Small ships:

* lower durability
* higher evasion

Large ships:

* greater durability
* slower response

---

# End of Document
