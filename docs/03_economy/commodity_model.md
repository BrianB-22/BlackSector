# Commodity Model Specification

## Version: 0.1

## Status: Draft

## Owner: Economy Systems

## Last Updated: 2026-03-03

---

# 1. Purpose

The Commodity Model defines the global structure and behavior of tradable goods within the universe economy.

Commodities are the core resources exchanged between:

* ports
* AI traders
* players
* industrial infrastructure

This subsystem provides the canonical definitions for commodities including:

* identity
* category
* base value
* weight
* legality
* rarity
* volatility
* economic tags

Commodity definitions are loaded from configuration files and used by all economic subsystems.

---

# 2. Scope

IN SCOPE:

* commodity definitions
* commodity categories
* legality classification
* weight system
* volatility properties
* rarity definitions
* commodity tagging
* regional origin metadata
* commodity configuration loading

OUT OF SCOPE:

* port inventory management (Port System)
* player cargo management (Ship System)
* mining production (Resource Extraction)
* manufacturing chains (Industrial Systems)
* AI trading behavior (AI Trader System)

---

# 3. Design Principles

Non-negotiable constraints:

* commodity definitions loaded from configuration
* server authoritative economic logic
* deterministic price baseline
* lightweight commodity metadata
* extensible tag system
* economy must encourage regional trade routes

Commodity definitions must remain immutable during runtime.

---

# 4. Core Concepts

Commodity
A tradable resource that can be bought, sold, transported, or consumed within the economy.

Commodity Definition
The static configuration describing a commodity's economic characteristics.

Commodity Category
Classification used to group commodities by economic function.

Legality
Determines whether a commodity is legal, restricted, or contraband.

Volatility
Defines how sensitive a commodity is to supply fluctuations.

Rarity
Describes how frequently the commodity appears in production systems.

Economic Tags
Flexible metadata used by other systems to determine behavior.

Regional Origin
The region type where a commodity is most commonly produced.

---

# 5. Data Model

## Entity: CommodityDefinition (Configuration)

Commodity definitions are loaded from JSON configuration at server startup.

Fields:

* commodity_id: string
* name: string
* category: string
* base_price: float
* weight: float
* rarity: float
* volatility: float
* legality: enum
* origin_region_type: string
* tags: list<string>

CommodityDefinition objects are immutable during runtime.

---

## Entity: CommodityRegistry (Runtime Cache)

In-memory registry containing all loaded commodity definitions.

Fields:

* commodities: map<string, CommodityDefinition>

Used for rapid lookup by commodity_id.

---

# 6. State Machine

Commodities themselves do not change state.

However commodities may exist in economic states within other systems.

Example states:

AVAILABLE
→ Commodity present in port inventory.

SCARCE
→ Commodity supply significantly below demand.

SURPLUS
→ Commodity supply exceeds demand.

CONTRABAND
→ Commodity illegal in certain security zones.

These states are derived from economic conditions rather than stored in commodity definitions.

---

# 7. Core Mechanics

Commodity definitions are loaded during server initialization.

Initialization sequence:

1. server loads commodity configuration file
2. JSON schema validated
3. commodity definitions parsed
4. registry created in memory
5. economic subsystems reference registry

All economic systems reference commodities using the `commodity_id`.

Example usage:

* port inventories store commodity_id
* cargo holds store commodity_id
* transactions reference commodity_id

Commodity definitions themselves never change during gameplay.

---

# 8. Mathematical Model

## Variables

BasePrice
Standard commodity value before supply/demand adjustments.

Weight
Cargo mass multiplier affecting ship capacity.

Volatility
Sensitivity of price changes to supply fluctuations.

Rarity
Probability modifier affecting commodity generation.

---

## Price Influence

Commodity volatility modifies how quickly price reacts to scarcity.

Example conceptual formula:

PriceAdjustmentFactor =

1 + (Volatility × ScarcityRatio)

Where:

ScarcityRatio =
1 - (CurrentSupply / MaximumSupply)

Higher volatility commodities experience larger price swings.

---

# 9. Tunable Parameters

DefaultCommodityCountTarget = 10

VolatilityRange:

0.2 → stable goods
1.0 → highly volatile goods

RarityRange:

0.1 → extremely rare
1.0 → very common

WeightRange:

0.25 → very light
5.0 → very heavy

---

# 10. Integration Points

Depends On:

* Configuration Loader
* Economy Initialization
* Tick Engine

Exposes:

* commodity metadata
* base price
* cargo weight
* legality classification
* economic tags

Used By:

* Port System
* AI Trader System
* Mission System
* Resource Extraction
* Black Market System
* Economic Simulation

---

# 11. Failure & Edge Cases

