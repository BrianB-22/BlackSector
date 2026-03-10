# BlackSector

A persistent, multiplayer text-based space trading simulation served over SSH. Navigate a dangerous universe, trade commodities across star systems, and survive encounters with pirates in this server-authoritative space game.

## What is BlackSector?

BlackSector is a strategic space trading game where you pilot a ship through a procedurally connected universe. The game runs continuously on a server—the universe doesn't pause when you disconnect. Trade commodities between ports, navigate through jump points, and manage your ship's resources while avoiding (or fighting) NPC pirates in low-security space.

This is not an arcade game. It's a simulation where information, planning, and risk management matter more than reflexes.

## Phase 1 Features (Current Release)

This is the **Phase 1 Vertical Slice**—a complete but intentionally narrow implementation of the core game loop:

### Universe & Navigation
- **15-20 star systems** connected by jump points
- **High Security** and **Low Security** zones (higher risk, higher profit)
- **Federated Space** origin system (safe starting area)
- **Jump system** for interstellar travel (consumes fuel)
- **Docking** at ports to access markets

### Economy & Trading
- **7 commodities** to trade: food supplies, fuel cells, raw ore, refined ore, machinery, electronics, luxury goods
- **Static pricing** with zone-based multipliers (Low Security pays ~18% more)
- **Buy/sell** at docked ports
- **Cargo management** (20-unit capacity on starter ship)
- **Credits tracking** (start with 10,000 credits)

### Ships & Status
- **Courier-class** starting ship (100 hull, 50 shields, 100 energy, 20 cargo)
- **Real-time status display** showing system, credits, hull/shields/energy
- **Fuel consumption** on jumps

### Interface
- **Full TEXT mode** terminal interface with ANSI colors
- **Command-line input** with clear feedback
- **System maps** showing jump connections
- **Market displays** with current prices
- **Cargo manifest** view
- **Help system** with command reference

### Persistence
- **SQLite database** stores all game state
- **Automatic snapshots** every 100 ticks
- **Survives server restarts** with full state recovery

## Connecting to the Game

BlackSector is played over SSH. Connect using any SSH client:

```bash
ssh -p 2222 username@your-server-address
```

Replace `username` with your chosen player name and `your-server-address` with the server's hostname or IP.

### First-Time Registration

When you connect for the first time, you'll be prompted to register:

1. The server will ask if you want to create a new account
2. Choose a password (hashed with bcrypt)
3. You'll receive a **player token**—save this securely
4. Your ship will spawn at the Federated Space origin starbase
5. You'll start with 10,000 credits

**Important**: Save your token! You'll need it to reconnect if you use a different SSH client or machine.

## How to Play

### Basic Commands

```
help              Show all available commands
system            Display current system info and jump connections
market            Show commodity prices at current port (must be docked)
cargo             Display your ship's cargo manifest
jump <system_id>  Jump to a connected system
dock <port_id>    Dock at a port in your current system
undock            Undock from current port
buy <commodity> <quantity>   Buy commodities (must be docked)
sell <commodity> <quantity>  Sell commodities (must be docked)
```

### Commodity Names

Use these exact names when buying/selling:
- `food_supplies`
- `fuel_cells`
- `raw_ore`
- `refined_ore`
- `machinery`
- `electronics`
- `luxury_goods`

### Example Trading Session

```bash
# Check where you are
> system
System: Federated Space (ID: 1)
Security: High
Ports: [1] Federated Starbase
Jump Connections: [2] Alpha Centauri, [3] Sirius

# Dock at the starbase
> dock 1
Docked at Federated Starbase

# Check the market
> market
=== Market Prices at Federated Starbase ===
food_supplies    Buy: 110cr  Sell: 90cr
fuel_cells       Buy: 165cr  Sell: 135cr
raw_ore          Buy: 88cr   Sell: 72cr
...

# Buy some raw ore
> buy raw_ore 10
Purchased 10 units of raw_ore for 880 credits

# Check your cargo
> cargo
=== Cargo Manifest ===
raw_ore: 10 units
Capacity: 10/20

# Undock and jump to another system
> undock
Undocked from port

> jump 2
Jumped to Alpha Centauri

# Find a port that buys raw ore at a higher price
> system
System: Alpha Centauri (ID: 2)
Security: Low
Ports: [5] Mining Station Alpha
...

> dock 5
> market
# (Check if they buy raw_ore at a good price)

> sell raw_ore 10
Sold 10 units of raw_ore for 850 credits
```

