# Mining Balancing Guidelines

## Version: 0.1

## Status: Draft

## Owner: Core Simulation

## Last Updated: 2026-03-05

---

# 1. Purpose

This document defines the design philosophy and balancing guidelines for the Mining subsystem.

Unlike subsystem specifications, which describe mechanics and implementation details, this document defines the intended **mining experience and gameplay dynamics**.

Mining in Black Sector is designed around the concept of **risk-rewarded resource extraction**:

* deliberate risk escalation across security zones
* never purely passive or safe
* every extraction cycle carries meaningful tradeoffs
* economic impact from individual mining actions

These guidelines ensure future development maintains the intended risk-reward feel.

---

# 2. Core Mining Philosophy

Mining should feel calculated, dangerous, and economically meaningful.

Key characteristics:

* yield scales directly with danger
* every extraction cycle has a cost (energy, heat, exposure)
* players are never guaranteed profit
* safe mining is intentionally low-margin

Mining should reward **informed risk-taking**, not passive grinding.

---

# 3. Security Zone Risk vs Reward

The four security zones define distinct mining experiences:

| Zone            | SecurityRating | Yield Modifier | Hazard Probability | PvP Risk |
| --------------- | -------------- | -------------- | ------------------ | -------- |
| High Security   | 0.7–1.0        | 0.7×           | Low                | Minimal  |
| Medium Security | 0.4–0.7        | 1.0×           | Moderate           | Present  |
| Low Security    | 0.0–0.4        | 1.3×           | High               | Frequent |
| Black Sector    | NULL           | 2.0×+          | Maximum            | Constant |

High security must feel **stable but unrewarding**.
Low security must feel **lucrative but dangerous**.
Black Sector must feel **extreme in both dimensions**.

Players should never be able to match low-security income by farming high-security at volume.

---

# 4. Yield Balance

## 4.1 Yield Formula

```
Yield =
(BaseYield × Density)
× RandomFactor
× (1 − InstabilityFactor)
× SecurityYieldModifier
```

## 4.2 RandomFactor Range

RandomFactor ∈ [0.6 – 1.4]

This spread must remain wide enough that:

* no two extractions feel identical
* high-instability fields produce occasional zero-yield cycles
* players cannot reliably predict exact yield per cycle

## 4.3 Depletion Behavior

Density depletion rate: 0.02 per cycle (default)

Fields must deplete meaningfully so that:

* no single field is indefinitely farmable
* players are encouraged to move between systems
* over-mined regions create economic supply imbalances

Density floor of 0.1 prevents complete field exhaustion in v1.

---

# 5. Energy and Heat Constraints

Mining costs per cycle:

* EnergyCost = 20
* HeatIncrease = +6

These values must ensure:

* sustained mining creates real heat pressure
* energy depletion is a genuine ceiling on mining runs
* players cannot mine indefinitely without resource management

Heat accumulation during mining interacts with the same heat system as combat. A miner caught in a PvP encounter while already heat-stressed faces compounding disadvantage. This is intentional.

---

# 6. Hazard Balancing

HazardProbability formula:

```
HazardProbability = (1 − SecurityRating) × 0.5
```

Hazard balance targets:

* High Security: hazards must be rare enough to be noteworthy when encountered
* Low Security: hazards should occur frequently enough to be a real decision factor
* Black Sector: hazards should be assumed present, not incidental

MineTriggerChance = 0.25 (when hazard present)

Hazards must create **tension**, not guaranteed punishment. Players who use drones and sensors properly should be able to mitigate but not eliminate risk.

---

# 7. Drone Balance

Drones enhance but never replace risk.

| Drone Type | AssistMultiplier |
| ---------- | ---------------- |
| None       | 1.0×             |
| Scout      | 1.5×             |

Design targets:

* drones must meaningfully improve hazard detection
* drones must not make mining safe in low-security zones
* DestructionChance of 0.30 when hazard present must feel like a real loss
* drone cost (10 energy + 2-tick channel) must matter at scale

A player running Scout drones in every cycle should feel enhanced, not immune.

---

# 8. PvP Exposure

Mining increases player vulnerability:

* SignatureRadius × 1.25
* Velocity −30%
* DetectionScore bonus to observers

These penalties must be significant enough that:

* a player who chooses to mine in dangerous space is making a real commitment
* aborting a mining run to deal with a threat is always a meaningful tradeoff
* camping mining hotspots is a viable piracy strategy

Mining hotspots naturally become PvP conflict zones. This is by design.

---

# 9. Economic Impact

Mining must not destabilize the economy in isolation.

Balance targets:

* a single miner cannot meaningfully crash a commodity price on their own
* a cluster of miners in the same system should produce visible supply pressure
* Rare Discovery events must be exceptional, not routine

RareMineralChance formula:

```
RareMineralChance = Density × (1 − SecurityRating) × RareSpawnMultiplier
```

Default RareSpawnMultiplier = 0.4

Rare discoveries (threshold ≥ 0.6) should:

* trigger market volatility that is noticeable but temporary
* increase pirate activity in the area
* create mission opportunities rather than permanent economic distortion

No miner or group of miners should create a permanent economic monopoly.

---

# 10. Field Regeneration

Regeneration rate: 0.005 per 100 ticks (default)

Fields must regenerate slowly enough that:

* depleted zones remain economically depressed for a meaningful period
* player movement between systems is incentivized
* no field resets to full between short play sessions

Field regeneration is not visible to players in real time. The economy should feel like it has memory.

---

# 11. Black Sector Mining

Black Sector systems (SecurityRating = NULL) represent the extreme case:

* yield modifiers at 2.0× and above
* hazard probability at maximum
* pirate spawn chance at maximum
* PvP is constant and unpoliced

Black Sector mining is not a casual activity. Players who mine there are accepting maximum risk for maximum reward. No game system should soften this tradeoff.

---

# 12. Anti-Patterns to Avoid

The following outcomes represent design failures:

* **Safe high-yield routes exist** — players must never find a high-security field that competes economically with low-security fields
* **Passive income loops** — mining must always require player attention and decisions per cycle
* **Zero-risk drone scouting** — drones reduce uncertainty; they do not eliminate hazards
* **Instant depletion** — fields depleting in a single session break economic geography
* **Guaranteed rare spawns** — rare minerals must remain rare

---

# 13. Monitoring and Balance

Mining telemetry should be monitored to maintain balance.

Key metrics:

* yield distribution by security tier
* instability failure rate
* field depletion frequency per system
* rare mineral discovery rate
* mining-related PvP encounter rate
* commodity supply injection per tick

Adjustments should occur gradually. Mining balance is tightly coupled to the broader economy — abrupt changes cascade into price shifts and trader behavior.

---

# 14. Non-Goals (v1)

Not intended for v1:

* cooperative fleet mining
* automated mining bots
* player-owned mining stations or territory claims
* mining specialization modules

Mining is individual and risk-based in v1.

---

# 15. Future Extensions

Potential expansions that must preserve the risk-reward philosophy:

* mining specialization modules (higher yield, higher heat cost)
* equipment degradation from instability
* cooperative extraction bonuses (require physical proximity, not passive)
* volatile mineral chain reactions
* player-deployed mining arrays (future territory mechanics)

All future mining features must preserve: **risk, volatility, and economic impact**.

---

# End of Document
