# Universe CLI Tool Specification

## Version: 0.1

## Status: Draft

## Owner: Core Architecture

## Last Updated: 2026-03-05

---

# 1. Purpose

Defines `bsctl` — the BlackSector control CLI tool.

`bsctl` is a standalone binary for universe creation, world management, player administration, and server health inspection. It operates directly against the SQLite database and configuration files, and can run whether the server is online or offline.

This tool is the primary interface for operators setting up a new server, managing the game world, and performing administrative tasks outside of live gameplay.

---

# 2. Scope

IN SCOPE:

* Universe and world creation
* System, region, port, and jump connection management
* World inspection and validation
* Player account management (offline)
* Economy inspection
* Server health and diagnostics
* Configuration validation
* Session and online player status

OUT OF SCOPE:

* Live gameplay commands (use server stdin admin interface — see `cli_management_tool.md`)
* Mission hot-reload during play (use `mission reload` in server admin)
* Direct tick engine control

---

# 3. Design Principles

* Works offline — operates directly against SQLite, no server connection required
* Read-only by default for inspection commands — destructive changes require `--confirm`
* Outputs human-readable tables by default; JSON output with `--json` flag
* All writes are wrapped in SQLite transactions — partial failures leave no corrupt state
* Does not require root or special OS permissions beyond file access to the data directory

---

# 4. Binary and Invocation

```
bsctl [--db <path>] [--config <dir>] <command> [subcommand] [arguments] [flags]
```

Defaults:

```
--db     ./data/blacksector.db
--config ./config/
```

Override for non-standard installations:

```
bsctl --db /opt/blacksector/data/blacksector.db --config /opt/blacksector/config/ world status
```

---

# 5. World Commands

## bsctl world status

High-level summary of the universe.

```
$ bsctl world status

Universe: Black Sector
Regions:  8
Systems:  512
Ports:    847
Jump Connections: 1,024
Asteroid Fields:  218
Anomalies:        94

Security Zone Distribution:
  High Security:    124 systems (24%)
  Medium Security:  201 systems (39%)
  Low Security:     143 systems (28%)
  Black Sector:      44 systems  (9%)
```

---

## bsctl world validate

Validates all world data for integrity issues.

```
$ bsctl world validate

Checking regions...        OK (8 regions)
Checking systems...        OK (512 systems)
Checking jump connections...  OK (1,024 connections)
  Checking connectivity...  OK (graph is fully connected)
Checking ports...          OK (847 ports)
Checking commodities...    OK (14 commodities in config)
Checking port inventories... OK

Validation PASSED. No issues found.
```

Reports specific errors with system/port IDs if issues are found:

```
ERROR: system_id 304 has no jump connections (isolated)
ERROR: port_id 112 references system_id 999 which does not exist
WARN:  region_id 5 has no ports
```

---

## bsctl world create

Interactive wizard to create a new universe from scratch.

```
$ bsctl world create

BlackSector Universe Creation Wizard
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

This will create a new universe in the database.
Existing world data will be cleared.

Enter universe name [Black Sector]: _
Number of regions [6]: _
Target system count [500]: _
Target port count [800]: _

Security zone balance:
  High Security  [25%]: _
  Medium Security[40%]: _
  Low Security   [25%]: _
  Black Sector   [10%]: _

Random seed [leave blank for random]: _

Generating universe...
  Creating 6 regions...        done
  Placing 500 systems...       done
  Building jump network...     done (994 connections)
  Placing 812 ports...         done
  Generating asteroid fields... done (204 fields)
  Placing anomalies...         done (87 anomalies)

Validating...                  PASSED

Universe created. Run 'bsctl world status' to verify.
```

Also supports non-interactive mode via flags:

```
bsctl world create --name "Black Sector" --regions 6 --systems 500 \
  --ports 800 --seed 42 --confirm
```

---

## bsctl world import \<file\>

Import world data from a JSON export file.

```
$ bsctl world import universe_backup.json

Importing world data...
  Regions:  8
  Systems:  512
  Ports:    847
  Connections: 1,024

Validating...  PASSED
Import complete.
```

---

## bsctl world export \<file\>

Export full world structure to a JSON file (without player data or economy state).

```
$ bsctl world export universe_backup.json

Exported 512 systems, 847 ports, 1024 connections to universe_backup.json
```

---

# 6. Region Commands

## bsctl region list