Invalid Commodity ID
System rejects request and logs error.

Duplicate Commodity ID
Configuration load fails.

Invalid JSON Schema
Server startup aborted.

Missing Required Fields
Commodity definition rejected.

---

# 12. Performance Constraints

Commodity registry must support:

* constant-time lookup
* minimal memory footprint
* rapid iteration for economic systems

Target scale:

* 10–50 commodities
* near-zero runtime overhead

---

# 13. Security Considerations

Commodity definitions must only load during server startup.

Clients cannot:

* modify commodity definitions
* inject new commodities
* alter base price values

Commodity configuration files must be validated before loading.

---

# 14. Telemetry & Logging

Log events:

* commodity configuration load
* invalid commodity references
* configuration schema violations

Track metrics:

* commodity trade volume
* price volatility
* commodity scarcity frequency

---

# 15. Balancing Guidelines

Commodity distribution should create natural trade routes.

Design goals:

* essential goods widely available
* industrial goods regionally concentrated
* high-value goods rare
* volatile commodities enable speculation

Economic depth should increase with distance from core regions.

---

# 16. Non-Goals (v1)

Not included in the initial version:

* commodity decay
* dynamic commodity mutation
* manufacturing chains
* commodity crafting

---

# 17. Future Extensions

Potential expansions include:

* player manufacturing
* industrial supply chains
* faction-controlled resources
* commodity futures markets
* smuggling systems
* economic event modifiers

---

# Commodity Configuration

Commodity definitions are stored in JSON configuration files.

Example location:

config/economy/commodities.json

Commodity definitions load during server startup and remain immutable for the lifetime of the server process.

---

## Example Commodity Configuration

```json
{
  "commodities": [
    {
      "commodity_id": "warp_fuel",
      "name": "Warp Fuel",
      "category": "energy",
      "base_price": 10,
      "weight": 1.0,
      "rarity": 0.8,
      "volatility": 0.3,
      "legality": "legal",
      "origin_region_type": "industrial",
      "tags": ["essential", "high_volume"]
    },
    {
      "commodity_id": "food_supplies",
      "name": "Food Supplies",
      "category": "agriculture",
      "base_price": 8,
      "weight": 1.2,
      "rarity": 0.9,
      "volatility": 0.2,
      "legality": "legal",
      "origin_region_type": "agricultural",
      "tags": ["consumable", "essential"]
    },
    {
      "commodity_id": "refined_ore",
      "name": "Refined Ore",
      "category": "industrial",
      "base_price": 12,
      "weight": 3.0,
      "rarity": 0.7,
      "volatility": 0.4,
      "legality": "legal",
      "origin_region_type": "asteroid",
      "tags": ["raw_material"]
    },
    {
      "commodity_id": "advanced_electronics",
      "name": "Advanced Electronics",
      "category": "technology",
      "base_price": 18,
      "weight": 0.5,
      "rarity": 0.5,
      "volatility": 0.7,
      "legality": "legal",
      "origin_region_type": "high_tech",
      "tags": ["high_value"]
    },
    {
      "commodity_id": "luxury_goods",
      "name": "Luxury Goods",
      "category": "luxury",
      "base_price": 30,
      "weight": 0.8,
      "rarity": 0.3,
      "volatility": 0.8,
      "legality": "legal",
      "origin_region_type": "core_world",
      "tags": ["luxury"]
    },
    {
      "commodity_id": "military_hardware",
      "name": "Military Hardware",
      "category": "military",
      "base_price": 40,
      "weight": 2.0,
      "rarity": 0.4,
      "volatility": 0.9,
      "legality": "restricted",
      "origin_region_type": "military",
      "tags": ["military"]
    },
    {
      "commodity_id": "medical_supplies",
      "name": "Medical Supplies",
      "category": "medical",
      "base_price": 22,
      "weight": 0.7,
      "rarity": 0.5,
      "volatility": 0.6,
      "legality": "legal",
      "origin_region_type": "industrial",
      "tags": ["essential"]
    },
    {
      "commodity_id": "nanotech_components",
      "name": "Nanotech Components",
      "category": "technology",
      "base_price": 45,
      "weight": 0.3,
      "rarity": 0.2,
      "volatility": 1.0,
      "legality": "restricted",
      "origin_region_type": "high_tech",
      "tags": ["high_value"]
    },
    {
      "commodity_id": "alien_artifacts",
      "name": "Alien Artifacts",
      "category": "exotic",
      "base_price": 80,
      "weight": 0.4,
      "rarity": 0.1,
      "volatility": 1.0,
      "legality": "contraband",
      "origin_region_type": "anomaly",
      "tags": ["rare", "black_market"]
    }
  ]
}
```

---

# End of Document
