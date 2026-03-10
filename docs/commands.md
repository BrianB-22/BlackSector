# BlackSector Command Reference

This document provides a complete reference for all player commands available in Phase 1 of BlackSector.

## Command Categories

Commands are organized into the following categories:
- **Navigation** - Moving between systems and docking
- **Trading** - Buying and selling commodities
- **Combat** - Fighting NPC pirates
- **Missions** - Accepting and managing delivery missions
- **Information** - Viewing game state and help

## Command Syntax Conventions

- `<parameter>` - Required parameter
- `[parameter]` - Optional parameter (not used in Phase 1)
- Parameters are case-insensitive unless otherwise noted
- Commodity names use underscores: `food_supplies`, `fuel_cells`, etc.

---

## Navigation Commands

### jump

**Syntax:** `jump <system_id>`

**Description:** Jump to a connected star system via jump point.

**Requirements:**
- Ship must be in space (not docked)
- Ship must not be in combat
- Target system must be connected to current system via jump point
- Sufficient fuel (automatically available in Phase 1)

**Parameters:**
- `system_id` - Integer ID of the destination system (must be positive)

**Examples:**
```
jump 5
jump 12
```

**Common Errors:**
- "jump requires exactly one argument" - Missing or too many parameters
- "invalid system_id: must be an integer" - Non-numeric system ID
- "invalid system_id: must be positive" - Zero or negative system ID
- "Cannot jump while docked" - Must undock first
- "Cannot jump while in combat" - Must flee or defeat pirate first
- "No jump connection to system X" - Systems not connected

**Notes:**
- Jumping to Low Security systems (SecurityLevel < 0.4) may trigger pirate encounters
- Use the `system` command to see available jump connections


### dock

**Syntax:** `dock <port_id>`

**Description:** Dock at a port in your current system.

**Requirements:**
- Ship must be in space (not already docked)
- Ship must not be in combat
- Port must exist in current system

**Parameters:**
- `port_id` - Integer ID of the port (must be positive)

**Examples:**
```
dock 101
dock 205
```

**Common Errors:**
- "dock requires exactly one argument" - Missing or too many parameters
- "invalid port_id: must be an integer" - Non-numeric port ID
- "invalid port_id: must be positive" - Zero or negative port ID
- "Cannot dock while in combat" - Must flee or defeat pirate first
- "Port X not in current system" - Port is in a different system
- "Already docked" - Must undock first

**Notes:**
- Docking is required to access the market, missions, and repairs
- Use the `system` command to see ports in your current system


### undock

**Syntax:** `undock`

**Description:** Undock from the current port and enter space.

**Requirements:**
- Ship must be docked at a port

**Parameters:** None

**Examples:**
```
undock
```

**Common Errors:**
- "undock takes no arguments" - Command does not accept parameters
- "Not docked at any port" - Ship is already in space

**Notes:**
- You cannot trade or accept missions while in space
- Undocking is required before jumping to another system


### system

**Syntax:** `system`

**Description:** Display information about your current star system, including ports and available jump connections.

**Requirements:** None (can be used anytime)

**Parameters:** None

**Examples:**
```
system
```

**Output Includes:**
- System name and ID
- Security level (Federated Space, High Security, Low Security)
- List of ports in the system with IDs and types
- Available jump connections to other systems

**Common Errors:**
- "system takes no arguments" - Command does not accept parameters

**Notes:**
- This is a local command (processed immediately, not queued)
- Use this to find port IDs for docking and system IDs for jumping

---

## Trading Commands

### buy

**Syntax:** `buy <commodity> <quantity>`

**Description:** Purchase commodities from the current port's market.

**Requirements:**
- Ship must be docked at a port
- Port must sell the specified commodity
- Player must have sufficient credits
- Ship must have sufficient cargo capacity

**Parameters:**
- `commodity` - Commodity name (see valid commodities below)
- `quantity` - Number of units to purchase (must be positive integer)

**Valid Commodities:**
- `food_supplies` - Essential provisions
- `fuel_cells` - Energy and propulsion fuel
- `raw_ore` - Unprocessed minerals
- `refined_ore` - Processed metals
- `machinery` - Industrial equipment
- `electronics` - Technology components
- `luxury_goods` - High-value consumer items

**Examples:**
```
buy food_supplies 10
buy refined_ore 5
buy luxury_goods 3
```

**Common Errors:**
- "buy requires two arguments" - Missing parameters
- "invalid quantity: must be an integer" - Non-numeric quantity
- "invalid quantity: must be positive" - Zero or negative quantity
- "invalid commodity: X" - Unrecognized commodity name
- "Not docked at any port" - Must dock first
- "Insufficient credits" - Cannot afford purchase
- "Insufficient cargo capacity" - Not enough cargo space
- "Port does not sell X" - Commodity not available at this port type
- "Insufficient port inventory" - Port out of stock

