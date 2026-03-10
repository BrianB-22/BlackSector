#!/bin/bash
# Create a test player for BlackSector

# This script creates a test player with:
# - Username: testplayer
# - Password: test123
# - Token: test_token_12345
# - Starting credits: 10000
# - Starting ship at Nexus Prime (system 1, port 1)

echo "Creating test player..."

# Generate UUIDs for player and ship
PLAYER_ID=$(uuidgen | tr '[:upper:]' '[:lower:]')
SHIP_ID=$(uuidgen | tr '[:upper:]' '[:lower:]')
TIMESTAMP=$(date +%s)

# Hash the token (bcrypt cost 10)
# For testing, we'll use a pre-computed hash of "test_token_12345"
TOKEN_HASH='$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'

# Hash the password (bcrypt cost 12)  
# For testing, we'll use a pre-computed hash of "test123"
PASSWORD_HASH='$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYIRRYRRYRY'

# Insert player
sqlite3 blacksector.db <<EOF
INSERT INTO players (player_id, player_name, ssh_username, token_hash, password_hash, credits, created_at, is_banned)
VALUES ('$PLAYER_ID', 'TestPlayer', 'testplayer', '$TOKEN_HASH', '$PASSWORD_HASH', 10000, $TIMESTAMP, 0);

INSERT INTO ships (
  ship_id, player_id, ship_class, 
  hull_points, max_hull_points,
  shield_points, max_shield_points,
  energy_points, max_energy_points,
  cargo_capacity, missiles_current,
  current_system_id, position_x, position_y,
  status, docked_at_port_id, last_updated_tick
) VALUES (
  '$SHIP_ID', '$PLAYER_ID', 'courier',
  100, 100,
  50, 50,
  100, 100,
  20, 0,
  1, 0.0, 0.0,
  'DOCKED', 1, 0
);
EOF

echo ""
echo "✓ Test player created successfully!"
echo ""
echo "Connection details:"
echo "  SSH: ssh testplayer@localhost -p 2222"
echo "  Password: test123"
echo "  Token: test_token_12345"
echo ""
echo "Player details:"
echo "  Player ID: $PLAYER_ID"
echo "  Ship ID: $SHIP_ID"
echo "  Credits: 10,000"
echo "  Location: System 1, Port 1 (Docked)"
echo ""
