# Drone System Specification

## Version: 0.1
## Status: Draft
## Owner: Core Simulation

## Last Updated: 2026-03-02

---

# 1. Purpose

Defines deployable drone mechanics supporting:

- Mining
- Hazard detection
- Exploration

---

# 2. Scope

IN SCOPE:

- Drone deployment
- Drone scan logic
- Drone destruction risk

OUT OF SCOPE:

- Combat drones
- Autonomous combat

---

# 3. Design Principles

- Consumable asset
- Enhances but does not replace player risk
- Destroyable

---

# 4. Core Concepts

- DronePower
- DroneAssistMultiplier
- DroneDestructionChance

---

# 5. Data Model

## Entity: Drone

- drone\_id
- owner\_id
- type
- power
- active

---

# 6. Core Mechanics

Deployment:

EnergyCost = 10

ChannelTime = 2 ticks

DroneAssistMultiplier:

- None: 1.0
- Scout: 1.5

---

# 7. Mathematical Model

DroneScanDetail =

(DronePower / 100)

× (1 − InstabilityFactor)

× (1 − DepletionLevel)

× SecurityYieldModifier

DestructionChance = 0.30 if hazard present.

---

# 8. Tunable Parameters

- DronePower scaling
- DestructionChance
- Assist multiplier

---

# 9. Integration Points

Depends On:

- Mining
- Exploration
- Hazard

---

# 10. Non-Goals (v1)

- Autonomous mining
- Swarm drones

---

# End of Document
