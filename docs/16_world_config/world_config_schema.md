# World Config Schema

## Version: 1.0
## Status: Draft
## Owner: Core Architecture
## Last Updated: 2026-03-05

---

# 1. Purpose

Defines the JSON format for world configuration files. World files define all star systems, jump connections, ports, and commodity assignments that make up a playable region.

World files are loaded at server startup. The active world file is set in `server.json` → `server.world_config_path`.

The Phase 1 sample world is at `config/world/alpha_sector.json`.

---

# 2. File Structure

```json
{
  "_schema_version": "1.0",
  "world": { ... },
  "commodities": [ ... ],
  "systems": [ ... ],
  "jump_connections": [ ... ],
  "ports": [ ... ]
}
```

---

# 3. world Object

```json
"world": {
  "name": "Alpha Sector",
  "description": "string — player-visible region description",
  "seed": 1701
}
```

| Field         | Type   | Required | Description                                            |
| ------------- | ------ | -------- | ------------------------------------------------------ |
| `name`        | string | yes      | Display name shown to players                          |
| `description` | string | no       | Flavor text for the region                             |
| `seed`        | int64  | yes      | World-specific seed. Combined with server seed for generation |

---

# 4. commodities Array

Defines all tradeable commodities in this world.

```json
"commodities": [
  {
    "commodity_id": "food_supplies",
    "name": "Food Supplies",
    "category": "essential",
    "base_price": 100,
    "unit": "crate"
  }
]
```

| Field          | Type   | Required | Description                                           |
| -------------- | ------ | -------- | ----------------------------------------------------- |
| `commodity_id` | string | yes      | Snake_case unique identifier                          |
| `name`         | string | yes      | Display name                                          |
| `category`     | string | yes      | `"essential"`, `"industrial"`, `"luxury"`             |
| `base_price`   | int    | yes      | Reference price in credits                            |
| `unit`         | string | no       | Unit label for display (default: "unit")              |

---

# 5. systems Array

Defines star systems.

```json
"systems": [
  {
    "system_id": "nexus_prime",
    "name": "Nexus Prime",
    "description": "The central hub of Federated Space.",
    "security_rating": 2.0,
    "security_zone": "federated",
    "x": 0,
    "y": 0
  }
]
```

| Field            | Type   | Required | Description                                                   |
| ---------------- | ------ | -------- | ------------------------------------------------------------- |
| `system_id`      | string | yes      | Snake_case unique identifier                                  |
| `name`           | string | yes      | Display name                                                  |
| `description`    | string | no       | Flavor text                                                   |
| `security_rating`| float  | yes      | Numeric rating. Federated=2.0, High=0.6-1.0, Low=0.1-0.2    |
| `security_zone`  | string | yes      | `"federated"`, `"high"`, `"medium"`, `"low"`, `"black"`      |
| `x`              | int    | yes      | Map coordinate X (arbitrary units, used for display)          |
| `y`              | int    | yes      | Map coordinate Y                                              |

---

# 6. jump_connections Array

Defines bidirectional jump corridors between systems. All connections are two-way.

```json
"jump_connections": [
  {
    "from_system_id": "nexus_prime",
    "to_system_id": "gateway_station",
    "fuel_cost": 5
  }
]
```

| Field            | Type   | Required | Description                                           |
| ---------------- | ------ | -------- | ----------------------------------------------------- |
| `from_system_id` | string | yes      | One end of the connection                             |
| `to_system_id`   | string | yes      | Other end of the connection                           |
| `fuel_cost`      | int    | yes      | Energy units required to traverse this jump point     |

---

# 7. ports Array

Defines all ports within each system.

```json
"ports": [
  {
    "port_id": "nexus_prime_starbase",
    "name": "Nexus Prime Starbase",
    "system_id": "nexus_prime",
    "port_type": "trading",
    "description": "The origin station for all new pilots.",
    "services": {
      "has_bank": true,
      "interest_rate_percent": 1.5,
      "has_shipyard": false,
      "has_upgrade_market": false,
      "has_drone_market": false,
      "has_missile_supply": false,
      "has_repair": true,
      "has_fuel": true
    },
    "commodities": {
      "produces": ["food_supplies", "luxury_goods", "electronics", "refined_ore", "machinery"],
      "consumes": ["raw_ore", "fuel_cells", "food_supplies", "luxury_goods", "electronics"]
    }
  }
]
```

### Port Fields

| Field         | Type   | Required | Description                                                 |
| ------------- | ------ | -------- | ----------------------------------------------------------- |
| `port_id`     | string | yes      | Snake_case unique identifier (must be unique across world)  |
| `name`        | string | yes      | Display name                                                |
| `system_id`   | string | yes      | Parent system                                               |
| `port_type`   | string | yes      | `"trading"`, `"mining"`, `"refueling"`, `"black_market"`   |
| `description` | string | no       | Flavor text                                                 |
| `services`    | object | yes      | Service flags (see below)                                   |
| `commodities` | object | yes      | What the port sells and buys                                |

### services Object

| Field                  | Type   | Default | Description                                             |
| ---------------------- | ------ | ------- | ------------------------------------------------------- |
| `has_bank`             | bool   | false   | Players can open bank accounts and deposit/withdraw     |
| `interest_rate_percent`| float  | 0.0     | Annual interest rate (applied per `interest_apply_interval_ticks`) |
| `has_shipyard`         | bool   | false   | Players can purchase new ships (Phase 2)                |
| `has_upgrade_market`   | bool   | false   | Players can buy ship upgrades (Phase 2)                 |
| `has_drone_market`     | bool   | false   | Players can buy drones (Phase 2)                        |
| `has_missile_supply`   | bool   | false   | Players can buy missiles                                |
| `has_repair`           | bool   | true    | Players can repair hull damage                          |
| `has_fuel`             | bool   | true    | Players can refuel                                      |

### commodities Object

| Field      | Type     | Required | Description                                                |
| ---------- | -------- | -------- | ---------------------------------------------------------- |
| `produces` | string[] | yes      | Commodity IDs the port sells to players (port has supply)  |
| `consumes` | string[] | yes      | Commodity IDs the port buys from players (port has demand) |

A commodity can appear in both arrays (port both sells and buys it, e.g. the origin starbase).

---

# 8. Validation Rules

The server validates the world config on load and exits if:

- Any `system_id` referenced in `jump_connections` or `ports` does not exist in `systems`
- Any `port_id` is duplicated
- Any `system_id` is duplicated
- A jump connection references itself (`from == to`)
- A commodity in `ports.commodities` is not in the top-level `commodities` array
- `starting_system_id` in `server.json` does not exist in this world

---

# End of Document
