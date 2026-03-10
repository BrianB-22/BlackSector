# Tasks: Milestone 2 - Vertical Slice (Phase 1)

## Task Breakdown

### 1. World System
- [x] 1.1 Create world package structure (internal/world/)
- [x] 1.2 Implement world configuration loader (alpha_sector.json)
- [x] 1.3 Implement topology validator (all systems reachable)
- [x] 1.4 Create world data structures (Universe, System, Port, JumpConnection)
- [x] 1.5 Implement world cache with RWMutex for concurrent reads
- [x] 1.6 Write unit tests for world loading and validation
- [x] 1.7 Create alpha_sector.json with 15-20 systems, Federated Space origin, High/Low Security zones

### 2. Navigation System
- [x] 2.1 Create navigation package (internal/navigation/)
- [x] 2.2 Implement jump connection lookup
- [x] 2.3 Implement ProcessJump() with validation and state updates
- [x] 2.4 Implement ProcessDock() with port validation
- [x] 2.5 Implement ProcessUndock() with status checks
- [x] 2.6 Add jump connection display (system map)
- [x] 2.7 Write unit tests for jump validation and state transitions
- [x] 2.8 Write integration tests for complete navigation flows

### 3. Economy System
- [x] 3.1 Create economy package (internal/economy/)
- [x] 3.2 Implement commodity definitions (7 commodities)
- [x] 3.3 Implement CalculatePrice() with zone multipliers and spread
- [x] 3.4 Implement BuyCommodity() with validation and transaction
- [x] 3.5 Implement SellCommodity() with validation and transaction
- [x] 3.6 Implement GetMarketPrices() for port inventory display
- [x] 3.7 Implement cargo management (add/remove/display)
- [x] 3.8 Write unit tests for price calculation (table-driven)
- [x] 3.9 Write unit tests for trade validation
- [x] 3.10 Write integration tests for complete trade flows


### 4. Combat System
- [x] 4.1 Create combat package (internal/combat/)
- [x] 4.2 Implement pirate tier definitions (raider, marauder)
- [x] 4.3 Implement pirate spawn probability calculation
- [x] 4.4 Implement CheckPirateSpawns() with security zone filtering
- [x] 4.5 Implement SpawnPirate() creating ephemeral pirate entities
- [x] 4.6 Implement resolveDamage() with shield/hull mechanics
- [x] 4.7 Implement ProcessAttack() for player attacks
- [x] 4.8 Implement pirate counter-attack logic
- [x] 4.9 Implement pirate flee behavior at threshold
- [x] 4.10 Implement ProcessFlee() for player escape attempts
- [x] 4.11 Implement ProcessSurrender() with credit penalty
- [x] 4.12 Implement handleShipDestruction() with respawn and insurance
- [x] 4.13 Implement ResolveCombatTick() processing all active combats
- [x] 4.14 Write unit tests for damage calculation (table-driven)
- [x] 4.15 Write unit tests for pirate spawn probability
- [x] 4.16 Write unit tests for flee threshold logic
- [x] 4.17 Write integration tests for complete combat encounters

### 5. Mission System
- [x] 5.1 Create missions package (internal/missions/)
- [x] 5.2 Implement mission JSON schema validation
- [x] 5.3 Implement LoadMissions() with error handling for invalid files
- [x] 5.4 Implement mission registry with in-memory cache
- [x] 5.5 Implement GetAvailableMissions() with security zone filtering
- [x] 5.6 Implement AcceptMission() with one-active-mission validation
- [x] 5.7 Implement objective progress tracking
- [x] 5.8 Implement EvaluateObjectives() for deliver_commodity type
- [x] 5.9 Implement mission completion and reward distribution
- [x] 5.10 Implement mission expiry checking
- [x] 5.11 Implement AbandonMission()
- [x] 5.12 Write unit tests for mission validation
- [x] 5.13 Write unit tests for objective evaluation
- [x] 5.14 Write integration tests for complete mission lifecycle
- [x] 5.15 Create phase1_delivery.json with starter missions

### 6. Registration System
- [x] 6.1 Extend existing registration to check SSH username
- [x] 6.2 Implement registration prompt flow (TEXT mode)
- [x] 6.3 Implement player token generation (32 bytes, base64url)
- [x] 6.4 Implement bcrypt hashing for passwords (cost 12) and tokens (cost 10)
- [x] 6.5 Implement starting ship provisioning at Federated Space origin
- [x] 6.6 Implement starting credits grant (10,000 configurable)
- [x] 6.7 Display player token once with save instructions
- [x] 6.8 Implement registration rate limiting (3 per IP per hour)
- [x] 6.9 Write unit tests for token generation and hashing
- [ ] 6.10 Write integration tests for complete registration flow


### 7. TUI System
- [x] 7.1 Create tui package (internal/tui/)
- [x] 7.2 Implement GameView bubbletea model
- [x] 7.3 Implement status bar rendering (system, credits, hull/shield/energy)
- [x] 7.4 Implement command prompt and input handling
- [x] 7.5 Create styles.go with lipgloss style definitions
- [x] 7.6 Implement system map view (ASCII list with jump indicators)
- [x] 7.7 Implement market view (commodity listings with prices)
- [x] 7.8 Implement combat view (attack/flee/surrender options)
- [x] 7.9 Implement mission board view
- [x] 7.10 Implement cargo manifest view
- [x] 7.11 Implement help command display
- [x] 7.12 Implement command parser and validator
- [x] 7.13 Implement state update handling from tick engine broadcasts
- [x] 7.14 Implement error message display
- [ ] 7.15 Write manual test plan for TUI interactions

