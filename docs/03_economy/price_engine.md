# Price Engine Specification

## Version: 0.1

## Status: Draft

## Owner: Economy Systems

## Last Updated: 2026-03-03

---

# 1. Purpose

The Price Engine calculates commodity prices within the galactic economy.

Prices are derived from:

* base commodity value
* supply-demand signals
* economic events
* regional modifiers

The system converts market signals into **stable, predictable price adjustments** while preventing extreme volatility.

---

# 2. Scope

IN SCOPE:

* commodity price calculation
* price smoothing
* price clamping
* event-based price modifiers
* scarcity-based pricing

OUT OF SCOPE:

* commodity supply calculation
* port inventory updates
* economic event generation
* AI trading behavior

---

# 3. Design Principles

Non-negotiable constraints:

* deterministic price calculation
* server authoritative pricing
* smoothed price transitions
* bounded price ranges
* minimal computational overhead

Prices must remain stable enough to avoid economic exploitation.

---

# 4. Core Concepts

Base Price
Default commodity value defined in the Commodity Model.

Target Price
Theoretical price derived from market conditions.

Current Price
The actual market price used for trading.

Price Smoothing
Gradual movement toward target price to prevent sudden spikes.

Price Clamp
Hard limits preventing extreme prices.

---

# 5. Data Model

## Entity: CommodityPriceState (Persistent)

Fields:

* commodity_id: string
* port_id: uint64
* current_price: float
* target_price: float
* last_update_tick: int

---

# 6. State Machine

STABLE
→ Price within normal range.

ADJUSTING
→ Price moving toward target value.

CLAMPED
→ Price reached configured maximum or minimum.

Prices always trend toward equilibrium.

---

# 7. Core Mechanics

Price calculation sequence:

1. retrieve base price
2. retrieve supply-demand signals
3. apply scarcity modifiers
4. apply event modifiers
5. calculate target price
6. move current price toward target price

Example:

```
TargetPrice = BasePrice × ScarcityModifier × EventModifier
```

Price smoothing:

```
CurrentPrice =
CurrentPrice + (TargetPrice - CurrentPrice) × SmoothingFactor
```

---

# 8. Mathematical Model

## Variables

BasePrice
Commodity base value.

ScarcityModifier
Derived from supply ratio.

EventModifier
Applied during economic events.

SmoothingFactor
Rate at which prices approach target value.

---

## Scarcity Modifier

Example:

```
ScarcityModifier =
1 + (ScarcityScore × ScarcityImpact)
```

---

## Price Clamp

```
MinPrice = BasePrice × 0.5
MaxPrice = BasePrice × 5.0
```

---

# 9. Tunable Parameters

PriceSmoothingFactor = 0.15

ScarcityImpact = 3.0

MinimumPriceMultiplier = 0.5

MaximumPriceMultiplier = 5.0

PriceUpdateInterval = 60 seconds

---

# 10. Integration Points

Depends On:

* Commodity Model
* Supply Demand Model
* Economic Events
* Economic Tick

Exposes:

* commodity prices
* price trend signals

Used By:

* Port System
* AI Trader Model
* Economic Analytics
* Player Trading Interface

---

# 11. Failure & Edge Cases

Negative Price
Automatically clamped.

Invalid Commodity ID
Request rejected.

Extreme Scarcity
Price limited by clamp values.

Price Desynchronization
Corrected during next update cycle.

---

# 12. Performance Constraints

System must support:

* up to 1,000 ports
* 10–50 commodities
* frequent price updates

Price calculation must remain under 5 ms per economic tick.

---

# 13. Security Considerations

All price calculations must occur server-side.

Clients cannot:

* alter price values
* inject price updates
* bypass price clamps

---

# 14. Telemetry & Logging

Track metrics including:

* commodity price volatility
* price spike frequency
* regional price variance
* trade profitability

---

# 15. Balancing Guidelines

Prices should:

* respond to scarcity
* reward long-distance trade
* remain stable enough for players to predict trends

Large price swings should occur mainly during economic events.

---

# 16. Non-Goals (v1)

Not included initially:

* player price manipulation
* speculative market trading
* futures markets

---

# 17. Future Extensions

Possible expansions include:

* player-driven markets
* faction economic influence
* regional currency systems
* market speculation

---

# End of Document
