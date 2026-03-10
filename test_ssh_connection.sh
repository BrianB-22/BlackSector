#!/bin/bash
# Test script for SSH connection and handshake

# Add a test player to the database
echo "Adding test player to database..."
sqlite3 blacksector.db <<EOF
INSERT OR REPLACE INTO players (player_id, player_name, token_hash, credits, created_at, is_banned)
VALUES ('test-player-1', 'TestPlayer', '\$2a\$10\$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 10000, $(date +%s), 0);
EOF

echo "Test player added (token: test-token)"
echo ""
echo "To test the SSH connection manually:"
echo "1. Start the server: ./blacksector-server"
echo "2. In another terminal, connect: ssh -p 2222 testplayer@localhost"
echo "3. Use any password (authentication happens in handshake)"
echo "4. Send handshake_response JSON:"
echo '{"type":"handshake_response","timestamp":'$(date +%s)',"protocol_version":"1.0","correlation_id":"test-123","payload":{"player_token":"test-token"}}'
echo ""
echo "Expected response: handshake_ack with session_id"