**Notes:**
- Buy price = zone_price × 1.10 (10% markup)
- Low Security systems have 18% higher base prices
- Use `market` command to see current prices and availability
- Port types determine commodity availability:
  - Trading ports: sell food_supplies, luxury_goods, electronics
  - Mining ports: sell raw_ore, fuel_cells
  - Refueling ports: sell fuel_cells, food_supplies


### sell

**Syntax:** `sell <commodity> <quantity>`

**Description:** Sell commodities from your cargo to the current port's market.

**Requirements:**
- Ship must be docked at a port
- Port must buy the specified commodity
- Player must have the commodity in cargo

**Parameters:**
- `commodity` - Commodity name (see valid commodities in `buy` command)
- `quantity` - Number of units to sell (must be positive integer)

**Examples:**
```
sell food_supplies 10
sell raw_ore 8
sell machinery 2
```

**Common Errors:**
- "sell requires two arguments" - Missing parameters
- "invalid quantity: must be an integer" - Non-numeric quantity
- "invalid quantity: must be positive" - Zero or negative quantity
- "invalid commodity: X" - Unrecognized commodity name
- "Not docked at any port" - Must dock first
- "Commodity not in cargo" - You don't have this commodity
- "Insufficient quantity in cargo" - You don't have enough units
- "Port does not buy X" - Port won't purchase this commodity

**Notes:**
- Sell price = zone_price × 0.90 (10% markdown)
- Low Security systems have 18% higher base prices (better profit margins)
- Use `cargo` command to see what you're carrying
- Port types determine what they buy:
  - Trading ports: buy raw_ore, refined_ore, machinery
  - Mining ports: buy food_supplies, machinery
  - Refueling ports: buy nothing (sell only)


### market

**Syntax:** `market`

**Description:** Display the current port's commodity market with buy/sell prices and available quantities.

**Requirements:**
- Ship must be docked at a port

**Parameters:** None

**Examples:**
```
market
```

**Output Includes:**
- Commodity names
- Buy prices (what you pay to purchase)
- Sell prices (what you receive when selling)
- Available quantity in port inventory
- Which commodities the port buys vs. sells

**Common Errors:**
- "market takes no arguments" - Command does not accept parameters
- "Not docked at any port" - Must dock first

**Notes:**
- This is a local command (processed immediately)
- Prices are static in Phase 1 (no dynamic pricing)
- Use this before trading to check prices and availability


### cargo

**Syntax:** `cargo`

**Description:** Display your ship's cargo manifest showing all commodities and quantities.

**Requirements:** None (can be used anytime)

**Parameters:** None

**Examples:**
```
cargo
```

**Output Includes:**
- List of commodities in cargo with quantities
- Total cargo used / maximum cargo capacity
- Available cargo space

**Common Errors:**
- "cargo takes no arguments" - Command does not accept parameters

**Notes:**
- This is a local command (processed immediately)
- Courier-class ships have 20 cargo capacity
- Empty cargo holds show "Cargo hold empty"

---

## Combat Commands

Combat commands are only available when engaged with an NPC pirate. Combat is initiated automatically when a pirate spawns after jumping to a Low Security system.

### attack

**Syntax:** `attack`

**Description:** Attack the pirate with your ship's weapons.

**Requirements:**
- Ship must be in combat with a pirate
- Ship must not be destroyed

**Parameters:** None

**Examples:**
```
attack
```

**Combat Mechanics:**
- Courier ships deal 15 damage per attack
- Damage applies to pirate shields first, then hull
- Pirate counter-attacks if it survives
- Combat ends when pirate is destroyed or flees
- Pirates flee when hull drops below threshold (15% for raiders, 10% for marauders)

**Common Errors:**
- "attack takes no arguments" - Command does not accept parameters
- "Not in combat" - No active combat encounter

**Notes:**
- This is the primary combat action
- Pirate stats vary by tier:
  - Raider: 60 hull, 20 shield, 12-18 damage, 60% accuracy
  - Marauder: 90 hull, 40 shield, 18-28 damage, 65% accuracy


### flee

**Syntax:** `flee`

**Description:** Attempt to disengage from combat and escape.

**Requirements:**
- Ship must be in combat with a pirate

**Parameters:** None

**Examples:**
```
flee
```

**Flee Mechanics:**
- Flee attempts have a success chance (implementation-dependent)
- Successful flee ends combat immediately
- Failed flee may result in pirate counter-attack
- No credit or cargo penalty for fleeing

