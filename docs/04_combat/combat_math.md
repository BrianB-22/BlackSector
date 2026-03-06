# Combat Math Specification
## Version: 0.2
## Status: Draft
## Owner: Core Simulation
## Last Updated: 2026-03-05

---

# 1. Purpose

Defines all numerical formulas governing combat resolution.

This document contains no behavioral flow.

---

# 2. Scope

IN SCOPE:
- Detection
- Tracking
- Heat
- Energy
- Firing solution
- Hit probability
- Damage resolution

OUT OF SCOPE:
- UI interpretation
- Combat lifecycle
- Scanning mechanics (see scanning_system.md)

---

# 3. Core Variables

TrackingConfidence ∈ [0–1]  
SensorStrength ∈ [1–100]  
HeatLevel ∈ [0–100]  
SignatureRadius ∈ [0.5–2.0]  
Distance ∈ [0–10000]

---

# 4. Detection

DistanceFactor = 1 + (Distance / 5000)

DetectionScore =
(SensorStrength / 100)
× SignatureRadius
÷ DistanceFactor
− (JammingStrength / 100)

DetectionSuccess if ≥ 0.25

---

# 5. Tracking

TrackingGain =
(SensorStrength / 100)
× 0.08
× (1 − HeatPenalty)
÷ DistanceFactor

TrackingDisruption =
(TargetEvasionRating / 100)
× TargetManeuverFactor

TrackingConfidence(t+1) =
TrackingConfidence(t)
+ TrackingGain
− TrackingDisruption

Clamped [0–1]

---

# 6. Heat

HeatPenalty = (HeatLevel / 100)²

HeatDecay = 5 per tick

HeatPenalty affects:
- TrackingGain
- WeaponAccuracy
- ShieldRecharge

---

# 7. Firing Solution

SolutionQuality =
TrackingConfidence
× WeaponAccuracy
× (1 − HeatPenalty)

---

# 8. Hit Probability

HitChance =
SolutionQuality
÷ (1 + (TargetEvasionRating / 100))

Clamped [0.05–0.95]

---

# 9. Damage

RawDamage = WeaponBaseDamage × SolutionQuality

ShieldDamage =
min(ShieldLevel, RawDamage)

HullDamage =
(RawDamage − ShieldDamage)
× (1 − ArmorReduction)

ArmorReduction = ShipArmor / 200

---

# 10. Tunable Parameters

- Tracking multiplier (0.08)
- HeatDecay (5)
- Detection threshold (0.25)
- Heat exponent (2)
- Hit chance caps
- MissileSpeed

---

# 11. Balance Constraints

- Heat above 70 significantly reduces effectiveness
- Early firing discouraged
- No guaranteed invulnerability

---

---

# 12. Phase 1 Simplified Model

**Phase 1 does not use the full tracking/heat/solution quality pipeline.**

The Phase 1 vertical slice uses a direct simplified model to reduce implementation scope. The formulas below replace Sections 4–9 for Phase 1 only.

## 12.1 Phase 1 Hit Check

```
HitSuccess = random() < accuracy
```

No tracking confidence, no heat penalty, no solution quality modifier.

| Actor    | Base Accuracy |
| -------- | ------------- |
| Courier  | 0.65          |
| Raider   | 0.60          |
| Marauder | 0.65          |

`random()` draws from the tick's seeded PRNG (see `headless_server_spec.md` Section 15).

## 12.2 Phase 1 Damage

**Player (Courier):** fixed damage, no variance.

```
PlayerDamage = weapon_damage    // 15 for Courier
```

**NPC:** uniform random in damage range.

```
NPCDamage = min_damage + floor(random() × (max_damage - min_damage + 1))
```

| Tier     | min | max |
| -------- | --- | --- |
| Raider   | 12  | 18  |
| Marauder | 18  | 28  |

## 12.3 Phase 1 Turn Order

Both sides fire simultaneously. Resolution sequence per tick:

1. Pirate fires at player → resolve hit/damage
2. If player issued `attack`: player fires at pirate → resolve hit/damage
3. If player issued `flee`: resolve flee chance (see `engagement_flow.md` §13.3)
4. Apply damage to both sides
5. Check destruction (`hull_points <= 0`)

## 12.4 Phase 1 Shield/Hull Application

Unchanged from full model (Section 9). Shield absorbs first, overkill passes to hull.

## 12.5 Phase 1 Energy

Weapons have no energy cost in Phase 1. Energy is only consumed by jumps.

## 12.6 Mutual Kill

If both parties reach `hull_points <= 0` in the same tick:
- Player is DESTROYED (death/respawn applies)
- NPC is removed (no kill credit in mutual kill)

---

# End of Document
