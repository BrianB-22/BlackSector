# Communications System Specification

## Version: 0.3

## Status: Draft

## Owner: Core Architecture

## Last Updated: 2026-03-05

---

# 1. Purpose

Defines the player-to-player, player-to-all, and player-to-drone communications system for BlackSector.

Communication is asynchronous by design. Many players will not be online simultaneously. Messages are stored and delivered when recipients are available.

Communication is not consequence-free. Transmitting a signal exposes your position, increases your electromagnetic signature, and in combat, actively degrades your tactical situation.

---

# 2. Design Principles

* Messaging is async — no real-time chat. Messages are stored and delivered.
* All transmissions have a risk cost proportional to range and power
* The IRN (Interplanetary Relay Network) is a simulated in-universe infrastructure, not a game abstraction
* Communication can be disrupted, intercepted, and jammed
* Offline players receive messages in their mailbox on next login
* Messages are in-universe artifacts — they have sender, range, delay, and exposure
* The IRN carries not just messages but control signals to player-owned remote drones

---

# 3. Message Types

| Type              | Range          | Recipient         | Delay       | Exposure   |
| ----------------- | -------------- | ----------------- | ----------- | ---------- |
| Proximity         | Sensor range   | Ships in range    | Immediate   | Low        |
| System Broadcast  | Current system | All in system     | Immediate   | Medium     |
| IRN Direct        | Galaxy-wide    | Specific player   | Simulated   | Medium-High|
| IRN Broadcast     | Galaxy-wide    | All players       | Simulated   | High       |
| Distress Beacon   | Galaxy-wide+   | All players       | Immediate   | Maximum    |
| Dead Drop         | Port only      | Named recipient   | On pickup   | None       |
| Drone Command     | Galaxy-wide    | Owned drone       | Simulated   | Medium-High|
| Drone Telemetry   | Galaxy-wide    | Drone owner       | Simulated   | Medium     |

Drone Command and Drone Telemetry share the IRN infrastructure and are subject to the same delay, coverage, and disruption rules as IRN Direct.

---

# 4. Transmission Exposure

Every transmission emits an electromagnetic signature. This has two consequences:

## 4.1 Position Reveal

Transmitting briefly reveals your position to nearby sensors.

| Message Type      | Position Reveal                              |
| ----------------- | -------------------------------------------- |
| Proximity         | None beyond normal sensor visibility          |
| System Broadcast  | Exact position revealed to all in system     |
| IRN Direct        | Position revealed to nearby relay nodes (logged, interceptable) |
| IRN Broadcast     | Position revealed system-wide at origin      |
| Distress Beacon   | Position broadcast galaxy-wide               |
| Drone Command     | Same as IRN Direct (at command origin)       |
| Drone Telemetry   | Drone position revealed in drone's system    |

Revealed position persists for 1–3 ticks (configurable) after transmission ends.

## 4.2 Combat Tracking Penalty

Transmitting while in combat generates an EM spike that aids missile guidance systems.

| Message Type      | Tracking Accuracy Bonus to Attacker |
| ----------------- | ------------------------------------ |
| Proximity         | +0%                                  |
| System Broadcast  | +5%                                  |
| IRN Direct        | +10%                                 |
| IRN Broadcast     | +15%                                 |
| Distress Beacon   | +25%                                 |
| Drone Command     | +10%                                 |

The penalty applies for the tick in which the message is sent plus 2 subsequent ticks (EM dissipation window).

This creates a meaningful tactical decision: calling for help in combat may make your situation worse before it gets better. Commanding a drone while under fire carries the same risk.

---

# 5. Proximity Messaging

Sent to all ships within the player's current sensor range in the same system.

* No IRN infrastructure required
* Delivered immediately to online recipients
* Stored for offline recipients in the same system at time of send (they receive it on login if still in range — otherwise message is dropped)
* Very low exposure: only ships already capable of detecting you can receive it

```
[PROX] nova → ghost:  "Nice flying out there."
[PROX] nova → all nearby: "Abandon your cargo or be destroyed."
```

Recipient list is determined server-side at time of transmission. Sender does not know exactly who received it.

---

# 6. System Presence

## 6.1 Who Command