**Common Errors:**
- "flee takes no arguments" - Command does not accept parameters
- "Not in combat" - No active combat encounter

**Notes:**
- Use this when outmatched or low on hull/shields
- Fleeing is safer than surrendering (no credit loss)


### surrender

**Syntax:** `surrender`

**Description:** Surrender to the pirate, ending combat with a credit penalty.

**Requirements:**
- Ship must be in combat with a pirate

**Parameters:** None

**Examples:**
```
surrender
```

**Surrender Mechanics:**
- Immediately ends combat
- Deducts 40% of wallet credits (configurable)
- No damage to ship
- No cargo loss in Phase 1

**Common Errors:**
- "surrender takes no arguments" - Command does not accept parameters
- "Not in combat" - No active combat encounter

**Notes:**
- Use this as a last resort when flee fails
- Better than ship destruction (which costs insurance payout and cargo)
- If you have few credits, surrender penalty is minimal

---

## Mission Commands

### missions

**Syntax:** `missions` (list available missions)

**Description:** Display available missions at the current port.

**Requirements:**
- Ship must be docked at a port

**Parameters:** None

**Examples:**
```
missions
```

**Output Includes:**
- Mission IDs
- Mission names and descriptions
- Objectives (e.g., "Deliver 15 refined_ore to Port X")
- Credit rewards
- Expiry time (if applicable)
- Security zone requirements

**Common Errors:**
- "missions takes no arguments" - Command does not accept parameters
- "Not docked at any port" - Must dock first

**Notes:**
- This is a local command (processed immediately)
- Missions are filtered by security zone and port
- Only deliver_commodity missions available in Phase 1


### missions accept

**Syntax:** `missions accept <mission_id>`

**Description:** Accept a mission and add it to your active mission slot.

**Requirements:**
- Ship must be docked at a port
- Player must not have an active mission (Phase 1 limit: one mission at a time)
- Mission must be available at current port

**Parameters:**
- `mission_id` - String ID of the mission (e.g., "emergency_ore_run")

**Examples:**
```
missions accept emergency_ore_run
missions accept food_delivery_alpha
```

**Common Errors:**
- "missions accept requires one argument" - Missing mission ID
- "Not docked at any port" - Must dock first
- "Already have an active mission" - Must complete or abandon current mission first
- "Mission not found" - Invalid mission ID
- "Mission not available at this port" - Wrong port or security zone

**Notes:**
- Accepting a mission does NOT provide the required commodities
- You must purchase commodities separately
- Mission timer starts immediately upon acceptance
- Use `missions status` to track progress


### missions status

**Syntax:** `missions status`

**Description:** Display your active mission status and objective progress.

**Requirements:**
- Player must have an active mission

**Parameters:** None

**Examples:**
```
missions status
```

**Output Includes:**
- Mission name and description
- Current objective and progress
- Destination port and system
- Time remaining (if mission has expiry)
- Reward amount

**Common Errors:**
- "missions status takes no arguments" - Command does not accept parameters
- "No active mission" - You haven't accepted a mission

**Notes:**
- This is a local command (processed immediately)
- Objectives update automatically each tick
- Mission completes when you dock at destination with required commodities


### missions abandon

**Syntax:** `missions abandon`

**Description:** Abandon your active mission without completing it.

**Requirements:**
- Player must have an active mission

**Parameters:** None

**Examples:**
```
missions abandon
```

**Abandon Mechanics:**
- Immediately cancels the mission
- No reward granted
- No penalty (credits or reputation)
- Frees mission slot for new mission

**Common Errors:**
- "missions abandon takes no arguments" - Command does not accept parameters
- "No active mission" - You haven't accepted a mission

**Notes:**
- Use this if you can't complete the mission or want to accept a different one
- Commodities purchased for the mission remain in your cargo
- Repeatable missions become available again after cooldown

---

## Information Commands

### help

**Syntax:** `help`

**Description:** Display a list of all available commands with brief descriptions.

**Requirements:** None (can be used anytime)

**Parameters:** None

**Examples:**
```
help
```

**Output Includes:**
- Command names grouped by category
- Brief syntax for each command
- Quick reference for common actions

**Common Errors:**
- "help takes no arguments" - Command does not accept parameters

**Notes:**
- This is a local command (processed immediately)
- Use this document for detailed command information

---

## Command Processing

### Local vs. Engine Commands

Commands are processed in two ways:

**Local Commands** (processed immediately by your session):
- `system` - View current system info
- `market` - View market prices
- `cargo` - View cargo manifest
- `help` - View command list
- `missions` - View available missions (when docked)
- `missions status` - View active mission

