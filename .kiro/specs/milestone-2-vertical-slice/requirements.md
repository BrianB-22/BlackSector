# Requirements Document: Milestone 2 - Vertical Slice (Phase 1)

## 1. Overview

This document defines the functional and non-functional requirements for Phase 1 of BlackSector - the minimum playable game delivering a complete vertical slice through all core systems.

## 2. Functional Requirements

### 2.1 Player Registration and Authentication

**REQ-REG-001**: The system SHALL prompt new SSH users for registration when no account exists for their SSH username.

**REQ-REG-002**: The system SHALL collect display name (3-20 characters) and password (minimum 8 characters) during registration.

**REQ-REG-003**: The system SHALL generate a unique player_id (UUID) for each new player.

**REQ-REG-004**: The system SHALL generate a cryptographically random player token (32 bytes, base64url encoded) at registration.

**REQ-REG-005**: The system SHALL display the player token exactly once and require the player to save it.

**REQ-REG-006**: The system SHALL hash passwords with bcrypt (cost factor 12) before storage.

**REQ-REG-007**: The system SHALL hash tokens with bcrypt (cost factor 10) before storage.

**REQ-REG-008**: The system SHALL provision a courier-class starter ship at Federated Space origin for new players.

**REQ-REG-009**: The system SHALL grant 10,000 starting credits to new players (configurable).

**REQ-REG-010**: The system SHALL authenticate returning players via SSH username lookup without additional password prompt.

**REQ-REG-011**: The system SHALL rate limit registration to 3 new accounts per IP per hour.


### 2.2 World and Universe

**REQ-WORLD-001**: The system SHALL load a static universe configuration from `config/world/alpha_sector.json`.

**REQ-WORLD-002**: The universe SHALL contain exactly 1 region of core/industrial type.

**REQ-WORLD-003**: The universe SHALL contain 15-20 star systems connected by jump points.

**REQ-WORLD-004**: Each system SHALL have 1-3 ports.

**REQ-WORLD-005**: The universe SHALL include Federated Space systems with SecurityLevel = 2.0 as the safe starting zone.

**REQ-WORLD-006**: The universe SHALL include High Security systems (SecurityLevel 0.7-1.0).

**REQ-WORLD-007**: The universe SHALL include Low Security systems (SecurityLevel 0.0-0.4).

**REQ-WORLD-008**: All systems SHALL be reachable via jump connections (no isolated systems).

**REQ-WORLD-009**: The Federated Space origin starbase SHALL stock all 7 commodities at base prices.

**REQ-WORLD-010**: Port types SHALL be one of: trading, mining, or refueling.

### 2.3 Navigation System

**REQ-NAV-001**: The system SHALL allow players to jump between systems connected by jump points.

**REQ-NAV-002**: The system SHALL prevent jumps when ship status is DOCKED or IN_COMBAT.

**REQ-NAV-003**: The system SHALL update ship CurrentSystemID upon successful jump.

**REQ-NAV-004**: The system SHALL display available jump connections via the "jumps" command.

**REQ-NAV-005**: The system SHALL allow players to dock at ports in their current system.

**REQ-NAV-006**: The system SHALL set ship status to DOCKED and record DockedAtPortID upon docking.

**REQ-NAV-007**: The system SHALL allow players to undock from ports.

**REQ-NAV-008**: The system SHALL set ship status to IN_SPACE and clear DockedAtPortID upon undocking.

**REQ-NAV-009**: The system SHALL display system information via the "system" command showing current location and ports.

**REQ-NAV-010**: The system SHALL trigger pirate spawn checks after jumps to Low Security systems.


### 2.4 Economy and Trading

**REQ-ECON-001**: The system SHALL support 7 commodities: food_supplies, fuel_cells, raw_ore, refined_ore, machinery, electronics, luxury_goods.

**REQ-ECON-002**: Each commodity SHALL have a static base price (no dynamic pricing in Phase 1).

**REQ-ECON-003**: The system SHALL apply zone price multipliers: Federated Space = 1.0x, High Security = 1.0x, Low Security = 1.18x.