Players may query who else is currently in their system.

```
who                 — list all players present in the current star system
```

Output format:

```
Players in Vega Prime (Low Security):
  nova          [undocked — sector 14-C]
  ghost         [docked at Port Helios]
  xander        [undocked — sector 9-A]
```

Rules:

* Shows only players whose ships are in the **same star system** as the caller
* Ships in SILENT mode are **not listed** (their sensor signature is suppressed)
* Ships in combat are listed normally
* Docked ships are visible and their port is shown
* Undocked ships show their sector within the system
* No exposure penalty for using `who` — it reads passive sensor returns, not active ping

`who` output is generated server-side from current ship positions. It reflects the state at the tick the command is processed.

## 6.2 System Broadcast

Sent to all players currently in the same star system.

* Delivered immediately to online recipients in the system
* Stored in mailbox for offline recipients who were in the system at send time
* Medium exposure: exact position revealed to everyone in system

```
[SYS] xander → system: "Selling refined ore at port 22, best prices in the sector."
```

---

# 7. Interplanetary Relay Network (IRN)

The IRN is a network of automated relay buoys distributed throughout known space. Major relay nodes are co-located at ports. Minor relay nodes exist throughout systems. Player-deployed relay drones (see Section 14) can extend IRN coverage into dead zones.

The IRN enables galaxy-wide messaging and drone control at the cost of delay and exposure.

## 7.1 IRN Direct (Player to Player)

Sends a message to a specific player by name, regardless of their location.

* Recipient addressed by `player_name`
* If recipient is online: delivered to their session at end of simulated delay
* If recipient is offline: stored in mailbox, delivered on next login
* Simulated delay: 5–20 ticks depending on distance between origin system and nearest relay node (configurable)

```
[IRN] nova → ghost: "Meet me at Vega Station. I have a deal."
  [Delivered in 12 ticks via IRN relay]
```

## 7.2 IRN Broadcast (All Players)

Sends a message to all players in the galaxy.

* Delivered to all online players immediately (at end of simulated delay)
* Stored in mailbox for all offline players
* High exposure: significant EM signature at origin system

```
[IRN-BROADCAST] xander: "Trading consortium forming. Contact me for details."
```

Broadcast messages are rate-limited: maximum 2 per player per hour (configurable) to prevent spam.

## 7.3 IRN Delay Model

Delay is simulated, not real. It exists for immersion and to prevent instantaneous galaxy-wide alerts.

```
BaseDelay = 5 ticks
DistanceFactor = floor(distance_to_nearest_relay / 10) ticks
TotalDelay = BaseDelay + DistanceFactor
```

Maximum delay: 20 ticks (configurable).

Systems with no port have reduced relay access. Black Sector systems have degraded relay coverage (see Section 12). Systems without any relay infrastructure (no port, no relay drone) cannot send or receive IRN signals.

---

# 8. Distress Beacon

A special high-power emergency transmission. Maximum range. Immediate delivery. Maximum exposure.

* Broadcast to all players galaxy-wide — no delay
* Contains: sender name, system, position, optional short message (80 chars)
* Generates the largest EM signature in the game — position revealed galaxy-wide for 5 ticks
* In combat: +25% tracking bonus to attacker for 3 ticks

```
[DISTRESS] nova — System 14 (Vega Prime) — Position (42.5, -18.3)
  "Under attack by three pirates. Requesting assistance."
```

Use cases:
* Calling for help when under attack
* Alerting the community to a threat
* (Abuse case) False distress to lure players — carries full exposure risk regardless

Cooldown: 10 minutes between distress beacons per player (configurable).

---

# 9. Dead Drop

A message left at a specific port for a named recipient. No transmission — no exposure.

* Player docks at a port and deposits a message
* Message stored server-side tagged to the port and recipient
* When the named recipient docks at that port, they are notified of a waiting message
* No exposure because no signal is transmitted — the message travels with docking activity
* Contents remain private — not interceptable

```
[DEAD DROP — Vega Station]  nova left you a message.
  "The coordinates you wanted: (88.2, -44.1). Don't tell anyone."
```

Dead drops expire after 7 days (configurable) if uncollected.

Dead drops to unknown player names are rejected at deposit time.

