# Targeting and Tracking Model Specification

## Version: 0.1

## Status: Draft

## Owner: Combat Systems

## Last Updated: 2026-03-03

---

# 1. Purpose

The Targeting and Tracking Model governs how ships maintain combat locks on detected targets.

Tracking stability determines whether weapons can be effectively employed.

---

# 2. Core Concepts

Sensor Contact
Initial detection of a target.

Target Lock
Stable tracking enabling weapon targeting.

Lock Degradation
Loss of tracking due to distance, maneuvering, or environmental interference.

---

# 3. Tracking Mechanics

Tracking sequence:

```
detect target
→ acquire sensor lock
→ maintain tracking
→ degrade lock over time
→ lose contact
```

---

# 4. Tracking Factors

Tracking strength influenced by:

* distance
* ship size
* sensor quality
* environmental interference

Example environments:

* asteroid fields
* nebula clouds
* ion storms

---

# 5. Tactical Outcomes

Loss of tracking may cause:

* missile guidance failure
* targeting disruption
* disengagement opportunities

---

# End of Document
