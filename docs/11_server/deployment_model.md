# Deployment Model Specification

## Version: 0.1

## Status: Draft

## Owner: Core Architecture

## Last Updated: 2026-03-05

---

# 1. Purpose

Defines the deployment model for a BlackSector server instance.

This document describes how to install, configure, and operate the server on a Linux host.

---

# 2. Design Principles

* Single binary deployment — no runtime dependencies beyond the OS
* Operator-friendly — minimal setup steps to a running server
* Configurable — all parameters in `config/server.json`
* Self-contained — all data files under a single working directory
* Lightweight — runs comfortably on modest hardware

---

# 3. System Requirements

## Minimum

| Resource | Requirement                                  |
| -------- | -------------------------------------------- |
| OS       | Linux (amd64 or arm64)                       |
| RAM      | 512 MB                                       |
| CPU      | 1 core                                       |
| Disk     | 1 GB (game data + logs + snapshots)          |
| Network  | Public ports 2222 and 2223 accessible        |

## Recommended (100 concurrent players)

| Resource | Recommendation                               |
| -------- | -------------------------------------------- |
| OS       | Linux (amd64)                                |
| RAM      | 2 GB                                         |
| CPU      | 2 cores                                      |
| Disk     | 10 GB                                        |
| Network  | Low-latency connection, 10 Mbps+             |

---

# 4. Directory Structure

```
/opt/blacksector/
├── blacksector          # server binary
├── config/
│   ├── server.json      # server configuration
│   ├── economy/
│   │   ├── commodities.json
│   │   └── economic_events.json
│   ├── ships/
│   │   ├── ship_classes.json
│   │   └── upgrades.json
│   ├── ai/
│   │   └── trader_names.json
│   └── missions/
│       ├── core/
│       │   ├── combat_missions.json
│       │   ├── trade_missions.json
│       │   └── exploration_missions.json
│       └── community/   # community-contributed missions
├── data/
│   └── blacksector.db   # SQLite database
├── snapshots/
│   ├── snapshot_latest.json
│   └── snapshot_008205_1709612345.json
├── logs/
│   ├── events.log
│   └── server.log
└── certs/               # TLS certificates (for GUI port)
    ├── server.crt
    └── server.key
```

---

# 5. Installation

## Step 1: Create service user

```bash
useradd -r -s /bin/false -d /opt/blacksector blacksector
```

## Step 2: Create directory structure

```bash
mkdir -p /opt/blacksector/{config/economy,config/ships,config/ai,config/missions/core,config/missions/community}
mkdir -p /opt/blacksector/{data,snapshots,logs,certs}
```

## Step 3: Deploy binary

```bash
cp blacksector /opt/blacksector/
chmod +x /opt/blacksector/blacksector
```

## Step 4: Deploy configuration files

Copy all JSON configuration files to the appropriate directories under `config/`.

## Step 5: Set permissions

```bash
chown -R blacksector:blacksector /opt/blacksector/
chmod 750 /opt/blacksector/
chmod 640 /opt/blacksector/config/server.json
```

---

# 6. SSH Configuration

The server binds its own SSH listener on port 2222. This is a game-specific SSH server, separate from the host's SSH daemon (typically on port 22).

The server requires an SSH host key. Generate one if not present:

```bash
ssh-keygen -t ed25519 -f /opt/blacksector/certs/ssh_host_key -N ""
chown blacksector:blacksector /opt/blacksector/certs/ssh_host_key
chmod 600 /opt/blacksector/certs/ssh_host_key
```

Set `ssh_host_key_file` in `config/server.json`:

```json
{
  "ssh_host_key_file": "certs/ssh_host_key"
}
```

---

# 7. TLS Configuration (GUI Port)

The GUI port (2223) requires TLS. Generate or obtain a certificate:

## Self-Signed (development / LAN use)

```bash
openssl req -x509 -newkey rsa:4096 \
  -keyout /opt/blacksector/certs/server.key \
  -out /opt/blacksector/certs/server.crt \
  -days 365 -nodes \
  -subj "/CN=blacksector"
```

## CA-Signed (production / internet-facing)

Use Let's Encrypt or your organization's CA. Place the certificate chain and private key in `certs/`.

Set in `config/server.json`:

```json
{
  "tls_cert_file": "certs/server.crt",
  "tls_key_file": "certs/server.key"
}
```

If TLS files are absent or invalid, port 2223 is not opened. SSH (port 2222) is unaffected.

---

# 8. Systemd Service

Create `/etc/systemd/system/blacksector.service`:

```ini
[Unit]
Description=BlackSector Game Server
After=network.target

[Service]
Type=simple
User=blacksector
Group=blacksector
WorkingDirectory=/opt/blacksector
ExecStart=/opt/blacksector/blacksector
Restart=on-failure
RestartSec=10
StandardOutput=append:/opt/blacksector/logs/server.log
StandardError=append:/opt/blacksector/logs/server.log
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
systemctl daemon-reload
systemctl enable blacksector
systemctl start blacksector
```

---

# 9. Firewall Configuration

Open the game ports:

```bash
# Using ufw
ufw allow 2222/tcp   # SSH (TEXT mode)
ufw allow 2223/tcp   # TLS/TCP (GUI mode, optional)
```

Do not expose the SQLite database file or snapshot directory over the network.

---

# 10. Backups

The following files should be included in regular backups:

| Path                           | Description              | Frequency     |
| ------------------------------ | ------------------------ | ------------- |
| `data/blacksector.db`          | SQLite game database     | Daily minimum |
| `snapshots/`                   | Game state snapshots     | Daily minimum |
| `config/`                      | All configuration files  | On change     |

Stop the server before copying the SQLite file for a consistent backup:

```bash
systemctl stop blacksector
cp /opt/blacksector/data/blacksector.db /backup/blacksector_$(date +%Y%m%d).db
systemctl start blacksector
```

Alternatively, use SQLite's `.backup` command with the server running.

---

# 11. Log Management

Logs are written to `logs/events.log` (NDJSON event log) and `logs/server.log` (server stdout/stderr via systemd).

Rotate logs using logrotate:

```
/opt/blacksector/logs/*.log {
    daily
    rotate 30
    compress
    missingok
    notifempty
    copytruncate
}
```

---

# 12. Updating the Server

1. Download new binary
2. Stop server: `systemctl stop blacksector`
3. Replace binary: `cp blacksector_new /opt/blacksector/blacksector`
4. Start server: `systemctl start blacksector`

The server will load the most recent snapshot on startup and resume from the last saved tick.

Configuration file changes take effect on restart. Mission and event config changes may be hot-reloaded while running via `server reload` admin command.

---

# 13. Monitoring

The server exposes its health through:

* Exit code (non-zero on crash)
* `logs/server.log` (stdout/stderr)
* `logs/events.log` (NDJSON event log with `tick_slow` and error events)
* `server status` admin command

For automated monitoring, watch for `tick_slow` and error events in the event log.

---

# 14. Non-Goals (v1)

* Container-based deployment (Docker/Kubernetes)
* Auto-scaling
* High availability / failover
* Remote management API
* Health check HTTP endpoint

---

# End of Document