**REQ-ECON-004**: The system SHALL apply buy/sell spread: buy price = zone_price × 1.10, sell price = zone_price × 0.90.

**REQ-ECON-005**: The system SHALL display market prices via the "market" command when docked.

**REQ-ECON-006**: The system SHALL allow commodity purchases via "buy <commodity> <quantity>" when docked.

**REQ-ECON-007**: The system SHALL validate sufficient player credits before purchase.

**REQ-ECON-008**: The system SHALL validate sufficient cargo capacity before purchase.

**REQ-ECON-009**: The system SHALL validate sufficient port inventory before purchase.

**REQ-ECON-010**: The system SHALL deduct credits and add cargo atomically in a single transaction.

**REQ-ECON-011**: The system SHALL allow commodity sales via "sell <commodity> <quantity>" when docked.

**REQ-ECON-012**: The system SHALL validate player has commodity in cargo before sale.

**REQ-ECON-013**: The system SHALL add credits and remove cargo atomically in a single transaction.

**REQ-ECON-014**: The system SHALL display cargo manifest via the "cargo" command.

**REQ-ECON-015**: Port commodity availability SHALL match port type (trading/mining/refueling).

**REQ-ECON-016**: Trading ports SHALL sell: food_supplies, luxury_goods, electronics; buy: raw_ore, refined_ore, machinery.

**REQ-ECON-017**: Mining ports SHALL sell: raw_ore, fuel_cells; buy: food_supplies, machinery.

**REQ-ECON-018**: Refueling ports SHALL sell: fuel_cells, food_supplies; buy: none.


### 2.5 Combat System

**REQ-COMBAT-001**: The system SHALL spawn NPC pirates only in Low Security systems (SecurityLevel < 0.4).

**REQ-COMBAT-002**: The system SHALL check for pirate spawns every 5 ticks (configurable).

**REQ-COMBAT-003**: Pirate spawn probability SHALL be calculated as: ActivityBase × (1 - SecurityLevel).

**REQ-COMBAT-004**: The system SHALL select pirate tier randomly: 70% raider, 30% marauder.

**REQ-COMBAT-005**: Raider pirates SHALL have: Hull=60, Shield=20, Damage=12-18, Accuracy=60%, FleeThreshold=15%.

**REQ-COMBAT-006**: Marauder pirates SHALL have: Hull=90, Shield=40, Damage=18-28, Accuracy=65%, FleeThreshold=10%.

**REQ-COMBAT-007**: The system SHALL initiate combat immediately when a pirate spawns.

**REQ-COMBAT-008**: The system SHALL set ship status to IN_COMBAT when combat begins.

**REQ-COMBAT-009**: The system SHALL allow player attack via "attack" command during combat.

**REQ-COMBAT-010**: The system SHALL calculate hit based on random roll vs pirate accuracy.

**REQ-COMBAT-011**: The system SHALL apply damage to shields first, then hull.

**REQ-COMBAT-012**: The system SHALL process pirate counter-attack after player attack (if pirate survives).

**REQ-COMBAT-013**: The system SHALL end combat when pirate hull reaches 0 (destroyed).

**REQ-COMBAT-014**: The system SHALL end combat when pirate hull drops below flee threshold (pirate flees).

**REQ-COMBAT-015**: The system SHALL allow player flee attempts via "flee" command.

**REQ-COMBAT-016**: The system SHALL allow player surrender via "surrender" command.

**REQ-COMBAT-017**: Surrender SHALL deduct 40% of wallet credits (configurable) and end combat.

**REQ-COMBAT-018**: The system SHALL handle player ship destruction when hull reaches 0.

**REQ-COMBAT-019**: Ship destruction SHALL clear cargo, respawn ship at nearest port, restore hull/shields, and grant insurance payout (5000 credits default).

**REQ-COMBAT-020**: Pirates SHALL be ephemeral entities (not persisted after combat ends).

