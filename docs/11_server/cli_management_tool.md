# CLI Management Tool Specification

## Version: 0.1

## Status: Draft

## Owner: Core Architecture

## Last Updated: 2026-03-05

---

# 1. Purpose

Defines the admin command interface for the BlackSector server.

The CLI management tool is the primary mechanism for server administrators to manage, monitor, and control a running server instance.

Admin commands are submitted via stdin of the server process. No separate admin client is required.

---

# 2. Design Principles

* Commands are simple line-oriented text
* No authentication required — physical access to the process stdin is assumed
* Commands execute at the start of the next tick (non-blocking)
* Command output is written to stdout
* Commands are safe: no destructive operations without confirmation

---

# 3. Command Format

```
<category> <subcommand> [arguments]
```

Examples:

```
server status
player list
mission reload
```

Arguments are space-separated. Quoted strings are supported for names containing spaces.

---

# 4. Server Commands

## server status

Displays current server state.

```
> server status

Server: Black Sector
Uptime: 4h 32m
Current Tick: 8,204
Tick Interval: 2000ms
Last Tick Duration: 47ms
Connected Players: 12 / 100
Active Sessions: 12
Protocol Version: 1.0
```

---

## server shutdown

Initiates graceful shutdown.

```
> server shutdown

[WARNING] Initiating graceful shutdown. Connected players will be notified.
Type 'confirm' to proceed:
> confirm
Shutdown initiated. Waiting for tick to complete...
Final snapshot written.
Server stopped. (Tick 8,205)
```

---

## server snapshot

Forces an immediate snapshot write outside the normal interval.

```
> server snapshot

Snapshot written: snapshots/snapshot_008205_1709612345.json
```

---

## server reload

Hot-reloads mission definitions and economic event definitions.

```
> server reload

Reloading missions... 14 missions loaded, 0 errors.
Reloading economic events... 8 events loaded, 0 errors.
Reload complete.
```

---

## server config

Displays active server configuration values.

```
> server config

tick_interval_ms: 2000
max_concurrent_players: 100
snapshot_interval_ticks: 100
linger_timeout_seconds: 300
...
```

---

# 5. Player Commands

## player list

Lists all connected players.

```
> player list

ID                                    Name        Mode   System   Tick Connected
550e8400-e29b-41d4-a716-446655440000  Orion       TEXT   14       8200  4h 12m
...
(12 players connected)
```

---

## player info \<player_id_or_name\>

Displays detailed information about a player.

```
> player info Orion

Player: Orion
ID: 550e8400-e29b-41d4-a716-446655440000
Status: Connected (TEXT)
Session: abc123... (connected 4h 12m ago)
Ship: courier | Hull 80/100 | Shields 40/50 | Energy 75/100
Location: System 14 (Vega Prime) | Position (12.5, -8.3)
Cargo: food_supplies x10, refined_ore x5
Active Missions: pirate_hunt (IN_PROGRESS)
```

---

## player kick \<player_id_or_name\> [\<reason\>]

Disconnects a player session.

```
> player kick Orion "Server maintenance"

Player Orion kicked. Reason: Server maintenance
```

---

## player ban \<player_id_or_name\> [\<reason\>]

Bans a player account. Disconnects active session if present.

```
> player ban Orion "Terms of service violation"

Player Orion banned. Active session terminated.
```

---

## player unban \<player_id_or_name\>

Removes a ban from a player account.

```
> player unban Orion

Player Orion unbanned.
```

---

## player teleport \<player_id_or_name\> \<system_id\>

Moves a player's ship to the specified system. For admin use only.

```
> player teleport Orion 1

Player Orion teleported to System 1 (Sol).
```

---

# 6. Mission Commands

## mission list

Lists all loaded mission definitions.

```
> mission list

ID                 Name                   Enabled  Active Instances
pirate_hunt        Pirate Hunt            YES      3
escort_convoy      Escort the Convoy      YES      1
ore_delivery       Ore Delivery Run       YES      5
black_market_run   Black Market Run       NO       0

(4 missions loaded)
```

---

## mission info \<mission_id\>

Displays full details for a mission definition.