---

# 10. Mailbox

Each player has a persistent in-game mailbox.

* Unread message count displayed in HUD and on login
* Messages stored until read or expired (default 30 days)
* Maximum mailbox size: 200 messages (configurable — oldest deleted when full)
* Messages marked read when opened

Login notification:

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  You have 3 unread messages.
  Type 'mail' to read them.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

Mailbox commands:

```
mail                     — list messages
mail read <id>           — read a message
mail delete <id>         — delete a message
mail reply <id> <text>   — reply via IRN Direct
mail clear               — delete all read messages
```

---

# 11. IRN Coverage and Disruption

## 11.1 Coverage by Zone

| Security Zone  | IRN Reliability | Delay Modifier |
| -------------- | --------------- | -------------- |
| High Security  | 100%            | ×1.0           |
| Medium Security| 95%             | ×1.0           |
| Low Security   | 80%             | ×1.5           |
| Black Sector   | 50%             | ×2.0           |

In Black Sector, IRN messages have a 50% chance of delivery failure. Failed messages are logged as undeliverable and refunded (no exposure charged for failed sends). Drone commands sent to drones in Black Sector are subject to the same failure rate.

## 11.2 Relay Disruption Events

An economic/chaos event can knock out IRN relay infrastructure in a region.

Event: `irn_relay_outage`

Effects:
* IRN Direct, IRN Broadcast, and Drone Commands fail for all systems in the affected region
* Drones in the affected region enter autonomous standby (see Section 14.5)
* Proximity and System Broadcast still work (no relay required)
* Duration: 30–120 minutes (configurable)
* Publicly announced when it starts: `[IRN] Relay outage detected in Outer Rim. Long-range comms disrupted.`

This creates isolation as a gameplay mechanic — a region can be cut off from the rest of the galaxy.

---

# 12. IRN Interception

Players in Low Security or Black Sector systems with a **Signal Intercept** upgrade can intercept IRN messages passing through their current system.

Interception mechanics:
* Passive — no action required, just having the upgrade active
* Only intercepts IRN messages that route through your system (based on relay path simulation)
* Intercept probability: 40% per message in Low Security, 70% in Black Sector (configurable)
* Intercepted messages appear in a separate `intercept log`: `[INTERCEPTED] nova → ghost: "..."`
* Drone commands and drone telemetry are also interceptable — revealing drone positions and owner intent
* Sender is not notified of the interception
* Encrypted messages (future upgrade) cannot be intercepted

This creates an intelligence-gathering layer — operating in dangerous space with the right equipment lets you know what others are communicating, and where their drones are deployed.

---

# 13. Communications Jamming

Ships or installations with a **Comms Jammer** upgrade can suppress IRN communications in their local system.

Jamming mechanics:
* Active ability — player activates it, consumes energy per tick while active
* Blocks all IRN Direct, IRN Broadcast, and Drone Command transmission FROM the system
* Drones in a jammed system cannot receive commands or send telemetry — they enter autonomous standby
* Does not affect Proximity messaging
* Does not affect incoming messages
* Players attempting to transmit IRN from a jammed system receive: `[COMM ERROR] IRN signal blocked. Local jamming detected.`
* Jammer presence is detectable by sensors (it emits its own EM signature)

Use cases:
* Prevent a cornered enemy from calling for reinforcements
* Cut off a system during an ambush
* Disable enemy scout drones by jamming the system they patrol
* Area denial in Black Sector

Jammer energy cost: 5 energy per tick while active (configurable).

---

# 14. IRN Drone Control

Players purchase and deploy unmanned autonomous drones from trading ports. Drones are persistent physical objects in the game world that operate semi-autonomously — the player sets a directive, the drone executes it over subsequent ticks, and reports back progress via telemetry.

Drone control uses the IRN infrastructure — all commands and telemetry are subject to the same delay, exposure, and disruption rules as IRN Direct messages.

## 14.1 Acquisition

Drones are purchased at trading ports. Not every port stocks every type — availability depends on port size and security zone.

