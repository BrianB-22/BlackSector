# Weapon System Specification

## Version: 0.1
## Status: Draft
## Owner: Core Simulation

## Last Updated: 2026-03-02

---

# 1. Purpose

Defines all offensive ship-mounted weapon systems, including:

- Space torpedoes (long-range ordnance)
- Energy weapons (e.g., mining lasers repurposed for combat)
- EMF/EMP-style disabling weapons (non-lethal piracy tools)

Weapons are designed to support multiple playstyles:

- Long-range destruction
- Sustained pressure
- Precision subsystem targeting (future)
- Ship incapacitation for piracy

---

# 2. Scope

IN SCOPE:

- Weapon classes
- Weapon attributes
- Projectile lifecycle integration
- Damage vs disable mechanics
- Energy \& heat costs
- Combat integration

OUT OF SCOPE:

- Fleet weapons
- Planetary bombardment
- Boarding mechanics (future)
- Subsystem damage modeling (future v1.1)

---

# 3. Design Principles

- Weapons must reflect long-range, sensor-based combat.
- No weapon should be universally dominant.
- Torpedoes reward tracking discipline.
- Energy weapons reward sustained control.
- EMF weapons reward tactical piracy.
- Heat \& energy costs enforce tradeoffs.

---

# 4. Core Concepts

- WeaponClass
- WeaponAccuracy
- WeaponBaseDamage
- ProjectileSpeed
- DisableStrength
- ShieldInteractionType
- WeaponHeatCost
- WeaponEnergyCost

---

# 5. Data Model

## Entity: WeaponModule

- weapon\_id
- weapon\_class
- base\_damage
- weapon\_accuracy
- projectile\_speed
- energy\_cost
- heat\_cost
- cooldown\_ticks
- disable\_strength (if applicable)
- shield\_modifier
- hull\_modifier

Persistent ship module.

---

# 6. Weapon Classes

---

## 6.1 Space Torpedoes

Role:

Primary long-range destructive weapon.

Characteristics:

- High base damage
- Slower projectile speed
- Long intercept time
- High energy cost
- High heat cost
- High shield impact

Use Case:

- Finishing blows
- Breaking shields
- High-risk engagement escalation

Strengths:

- Massive burst damage
- Strong shield penetration

Weaknesses:

- Requires strong tracking
- Heat-heavy
- Slow travel time allows evasion

---

## 6.2 Rail / Kinetic Weapons

Role:

Mid-damage sustained weapon.

Characteristics:

- Moderate damage
- Faster projectile
- Lower heat cost
- Short cooldown

Use Case:

- Sustained pressure
- Efficient combat pacing

Strengths:

- Heat-efficient
- Reliable DPS

Weaknesses:

- Lower burst potential

---

## 6.3 Mining Lasers (Repurposed)

Role:

Dual-purpose tool weapon.

Characteristics:

- Low direct hull damage
- Higher shield interaction
- Continuous beam (no projectile travel)
- Moderate heat cost
- No travel delay

Use Case:

- Sustained shield pressure
- Close-range finishing
- Multi-role mining/combat loadout

Strengths:

- Reliable hit resolution
- No intercept window

Weaknesses:

- Low burst damage
- Inefficient vs armored hull

---

## 6.4 EMF / EMP Disruptors

Role:

Non-lethal disable weapon for piracy.

Characteristics:

- Low hull damage
- Moderate shield interaction
- Applies DisableEffect
- Medium energy cost
- Moderate heat cost

DisableEffect may:

- Reduce MaxVelocity
- Reduce EnergyRegen
- Disable Jump
- Increase Heat accumulation
- Reduce TrackingGain

Use Case:

- Ship capture
- Piracy
- Prevent disengagement

Strengths:

- Control-based playstyle
- Enables capture scenarios

Weaknesses:

- Low lethal potential
- Requires follow-up control

---

# 7. Core Mechanics

---

## 7.1 Firing Process

1\. Validate energy availability.

2\. Validate cooldown.

3\. Apply heat cost.

4\. Spawn projectile (if applicable).

5\. Projectile resolves per combat\_math.

Beam weapons resolve instantly within tick.

---

## 7.2 Projectile Lifecycle

CREATED  

→ IN\_FLIGHT  

→ IMPACT | MISS  

→ CLEANUP  

ProjectileSpeed determines intercept ticks.

---

## 7.3 Shield \& Hull Interaction

RawDamage =

WeaponBaseDamage × SolutionQuality

Apply ShieldModifier and HullModifier if defined.

Example:

- Torpedo: ShieldModifier 1.2
- Mining Laser: ShieldModifier 1.1, HullModifier 0.7
- EMP: ShieldModifier 0.8, HullModifier 0.3

---

## 7.4 Disable Effects (EMP)

DisableStrength ∈ \[0–1]

DisableDurationTicks =

BaseDisableDuration × DisableStrength × SolutionQuality

Effects stack but are capped.

Disable effects cannot fully freeze ship in v1.

---

# 8. Mathematical Integration

Hit probability, tracking, and damage resolution defined in `combat\_math.md`.

Disable effects applied after damage resolution.

---

# 9. Tunable Parameters

- WeaponBaseDamage
- ProjectileSpeed
- EnergyCost
- HeatCost
- CooldownTicks
- ShieldModifier
- DisableDuration base
- Disable cap limits

---

# 10. Integration Points

Depends On:

- Combat System
- Combat Math
- Heat \& Energy Model
- Engagement Flow

Exposes:

- WeaponFireEvent
- ProjectileSpawnEvent
- DisableEffectAppliedEvent

---

# 11. Failure \& Edge Cases

If Energy < EnergyCost:

Weapon does not fire.

If ship overheated:

Accuracy penalty applies.

If target destroyed mid-flight:

Projectile resolves harmlessly.

Disable effects removed upon destruction.

---

# 12. Performance Constraints

- Projectile tracking O(1) per instance
- No per-frame physics simulation
- No cross-system projectile persistence

---

# 13. Security Considerations

- All damage resolved server-side.
- Disable effects validated per tick.
- Cooldowns enforced server-side.

---

# 14. Telemetry \& Logging

Log:

- Weapon usage frequency
- Hit probability distribution
- EMP disable rates
- Damage breakdown by class

Used for balancing.

---

# 15. Balancing Guidelines

- Torpedoes must punish poor tracking.
- Mining lasers must not outperform combat weapons.
- EMP must enable piracy but not create permanent lockdown.
- Heat must discourage torpedo spam.
- Disable stacking must be capped.

---

# 16. Non-Goals (v1)

- Boarding mechanics
- Subsystem destruction
- Area-of-effect weapons
- Nuclear-scale weapons

---

# 17. Future Extensions

- Smart torpedoes
- Decoy countermeasures
- Subsystem targeting
- Capture mechanics tied to EMP threshold
- Armor penetration classes

---

# End of Document