```
$ bsctl region list

ID  Name              Type           Security  Systems  Ports
 1  Core Worlds       core           0.85       82       148
 2  Industrial Belt   industrial     0.62      124       201
 3  Agri Corridor     agricultural   0.71       96       174
 4  Outer Rim         frontier       0.22       88        91
 5  Far Frontier      frontier       0.12       78        67
 6  The Void          black         -1.00       44        46
```

---

## bsctl region info \<id\>

```
$ bsctl region info 4

Region: Outer Rim (ID: 4)
Type:    frontier
Security Level: 0.22 (Low Security)
Systems: 88
Ports:   91
Avg Port Commodity Count: 4.2
Active Economic Events: 1
```

---

# 7. System Commands

## bsctl system list

```
$ bsctl system list [--region <id>] [--security <high|medium|low|black>]

ID    Name              Region            Security  Ports  Jumps
  14  Vega Prime        Industrial Belt   0.72        3      4
  31  Kepler's Rest     Outer Rim         0.18        1      2
 104  The Wound         The Void         -1.00        2      2
...
(showing 20 of 512 — use --limit or --all)
```

---

## bsctl system info \<id\>

```
$ bsctl system info 14

System: Vega Prime (ID: 14)
Region: Industrial Belt (ID: 2)
Security Level: 0.72 (High Security)
Position: (42.5, -18.3)
Hazard Level: 0.1

Ports (3):
  Port 22 — Vega Station       trading      sec 0.72
  Port 23 — Vega Mining Hub    mining       sec 0.72
  Port 24 — Refuel Point Alpha refueling    sec 0.72

Jump Connections:
  → System 11 (Alderon)       bidirectional  fuel mod 1.0
  → System 17 (Dusk Point)    bidirectional  fuel mod 1.2
  → System 22 (Iron Gate)     bidirectional  fuel mod 0.9
  → System 38 (The Crossing)  one-way        fuel mod 1.5

Asteroid Fields: 2
Anomalies: 0
```

---

## bsctl system add

```
$ bsctl system add --name "New Outpost" --region 4 --security 0.15 \
  --x 88.2 --y -44.1 --confirm

System created: ID 513, New Outpost (region 4, security 0.15)
```

---

## bsctl system edit \<id\>

```
$ bsctl system edit 14 --security 0.68 --confirm

System 14 (Vega Prime) updated: security_level 0.72 → 0.68
```

---

# 8. Port Commands

## bsctl port list

```
$ bsctl port list [--system <id>] [--type <trading|mining|refueling|black_market>]

ID   Name                  System  Type        Security  Docking Fee
 22  Vega Station          14      trading     0.72       50
 23  Vega Mining Hub       14      mining      0.72       20
104  Frontier Depot        31      trading     0.18      100
...
```

---

## bsctl port info \<id\>

```
$ bsctl port info 22

Port: Vega Station (ID: 22)
System: Vega Prime (14)
Type: trading
Security: 0.72
Docking Fee: 50 credits

Inventory (8 commodities):
  food_supplies    qty:  800  buy: 120  sell: 100
  fuel_cells       qty:  500  buy: 160  sell: 140
  refined_ore      qty:  200  buy: 280  sell: 250
  machinery        qty:   80  buy: 640  sell: 600
  electronics      qty:   45  buy: 850  sell: 800
  luxury_goods     qty:   12  buy: 1600 sell: 1500
  medical_supplies qty:  320  buy: 200  sell: 180
  raw_ore          qty: 1200  buy:  90  sell:  75
```

---

## bsctl port add

```
$ bsctl port add --system 14 --name "Black Market Den" \
  --type black_market --security 0.1 --docking-fee 200 --confirm

Port created: ID 848, Black Market Den (system 14, black_market)
```

---

## bsctl port edit \<id\>

```
$ bsctl port edit 22 --docking-fee 75 --confirm

Port 22 (Vega Station) updated: docking_fee 50 → 75
```

---

## bsctl port inventory set \<port_id\> \<commodity_id\> \<quantity\>

Manually set a port's stock of a commodity.

```
$ bsctl port inventory set 22 food_supplies 1000 --confirm

Port 22 food_supplies quantity set to 1,000.
```

---

# 9. Jump Connection Commands

## bsctl jump list \<system_id\>