| Port Type          | Mapping Drone | Prospecting Drone | Decoy Drone | Relay Drone |
| ------------------ | ------------- | ----------------- | ----------- | ----------- |
| Large (High Sec)   | Yes           | Yes               | Yes         | Yes         |
| Medium (High Sec)  | Yes           | Yes               | No          | Yes         |
| Small (High Sec)   | Limited       | No                | No          | No          |
| Low Security       | Black market  | Black market      | Black market| No          |
| Black Sector       | No            | No                | No          | No          |

Drone inventory restocks with the port's standard trade cycle. Prices fluctuate based on supply.

Once purchased, the drone is stored as ship cargo (occupies cargo space) until deployed. Deploying a drone removes it from cargo and creates a persistent drone entity in the current system.

## 14.2 Drone Types

| Type              | Purpose                                          | Cargo | Active Sensors | IRN Required |
| ----------------- | ------------------------------------------------ | ----- | -------------- | ------------ |
| Mapping Drone     | Survey, scan, map anomalies and system features  | No    | Yes            | For commands |
| Prospecting Drone | Sample asteroid fields, assess resource quality  | Small | Limited        | For commands |
| Decoy Drone       | Emit false EM signatures to fool sensors         | No    | No             | For commands |
| Relay Drone       | Extend IRN coverage into dead zones (passive)    | No    | No             | No           |

## 14.3 Mapping Drones

Mapping drones are mobile sensor platforms designed for systematic survey work. They fly semi-autonomous search patterns and report discoveries back to the owner.

Directives:
* **Survey area**: Grid-sweep a defined bounding box over multiple ticks. Reports anomalies, asteroid fields, hazard zones, and jump points found.
* **Point scan**: Navigate to a specific position and perform a deep scan. Higher resolution than a ship's passive sensors.
* **Perimeter sweep**: Orbit a point at a given radius, scanning outward.
* **Goto**: Navigate to specific coordinates and hold.

Capabilities:
* Detects everything a player ship can detect, at comparable range
* Discoveries are added to the owner's `player_map_data` — same as if the player flew there
* Reports back with findings summary on directive completion or on `drone report` command

Limitations:
* Single-system only — cannot navigate across jump points
* Energy depletes over time; enters standby at 10% energy
* Fragile hull — destroyed easily if targeted

```
[DRONE] Eye-1 — System 22 (Proxima Cross) — Pos (33.2, -12.1)
  Directive: SURVEY (18% complete)   Hull: 45/50   Energy: 80/100
  Findings so far: 1 asteroid field, 0 anomalies
```

## 14.4 Prospecting Drones

Prospecting drones are not mining platforms — they are survey instruments for assessing resource value before a player commits time to manual mining.

Directives:
* **Prospect field**: Navigate to a designated asteroid field and run a full resource composition scan over several ticks.
* **Sample**: Take a physical micro-sample from the current asteroid field (stores 1–2 units — confirms grade, does not substitute for real mining).
* **Goto**: Navigate to coordinates.

Report output after prospecting:
```
[DRONE] Sampler-1 — Asteroid Field ID 8 (System 22)
  Composition scan complete:
    Primary:   Titanite (grade A, density 7/10)
    Secondary: Iron Silicate (grade C, density 4/10)
  Estimated yield (manual mining, 10 ticks): ~140 units titanite
  Sample collected: 1 unit titanite (grade A confirmed)
```

Limitations:
* Cannot perform sustained extraction — exploratory only
* Cargo capacity: 2 units (sample confirmation only)
* Must dock at port or rendezvous with player ship to offload sample

## 14.5 Decoy Drones

Decoy drones emit a crafted electromagnetic signature designed to register as a ship on enemy sensors. They are tactical deception tools.

Directives:
* **Mimic**: Configure the EM signature profile (courier, freighter, or fighter class).
* **Pattern — static**: Hold position and broadcast the false signature.
* **Pattern — drift**: Move slowly in a set direction (simulates a coasting ship).
* **Pattern — orbit**: Circle a point at a set radius (simulates an active patrol).
* **Activate / Deactivate**: Toggle the signature emission without recalling the drone.