**REQ-COMBAT-021**: No loot drops in Phase 1 (pirate destruction grants no items).


### 2.6 Mission System

**REQ-MISSION-001**: The system SHALL load mission definitions from `config/missions/*.json` at startup.

**REQ-MISSION-002**: The system SHALL validate mission JSON against schema and reject invalid files.

**REQ-MISSION-003**: The system SHALL support deliver_commodity mission type in Phase 1.

**REQ-MISSION-004**: The system SHALL display available missions via "missions" command when docked.

**REQ-MISSION-005**: The system SHALL filter missions by security zone and port.

**REQ-MISSION-006**: The system SHALL allow mission acceptance via "missions accept <id>" command.

**REQ-MISSION-007**: The system SHALL enforce one active mission per player (Phase 1 limit).

**REQ-MISSION-008**: The system SHALL reject mission acceptance if player already has an active mission.

**REQ-MISSION-009**: The system SHALL track objective progress for active missions.

**REQ-MISSION-010**: The system SHALL evaluate mission objectives each tick.

**REQ-MISSION-011**: For deliver_commodity objectives, the system SHALL check if player is docked at destination port with required commodity quantity.

**REQ-MISSION-012**: The system SHALL mark mission COMPLETED when final objective is satisfied.

**REQ-MISSION-013**: The system SHALL distribute credit rewards immediately upon mission completion.

**REQ-MISSION-014**: The system SHALL mark mission EXPIRED if ExpiryTicks is reached before completion.

**REQ-MISSION-015**: The system SHALL allow mission abandonment via "missions abandon" command.

**REQ-MISSION-016**: The system SHALL display active mission status via "missions status" command.

**REQ-MISSION-017**: Repeatable missions SHALL become available again after cooldown period.

**REQ-MISSION-018**: The player SHALL purchase required commodities (missions do not provide them).


### 2.7 TUI Interface

**REQ-TUI-001**: The system SHALL render a TEXT mode terminal interface using bubbletea.

**REQ-TUI-002**: The system SHALL display a status bar showing: system name, credits, hull/max, shield/max, energy/max.

**REQ-TUI-003**: The system SHALL display a command prompt for user input.

**REQ-TUI-004**: The system SHALL support ANSI color rendering via lipgloss.

**REQ-TUI-005**: The system SHALL display system map as ASCII list with jump connection indicators.

**REQ-TUI-006**: The system SHALL display market listings with commodity names, buy/sell prices, and available quantity.

**REQ-TUI-007**: The system SHALL display combat interface with attack/flee/surrender options.

**REQ-TUI-008**: The system SHALL display mission board with available missions.

**REQ-TUI-009**: The system SHALL display active mission status with objective progress.

**REQ-TUI-010**: The system SHALL provide help via "help" command listing available commands.

**REQ-TUI-011**: The system SHALL update display in response to state broadcasts from tick engine.

**REQ-TUI-012**: The system SHALL parse and validate commands before sending to tick engine.

**REQ-TUI-013**: The system SHALL display error messages for invalid commands or failed operations.

### 2.8 Ship System

**REQ-SHIP-001**: All players SHALL start with a courier-class ship.

**REQ-SHIP-002**: Courier ships SHALL have: Hull=100, Shield=50, Energy=100, Cargo=20, WeaponDamage=15.

**REQ-SHIP-003**: The system SHALL track ship status: DOCKED, IN_SPACE, IN_COMBAT, DESTROYED.

**REQ-SHIP-004**: The system SHALL prevent actions incompatible with current status (e.g., no jump while DOCKED).

**REQ-SHIP-005**: The system SHALL track cargo in slots with commodity_id and quantity.

**REQ-SHIP-006**: Total cargo quantity SHALL NOT exceed CargoCapacity.

**REQ-SHIP-007**: The system SHALL track ship position (CurrentSystemID, PositionX, PositionY).

**REQ-SHIP-008**: The system SHALL update LastUpdatedTick on every ship state change.


## 3. Non-Functional Requirements

### 3.1 Performance

