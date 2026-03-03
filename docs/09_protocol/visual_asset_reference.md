# Visual Asset Reference Specification
## Version: 0.1
## Status: Draft
## Owner: Core Architecture
## Last Updated: 2026-03-02

---

# 1. Purpose

Defines the canonical visual reference model for:

- Planets
- Space stations
- Star types
- Asteroid fields
- Anomalies
- Ship hull classes
- Environmental objects

This specification ensures:

- Stable visual identifiers
- GUI compatibility
- Asset naming consistency
- Deterministic visual mapping
- Protocol-safe rendering abstraction

The server does not render visuals.  
It emits visual reference IDs only.

---

# 2. Scope

## IN SCOPE

- Visual asset ID conventions
- Naming schema
- Asset categorization
- Versioning rules
- Server-to-client reference behavior

## OUT OF SCOPE

- Actual image files
- 3D models
- Rendering engine logic
- Animation systems
- Shader configuration

---

# 3. Design Principles

- Server emits stable visual identifiers only.
- Visual identifiers are deterministic.
- Asset references must not change across minor versions.
- GUI client responsible for rendering.
- TEXT client may render symbolic representation.
- Visual identifiers must be predictable and structured.

---

# 4. Visual Asset ID Model

All visual references follow naming convention:

<category>-<type>-<variant>

Examples:

- planet-desert-01
- planet-ice-02
- station-industrial-01
- star-blue-giant-01
- anomaly-gravitational-01
- asteroid-field-dense-02

IDs must be lowercase and hyphen-separated.

---

# 5. Planet Visual References

Planets are categorized by biome or classification.

Supported v1 categories:

- planet-desert-01
- planet-desert-02
- planet-ice-01
- planet-lava-01
- planet-ocean-01
- planet-gas-giant-01
- planet-rocky-01
- planet-habitable-01

Generation rule:

PlanetVisualID derived from:

SystemSeed + PlanetIndex

Must remain deterministic.

---

# 6. Space Station Visual References

Station types include:

- station-industrial-01
- station-trade-hub-01
- station-military-01
- station-frontier-01
- station-mining-01

Station visual type derived from:

- System security
- Region classification
- Economic bias

Example mapping:

High security core → station-trade-hub-01  
Low security frontier → station-frontier-01  

---

# 7. Star Visual References

Star classifications:

- star-yellow-dwarf-01
- star-red-dwarf-01
- star-blue-giant-01
- star-white-dwarf-01
- star-neutron-01

Star type derived from:

SystemSeed deterministic mapping.

Star type may influence:

- Anomaly probability
- Hazard types
- Exploration value

---

# 8. Asteroid Field Visual References

Visual density categories:

- asteroid-field-sparse-01
- asteroid-field-medium-01
- asteroid-field-dense-01
- asteroid-field-volatile-01

Mapped from:

Density value bands.

Example:

Density < 0.4 → sparse  
Density 0.4–0.7 → medium  
Density > 0.7 → dense  

---

# 9. Anomaly Visual References

Examples:

- anomaly-energy-01
- anomaly-gravitational-01
- anomaly-rift-01
- anomaly-derelict-01

Anomaly visual must not reveal full gameplay value.

Visual is descriptive, not deterministic of rarity.

---

# 10. Ship Visual References

Ship hull types:

- ship-miner-01
- ship-trader-01
- ship-interceptor-01
- ship-explorer-01
- ship-pirate-01

Weapon modules may have optional visual overlays:

- weapon-torpedo-01
- weapon-railgun-01
- weapon-emp-01

Server only emits hull and module identifiers.

---

# 11. Protocol Integration

All visual references transmitted as part of payload.

Example:

{
  "type": "system_view",
  "timestamp": 10452,
  "correlation_id": null,
  "payload": {
    "system_id": "SYS-104",
    "star_visual": "star-yellow-dwarf-01",
    "planets": [
      { "id": "P1", "visual": "planet-desert-01" },
      { "id": "P2", "visual": "planet-gas-giant-01" }
    ],
    "stations": [
      { "id": "ST1", "visual": "station-industrial-01" }
    ]
  }
}

TEXT client may map these to symbolic icons.

GUI client resolves them to actual assets.

---

# 12. Versioning Rules

Visual IDs are considered stable API surface.

Rules:

- Do not rename existing IDs.
- Do not repurpose IDs.
- New variants must increment variant suffix.
- Removal requires major protocol version change.

---

# 13. Asset Registry

Server maintains a canonical registry of valid visual IDs.

Invalid IDs must:

- Trigger validation error in development
- Be logged as ERROR

No arbitrary strings allowed.

---

# 14. Determinism Requirements

Given:

- UniverseSeed
- SystemID

Visual references must be reproducible.

Visual assignment must not depend on runtime randomness.

---

# 15. TEXT Client Rendering Policy

TEXT mode may render:

- Unicode icons
- ASCII representations
- Color-coded symbols

Example mappings:

planet-desert-01 → 🟠  
planet-ice-01 → 🔵  
station-industrial-01 → ⚙  

TEXT mapping must remain independent from simulation logic.

---

# 16. Non-Goals (v1)

- Dynamic visual mutation
- Procedural texture generation
- Player-customizable station visuals
- Real-time environmental animations

---

# 17. Future Extensions

- Asset metadata (size, color theme)
- Animated anomaly references
- Visual damage states
- Region-specific visual themes
- Seasonal event overlays

All future additions must preserve stable ID structure.

---

# End of Document
