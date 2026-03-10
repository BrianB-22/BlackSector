# Phase 2 Deferred Features

This document tracks features that are intentionally excluded from Phase 1 (Milestone 2) and deferred to Phase 2 or later milestones.

## Combat System Enhancements

### Detection & Tracking System
- **Sensor strength** and target signature-based detection
- **Tracking confidence** (0-100%) that builds over time
- Tracking increases with sensor lock, decreases with maneuvers/jamming
- Detection must succeed before tracking begins
- Upgrades improve sensor strength and tracking growth rate

### Energy & Heat Management
- **Shared energy pool** across engines, shields, sensors, weapons, electronic warfare
- Energy reallocation via explicit commands
- **Heat generation** from weapon fire, engine burn, sensor boost, jamming
- Heat increases signature radius and reduces tracking stability
- Gradual heat dissipation per tick

### Maneuver System
- **Engine Burn** - increase/decrease range
- **Hard Evasion** - disrupt enemy tracking
- **Silent Mode** - reduce signature, limit detection
- **Sensor Boost** - improve tracking rate
- **Energy Reallocation** - shift power between subsystems
- **Jam Target** - degrade enemy sensors
- Each maneuver consumes energy, generates heat, has cooldown

### Projectile & Weapon System
- **Missile system** - delayed impact with time-to-impact tracking
- **Railgun** - instant resolution
- **EMP** - system disruption
- Projectile instances with origin, target, ETA, damage profile
- Projectiles resolved on tick when time_to_impact = 0

### Firing Solution Model
- **Solution Quality** calculation based on tracking confidence
- **Intercept Time** (ticks until impact)
- **Hit Probability** (abstracted from player)
- **Expected Damage Range**
- Player sees: Solution Confidence (Low/Moderate/High), ETA, Tracking %

### Electronic Warfare
- **Countermeasures** - decoys, missile interception, sensor jamming
- **Signature radius** affected by ship size, heat, active systems
- **Environmental concealment** - asteroid fields, nebula, ion storms reduce detection
- Large ships rely on countermeasures; small ships rely on stealth

### Disengagement Mechanics
- Maintain separation threshold for X ticks
- No weapon fire during disengagement attempt
- Tracking must drop below threshold
- Successful disengage ends CombatInstance with cooldown

### Combat Loot
- Pirate destruction drops cargo/credits/contraband
- Cargo drops attract scavengers
- Loot collection mechanics

## Economy System Enhancements

### Dynamic Pricing
- Price fluctuations based on supply/demand
- Port inventory affects prices
- Economic events trigger price changes
- Volatility per commodity affects price variance

### Economic Events
- Supply shortages
- Demand spikes
- Trade route disruptions
- Regional economic shifts

### AI Trader NPCs
- Autonomous traders moving commodities between ports
- AI traders as piracy targets
- Escort ships for high-value cargo
- Distress signals on attack

### Banking System
- Bank accounts separate from wallet
- Interest on deposits
- Loans and credit
- Inter-player credit transfers (wallet only in Phase 1)

## Mining System
- Resource extraction from asteroid fields
- Mining equipment and ships
- Ore processing
- Hazard system during mining
- Resource generation and depletion

## Exploration System
- Scanning and sensor mechanics
- Mapping data economy
- Rare anomaly events
- Exploration ships with specialized sensors
- Data trading

## Mission System Enhancements

### Complex Mission Types
- **Kill missions** - destroy specific targets
- **Scan missions** - gather intelligence
- **Multi-step missions** - sequential objectives
- **Escort missions** - protect AI traders
- **Reconnaissance missions** - explore systems

### Mission Features
- Mission chains and storylines
- Faction-specific missions
- Reputation requirements
- Mission failure consequences
- Time-sensitive missions with dynamic expiry

## Universe Expansion

### Multiple Regions
- Expand beyond single region
- Inter-region travel mechanics
- Regional economic differences
- Regional faction control

### Security Zones
- **Medium Security** (0.4-0.7) - moderate pirate activity
- **Black Sector** (0.0) - maximum danger, maximum reward
- Zone-specific mechanics and restrictions

### Procedural Generation
- Procedurally generated systems beyond hand-crafted core
- Seed-based world generation
- Dynamic system discovery

## Ship System Enhancements

### Ship Upgrades
- Hull reinforcement
- Shield upgrades
- Weapon upgrades
- Sensor upgrades
- Engine upgrades
- Cargo expansion

### Multiple Ship Classes
- Beyond courier: freighter, combat, exploration, mining
- Ship specialization and roles
- Multiple ships per player
- Ship storage and switching

### Ship Customization
- Loadout configuration
- Module slots
- Paint/cosmetics

## Communications System
- **IRN (Interstellar Relay Network)** - in-game messaging
- Player-to-player messaging
- Faction channels
- Trade advertisements
- Distress signals

## Advanced Features (Out of Scope / Future)

### Faction Systems
- Faction reputation
- Faction territories
- Faction missions and rewards
- Faction warfare

### Fleet Combat
- Multi-ship engagements
- Fleet coordination
- Carrier-based fighters
- Capital ships

### Player Trading
- Direct player-to-player cargo trading (credits only in Phase 1)
- Player-run markets
- Contracts and escrow

### Drones
- Drone deployment
- Drone types (combat, mining, exploration)
- Drone control mechanics

### GUI Client
- Port 2223 GUI client
- Visual interface
- Enhanced graphics
- Mouse controls

### Advanced Combat
- Cloaking technology
- Minefields
- Drone swarms
- Subsystem targeting
- Critical damage system

## Implementation Notes

### Phase 1 Simplifications
The Phase 1 combat system intentionally uses:
- **Simple turn-based combat** instead of predictive/tracking system
- **Basic hit/miss with pirate accuracy** instead of firing solution model
- **Fixed weapon damage** instead of energy/heat constrained system
- **No projectile tracking** - instant resolution
- **No maneuvers** - only attack/flee/surrender
- **No electronic warfare** - no jamming, countermeasures, or signature management

These simplifications validate the core architecture and game loop. Phase 2 will add the sophisticated submarine-style combat from `docs/04_combat/combat_requirements.md`.

### Architecture Considerations
The Phase 1 architecture is designed to support Phase 2 features:
- Tick engine can handle complex state updates
- Command queue supports arbitrary command types
- Database schema can be extended with new tables
- Subsystem interfaces can be expanded
- Configuration system supports new parameters

### Migration Strategy
When implementing Phase 2 features:
1. Extend existing interfaces rather than replacing them
2. Add new database tables/columns via migrations
3. Maintain backward compatibility where possible
4. Update configuration schema incrementally
5. Add new command types to tick engine routing
6. Expand TUI views for new features

---

**Last Updated**: 2026-03-06  
**Milestone**: Phase 1 (Milestone 2) Complete  
**Next Milestone**: Phase 2 (Milestone 3+)