```
$ bsctl jump list 14

Connections from System 14 (Vega Prime):

ID   From  To    Name              Bidirectional  Fuel Mod
 44    14   11   Vega-Alderon      YES            1.0
 45    14   17   Vega-Dusk Point   YES            1.2
 46    14   22   Vega-Iron Gate    YES            0.9
 47    14   38   Vega-The Crossing NO             1.5
```

---

## bsctl jump add

```
$ bsctl jump add --from 14 --to 55 --bidirectional --fuel-mod 1.1 --confirm

Jump connection created: ID 1025, System 14 ↔ System 55 (fuel mod 1.1)
```

---

## bsctl jump remove \<connection_id\>

```
$ bsctl jump remove 1025 --confirm

Jump connection 1025 removed. WARNING: re-run 'bsctl world validate' to check connectivity.
```

---

# 10. Player Commands

## bsctl player list

```
$ bsctl player list [--online] [--offline] [--banned]

Name        ID (short)  Status    Last Login       Ship      Location
nova        550e8400    online    just now         courier   Vega Prime (14)
xander      a3f12900    online    just now         freighter Outer Rim (31)
ghost       bd44e200    offline   2h ago           fighter   Dusk Point (17)
banned_guy  cc99f100    banned    3 days ago       —         —

(4 players total — 2 online, 1 offline, 1 banned)
```

---

## bsctl player info \<name\>

```
$ bsctl player info nova

Player: nova
ID: 550e8400-e29b-41d4-a716-446655440000
Status: online
Last Login: just now (session: TEXT)
Created: 2026-02-14

Ship: courier
  Hull: 80/100   Shields: 40/50   Energy: 75/100
  Cargo: food_supplies x10, refined_ore x5
  Location: Vega Prime (system 14) — docked at port 22
  Upgrades: none

Active Missions: 1
  pirate_hunt — IN_PROGRESS (started tick 8100)

Credits: 24,500
```

---

## bsctl player sessions \<name\>

```
$ bsctl player sessions nova

Recent Sessions (nova):

Session ID (short)  Mode  Connected           Disconnected        Duration
abc12300            TEXT  2026-03-05 14:22    —                   active
def45600            TEXT  2026-03-04 10:11    2026-03-04 12:44    2h 33m
ghi78900            TEXT  2026-03-03 08:01    2026-03-03 08:45    44m
```

---

## bsctl player ban \<name\> [\<reason\>]

```
$ bsctl player ban ghost "Exploiting economy bug" --confirm

Player ghost banned. Reason: Exploiting economy bug
Active session terminated (if any).
```

---

## bsctl player unban \<name\>

```
$ bsctl player unban ghost --confirm

Player ghost unbanned.
```

---

## bsctl player token-reset \<name\>

```
$ bsctl player token-reset nova --confirm

New token generated for nova:

  bXlzZWNyZXR0b2tlbmhlcmVmb3JleGFtcGxl

Provide this to the player securely. Old token invalidated immediately.
```

---

## bsctl player teleport \<name\> \<system_id\>

Moves a player's ship to a system. Works offline (modifies DB directly).

```
$ bsctl player teleport ghost 1 --confirm

Player ghost teleported to System 1 (Sol).
Note: If server is running, the player will see the change on next tick.
```

---

# 11. Economy Commands

## bsctl economy status

```
$ bsctl economy status

Active Economic Events: 2
  food_shortage   region 3  ends tick 9200  [PUBLIC]
  industrial_boom system 14 ends tick 8900  [PUBLIC]

AI Traders Active: 62 / 80

Recent Market Activity (last 100 ticks):
  Top traded commodity: refined_ore  (2,840 units)
  Highest price spike:  alien_artifacts (+240% in Black Sector)
  Most active port:     Vega Station (port 22)
```

---

## bsctl economy prices \<port_id\>

Alias: `bsctl port info <id>` — shows live prices with event modifiers applied.

---

## bsctl economy reset \<port_id\>

Resets a port's commodity prices to base values. Useful if an exploit inflated prices.

```
$ bsctl economy reset 22 --confirm

Port 22 (Vega Station) prices reset to base values.
```

---

# 12. Health Commands

## bsctl health

Runs a full health check across all subsystems.