**REQ-PERF-001**: The tick engine SHALL complete all processing within 100ms per tick.

**REQ-PERF-002**: The system SHALL support 5 concurrent players without exceeding 100ms tick duration.

**REQ-PERF-003**: The system SHALL process up to 50 commands per tick within the performance budget.

**REQ-PERF-004**: Database queries SHALL use prepared statements and indexes for optimal performance.

**REQ-PERF-005**: The tick interval SHALL be 2 seconds (configurable).

**REQ-PERF-006**: State broadcasts to sessions SHALL complete within 20ms.

**REQ-PERF-007**: The system SHALL log warnings when tick duration exceeds 100ms.

### 3.2 Reliability

**REQ-REL-001**: The system SHALL persist game state to SQLite database every tick.

**REQ-REL-002**: The system SHALL save snapshots every 100 ticks.

**REQ-REL-003**: The system SHALL restore from snapshot on server restart with no data loss.

**REQ-REL-004**: All database writes SHALL occur in atomic transactions.

**REQ-REL-005**: The system SHALL rollback transactions on error and restore from last snapshot.

**REQ-REL-006**: The system SHALL handle database connection loss gracefully without crashing.

**REQ-REL-007**: The system SHALL recover from panics in tick processing without server crash.

**REQ-REL-008**: The system SHALL validate all command parameters before processing.

### 3.3 Security

**REQ-SEC-001**: The system SHALL never store plaintext passwords or tokens.

**REQ-SEC-002**: The system SHALL use bcrypt for password hashing (cost factor 12).

**REQ-SEC-003**: The system SHALL use bcrypt for token hashing (cost factor 10).

**REQ-SEC-004**: The system SHALL use prepared statements for all SQL queries (no string concatenation).

**REQ-SEC-005**: The system SHALL validate player ownership before processing ship commands.

**REQ-SEC-006**: The system SHALL rate limit registration attempts (3 per IP per hour).

**REQ-SEC-007**: The system SHALL rate limit failed authentication attempts (10 per IP per minute).

**REQ-SEC-008**: The system SHALL rate limit commands (10 per session per tick).

**REQ-SEC-009**: Database file permissions SHALL be 0600 (owner read/write only).

**REQ-SEC-010**: Snapshot file permissions SHALL be 0600 (owner read/write only).


### 3.4 Scalability

**REQ-SCALE-001**: The system SHALL support up to 10 concurrent SSH sessions.

**REQ-SCALE-002**: The system SHALL use buffered channels for session communication to prevent blocking.

**REQ-SCALE-003**: The system SHALL cache world data in memory (systems, ports, commodities).

**REQ-SCALE-004**: The system SHALL limit command queue size to prevent unbounded growth.

**REQ-SCALE-005**: The system SHALL use SQLite WAL mode for concurrent read access.

### 3.5 Maintainability

**REQ-MAINT-001**: All game logic packages SHALL have unit tests.

**REQ-MAINT-002**: Critical systems (combat, economy, engine) SHALL have 80%+ test coverage.

**REQ-MAINT-003**: The system SHALL use structured logging with zerolog.

**REQ-MAINT-004**: All errors SHALL be wrapped with context using fmt.Errorf.

**REQ-MAINT-005**: The system SHALL log all player actions at DEBUG level.

**REQ-MAINT-006**: The system SHALL log connection events, errors, and warnings at INFO level.

**REQ-MAINT-007**: Configuration SHALL be loaded from `config/server.json` (no hardcoded values).

**REQ-MAINT-008**: The system SHALL support hot-reload of mission definitions.

### 3.6 Usability

**REQ-USE-001**: New players SHALL be able to register and start playing within 2 minutes.

**REQ-USE-002**: The system SHALL provide clear error messages for invalid commands.

**REQ-USE-003**: The system SHALL provide a help command listing all available commands.

**REQ-USE-004**: The system SHALL display current game state in the status bar.

**REQ-USE-005**: The system SHALL notify players of important events (pirate encounters, mission completion).