**Engine Commands** (queued and processed on next tick):
- `jump` - Jump to another system
- `dock` - Dock at a port
- `undock` - Undock from port
- `buy` - Purchase commodities
- `sell` - Sell commodities
- `attack` - Attack pirate
- `flee` - Flee from combat
- `surrender` - Surrender to pirate
- `missions accept` - Accept a mission
- `missions abandon` - Abandon a mission

### Rate Limiting

- Maximum 10 commands per session per tick (2-second tick interval)
- Exceeding this limit will result in commands being dropped
- Local commands do not count toward rate limit

### Error Handling

When a command fails:
1. An error message is displayed explaining the problem
2. No game state changes occur
3. You can retry with corrected parameters
4. The server continues running normally

Common error categories:
- **Syntax errors** - Invalid command format or parameters
- **State errors** - Action not allowed in current state (e.g., trading while in space)
- **Resource errors** - Insufficient credits, cargo, or inventory
- **Validation errors** - Invalid IDs or non-existent entities

---

## Quick Reference

### Starting Out
1. `system` - See where you are and what ports are available
2. `dock <port_id>` - Dock at a port
3. `market` - Check commodity prices
4. `buy <commodity> <quantity>` - Purchase goods
5. `undock` - Leave the port
6. `jump <system_id>` - Travel to another system

### Trading Loop
1. Dock at a trading port in Federated Space
2. Buy low-price commodities (e.g., food_supplies)
3. Undock and jump to a Low Security system
4. Dock at a mining port
5. Sell commodities at higher prices (18% markup in Low Sec)
6. Buy different commodities to sell elsewhere
7. Return to High Security or Federated Space

### Combat Survival
1. When pirate spawns: assess threat (raider vs. marauder)
2. `attack` - Fight if you have high hull/shields
3. `flee` - Escape if outmatched (no penalty)
4. `surrender` - Last resort (40% credit loss)
5. If destroyed: respawn at nearest port with insurance payout

### Mission Completion
1. Dock at a port with missions
2. `missions` - View available missions
3. `missions accept <id>` - Accept a delivery mission
4. `buy <commodity> <quantity>` - Purchase required goods
5. Undock and jump to destination system
6. Dock at destination port
7. Mission auto-completes, reward granted
8. `missions status` - Check progress anytime

---

## Tips and Best Practices

### Navigation
- Always check `system` before jumping to see available connections
- Low Security systems (SecurityLevel < 0.4) have pirate spawns
- Federated Space (SecurityLevel 2.0) is completely safe
- Plan multi-jump routes to minimize pirate encounters

### Trading
- Low Security systems offer 18% higher prices (better profit margins)
- Port types determine what they buy/sell - check with `market`
- Always check cargo capacity before buying
- Buy low in Federated Space, sell high in Low Security

### Combat
- Raiders (70% spawn rate) are easier than marauders (30%)
- Shields regenerate between combats, hull does not
- Fleeing has no penalty - use it liberally
- Surrendering costs 40% of credits - only use when desperate
- Ship destruction costs all cargo + respawn at nearest port

### Missions
- Only one mission at a time in Phase 1
- Missions do NOT provide commodities - buy them yourself
- Check mission expiry time before accepting
- Deliver missions are straightforward: buy, transport, dock
- Abandoning missions has no penalty

### Economy
- Courier ships have 20 cargo capacity
- Starting credits: 10,000
- Insurance payout on death: 5,000
- Buy/sell spread: ±10% from zone price
- No dynamic pricing in Phase 1 (prices are static)

---

## Troubleshooting

### "Command not recognized"
- Check spelling and capitalization
- Use `help` to see available commands
- Ensure you're using underscores in commodity names (e.g., `food_supplies` not `food supplies`)

### "Not docked at any port"
- Trading and mission commands require docking
- Use `system` to find port IDs
- Use `dock <port_id>` to dock

### "Cannot jump while docked"
- Use `undock` before jumping
- Check ship status in status bar

### "Insufficient credits"
- Check wallet balance in status bar
- Sell commodities to raise funds
- Complete missions for credit rewards

### "Insufficient cargo capacity"
- Use `cargo` to check current capacity
- Sell or deliver commodities to free space
- Courier ships have 20 capacity maximum

### "Port does not sell/buy X"
- Port types have different inventories
- Trading ports: different stock than mining/refueling ports
- Use `market` to see what's available

### "Already have an active mission"
- Phase 1 limit: one mission at a time
- Complete current mission or use `missions abandon`
- Check `missions status` for current mission

---

**End of Command Reference**
