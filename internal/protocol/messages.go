package protocol

// HandshakeInit is sent by the server after SSH authentication.
// It initiates the application-level handshake protocol.
type HandshakeInit struct {
	Type            string                 `json:"type"`              // "handshake_init"
	Timestamp       int64                  `json:"timestamp"`         // Unix timestamp
	ProtocolVersion string                 `json:"protocol_version"`  // "1.0"
	InterfaceMode   string                 `json:"interface_mode"`    // "TEXT" or "GUI"
	ServerName      string                 `json:"server_name"`       // Server display name
	MOTD            string                 `json:"motd"`              // Message of the day
	Payload         map[string]interface{} `json:"payload"`           // Empty object for consistency
}

// HandshakeResponse is sent by the client in response to HandshakeInit.
// It contains the player's authentication token.
type HandshakeResponse struct {
	Type            string                      `json:"type"`              // "handshake_response"
	Timestamp       int64                       `json:"timestamp"`         // Unix timestamp
	ProtocolVersion string                      `json:"protocol_version"`  // "1.0"
	CorrelationID   string                      `json:"correlation_id"`    // Matches request
	Payload         HandshakeResponsePayload    `json:"payload"`
}

// HandshakeResponsePayload contains the authentication data.
type HandshakeResponsePayload struct {
	PlayerToken string `json:"player_token"` // Player authentication token
}

// HandshakeAck is sent by the server on successful authentication.
// It confirms the session is established and provides session details.
type HandshakeAck struct {
	Type          string                `json:"type"`           // "handshake_ack"
	Timestamp     int64                 `json:"timestamp"`      // Unix timestamp
	CorrelationID string                `json:"correlation_id"` // Matches HandshakeResponse
	Payload       HandshakeAckPayload   `json:"payload"`
}

// HandshakeAckPayload contains session establishment details.
type HandshakeAckPayload struct {
	SessionID      string `json:"session_id"`       // Unique session identifier (UUID)
	PlayerID       string `json:"player_id"`        // Player's unique identifier
	TickIntervalMs int    `json:"tick_interval_ms"` // Server tick interval in milliseconds
	InterfaceMode  string `json:"interface_mode"`   // "TEXT" or "GUI"
}

// HandshakeReject is sent by the server when authentication fails.
// It provides a reason for the rejection.
type HandshakeReject struct {
	Type          string                   `json:"type"`           // "handshake_reject"
	Timestamp     int64                    `json:"timestamp"`      // Unix timestamp
	CorrelationID string                   `json:"correlation_id"` // Matches HandshakeResponse (if available)
	Payload       HandshakeRejectPayload   `json:"payload"`
}

// HandshakeRejectPayload contains the rejection reason.
type HandshakeRejectPayload struct {
	Reason string `json:"reason"` // Rejection reason code
}

// Common rejection reasons:
const (
	RejectReasonHandshakeTimeout    = "handshake_timeout"
	RejectReasonVersionMismatch     = "version_mismatch"
	RejectReasonInvalidToken        = "invalid_token"
	RejectReasonSessionAlreadyActive = "session_already_active"
)

// ServerShutdown is sent by the server when it is shutting down gracefully.
// It notifies all connected clients that the server is stopping.
type ServerShutdown struct {
	Type      string                  `json:"type"`      // "server_shutdown"
	Timestamp int64                   `json:"timestamp"` // Unix timestamp
	Payload   ServerShutdownPayload   `json:"payload"`
}

// ServerShutdownPayload contains the shutdown notification message.
type ServerShutdownPayload struct {
	Message string `json:"message"` // Human-readable shutdown message
}