### Tips for New Players

1. **Start in High Security**: Federated Space and High Security systems are safe from pirates
2. **Low Security = Higher Profit**: Prices are ~18% better in Low Security zones, but pirates spawn there
3. **Watch Your Fuel**: Jumping consumes fuel—buy fuel_cells at ports
4. **Trade Loops**: Look for systems that produce cheap commodities and systems that buy them expensive
5. **Cargo Capacity**: Your starter ship holds 20 units—plan your trades accordingly

## Development Setup

### Prerequisites

- Go 1.22 or higher
- Linux/macOS (Windows via WSL)
- SQLite (embedded via modernc.org/sqlite)

### Building from Source

```bash
# Clone the repository
git clone https://github.com/BrianB-22/BlackSector.git
cd BlackSector

# Install dependencies
go mod download

# Build the server
go build -o blacksector-server cmd/server/main.go

# Run the server
./blacksector-server
```

The server will:
- Listen on port 2222 for SSH connections
- Create `blacksector.db` for game state
- Create `snapshots/` directory for periodic saves
- Log to `server.log`

### Configuration

Server configuration is in `config/server.json`:

```json
{
  "ssh_port": 2222,
  "tick_interval_ms": 2000,
  "snapshot_interval_ticks": 100,
  "starting_credits": 10000,
  "world_config": "config/world/alpha_sector.json"
}
```

World configuration (systems, ports, jump connections) is in `config/world/alpha_sector.json`.

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for a specific package
go test ./internal/economy/
```

### Project Structure

```
cmd/
  server/          Main game server executable
  testclient/      Test SSH client for development
  hashtoken/       Utility for hashing player tokens
internal/
  config/          Server configuration loading
  db/              SQLite database layer
  engine/          Tick loop and command processing
  economy/         Trading and commodity pricing
  navigation/      Jump system and docking
  registration/    Player registration and authentication
  session/         SSH session management
  tui/             Terminal UI (bubbletea)
  world/           World configuration and topology
config/
  server.json      Server settings
  world/           World definition files
docs/              Complete specification documents
migrations/        SQL migration files
```

## Known Limitations (Phase 1)

These are intentional limitations that will be addressed in Phase 2:

- **Static prices**: Commodity prices don't change dynamically
- **No AI traders**: Markets don't rebalance automatically
- **Small universe**: Only 15-20 systems (proof of concept scale)
- **One ship class**: All players start with the same courier ship
- **No ship upgrades**: Can't modify or upgrade your ship
- **No combat loot**: Pirates don't drop anything when destroyed
- **No missions**: Delivery missions and other mission types are planned for Phase 2

## Roadmap

### Phase 2 (Planned)
- Dynamic commodity pricing based on supply/demand
- AI trader NPCs that rebalance markets
- Delivery missions with rewards
- Ship upgrades and equipment
- Combat loot system
- Expanded universe (50+ systems)
- Multiple ship classes

### Future Phases
- Mining and resource extraction
- Exploration and scanning mechanics
- Player-to-player interactions
- Economic events and market volatility
- Multiple regions and the Black Sector
- GUI client option

## Documentation

Comprehensive documentation is available in the `docs/` directory:

- `docs/00_overview/vision.md` - Game vision and design philosophy
- `docs/01_architecture/` - System architecture and design
- `docs/10_data_models/` - Database schema and data structures
- `docs/15_roadmap/` - Development roadmap and milestones

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines and coding conventions.

## License

See [LICENSE](LICENSE) for license information.

## Support

For bugs, feature requests, or questions:
- Open an issue on GitHub
- Check the `docs/` directory for detailed documentation
- Use the in-game `help` command for command reference

---

**Current Status**: Phase 1 Vertical Slice - Core trading and navigation loop complete

**Server Version**: 0.1.0

**Last Updated**: 2026-03-05