**REQ-USE-006**: Command syntax SHALL be intuitive and consistent across all commands.


## 4. Success Criteria

### 4.1 Functional Success Criteria

**SC-FUNC-001**: A new player can SSH in, register, and have a functional character within 2 minutes.

**SC-FUNC-002**: A player can complete a profitable trade run between two ports (buy low, sell high).

**SC-FUNC-003**: A player can accept a delivery mission, complete it, and receive the reward.

**SC-FUNC-004**: A player can survive a pirate encounter by defeating or fleeing from the pirate.

**SC-FUNC-005**: A player can die to a pirate and continue playing after respawn.

**SC-FUNC-006**: All 7 commodities are tradeable at appropriate ports.

**SC-FUNC-007**: Jump connections work bidirectionally between all systems.

**SC-FUNC-008**: Mission objectives are evaluated correctly each tick.

**SC-FUNC-009**: Combat damage is applied correctly (shields first, then hull).

**SC-FUNC-010**: Pirates flee when hull drops below threshold.

### 4.2 Performance Success Criteria

**SC-PERF-001**: 5 concurrent players can play simultaneously without server errors.

**SC-PERF-002**: Tick duration remains under 100ms with 5 concurrent players.

**SC-PERF-003**: Average tick duration is under 50ms with 5 concurrent players.

**SC-PERF-004**: Database queries complete within 30ms per tick.

**SC-PERF-005**: State broadcasts complete within 20ms per tick.

### 4.3 Reliability Success Criteria

**SC-REL-001**: Server survives restart and restores state from snapshot with no data loss.

**SC-REL-002**: All player credits, cargo, and positions are preserved across restart.

**SC-REL-003**: Active missions are preserved across restart.

**SC-REL-004**: No data corruption occurs during normal operation.

**SC-REL-005**: Server runs for 24 hours without crashes or memory leaks.

### 4.4 Integration Success Criteria

**SC-INT-001**: All subsystems (navigation, economy, combat, missions) integrate correctly.

**SC-INT-002**: Commands flow correctly from sessions through tick engine to subsystems.

**SC-INT-003**: State updates flow correctly from tick engine back to sessions.

**SC-INT-004**: Database transactions commit atomically with no partial updates.

**SC-INT-005**: Pirate spawns trigger correctly after jumps to Low Security systems.


## 5. Out of Scope (Phase 1)

The following features are explicitly excluded from Phase 1 and deferred to Phase 2 or later:

**OUT-001**: Dynamic commodity pricing (static prices only in Phase 1)

**OUT-002**: Economic events affecting prices

**OUT-003**: AI trader NPCs

**OUT-004**: Mining and resource extraction

**OUT-005**: Exploration and scanning

**OUT-006**: Complex missions (kill, scan, multi-step)

**OUT-007**: Multiple regions (only 1 region in Phase 1)

**OUT-008**: Black Sector systems

**OUT-009**: Medium Security space

**OUT-010**: Ship upgrades and customization

**OUT-011**: Combat loot drops

**OUT-012**: GUI client (port 2223)

**OUT-013**: Faction systems

**OUT-014**: Fleet combat

**OUT-015**: Player-to-player cargo trading

**OUT-016**: Banking system

**OUT-017**: Communications (IRN, messaging)

**OUT-018**: Drones

**OUT-019**: Multiple ships per player

**OUT-020**: Ship classes beyond courier

## 6. Acceptance Tests

### 6.1 Registration Flow Test

```
GIVEN: A new player connects via SSH with username "testplayer"
WHEN: The player completes registration with display name "TestPlayer" and password "secure123"
THEN: 
  - Player receives a unique player token
  - Player has 10,000 starting credits
  - Player has a courier-class ship docked at Federated Space origin
  - Player can immediately start playing
```

### 6.2 Trade Flow Test

