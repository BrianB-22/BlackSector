# Combat Math Specification
## Version: 0.1
## Status: Draft
## Owner: Core Simulation
## Last Updated: 2026-03-02

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

# End of Document
