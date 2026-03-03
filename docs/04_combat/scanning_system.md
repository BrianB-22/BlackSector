# Scanning System Specification
## Version: 0.1
## Status: Draft
## Owner: Core Simulation
## Last Updated: 2026-03-02

---

# 1. Purpose

Defines passive and active scanning mechanics used during combat.

Scanning provides information at cost of exposure.

---

# 2. Scope

IN SCOPE:
- Passive scan
- Active channelled scan
- Scan detail resolution
- Exposure penalties

OUT OF SCOPE:
- Exploration mapping
- Mining scans

---

# 3. Design Principles

- Information is partial
- Scanning increases vulnerability
- Scanning requires tracking

---

# 4. Core Concepts

- ScanPower
- ScanDetailLevel
- SignatureMultiplier
- TargetHeatPenalty

---

# 5. Data Model

ScanAction:
- owner_id
- channel_ticks_remaining
- scan_power

---

# 6. State Machine

IDLE → CHANNELING → RESOLVED | FAILED

---

# 7. Core Mechanics

Active Scan:
- EnergyCost = 15
- Heat +8
- Signature ×1.3
- ChannelTime = 2 ticks

Interrupted if:
- Hull damage
- Energy < threshold
- Major maneuver

---

# 8. Mathematical Model

ScanDetailLevel =

(ScanPower / 100)
× TrackingConfidence
÷ DistanceFactor
× (1 − TargetHeatPenalty)

---

# 9. Detail Thresholds

<0.20 → Ship size  
0.20–0.40 → Ship class  
0.40–0.60 → Shield estimate  
0.60–0.80 → Weapon type  
>0.80 → Cargo + heat estimate  

---

# 10. Tunable Parameters

- ScanPower base
- Signature multiplier
- Channel duration
- Detail thresholds

---

# 11. Integration Points

Depends On:
- Combat System
- Combat Math
- Heat model

---

# 12. Balance Guidelines

- Long-range scans limited
- High heat reduces scan clarity
- Close range required for deep intel

---

# 13. Non-Goals (v1)

- Subsystem breakdown scans
- Full stat reveal
- Real-time continuous scan

---

# End of Document
