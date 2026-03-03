# Combat System Requirements Document (CSR)
## Version 1.0
## Scope: 1v1 Tactical Predictive Combat
## Architecture: Server-Authoritative, Tick-Based

---

# 1. Purpose

This document defines the functional and non-functional requirements for the 1v1 tactical combat system used within the persistent multiplayer space trading simulation.

Combat is:
- Predictive, not twitch-based
- Information-imperfect
- Tick-resolved
- Server-authoritative
- Energy and heat constrained
- Skill-expressive through tracking and planning

---

# 2. Design Principles

1. Combat must reward anticipation and planning.
2. Information must be partial but actionable.
3. Upgrades must increase clarity, not eliminate decision-making.
4. No client-side authority over combat resolution.
5. All combat state transitions occur within the world tick engine.

---

# 3. Engagement Model

## 3.1 Combat Scope

- Primary mode: 1v1 engagements
- Combat occurs within a CombatInstance
- No fleet mechanics in v1

## 3.2 Combat Initiation

Combat begins when:
- Player initiates hostile action
- PvP interdiction triggers
- Mission event forces engagement
- Mutual aggression confirmed

Upon engagement:
- Jump disabled
- Docking disabled
- Trade disabled
- Combat state overlay applied

---

# 4. Tick Engine Integration

Combat is resolved in discrete world ticks.

## 4.1 Tick Duration

- Default: 2 seconds (configurable)

## 4.2 Tick Resolution Order

1. Process queued maneuver commands
2. Update velocity and range vectors
3. Recalculate detection and tracking
4. Apply electronic warfare effects
5. Advance projectile positions
6. Resolve impacts
7. Apply damage
8. Update energy and heat
9. Emit combat snapshot

All resolution is deterministic and server-controlled.

---

# 5. Combat State Model

## 5.1 CombatInstance

- combat_id
- player_a_id
- player_b_id
- start_tick
- engagement_range
- active_projectiles[]
- environment_modifiers
- status

## 5.2 ShipCombatState

- range (scalar)
- relative_velocity
- bearing_offset
- sensor_strength
- tracking_confidence (0–100%)
- shield_level
- hull_level
- energy_level
- heat_level
- signature_radius
- evasion_modifier
- active_effects[]

---

# 6. Detection & Tracking

## 6.1 Detection

Detection must succeed before tracking begins.

Detection depends on:
- Sensor strength
- Target signature
- Distance
- Electronic warfare
- Environmental modifiers

## 6.2 Tracking Confidence

TrackingConfidence ∈ [0–100]

Tracking increases over time if:
- Target detected
- Sensors active
- No major maneuver disruption

Tracking decreases when:
- Target performs hard burn
- Target jams
- Target enters silent mode
- Distance increases significantly

Tracking drives:
- Firing solution quality
- Hit probability
- Intercept confidence

---

# 7. Maneuver System

Players may issue one maneuver per tick.

Available maneuvers:

- Engine Burn
- Hard Evasion
- Silent Mode
- Sensor Boost
- Energy Reallocation
- Jam Target

Each maneuver:
- Consumes energy
- Increases heat
- Has cooldown
- Alters tracking or evasion variables

No maneuver may bypass tick system.

---

# 8. Energy Model

EnergyPool is shared between:

- Engines
- Shields
- Sensors
- Weapons
- Electronic warfare

Energy reallocation must:
- Occur via explicit command
- Persist until changed
- Affect subsystem effectiveness

Energy may not exceed maximum ship capacity.

---

# 9. Heat Model

Heat increases from:
- Weapon fire
- Engine burn
- Sensor boost
- Jamming

Heat decreases gradually per tick.

High heat:
- Increases signature radius
- Reduces tracking stability
- May reduce subsystem efficiency (future expansion)

Heat cannot be ignored.

---

# 10. Weapon System

## 10.1 Weapon Types (v1)

- Missile (delayed impact)
- Railgun (instant resolution)
- EMP (system disruption)

## 10.2 Firing Requirements

Player must:
- Have sufficient tracking
- Have sufficient energy
- Have weapon cooldown ready

Upon firing:
- Projectile instance created
- Impact ETA calculated
- Energy deducted
- Heat increased

---

# 11. Firing Solution Model

When firing:

System calculates:
- SolutionQuality
- InterceptTime (ticks)
- HitProbability (abstracted)
- ExpectedDamageRange

Player sees:
- Solution Confidence (Low / Moderate / High)
- ETA
- Tracking %

Exact formulas remain hidden.

---

# 12. Projectile Model

Projectile has:
- origin_ship_id
- target_ship_id
- time_to_impact
- damage_profile
- accuracy_modifier

Projectile resolved on tick when time_to_impact = 0.

---

# 13. Damage Model

Damage resolves:

1. Shield absorption
2. Remaining damage to hull

No direct hull bypass unless explicitly designed weapon type.

Ship destroyed when:
- hull_level <= 0

---

# 14. Destruction & Outcome

On destruction:

- Ship destroyed
- Cargo dropped
- Escape pod auto-triggered
- CombatInstance ends
- Reputation updated
- Bounty logic applied (if applicable)

No account deletion.

---

# 15. Disengagement

Disengagement requires:

- Maintain separation threshold
- No weapon fire for X ticks
- Tracking below Y threshold

Successful disengage:
- CombatInstance ends
- Cooldown applied

---

# 16. UI Data Requirements

Combat snapshot must include:

- Target ID
- Range
- Relative velocity
- Tracking %
- Signature estimate
- Shield %
- Hull %
- Energy %
- Heat %
- Incoming projectiles + ETA
- Suggested lead (if targeting computer >= tier 2)

UI rendering handled by interface adapter.

---

# 17. Upgrade Impact Rules

Upgrades may:

- Increase sensor strength
- Increase tracking growth rate
- Reduce energy cost
- Improve shield capacity
- Reduce heat generation
- Improve solution clarity

Upgrades must NOT:

- Remove need for tracking
- Guarantee hits
- Eliminate evasion
- Bypass tick engine

---

# 18. Anti-Exploit Requirements

- No client-side damage resolution
- No direct state mutation outside tick loop
- Unique projectile IDs
- Atomic state updates
- Command queue validation
- Cooldown enforcement

---

# 19. Performance Constraints

- Combat resolution per tick < 5ms per instance
- Scalable to multiple simultaneous 1v1 engagements
- No blocking operations inside tick loop
- Deterministic resolution order

---

# 20. Extensibility

Future-ready for:

- Small squad engagements
- Environmental hazards
- Advanced electronic warfare
- Critical subsystem damage
- Combat analytics

---

# 21. Non-Goals (v1)

- Fleet battles
- Manual vector coordinate entry
- Real-time twitch combat
- True orbital mechanics
- Player-owned combat stations

---

# End of Document