```
> mission info pirate_hunt

Mission: pirate_hunt
Name: Pirate Hunt
Enabled: YES
Source: config/missions/combat.json
Objectives: 2
  [0] kill - npc_pirate x3 (ACTIVE for 3 players)
  [1] navigate_to - system 14 (PENDING for 3 players)
Rewards: 3000 credits
Active Instances: 3
```

---

## mission enable \<mission_id\>

Makes a mission available for players to accept.

```
> mission enable black_market_run

Mission black_market_run enabled.
```

---

## mission disable \<mission_id\>

Prevents new players from accepting a mission. Existing instances continue.

```
> mission disable black_market_run

Mission black_market_run disabled. 0 active instances unaffected.
```

---

## mission validate \<file_path\>

Validates a mission JSON file without loading it.

```
> mission validate config/missions/community/new_mission.json

Validation: PASS
  3 missions found.
  All objective types valid.
  All reward fields valid.
```

---

## mission status \<player_id_or_name\>

Displays active mission instances for a player.

```
> mission status Orion

Player: Orion
Active Missions:
  pirate_hunt (IN_PROGRESS)
    [0] kill npc_pirate: 2/3 (ACTIVE)
    [1] navigate_to system 14: 0/1 (PENDING)
    Expires: tick 9200
```

---

## mission reset \<player_id_or_name\> \<mission_id\>

Cancels a player's active mission instance. The player may re-accept the mission.

```
> mission reset Orion pirate_hunt

Mission pirate_hunt reset for player Orion. Instance abandoned.
```

---

# 7. Economy Commands

## economy status

Displays current economic state summary.

```
> economy status

Active Economic Events: 2
  food_shortage   region:3  ends tick 9200  [PUBLIC]
  industrial_boom system:14 ends tick 8900  [PUBLIC]

Event Spawn Counter: 47
Last Spawn Tick: 8100
Next Eligible Spawn: minor=8320 regional=8500 global=9020
```

---

## economy event list

Lists all loaded economic event definitions.

```
> economy event list

ID                  Scope    Visibility  Duration (min)  Cooldown
food_shortage       region   public      60–180          360
industrial_boom     system   public      90–240          480
black_market_surge  region   hidden      120–240         600
```

---

## economy event trigger \<event_id\> [\<region_id_or_system_id\>]

Manually triggers an economic event. For testing and emergency use only.

```
> economy event trigger food_shortage 3

Economic event food_shortage triggered in region 3.
Duration: 120 minutes. End tick: 9260.
```

---

# 8. World Commands

## world status

Displays summary statistics about the game world.

```
> world status

Regions:  8
Systems:  512
Ports:    847
Jump Connections: 1,024
AI Traders Active: 62 / 80
```

---

## world system info \<system_id\>

Displays information about a specific system.

```
> world system info 14

System 14: Vega Prime
Region: Industrial Core (region 2)
Security Level: 0.72 (High Security)
Ports: 3 (trading x2, refueling x1)
Connections: systems 11, 17, 22
Players Present: 4
AI Traders Present: 2
```

---

# 9. Log Commands

## log tail [\<lines\>]

Displays the most recent event log entries.

```
> log tail 5

[8203] player_jump      player=Orion from=11 to=14
[8203] player_trade     player=Orion port=22 sell refined_ore x50 17000cr
[8204] combat_start     player=Orion vs npc_pirate system=14
[8204] combat_end       player=Orion victory ticks=12
[8205] mission_completed player=Orion pirate_hunt 3000cr
```

---

## log errors [\<count\>]

Displays recent error-level log entries.

```
> log errors 10

[8100] tick_slow tick=8100 duration=612ms
[7840] player_auth_fail addr=10.0.0.5 reason=invalid_token
```

---

# 10. Help Command

## help [\<category\>]

Displays available commands.

```
> help

Available command categories:
  server    — server management
  player    — player management
  mission   — mission management
  economy   — economy management
  world     — world information
  log       — log viewing

Type 'help <category>' for commands in a category.
```

---

# 11. Non-Goals

* Remote admin access (use SSH to the host machine)
* Web-based admin interface
* Role-based admin permissions
* Admin audit log (admin commands are logged in the event log)

---

# End of Document