```
GIVEN: A player is docked at Federated Space origin with 10,000 credits
WHEN: The player executes the following sequence:
  1. buy food_supplies 15
  2. undock
  3. jump <low_sec_system>
  4. dock <mining_port>
  5. sell food_supplies 15
THEN:
  - Player credits change correctly based on buy/sell prices
  - Cargo is added after buy and removed after sell
  - Ship position updates to Low Security system
  - All state changes persist to database
```


### 6.3 Combat Flow Test

```
GIVEN: A player jumps to a Low Security system
WHEN: A pirate spawns and initiates combat
AND: The player executes "attack" commands until combat ends
THEN:
  - Damage is applied correctly to pirate shields and hull
  - Pirate counter-attacks and damages player ship
  - Combat ends when pirate is destroyed or flees
  - Player ship status returns to IN_SPACE after combat
  - If player dies: ship respawns at nearest port with insurance payout
```

### 6.4 Mission Flow Test

```
GIVEN: A player is docked at a trading port
WHEN: The player executes the following sequence:
  1. missions (view available missions)
  2. missions accept emergency_ore_run
  3. buy refined_ore 15
  4. undock
  5. jump <destination_system>
  6. dock <destination_port>
THEN:
  - Mission status changes from AVAILABLE to IN_PROGRESS
  - Objective progress updates when commodity is purchased
  - Mission completes when player docks at destination with commodity
  - Reward credits are added to player account
  - Mission status changes to COMPLETED
```

### 6.5 Persistence Test

```
GIVEN: A server is running with 3 active players
AND: Players have executed various actions (trades, jumps, missions)
WHEN: The server is stopped and restarted
THEN:
  - All player accounts are restored
  - All ship positions and statuses are restored
  - All cargo contents are restored
  - All active missions are restored
  - All credits are preserved
  - Players can reconnect and continue playing
```

### 6.6 Concurrent Players Test

```
GIVEN: 5 players are connected and playing simultaneously
WHEN: All players execute commands concurrently for 60 seconds
THEN:
  - All commands are processed correctly
  - No commands are lost or duplicated
  - Tick duration remains under 100ms
  - No database errors occur
  - No race conditions or deadlocks occur
  - All state updates are broadcast correctly to all sessions
```

### 6.7 Error Handling Test

```
GIVEN: A player attempts invalid operations
WHEN: The player executes:
  1. buy food_supplies 1000 (insufficient credits)
  2. jump 999 (invalid system)
  3. sell machinery 10 (commodity not in cargo)
  4. missions accept mission_1 (already has active mission)
THEN:
  - Each command returns appropriate error message
  - No state changes occur for failed commands
  - Server continues running normally
  - Player can retry with valid commands
```

## 7. Traceability Matrix

| Requirement ID | Design Component | Test Coverage |
|----------------|------------------|---------------|
| REQ-REG-001 to REQ-REG-011 | Registration System | TestRegistrationFlow |
| REQ-WORLD-001 to REQ-WORLD-010 | World Generator | TestWorldLoading |
| REQ-NAV-001 to REQ-NAV-010 | Navigation System | TestJumpSystem, TestDocking |
| REQ-ECON-001 to REQ-ECON-018 | Economy System | TestPriceCalculation, TestTrading |
| REQ-COMBAT-001 to REQ-COMBAT-021 | Combat System | TestPirateSpawn, TestCombatResolution |
| REQ-MISSION-001 to REQ-MISSION-018 | Mission System | TestMissionLifecycle |
| REQ-TUI-001 to REQ-TUI-013 | TUI System | Manual testing |
| REQ-SHIP-001 to REQ-SHIP-008 | Ship Models | TestShipState |
| REQ-PERF-001 to REQ-PERF-007 | Tick Engine | TestTickPerformance |
| REQ-REL-001 to REQ-REL-008 | Snapshot System | TestPersistence |
| REQ-SEC-001 to REQ-SEC-010 | Security Layer | TestAuthentication, TestRateLimiting |

## 8. Configuration Requirements

### 8.1 server.json Configuration

The following configuration parameters SHALL be supported in `config/server.json`:

```json
{
  "server": {
    "ssh_port": 2222,
    "tick_interval_ms": 2000,
    "snapshot_interval_ticks": 100,
    "max_sessions": 10
  },
  "game": {
    "starting_credits": 10000,
    "insurance_payout": 5000
  },
  "economy": {
    "low_sec_price_multiplier": 1.18,
    "buy_markup": 1.10,
    "sell_markdown": 0.90
  },
  "combat": {
    "pirate_activity_base": 0.10,
    "spawn_check_interval_ticks": 5,
    "surrender_loss_percent": 40
  },
  "security": {
    "registration_rate_limit_per_hour": 3,
    "auth_failure_rate_limit_per_minute": 10,
    "command_rate_limit_per_tick": 10
  },
  "database": {
    "path": "blacksector.db",
    "wal_mode": true,
    "busy_timeout_ms": 5000
  },
  "logging": {
    "level": "info",
    "debug_commands": false
  }
}
```

### 8.2 World Configuration

The world configuration file `config/world/alpha_sector.json` SHALL define:

- Region definitions (id, name, type, security_level)
- System definitions (id, name, region_id, security_level, position)
- Port definitions (id, system_id, name, type, services)
- Jump connections (from_system_id, to_system_id, bidirectional)
- Port inventories (port_id, commodity_id, initial_quantity)

### 8.3 Mission Configuration

Mission files in `config/missions/` SHALL follow the schema defined in `docs/08_missions/content_schema.md`:

- mission_id (unique identifier)
- name, description, version, author
- enabled, repeatable, repeat_cooldown_ticks
- security_zones (filter by zone)
- expiry_ticks (optional timeout)
- objectives (array of objective definitions)
- rewards (credits, items)

## 9. Dependencies and Prerequisites

### 9.1 Milestone 1 Completion

Phase 1 requires that Milestone 1 (Foundation) is COMPLETE with the following verified:

- Server boots successfully
- SSH connectivity works
- Handshake protocol implemented
- Tick loop running at 2-second intervals
- Snapshot save/load functional
- Session management operational
- Database schema applied

### 9.2 Configuration Files

The following configuration files MUST exist before server start:

- `config/server.json` - Server configuration
- `config/world/alpha_sector.json` - Universe definition
- `config/missions/phase1_delivery.json` - Starter missions
- `migrations/001_initial_schema.sql` - Database schema

### 9.3 External Dependencies

All external Go modules MUST be available:

- github.com/gliderlabs/ssh
- github.com/charmbracelet/wish
- github.com/charmbracelet/bubbletea
- github.com/charmbracelet/lipgloss
- modernc.org/sqlite
- github.com/rs/zerolog
- github.com/google/uuid
- github.com/stretchr/testify (testing only)

## 10. Glossary

**Courier**: The starter ship class with balanced stats suitable for trading and light combat.

**Federated Space**: The safe starting zone with SecurityLevel = 2.0 where no pirates spawn and PvP is disabled.

**High Security**: Systems with SecurityLevel 0.7-1.0, relatively safe with rare pirate encounters.

**Low Security**: Systems with SecurityLevel 0.0-0.4, dangerous with frequent pirate spawns and higher commodity prices.

**Pirate Tier**: Classification of NPC pirates by difficulty (raider = easy, marauder = medium).

**Tick**: One iteration of the game loop (default 2 seconds), when all game state updates occur.

**Snapshot**: Complete serialization of game state saved periodically for crash recovery.

**Session**: An SSH connection representing one player's interaction with the server.

**Command Queue**: Buffered channel where sessions submit commands for tick engine processing.

**State Broadcast**: Distribution of updated game state from tick engine to all active sessions.

**Zone Multiplier**: Price adjustment factor based on system security level (Low Sec = 1.18x).

**Buy/Sell Spread**: Price difference between port buy and sell prices (±10% from zone price).

**Ephemeral Entity**: Game object that exists only in memory and is not persisted (e.g., pirates).

**WAL Mode**: Write-Ahead Logging mode for SQLite enabling concurrent reads during writes.

---

**End of Requirements Document**