### 8. Tick Engine Integration
- [x] 8.1 Extend tick engine to integrate navigation subsystem
- [x] 8.2 Extend tick engine to integrate economy subsystem
- [x] 8.3 Extend tick engine to integrate combat subsystem
- [x] 8.4 Extend tick engine to integrate mission subsystem
- [x] 8.5 Implement command routing by type (jump, buy, sell, attack, etc.)
- [x] 8.6 Implement command validation before processing
- [x] 8.7 Implement state broadcast to all active sessions
- [x] 8.8 Add tick duration monitoring and warnings (>100ms threshold)
- [x] 8.9 Implement graceful error handling for subsystem failures
- [ ] 8.10 Write integration tests for complete tick processing
- [ ] 8.11 Write performance tests for 5 concurrent players

### 9. Database Extensions
- [x] 9.1 Add combat_instances table for active combats
- [x] 9.2 Add mission_instances table for player missions
- [x] 9.3 Add objective_progress table for mission tracking
- [x] 9.4 Add indexes for performance (ships_system, missions_player, etc.)
- [x] 9.5 Implement combat persistence queries
- [x] 9.6 Implement mission persistence queries
- [x] 9.7 Update snapshot system to include combat and mission state
- [x] 9.8 Write migration script for new tables
- [ ] 9.9 Write unit tests for new queries with go-sqlmock
- [ ] 9.10 Write integration tests with in-memory SQLite


### 10. Configuration
- [x] 10.1 Extend server.json with economy settings (low_sec_multiplier, spread)
- [x] 10.2 Extend server.json with combat settings (pirate_activity, spawn_interval)
- [x] 10.3 Extend server.json with game settings (starting_credits, insurance_payout)
- [x] 10.4 Extend server.json with security settings (rate limits)
- [x] 10.5 Implement config validation at startup
- [x] 10.6 Write unit tests for config loading

### 11. Testing and Quality Assurance
- [ ] 11.1 Achieve 80%+ test coverage for internal/combat
- [x] 11.2 Achieve 80%+ test coverage for internal/economy
- [x] 11.3 Achieve 80%+ test coverage for internal/engine
- [ ] 11.4 Write acceptance test: registration flow
- [x] 11.5 Write acceptance test: trade flow
- [ ] 11.6 Write acceptance test: combat flow
- [ ] 11.7 Write acceptance test: mission flow
- [ ] 11.8 Write acceptance test: persistence (restart)
- [ ] 11.9 Write load test: 5 concurrent players for 60 seconds
- [ ] 11.10 Verify tick duration <100ms under load
- [ ] 11.11 Verify no memory leaks during 24-hour run
- [ ] 11.12 Manual testing: complete playthrough from registration to mission completion

### 12. Documentation
- [x] 12.1 Update README with Phase 1 features and how to play
- [x] 12.2 Create QUICKSTART.md with setup instructions
- [x] 12.3 Document all available commands in docs/
- [ ] 12.4 Document configuration options in docs/
- [ ] 12.5 Create example mission files with comments
- [ ] 12.6 Document known limitations and Phase 2 roadmap

### 13. Deployment Preparation
- [ ] 13.1 Create build script for single binary
- [ ] 13.2 Create systemd service file for Linux deployment
- [ ] 13.3 Create backup script for database and snapshots
- [ ] 13.4 Document server requirements (RAM, disk, ports)
- [ ] 13.5 Create admin guide for server operation
- [ ] 13.6 Test deployment on clean Linux system

## Estimated Effort

| Category | Tasks | Estimated Days |
|----------|-------|----------------|
| World System | 7 | 3 |
| Navigation System | 8 | 4 |
| Economy System | 10 | 5 |
| Combat System | 17 | 8 |
| Mission System | 15 | 7 |
| Registration System | 10 | 4 |
| TUI System | 15 | 7 |
| Tick Engine Integration | 11 | 5 |
| Database Extensions | 10 | 4 |
| Configuration | 6 | 2 |
| Testing & QA | 12 | 6 |
| Documentation | 6 | 2 |
| Deployment | 6 | 2 |
| **Total** | **133** | **59 days** |

## Critical Path

The following tasks are on the critical path and must be completed in order:

1. World System (1.1-1.7) - Foundation for all other systems
2. Database Extensions (9.1-9.4) - Required for persistence
3. Navigation System (2.1-2.5) - Core movement mechanics
4. Economy System (3.1-3.7) - Core trading mechanics
5. Combat System (4.1-4.13) - Core combat mechanics
6. Mission System (5.1-5.11) - Core mission mechanics
7. Tick Engine Integration (8.1-8.9) - Brings all systems together
8. TUI System (7.1-7.14) - Player interface
9. Testing & QA (11.1-11.12) - Verification of success criteria

## Dependencies

```
World System → Navigation System
World System → Economy System
World System → Combat System
Database Extensions → All subsystems
Navigation System → Tick Engine Integration
Economy System → Tick Engine Integration
Combat System → Tick Engine Integration
Mission System → Tick Engine Integration
Tick Engine Integration → TUI System
All subsystems → Testing & QA
```

## Milestones

**M2.1 - Core Systems (Week 1-2)**: World, Navigation, Economy basic functionality

**M2.2 - Combat & Missions (Week 3-4)**: Combat system and mission framework

**M2.3 - Integration (Week 5-6)**: Tick engine integration and TUI

**M2.4 - Polish & Testing (Week 7-8)**: Testing, documentation, deployment prep

**M2.5 - Release (Week 9)**: Final testing and Phase 1 release

---

**End of Tasks Document**