```
$ bsctl health

Database:
  File:          OK  (data/blacksector.db, 48 MB)
  Integrity:     OK  (SQLite PRAGMA integrity_check passed)
  Schema:        OK  (all tables present, no missing indexes)

World Data:
  Regions:       OK  (8)
  Systems:       OK  (512)
  Ports:         OK  (847)
  Jump Graph:    OK  (fully connected)
  Commodities:   OK  (14 loaded)

Configuration:
  server.json:         OK
  commodities.json:    OK
  ship_classes.json:   OK
  economic_events.json:OK
  missions (core):     OK  (14 missions, 0 errors)
  missions (community):WARN  (1 file with validation warnings — run 'bsctl mission validate')

Snapshot:
  Latest:  snapshot_008205_1709612345.json  OK
  Age:     3 minutes

Overall: WARN (1 non-critical issue)
```

---

## bsctl health db

SQLite-specific database health check.

```
$ bsctl health db

Running SQLite integrity check...  PASSED
Checking for orphaned records...   OK
Checking index coverage...         OK
Database size: 48 MB
WAL file: none (clean)
```

---

## bsctl health config

Validates all JSON configuration files.

```
$ bsctl health config

config/server.json:            OK
config/economy/commodities.json: OK  (14 commodities)
config/economy/economic_events.json: OK  (8 events)
config/ships/ship_classes.json:  OK  (3 classes)
config/ships/upgrades.json:      OK  (6 upgrades)
config/ai/trader_names.json:     OK  (200 names)
config/missions/core/:           OK  (14 missions)
config/missions/community/:      WARN  new_mission.json line 42: unknown objective type "collect_data"
```

---

## bsctl health snapshot

Validates the latest snapshot file.

```
$ bsctl health snapshot [<file>]

Snapshot: snapshots/snapshot_latest.json
  Version:  1.0  (compatible)
  Tick:     8,205
  Players:  4
  Sessions: 2
  Ships:    4

Validation: PASSED
```

---

# 13. Mission Commands

## bsctl mission list

```
$ bsctl mission list

ID                  Name                   File                            Enabled
pirate_hunt         Pirate Hunt            core/combat_missions.json       YES
escort_convoy       Escort the Convoy      core/trade_missions.json        YES
ore_delivery        Ore Delivery Run       core/trade_missions.json        YES
black_market_run    Black Market Run       core/black_market.json          YES
new_mission         Data Collection        community/new_mission.json      NO (validation error)
```

---

## bsctl mission validate \<file\>

```
$ bsctl mission validate config/missions/community/new_mission.json

Parsing... 3 missions found.

  patrol_duty         OK
  rescue_survivor     OK
  data_collection     WARN: unknown objective type "collect_data"
                            valid types: kill, deliver_commodity, acquire_commodity,
                            navigate_to, scan_object, dock_at, survive

Validation: 2 passed, 1 warning, 0 errors.
```

---

# 14. Online Status Commands

## bsctl online

Quick view of who is currently connected.

```
$ bsctl online

2 players online (of 4 registered)

Name     Mode  System          Since     Commands/Tick
nova     TEXT  Vega Prime      14m ago   1.2 avg
xander   TEXT  Outer Rim       2h ago    0.4 avg
```

---

## bsctl online history [\<hours\>]

Shows connection history over the past N hours (default 24).

```
$ bsctl online history 24

Last 24 Hours — Connection Summary

Player  Sessions  Total Play Time  Peak Time
nova    3         4h 12m           14:00–16:30
xander  1         2h 04m           10:15–12:19
ghost   2         55m              08:00–08:45
```

---

# 15. Server Status (Read-Only, Works Offline)

## bsctl server status

When run against a live database, shows current server state.

```
$ bsctl server status

Server State (from database):
  Last Tick:    8,205
  Last Snapshot: 3 minutes ago
  Active Sessions: 2
  Active Economic Events: 2
  AI Traders: 62 active

Note: Server appears to be RUNNING (snapshot written recently).
```

If the server has been stopped:

```
Note: Server appears to be STOPPED (last snapshot 14 hours ago).
```

---

# 16. Global Flags

| Flag          | Description                                    |
| ------------- | ---------------------------------------------- |
| `--db <path>` | Path to SQLite database file                   |
| `--config <dir>` | Path to config directory                    |
| `--json`      | Output results as JSON instead of tables       |
| `--confirm`   | Required for destructive or modifying commands |
| `--dry-run`   | Show what would change without writing         |
| `--quiet`     | Suppress non-essential output                  |
| `--verbose`   | Show extra detail                              |

---

# 17. Non-Goals

* Live server control (use server stdin admin — see `cli_management_tool.md`)
* GUI-based world editor
* Procedural generation beyond `world create`
* Automated balancing adjustments

---

# End of Document
