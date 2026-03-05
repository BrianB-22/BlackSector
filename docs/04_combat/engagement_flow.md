# Engagement Flow Specification

## Version: 0.3
## Status: Draft
## Owner: Core Simulation

## Last Updated: 2026-03-05

---

# 1. Purpose

Defines the temporal lifecycle of a combat encounter, from first detection through disengagement or destruction.

This document specifies:

- Detection-only state
- Tracking buildup prior to formal engagement
- Scan-only interactions
- Silent shadowing behavior
- Pre-attack jump mechanics
- Engagement lock
- Disengagement flow

No mathematical formulas are defined here. See `combat\_math.md` for numeric models.

---

# 2. Scope

IN SCOPE:

- Engagement state transitions
- Pre-combat interaction rules
- Combat lock conditions
- Jump escape constraints
- Disengagement sequence
- Destruction sequence

OUT OF SCOPE:

- Combat math formulas
- Economic consequences
- Fleet mechanics
- NPC AI decision logic

---

# 3. Design Principles

- Space is vast; ships are small.
- Visual-range dogfighting is not the baseline.
- Combat is sensor-driven, not visually driven.
- Long-range standoffs are normal.
- Engagement is information warfare before weapons fire.
- Combat must be avoidable prior to formal lock.
- All transitions are deterministic and tick-driven.

---

# 4. Spatial \& Engagement Philosophy

## 4.1 Scale Assumption

The simulation assumes:

- Engagement distances are measured in thousands of kilometers.
- Ships cannot visually identify each other at combat ranges.
- All meaningful interaction occurs through:

     - Sensor returns

     - Signature analysis

     - Tracking confidence

     - Electronic warfare

Close-range combat is rare and typically the result of:

- Successful pursuit
- Failed disengagement
- Deliberate high-risk maneuvering

---

## 4.2 Sensor-Driven Interaction Model

Before weapons fire:

- Ships interact through detection and tracking.
- Targeting solutions rely on predicted vectors.
- Scan resolution depends on signal clarity.
- Heat, jamming, and silent mode shape information quality.

Visual confirmation is not a gameplay mechanic.

All UI representations are abstractions of sensor data.

---

# 5. Engagement State Machine

NO\_CONTACT  

→ DETECTION\_ONLY  

→ TRACKING  

→ ENGAGED  

→ RESOLVING  

→ DISENGAGED | DESTROYED

Each state transition occurs inside the tick engine.

---

# 6. Detection-Only State

## 6.1 Entry Condition

- DetectionSuccess = true
- No hostile action initiated

## 6.2 Characteristics

- Ships aware of sensor contact
- Passive scan allowed
- Active scan allowed
- No weapon fire permitted
- Jump remains available
- Long-range standoff typical

At this stage, ships may be thousands of kilometers apart.

---

# 7. Tracking State (Pre-Engagement)

## 7.1 Entry Condition

- Detection persists
- TrackingConfidence > 0
- No weapons fired

## 7.2 Characteristics

- Tracking builds over time
- Scan-only interactions permitted
- Silent shadowing possible
- Jump remains available
- Ships still separated by long-range distances

Tracking represents predictive modeling, not visual aim.

---

# 8. Scan-Only Interaction Phase

During DETECTION\_ONLY or TRACKING:

Players may:

- Channel active scans
- Boost sensors
- Jam target
- Enter silent mode
- Adjust vector for future intercept

Weapons may NOT be fired.

This phase represents electronic maneuvering at distance.

---

# 9. Silent Shadowing

Silent shadowing represents:

- Maintaining sensor lock while minimizing signature.
- Avoiding escalation.
- Remaining outside optimal firing solution windows.

Characteristics:

- Reduced signature
- Reduced sensor strength
- Reduced velocity
- Harder to detect
- Slower tracking gain

Silent ships are not invisible — only harder to model precisely.

If a silent ship fires:

Immediate transition to ENGAGED.

---

# 10. Pre-Attack Jump Mechanics

Because space is vast, disengagement is often achieved via jump.

A ship may jump if:

- Not in ENGAGED state
- Energy ≥ JumpEnergyCost
- Valid jump corridor exists
- No active hostile fire exchange initiated

Jump represents strategic repositioning, not instant teleportation.

Jump consumes:

- Significant energy
- Generates heat
- Locks ship for 1 tick during transition

If weapon fire has occurred:

Jump is disabled until disengagement conditions met.

---

# 11. Engagement Lock

## 11.1 Trigger Conditions

ENGAGED begins when:

- A weapon is fired
- Offensive module activated
- Pirate ambush occurs
- Mutual aggression confirmed

## 11.2 Effects

- Jump disabled
- Docking disabled
- Trade disabled
- CombatInstance created
- Engagement distance fixed to combat abstraction
- Sensor warfare transitions to full combat mode

Even during ENGAGED, ships are still separated by long-range distance.

Close-range engagement is an abstraction of tracking convergence.

---

# 12. Projectile Lifecycle

Projectiles represent long-range ordnance.

States:

CREATED  

→ IN\_FLIGHT  

→ IMPACT | MISS  

→ CLEANUP  

Travel time represents long-range intercept windows.

Disengagement does not cancel in-flight projectiles.

---

# 13. Disengagement Flow

## 13.1 Conditions

- Range exceeds disengage threshold
- TrackingConfidence below threshold
- No weapons fired for X ticks

## 13.2 Countdown

Conditions must persist for N ticks.

Any weapon fire or hard maneuver resets countdown.

## 13.3 Flee Command (Phase 1 — NPC Combat)

During NPC pirate combat, the player may issue the `flee` command to attempt an emergency disengage.

```
flee                — attempt to break combat immediately
```

Flee is resolved once per tick as a probability check:

```
FleeChance = flee_base_chance + (energy_current / max_energy × flee_energy_bonus)
```

Default values (configured in `server.json`):

```json
"combat": {
  "flee_base_chance": 0.35,
  "flee_energy_bonus": 0.25
}
```

This gives a range of ~35% (no energy) to ~60% (full energy).

**On flee success:**

- CombatInstance → DISENGAGED
- Jump re-enabled (no cooldown on successful flee)
- Player notified: `You have broken off the engagement and escaped.`

**On flee failure:**

- Combat continues for this tick
- Pirate fires normally
- Player may attempt `flee` again next tick

**Flee is not available** while a jump is already in progress or after `surrender` has been issued.

**Flee does not prevent pirate weapon fire on the same tick** — the pirate fires before the flee check resolves.

## 13.4 Resolution

On success (flee or natural disengage):

- CombatInstance → DISENGAGED
- Jump re-enabled
- Short cooldown applied (natural disengage only)

Disengagement represents successful break of predictive model, not visual escape.

---

# 14. Destruction Flow

If HullLevel ≤ 0:

- ShipDestroyed event emitted
- Cargo drop generated
- Escape pod triggered
- CombatInstance terminated
- Bounty/reputation updated

Destruction occurs within long-range engagement context.

---

# 15. Edge Case Handling

## 15.1 Disconnect During Combat

- Ship remains in simulation
- AI defensive behavior optional
- Destruction possible

## 15.2 Projectile Impact During Disengage

Impact cancels disengagement if conditions violated.

---

# 16. Integration Points

Depends On:

- Combat System
- Combat Math
- Scanning System
- Tick Engine
- Command Queue

Exposes:

- CombatStartEvent
- CombatLockEvent
- CombatDisengagedEvent
- ShipDestroyedEvent

---

# 17. Balancing Guidelines

- Long-range standoffs should be common.
- Early firing should be inefficient.
- Sensor play should dominate early phase.
- Jump escape viable before aggression.
- Silent shadowing viable but not dominant.

---

# 18. Non-Goals (v1)

- Visual dogfighting
- Manual aiming interfaces
- Close-quarters arcade combat
- Cinematic camera-based targeting

Combat is strategic, not cinematic.

---

# 19. Future Extensions

- Warp disruption modules
- Fleet-level sensor nets
- Sector-wide engagement alerts

---

# End of Document