Behavior:
* While active, other players' sensors detect the decoy as a ship of the configured class
* Players with advanced sensors (Signal Intercept upgrade) have a chance to classify it as a decoy rather than a real ship — probability increases at close range
* The decoy's EM signature is subject to interception — an enemy watching the system may see that a new ship appeared from nowhere
* In combat, missiles will target decoys if they are the strongest EM source in range — gives the real ship one missile redirection per decoy

Limitations:
* High energy consumption while active: 8 energy per tick (configurable)
* Cannot navigate while emitting — must deactivate to move, then reactivate
* No sensors — blind to surroundings
* Destroyed in one hit from any weapon

```
[DRONE] Ghost-1 — System 14 — Pos (51.0, -8.5)
  Mode: ACTIVE DECOY (Courier profile)   Energy: 62/100
  Pattern: ORBIT (radius 5.0 around 50.0, -9.0)
```

## 14.6 Relay Drones

Relay drones are passive infrastructure devices that extend IRN coverage into systems with no port.

Mechanics:
* Deployed by the owner from within the target system
* Stationary after deployment — no navigation capability
* Extends IRN coverage for the system: treated as Low Security reliability (80%) for IRN purposes
* Emits a constant low-level EM signature — detectable by sensors at medium range
* Can be destroyed by other players (small hull, no weapons)
* Can be recalled by the owner at any time

Use cases:
* Establish a comms link into an unmapped or Black Sector system
* Maintain IRN contact with a mapping or prospecting drone in a dead zone
* Build forward operating infrastructure for extended Black Sector operations

```
[DRONE] Relay-A — System 47 (Dark Reach) — Status: PASSIVE RELAY
  IRN Coverage: extended to this system (80% reliability).
```

## 14.7 Autonomous Standby Behavior

When a drone loses IRN contact (jammed system, relay outage, Black Sector delivery failure, or low energy), it enters autonomous standby:

* Ceases active operations
* Mapping and prospecting drones maintain passive sensors
* Decoy drones deactivate their emission (no energy wasted)
* Does not transmit (no exposure, no telemetry)
* Continues executing any command already received and not yet completed
* Resumes normal operation and queued commands when IRN contact is re-established

A player cannot instantly cut off an enemy's drone by jamming — a directive already executing will run to completion.

## 14.8 Drone Exposure

| Event                           | Exposure at Owner's Location        | Exposure at Drone's Location       |
| ------------------------------- | ----------------------------------- | ---------------------------------- |
| Sending drone command           | Same as IRN Direct (+10% tracking)  | None                               |
| Receiving drone telemetry       | None                                | Drone position briefly revealed    |
| Relay drone passive operation   | None                                | Low constant EM (detectable)       |
| Mapping drone active scan       | None                                | Spike at drone's location          |
| Decoy drone active emission     | None                                | Ship-class EM signature (constant) |

A skilled adversary with Signal Intercept can infer drone positions and the owner's tactical intent from intercepted telemetry and commands.

## 14.9 Drone Limits

| Resource              | Default | Configurable |
| --------------------- | ------- | ------------ |
| Max drones per player | 5       | Yes          |
| Max relay drones      | 2       | Yes          |
| Max decoys active     | 2       | Yes          |

* Destroyed drones are permanently lost — must be replaced at a port
* Drones persist across server restarts via snapshot recovery
* Drones cannot be traded between players (bound to owner on purchase)

---

# 15. Transmission Signature and Sensors

Transmissions are detectable by sensors beyond just their position reveal.

Players with enhanced sensors can detect:
* That a transmission occurred in a nearby system (not content, just that a signal was sent)
* Rough classification of transmission type (proximity vs. IRN vs. drone telemetry)
* This is part of the broader sensor and intelligence system

Signal detection range:

| Message Type      | Detectable at Range     |
| ----------------- | ----------------------- |
| Proximity         | Within sensor range only |
| System Broadcast  | Same system             |
| IRN Direct        | Same system + adjacent  |
| IRN Broadcast     | Same system + 2 hops    |
| Distress Beacon   | Galaxy-wide             |
| Drone Command     | Same system + adjacent  |
| Drone Telemetry   | Same system (drone's location) |
| Relay Drone       | Same system + adjacent (passive EM) |

---

# 16. Data Model

The canonical schema for all communications and drone tables is in `docs/10_data_models/database_schema.md`. That file is the source of truth. If this document and the schema file ever conflict, the schema file wins.

Tables defined there that belong to this system:

| Table                  | Schema Section | Purpose                                      |
| ---------------------- | -------------- | -------------------------------------------- |
| `messages`             | Section 14     | All player-to-player messages, IRN delivery  |
| `drones`               | Section 14     | Deployed drone state                         |
| `drone_commands`       | Section 14     | Queued commands with IRN delay               |
| `drone_telemetry`      | Section 14     | Drone reports delivered back to owner        |
| `ship_drone_inventory` | Section 5      | Undeployed drones in ship bays               |

---

# 17. Tick Integration

Communications are processed during the tick loop.

Tick steps (added to headless_server_spec.md tick loop):

```
After step 12 (emit protocol messages to sessions):

13a. Process pending IRN deliveries (deliver_at_tick <= current_tick)
13b. Notify online recipients of new messages
13c. Expire old messages (expires_at_tick <= current_tick)
13d. Apply transmission exposure effects (position reveal, tracking bonus)
13e. Process pending drone commands (deliver_at_tick <= current_tick)
13f. Execute drone autonomous behavior (movement, mining, standby checks)
13g. Generate and queue drone telemetry for pending report responses
13h. Deliver pending drone telemetry to owners
```

Transmission exposure effects are applied in the tick the message is sent (step 7 of existing tick loop, as a game event).

---

# 18. Protocol Messages

## Incoming message notification (Server → Client)

```json
{
  "type": "message_received",
  "timestamp": 8205,
  "correlation_id": null,
  "payload": {
    "message_id": "...",
    "message_type": "irn_direct",
    "sender_name": "nova",
    "content": "Meet me at Vega Station.",
    "sent_tick": 8190,
    "origin_system_id": 14
  }
}
```

## Send message command (Client → Server)

```json
{
  "type": "command_submit",
  "timestamp": 8200,
  "correlation_id": "...",
  "payload": {
    "command": "send_message",
    "parameters": {
      "message_type": "irn_direct",
      "recipient_name": "ghost",
      "content": "I have what you're looking for."
    }
  }
}
```

## Drone command (Client → Server)

```json
{
  "type": "command_submit",
  "timestamp": 8200,
  "correlation_id": "...",
  "payload": {
    "command": "drone_command",
    "parameters": {
      "drone_id": "d-001",
      "command_type": "scan",
      "parameters": {}
    }
  }
}
```

## Drone command with parameters (Client → Server)

```json
{
  "type": "command_submit",
  "timestamp": 8200,
  "correlation_id": "...",
  "payload": {
    "command": "drone_command",
    "parameters": {
      "drone_id": "d-001",
      "command_type": "move",
      "parameters": { "x": 44.5, "y": -21.0 }
    }
  }
}
```

## Drone telemetry (Server → Client)

```json
{
  "type": "drone_telemetry",
  "timestamp": 8225,
  "correlation_id": null,
  "payload": {
    "drone_id": "d-001",
    "drone_name": "Eye-1",
    "report_type": "scan_result",
    "origin_system_id": 22,
    "generated_tick": 8210,
    "data": {
      "ships_detected": 2,
      "signal_detected": true,
      "anomalies": [],
      "hull_points": 45,
      "energy_points": 80
    }
  }
}
```

## Drone destroyed notification (Server → Client)

```json
{
  "type": "drone_telemetry",
  "timestamp": 8300,
  "correlation_id": null,
  "payload": {
    "drone_id": "d-001",
    "drone_name": "Eye-1",
    "report_type": "destroyed",
    "origin_system_id": 22,
    "generated_tick": 8295,
    "data": {
      "cause": "weapons_fire"
    }
  }
}
```

## Transmission exposure event (Server → Client, combat)

```json
{
  "type": "combat_update",
  "timestamp": 8200,
  "correlation_id": null,
  "payload": {
    "event": "transmission_detected",
    "tracking_bonus": 0.10,
    "duration_ticks": 3,
    "source": "irn_direct_transmission"
  }
}
```

## IRN outage notification (Server → Client)

```json
{
  "type": "irn_outage",
  "timestamp": 8100,
  "correlation_id": null,
  "payload": {
    "affected_region_id": 4,
    "affected_region_name": "Outer Rim",
    "estimated_duration_ticks": 1800
  }
}
```

---

# 19. Player Commands

```
msg <player_name> <text>         — send IRN Direct message
msg all <text>                   — send IRN Broadcast
msg nearby <text>                — send Proximity message
msg system <text>                — send System Broadcast
distress [<text>]                — send Distress Beacon
drop <player_name> <text>        — leave Dead Drop at current port (must be docked)
mail                             — open mailbox
mail read <id>
mail reply <id> <text>
mail delete <id>
intercept log                    — view intercepted messages (requires Signal Intercept upgrade)
jam start / jam stop             — toggle comms jammer (requires Comms Jammer upgrade)

drone list                              — list all your drones and last known status
drone info <name>                       — show last telemetry report for a drone
drone report <name>                     — request fresh status telemetry from drone
drone goto <name> <x> <y>              — navigate drone to coordinates
drone return <name>                     — navigate drone to your current position
drone standby <name>                    — put drone into low-power standby
drone recall <name>                     — command drone to dock at nearest port
drone deploy <name>                     — deploy drone from cargo at current position

— Mapping drone directives:
drone survey <name> <x1> <y1> <x2> <y2>  — survey a grid area
drone scan <name>                         — perform a deep point scan at current position
drone sweep <name> <cx> <cy> <radius>     — perimeter sweep around a point

— Prospecting drone directives:
drone prospect <name> <field_id>       — navigate to asteroid field and run composition scan
drone sample <name>                    — take a physical sample at current position

— Decoy drone directives:
drone mimic <name> <courier|freighter|fighter>  — set EM signature profile
drone pattern <name> static                     — hold position and emit
drone pattern <name> drift <dx> <dy>            — drift in direction while emitting
drone pattern <name> orbit <cx> <cy> <radius>   — orbit a point while emitting
drone activate <name>                           — start emitting decoy signature
drone deactivate <name>                         — stop emitting (drone stays deployed)

— Relay drone:
drone deploy relay <name>              — deploy relay drone at current position (must be in system)
```

---

# 20. Balancing Guidelines

* IRN Broadcast spam must be prevented — 2 per hour rate limit is the baseline
* Distress Beacon must feel meaningful — cooldown and maximum exposure make it a last resort
* Transmission exposure should be noticeable in combat but not automatically fatal — it tips the balance, not decides it
* Interception should feel like a reward for operating in dangerous space, not a routine occurrence
* Dead Drops should encourage player-to-player intrigue and information trade
* Drone control adds strategic depth but must not replace active play — drones cannot fight or trade on the player's behalf
* Relay drones should feel valuable but destroyable — infrastructure that can be contested
* Drone commands should feel consequential due to IRN delay — not a real-time remote control
* Decoy drones should be a tool of deception, not a hard counter — advanced sensors should reliably detect them at close range
* Prospecting drones give intel, not yield — the player still has to do the actual mining
* Port availability and cost should make drones feel like an investment, not throwaway items

---

# 21. Non-Goals (v1)

* Real-time chat or voice
* Group channels or guilds
* Encrypted messages (future upgrade)
* In-game forum or bulletin boards
* Message forwarding chains
* Automated NPC message responses
* Armed or combat drones
* Drone-to-drone communication
* Drone swarms or formation control
* Cross-system drone navigation (drones are single-system only in v1)
* Drone cargo delivery to ports without player involvement

---

# 22. Future Extensions

* Encrypted transmission upgrade (blocks interception)
* Faction broadcast channels
* Message forwarding via trusted intermediary
* IRN message spoofing (fake sender identity — requires upgrade)
* Bulletin boards at ports (public message boards per port)
* Automated distress response from nearby NPC vessels
* Message-based mission triggers (NPC contacts player via IRN)
* Heavy mining drones (full extraction, larger cargo, requires retrieval)
* Armed guard drone variant (limited weapons, patrols a waypoint path)
* Drone upgrade slots (enhanced sensors, extended range, larger cargo)
* Drone-to-drone relay chains for deep Black Sector coverage
* Cross-system drone transfer via jump point (advanced drone class)

---

# End of Document
