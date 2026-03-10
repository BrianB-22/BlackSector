# BlackSector Quick Start

Get BlackSector running in 5 minutes. For detailed information, see [README.md](README.md).

## Prerequisites

- Go 1.22+ installed
- Linux/macOS (Windows users: use WSL)
- SSH client (built into most systems)

## Quick Setup (Development)

```bash
# 1. Clone and enter directory
git clone https://github.com/BrianB-22/BlackSector.git
cd BlackSector

# 2. Download dependencies
go mod download

# 3. Build and start server
./start-server.sh
```

The server will start on port 2222. Leave this terminal running.

## Connect and Play

Open a new terminal:

```bash
# Connect to local server
ssh localhost -p 2222

# Or specify a username
ssh yourname@localhost -p 2222
```

First-time connection will prompt you to register. Choose a password and **save your player token**.

## Basic Commands

```bash
help              # Show all commands
system            # See current location and jump connections
dock 1            # Dock at port 1
market            # View prices (must be docked)
buy fuel_cells 5  # Buy 5 fuel cells
cargo             # Check your cargo
undock            # Leave port
jump 2            # Jump to system 2
```

## Quick Trading Loop

```bash
dock 1                    # Dock at starting port
market                    # Check prices
buy raw_ore 10            # Buy cheap commodity
undock                    # Leave port
jump 2                    # Jump to another system
dock 5                    # Dock at new port
market                    # Check if prices are better
sell raw_ore 10           # Sell for profit
```

## Production Deployment

### Build for Production

```bash
# Build optimized binary
go build -ldflags="-s -w" -o blacksector-server ./cmd/server

# Optional: compress binary
upx --best blacksector-server
```

### Server Setup

```bash
# 1. Copy files to server
scp blacksector-server user@server:/opt/blacksector/
scp -r config/ user@server:/opt/blacksector/
scp -r migrations/ user@server:/opt/blacksector/

# 2. SSH into server
ssh user@server

# 3. Create directories
cd /opt/blacksector
mkdir -p snapshots logs

# 4. Configure server (optional)
nano config/server.json
# Adjust ssh_port, starting_credits, etc.

# 5. Start server
./blacksector-server
```

### Run as System Service (systemd)

Create `/etc/systemd/system/blacksector.service`:

```ini
[Unit]
Description=BlackSector SSH Game Server
After=network.target

[Service]
Type=simple
User=blacksector
WorkingDirectory=/opt/blacksector
ExecStart=/opt/blacksector/blacksector-server
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable blacksector
sudo systemctl start blacksector
sudo systemctl status blacksector
```

View logs:

```bash
sudo journalctl -u blacksector -f
```

## Configuration

Edit `config/server.json` to customize:

```json
{
  "server": {
    "ssh_port": 2222,              // Change SSH port
    "tick_interval_ms": 2000,      // Game tick speed
    "max_concurrent_players": 50   // Player limit
  },
  "player": {
    "starting_credits": 1000,      // Starting money
    "starting_system_id": "nexus_prime"
  }
}
```

Changes require server restart.

## Database Management

### Backup Database

```bash
# Stop server first
sqlite3 blacksector.db ".backup blacksector-backup.db"
```

### Reset Database

```bash
# WARNING: Deletes all player data
rm blacksector.db
# Server will recreate on next start
```

### View Database

```bash
sqlite3 blacksector.db
sqlite> .tables
sqlite> SELECT * FROM players;
sqlite> .quit
```

## Testing Your Installation

### Run Test Suite

```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./internal/economy/
```

### Test Client Connection

```bash
# Build test client
go build -o testclient ./cmd/testclient

# Connect
./testclient
```

## Troubleshooting

### Server won't start

**Error: "address already in use"**
```bash
# Check if port 2222 is in use
lsof -i :2222
# Kill existing process or change ssh_port in config/server.json
```

**Error: "database is locked"**
```bash
# Another process is using the database
pkill blacksector-server
rm blacksector.db-shm blacksector.db-wal
```

### Can't connect via SSH

**Error: "Connection refused"**
```bash
# Check server is running
ps aux | grep blacksector-server

# Check firewall (production servers)
sudo ufw allow 2222/tcp
```

**Error: "Permission denied (publickey)"**
```bash
# Server uses password auth, not SSH keys
# Make sure you're entering a username:
ssh username@localhost -p 2222
```

### Game Issues

**Lost player token**
- Tokens are stored in the database
- Use `cmd/hashtoken` to generate a new hash if needed
- Or create a new account with a different username

**Ship stuck/broken state**
```bash
# Check server logs
tail -f server.log

# Check debug logs (if enabled)
tail -f debug.log
```

**Market prices seem wrong**
- Phase 1 uses static pricing with zone multipliers
- Low Security systems pay ~18% more
- This is intentional; dynamic pricing comes in Phase 2

## Next Steps

- Read [README.md](README.md) for gameplay guide
- Check `docs/` for complete specifications
- See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines
- Join the community (add your Discord/forum link here)

## Quick Reference

| Command | Description |
|---------|-------------|
| `help` | Show all commands |
| `system` | Current location info |
| `dock <id>` | Dock at port |
| `undock` | Leave port |
| `market` | View prices (docked) |
| `buy <commodity> <qty>` | Purchase goods |
| `sell <commodity> <qty>` | Sell goods |
| `cargo` | View cargo hold |
| `jump <id>` | Jump to system |

**Commodities**: `food_supplies`, `fuel_cells`, `raw_ore`, `refined_ore`, `machinery`, `electronics`, `luxury_goods`

---

**Need help?** Open an issue on GitHub or check the full [README.md](README.md)
