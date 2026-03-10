package protocol

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandshakeInitSerialization(t *testing.T) {
	msg := HandshakeInit{
		Type:            "handshake_init",
		Timestamp:       time.Now().Unix(),
		ProtocolVersion: "1.0",
		InterfaceMode:   "TEXT",
		ServerName:      "Black Sector",
		MOTD:            "Welcome. Watch your back out there.",
		Payload:         map[string]interface{}{},
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify all required fields are present
	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "handshake_init", decoded["type"])
	assert.NotNil(t, decoded["timestamp"])
	assert.Equal(t, "1.0", decoded["protocol_version"])
	assert.Equal(t, "TEXT", decoded["interface_mode"])
	assert.Equal(t, "Black Sector", decoded["server_name"])
	assert.Equal(t, "Welcome. Watch your back out there.", decoded["motd"])
	assert.NotNil(t, decoded["payload"])
}

func TestHandshakeResponseSerialization(t *testing.T) {
	msg := HandshakeResponse{
		Type:            "handshake_response",
		Timestamp:       time.Now().Unix(),
		ProtocolVersion: "1.0",
		CorrelationID:   "test-correlation-123",
		Payload: HandshakeResponsePayload{
			PlayerToken: "test-token-abc",
		},
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify all required fields are present
	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "handshake_response", decoded["type"])
	assert.NotNil(t, decoded["timestamp"])
	assert.Equal(t, "1.0", decoded["protocol_version"])
	assert.Equal(t, "test-correlation-123", decoded["correlation_id"])
	
	payload, ok := decoded["payload"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test-token-abc", payload["player_token"])
}

func TestHandshakeAckSerialization(t *testing.T) {
	msg := HandshakeAck{
		Type:          "handshake_ack",
		Timestamp:     time.Now().Unix(),
		CorrelationID: "test-correlation-123",
		Payload: HandshakeAckPayload{
			SessionID:      "session-uuid-456",
			PlayerID:       "player-uuid-789",
			TickIntervalMs: 2000,
			InterfaceMode:  "TEXT",
		},
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify all required fields are present
	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "handshake_ack", decoded["type"])
	assert.NotNil(t, decoded["timestamp"])
	assert.Equal(t, "test-correlation-123", decoded["correlation_id"])
	
	payload, ok := decoded["payload"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "session-uuid-456", payload["session_id"])
	assert.Equal(t, "player-uuid-789", payload["player_id"])
	assert.Equal(t, float64(2000), payload["tick_interval_ms"]) // JSON numbers are float64
	assert.Equal(t, "TEXT", payload["interface_mode"])
}

func TestHandshakeRejectSerialization(t *testing.T) {
	msg := HandshakeReject{
		Type:          "handshake_reject",
		Timestamp:     time.Now().Unix(),
		CorrelationID: "test-correlation-123",
		Payload: HandshakeRejectPayload{
			Reason: RejectReasonInvalidToken,
		},
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify all required fields are present
	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "handshake_reject", decoded["type"])
	assert.NotNil(t, decoded["timestamp"])
	assert.Equal(t, "test-correlation-123", decoded["correlation_id"])
	
	payload, ok := decoded["payload"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "invalid_token", payload["reason"])
}

func TestHandshakeRejectReasons(t *testing.T) {
	// Verify all rejection reason constants are defined
	assert.Equal(t, "handshake_timeout", RejectReasonHandshakeTimeout)
	assert.Equal(t, "version_mismatch", RejectReasonVersionMismatch)
	assert.Equal(t, "invalid_token", RejectReasonInvalidToken)
	assert.Equal(t, "session_already_active", RejectReasonSessionAlreadyActive)
}

func TestHandshakeResponseDeserialization(t *testing.T) {
	jsonData := `{
		"type": "handshake_response",
		"timestamp": 1234567890,
		"protocol_version": "1.0",
		"correlation_id": "test-123",
		"payload": {
			"player_token": "my-token"
		}
	}`

	var msg HandshakeResponse
	err := json.Unmarshal([]byte(jsonData), &msg)
	require.NoError(t, err)

	assert.Equal(t, "handshake_response", msg.Type)
	assert.Equal(t, int64(1234567890), msg.Timestamp)
	assert.Equal(t, "1.0", msg.ProtocolVersion)
	assert.Equal(t, "test-123", msg.CorrelationID)
	assert.Equal(t, "my-token", msg.Payload.PlayerToken)
}

func TestHandshakeInitDeserialization(t *testing.T) {
	jsonData := `{
		"type": "handshake_init",
		"timestamp": 1234567890,
		"protocol_version": "1.0",
		"interface_mode": "TEXT",
		"server_name": "Test Server",
		"motd": "Welcome!",
		"payload": {}
	}`

	var msg HandshakeInit
	err := json.Unmarshal([]byte(jsonData), &msg)
	require.NoError(t, err)

	assert.Equal(t, "handshake_init", msg.Type)
	assert.Equal(t, int64(1234567890), msg.Timestamp)
	assert.Equal(t, "1.0", msg.ProtocolVersion)
	assert.Equal(t, "TEXT", msg.InterfaceMode)
	assert.Equal(t, "Test Server", msg.ServerName)
	assert.Equal(t, "Welcome!", msg.MOTD)
	assert.NotNil(t, msg.Payload)
}
