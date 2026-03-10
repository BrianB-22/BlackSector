package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

type HandshakeInit struct {
	Type            string                 `json:"type"`
	Timestamp       int64                  `json:"timestamp"`
	ProtocolVersion string                 `json:"protocol_version"`
	InterfaceMode   string                 `json:"interface_mode"`
	ServerName      string                 `json:"server_name"`
	MOTD            string                 `json:"motd"`
	Payload         map[string]interface{} `json:"payload"`
}

type HandshakeResponse struct {
	Type            string                  `json:"type"`
	Timestamp       int64                   `json:"timestamp"`
	CorrelationID   string                  `json:"correlation_id"`
	ProtocolVersion string                  `json:"protocol_version"`
	Payload         HandshakeResponsePayload `json:"payload"`
}

type HandshakeResponsePayload struct {
	PlayerToken string `json:"player_token"`
}

type HandshakeAck struct {
	Type          string `json:"type"`
	Timestamp     int64  `json:"timestamp"`
	CorrelationID string `json:"correlation_id"`
	Payload       struct {
		SessionID      string `json:"session_id"`
		PlayerID       string `json:"player_id"`
		TickIntervalMs int    `json:"tick_interval_ms"`
		InterfaceMode  string `json:"interface_mode"`
	} `json:"payload"`
}

func main() {
	token := flag.String("token", "test_token_12345", "Player token")
	host := flag.String("host", "localhost:2222", "Server host:port")
	flag.Parse()

	fmt.Printf("Connecting to %s...\n", *host)

	// SSH client config
	config := &ssh.ClientConfig{
		User: "testplayer",
		Auth: []ssh.AuthMethod{
			ssh.Password("test123"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	// Connect
	client, err := ssh.Dial("tcp", *host, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to dial: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	// Open session
	session, err := client.NewSession()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create session: %v\n", err)
		os.Exit(1)
	}
	defer session.Close()

	// Get stdin/stdout pipes
	stdin, err := session.StdinPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get stdin: %v\n", err)
		os.Exit(1)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get stdout: %v\n", err)
		os.Exit(1)
	}

	// Start the session
	if err := session.Shell(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start shell: %v\n", err)
		os.Exit(1)
	}

	// Read handshake_init
	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() {
		fmt.Fprintf(os.Stderr, "Failed to read handshake_init\n")
		os.Exit(1)
	}

	var initMsg HandshakeInit
	if err := json.Unmarshal(scanner.Bytes(), &initMsg); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse handshake_init: %v\n", err)
		fmt.Fprintf(os.Stderr, "Received: %s\n", scanner.Text())
		os.Exit(1)
	}

	fmt.Printf("Server: %s\n", initMsg.ServerName)
	fmt.Printf("MOTD: %s\n", initMsg.MOTD)
	fmt.Printf("Protocol: %s\n", initMsg.ProtocolVersion)

	// Send handshake_response
	response := HandshakeResponse{
		Type:            "handshake_response",
		Timestamp:       time.Now().Unix(),
		CorrelationID:   fmt.Sprintf("client-%d", time.Now().UnixNano()),
		ProtocolVersion: "1.0",
		Payload: HandshakeResponsePayload{
			PlayerToken: *token,
		},
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal response: %v\n", err)
		os.Exit(1)
	}

	if _, err := stdin.Write(append(responseJSON, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send response: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Sent handshake response...")

	// Read handshake_ack
	if !scanner.Scan() {
		fmt.Fprintf(os.Stderr, "Failed to read handshake_ack\n")
		os.Exit(1)
	}

	var ackMsg HandshakeAck
	if err := json.Unmarshal(scanner.Bytes(), &ackMsg); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse handshake_ack: %v\n", err)
		fmt.Fprintf(os.Stderr, "Received: %s\n", scanner.Text())
		os.Exit(1)
	}

	if ackMsg.Payload.SessionID == "" || ackMsg.Payload.PlayerID == "" {
		fmt.Fprintf(os.Stderr, "\n✗ Handshake failed - empty session/player ID\n")
		fmt.Fprintf(os.Stderr, "This usually means authentication failed.\n")
		fmt.Fprintf(os.Stderr, "Check that the player token is correct and the player exists in the database.\n")
		os.Exit(1)
	}

	fmt.Printf("\n✓ Connected successfully!\n")
	fmt.Printf("Session ID: %s\n", ackMsg.Payload.SessionID)
	fmt.Printf("Player ID: %s\n", ackMsg.Payload.PlayerID)
	fmt.Printf("\nStarting TUI...\n\n")

	// Now we're connected - just read and display output
	// Set a timeout for reading to detect if TUI isn't sending anything
	go func() {
		time.Sleep(5 * time.Second)
		fmt.Println("\n[No output from TUI after 5 seconds - this might indicate an issue]")
	}()

	lineCount := 0
	for scanner.Scan() {
		lineCount++
		fmt.Println(scanner.Text())
	}

	if lineCount == 0 {
		fmt.Fprintf(os.Stderr, "\n✗ No output received from TUI\n")
		fmt.Fprintf(os.Stderr, "The TUI may have crashed or exited immediately.\n")
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading: %v\n", err)
	}
}
