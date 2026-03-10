# Requirements Document: Milestone 1 Foundation

## Introduction

This document specifies the functional and non-functional requirements for BlackSector Milestone 1 (Foundation). Milestone 1 establishes the core server infrastructure for a multiplayer text-based space trading game served over SSH. The system must provide a running headless server that accepts SSH connections, manages player sessions, executes a deterministic tick loop, persists state to SQLite, and provides administrative controls.

Success criteria: A player can SSH in, complete the handshake protocol, receive a handshake acknowledgment, and the server can survive restart by restoring state from snapshot.

## Glossary

- **Server**: The BlackSector game server process
- **Player**: A registered user account with credentials
- **Session**: An active connection between a player and the server
- **Tick_Engine**: The authoritative simulation loop that processes game state
- **Handshake**: The authentication protocol executed after SSH connection
- **Snapshot**: A serialized copy of complete server state at a specific tick
- **Command_Queue**: A FIFO buffer of player commands awaiting processing
- **Admin_CLI**: The administrative command-line interface accessible via stdin
- **Database**: The SQLite database storing persistent game state
- **Session_Manager**: The component managing active player sessions
- **Config**: The server configuration loaded from server.json

## Requirements

### Requirement 1: Server Initialization

**User Story:** As a system administrator, I want the server to initialize correctly from configuration, so that the game can start with proper settings.

#### Acceptance Criteria

1. WHEN the server starts, THE Server SHALL load configuration from server.json
2. IF server.json is missing or invalid, THEN THE Server SHALL log an error and exit with code 1
3. WHEN configuration is loaded, THE Server SHALL validate all required fields are present
4. WHEN configuration is loaded, THE Server SHALL validate tick_interval_ms is greater than 0
5. WHEN configuration is loaded, THE Server SHALL validate ssh_port is between 1024 and 65535
6. WHEN configuration is loaded, THE Server SHALL validate max_concurrent_players is greater than 0

### Requirement 2: Database Initialization

**User Story:** As a system administrator, I want the database to initialize with proper settings, so that data is persisted reliably.

#### Acceptance Criteria

1. WHEN the server starts, THE Server SHALL open the SQLite database at the configured db_path
2. IF the database cannot be opened, THEN THE Server SHALL log an error and exit with code 1
3. WHEN the database connection is established, THE Server SHALL set journal_mode to WAL
4. WHEN the database connection is established, THE Server SHALL set synchronous to NORMAL
5. WHEN the database connection is established, THE Server SHALL set foreign_keys to ON
6. WHEN the database connection is established, THE Server SHALL set busy_timeout to 5000
7. WHEN the database file is new, THE Server SHALL initialize the schema from migrations

### Requirement 3: Snapshot Recovery

**User Story:** As a system administrator, I want the server to restore from snapshots, so that game state persists across restarts.

#### Acceptance Criteria

1. WHEN the server starts, THE Server SHALL check for existing snapshots in the snapshots directory
2. WHEN a valid snapshot exists, THE Server SHALL load the most recent snapshot
3. WHEN a snapshot is loaded, THE Server SHALL restore tick number from snapshot.tick plus 1
4. WHEN a snapshot is loaded, THE Server SHALL restore all player data from the snapshot
5. WHEN a snapshot is loaded, THE Server SHALL restore all session data from the snapshot
6. WHEN no snapshot exists, THE Server SHALL initialize with empty state and tick number 0
7. IF a snapshot is corrupted or invalid, THEN THE Server SHALL log an error and exit with code 1
8. WHEN a snapshot is loaded, THE Server SHALL validate snapshot_version matches the server version

### Requirement 4: SSH Listener

**User Story:** As a player, I want to connect to the server via SSH, so that I can play the game.

#### Acceptance Criteria

1. WHEN the server starts, THE Server SHALL bind an SSH listener to the configured ssh_port
2. IF the SSH port cannot be bound, THEN THE Server SHALL log an error and exit with code 1
3. WHEN the SSH listener is active, THE Server SHALL accept incoming SSH connections
4. WHEN an SSH connection is established, THE Server SHALL create a new session handler
5. WHEN the number of active sessions reaches max_concurrent_players, THE Server SHALL reject new connections
6. WHEN the SSH listener is active, THE Server SHALL log connection events at INFO level

### Requirement 5: Handshake Protocol - Initialization

**User Story:** As a player, I want to receive handshake initialization, so that I know the server is ready for authentication.

#### Acceptance Criteria

1. WHEN an SSH connection is established, THE Server SHALL send a handshake_init message
2. WHEN sending handshake_init, THE Server SHALL include the current timestamp
3. WHEN sending handshake_init, THE Server SHALL include protocol_version "1.0"
4. WHEN sending handshake_init, THE Server SHALL include interface_mode "TEXT"
5. WHEN sending handshake_init, THE Server SHALL include the server_name from configuration
6. WHEN sending handshake_init, THE Server SHALL include the message of the day (MOTD)
7. WHEN sending handshake_init, THE Server SHALL format the message as valid JSON

### Requirement 6: Handshake Protocol - Authentication

**User Story:** As a player, I want to authenticate with my player token, so that I can access my account.

#### Acceptance Criteria

1. WHEN the server sends handshake_init, THE Server SHALL wait for handshake_response with a 30-second timeout
2. IF no handshake_response is received within 30 seconds, THEN THE Server SHALL send handshake_reject with reason "handshake_timeout" and close the connection
3. WHEN handshake_response is received, THE Server SHALL validate the protocol_version matches "1.0"
4. IF protocol_version does not match, THEN THE Server SHALL send handshake_reject with reason "version_mismatch" and close the connection
5. WHEN handshake_response is received, THE Server SHALL validate the player_token against the database
6. IF player_token is invalid, THEN THE Server SHALL send handshake_reject with reason "invalid_token" and close the connection
7. WHEN player_token is valid, THE Server SHALL check for existing active sessions for that player
8. IF an active session exists for the player, THEN THE Server SHALL send handshake_reject with reason "session_already_active" and close the connection

### Requirement 7: Handshake Protocol - Session Creation

**User Story:** As a player, I want to receive session confirmation, so that I know I'm connected and can start playing.

#### Acceptance Criteria

1. WHEN authentication succeeds, THE Server SHALL generate a unique session_id using UUID
2. WHEN authentication succeeds, THE Server SHALL create a session record with state "CONNECTED"
3. WHEN a session is created, THE Server SHALL set connected_at to the current Unix timestamp
4. WHEN a session is created, THE Server SHALL set last_activity_at to the current Unix timestamp
5. WHEN a session is created, THE Server SHALL persist the session to the database
6. WHEN a session is created, THE Server SHALL send handshake_ack with session_id and player_id
7. WHEN sending handshake_ack, THE Server SHALL include tick_interval_ms from configuration
8. WHEN sending handshake_ack, THE Server SHALL include interface_mode "TEXT"
9. WHEN sending handshake_ack, THE Server SHALL include the correlation_id from handshake_response
10. WHEN a session is created, THE Server SHALL log the player name and session_id at INFO level

### Requirement 8: Tick Loop Execution

**User Story:** As a system administrator, I want the tick loop to run reliably, so that game simulation progresses consistently.

#### Acceptance Criteria

1. WHEN the server starts, THE Tick_Engine SHALL begin the tick loop
2. WHEN the tick loop runs, THE Tick_Engine SHALL increment tick_number by 1 each iteration
3. WHEN the tick loop runs, THE Tick_Engine SHALL execute ticks at the configured tick_interval_ms
4. WHEN a tick begins, THE Tick_Engine SHALL record the start time
5. WHEN a tick completes, THE Tick_Engine SHALL calculate the tick duration
6. IF tick duration exceeds 500ms, THEN THE Tick_Engine SHALL log a warning with the duration
7. WHEN a tick completes, THE Tick_Engine SHALL sleep for the remaining interval time
8. WHEN the tick loop runs, THE Tick_Engine SHALL continue until the running flag is set to false
9. WHEN the tick loop runs, THE Tick_Engine SHALL ensure tick numbers increase monotonically without gaps

### Requirement 9: Command Queue Processing

**User Story:** As a player, I want my commands to be processed in order, so that my actions are executed fairly.

#### Acceptance Criteria

1. WHEN a tick begins, THE Tick_Engine SHALL drain all commands from the command queue
2. WHEN commands are drained, THE Tick_Engine SHALL process them in FIFO order
3. WHEN a command is received, THE Tick_Engine SHALL validate the command structure
4. IF a command is invalid, THEN THE Tick_Engine SHALL emit a command_reject event to the session
5. WHERE Milestone 1, WHEN a valid command is received, THE Tick_Engine SHALL log it at DEBUG level but not process it
6. WHEN commands are drained, THE Tick_Engine SHALL ensure no commands are lost or duplicated

### Requirement 10: Snapshot Creation

**User Story:** As a system administrator, I want snapshots to be created periodically, so that server state can be recovered after restart.

#### Acceptance Criteria

1. WHEN tick_number modulo snapshot_interval_ticks equals 0, THE Tick_Engine SHALL trigger snapshot creation
2. WHEN a snapshot is triggered, THE Tick_Engine SHALL create a snapshot containing current tick number
3. WHEN a snapshot is created, THE Tick_Engine SHALL include all player data in the snapshot
4. WHEN a snapshot is created, THE Tick_Engine SHALL include all session data in the snapshot
5. WHEN a snapshot is created, THE Tick_Engine SHALL include the current Unix timestamp
6. WHEN a snapshot is created, THE Tick_Engine SHALL include snapshot_version and protocol_version
7. WHEN a snapshot is created, THE Tick_Engine SHALL write it asynchronously to avoid blocking the tick loop
8. WHEN a snapshot is written, THE Server SHALL save it to snapshots/snapshot_{tick}_{timestamp}.json
9. WHEN a snapshot is written, THE Server SHALL update the snapshot_latest.json symlink
10. WHEN a snapshot is written, THE Server SHALL delete old snapshots beyond the retention count
11. IF snapshot write fails, THEN THE Server SHALL log an error but continue simulation

### Requirement 11: Session State Management

**User Story:** As a player, I want my session state to be tracked, so that the server knows when I'm connected.

#### Acceptance Criteria

1. WHEN a session is created, THE Session_Manager SHALL set the session state to "CONNECTED"
2. WHEN a player disconnects, THE Session_Manager SHALL update the session state to "DISCONNECTED_LINGERING"
3. WHEN a session enters DISCONNECTED_LINGERING, THE Session_Manager SHALL set disconnected_at timestamp
4. WHEN a session enters DISCONNECTED_LINGERING, THE Session_Manager SHALL calculate linger_expiry_at
5. WHEN a session is terminated, THE Session_Manager SHALL set the session state to "TERMINATED"
6. WHEN session state changes, THE Session_Manager SHALL persist the update to the database
7. WHEN querying active sessions, THE Session_Manager SHALL return only sessions with state "CONNECTED"

### Requirement 12: Graceful Shutdown

**User Story:** As a system administrator, I want the server to shut down gracefully, so that no data is lost.

#### Acceptance Criteria

1. WHEN a shutdown signal is received, THE Server SHALL set the Tick_Engine running flag to false
2. WHEN shutdown is initiated, THE Server SHALL wait for the current tick to complete
3. WHEN the current tick completes, THE Server SHALL create a final snapshot
4. WHEN the final snapshot is created, THE Server SHALL write it to disk synchronously
5. WHEN the final snapshot is written, THE Server SHALL send server_shutdown messages to all active sessions
6. WHEN server_shutdown messages are sent, THE Server SHALL close all SSH connections
7. WHEN all connections are closed, THE Server SHALL update all session states to "TERMINATED"
8. WHEN all sessions are terminated, THE Server SHALL close the database connection
9. WHEN the database is closed, THE Server SHALL log "Server stopped" at INFO level
10. WHEN shutdown completes, THE Server SHALL exit with code 0

### Requirement 13: Admin CLI

**User Story:** As a system administrator, I want to control the server via stdin, so that I can manage it without network access.

#### Acceptance Criteria

1. WHEN the server starts, THE Admin_CLI SHALL begin reading from stdin
2. WHEN "status" is entered, THE Admin_CLI SHALL display current tick number and active session count
3. WHEN "sessions" is entered, THE Admin_CLI SHALL display a list of all active sessions with player names
4. WHEN "shutdown" is entered, THE Admin_CLI SHALL initiate graceful shutdown
5. WHEN an unknown command is entered, THE Admin_CLI SHALL display "Unknown command" and continue
6. WHEN the Admin_CLI runs, THE Admin_CLI SHALL not block the tick loop

### Requirement 14: Logging

**User Story:** As a system administrator, I want comprehensive logging, so that I can monitor and debug the server.

#### Acceptance Criteria

1. WHEN the server starts, THE Server SHALL initialize zerolog with the configured log level
2. WHEN the server starts, THE Server SHALL write logs to the configured log_file
3. WHEN debug_log_enabled is true, THE Server SHALL log at DEBUG level
4. WHEN debug_log_enabled is false, THE Server SHALL log at INFO level or higher
5. WHEN logging events, THE Server SHALL use structured fields (player_id, tick, session_id, etc.)
6. WHEN a player connects, THE Server SHALL log the connection event at INFO level
7. WHEN a player disconnects, THE Server SHALL log the disconnection event at INFO level
8. WHEN a tick completes, THE Server SHALL log the tick number at DEBUG level
9. WHEN an error occurs, THE Server SHALL log the error with context at ERROR level
10. WHEN the server starts or stops, THE Server SHALL log the event at INFO level

### Requirement 15: Error Handling

**User Story:** As a system administrator, I want errors to be handled gracefully, so that the server remains stable.

#### Acceptance Criteria

1. WHEN a fatal error occurs during startup, THEN THE Server SHALL log the error and exit with code 1
2. WHEN a non-fatal error occurs during operation, THEN THE Server SHALL log the error and continue
3. WHEN a database query fails, THEN THE Server SHALL wrap the error with context
4. WHEN a snapshot write fails, THEN THE Server SHALL log the error but continue simulation
5. WHEN a session encounters an error, THEN THE Server SHALL close that session without affecting others
6. WHEN an invalid command is received, THEN THE Server SHALL reject it without crashing
7. WHEN a panic occurs, THEN THE Server SHALL recover, log the panic, and attempt graceful shutdown

### Requirement 16: Session Uniqueness

**User Story:** As a player, I want to be prevented from having multiple simultaneous sessions, so that my account state remains consistent.

#### Acceptance Criteria

1. WHEN a player attempts to connect, THE Session_Manager SHALL check for existing active sessions
2. WHEN an active session exists for a player, THE Session_Manager SHALL reject new connection attempts
3. WHEN rejecting a duplicate connection, THE Session_Manager SHALL send handshake_reject with reason "session_already_active"
4. WHEN a session is terminated, THE Session_Manager SHALL allow the player to connect again
5. WHEN checking for active sessions, THE Session_Manager SHALL only consider sessions with state "CONNECTED"

### Requirement 17: Database Transactions

**User Story:** As a developer, I want database writes to be transactional, so that data integrity is maintained.

#### Acceptance Criteria

1. WHEN writing related records, THE Database SHALL wrap them in a single transaction
2. WHEN a transaction fails, THE Database SHALL roll back all changes
3. WHEN a transaction succeeds, THE Database SHALL commit all changes atomically
4. WHEN using prepared statements, THE Database SHALL parameterize all query values
5. WHEN executing queries, THE Database SHALL never concatenate user input into SQL strings

### Requirement 18: Performance Monitoring

**User Story:** As a system administrator, I want performance metrics logged, so that I can identify bottlenecks.

#### Acceptance Criteria

1. WHEN a tick completes, THE Tick_Engine SHALL log the tick duration at DEBUG level
2. WHEN tick duration exceeds 500ms, THE Tick_Engine SHALL log a warning at WARN level
3. WHEN a snapshot is written, THE Server SHALL log the write duration at DEBUG level
4. WHEN a database query completes, THE Database SHALL log queries exceeding 10ms at WARN level
5. WHEN the server runs, THE Server SHALL track the count of active sessions

### Requirement 19: Protocol Message Format

**User Story:** As a client developer, I want all protocol messages to follow a consistent format, so that I can parse them reliably.

#### Acceptance Criteria

1. WHEN sending any protocol message, THE Server SHALL format it as valid JSON
2. WHEN sending any protocol message, THE Server SHALL include a "type" field identifying the message type
3. WHEN sending any protocol message, THE Server SHALL include a "timestamp" field with Unix time
4. WHEN sending a response message, THE Server SHALL include a "correlation_id" matching the request
5. WHEN sending a message with data, THE Server SHALL include a "payload" object
6. WHEN sending handshake_init, THE Server SHALL include protocol_version, interface_mode, server_name, and motd
7. WHEN sending handshake_ack, THE Server SHALL include session_id, player_id, tick_interval_ms, and interface_mode
8. WHEN sending handshake_reject, THE Server SHALL include a reason string

### Requirement 20: Configuration Validation

**User Story:** As a system administrator, I want configuration to be validated at startup, so that I catch errors early.

#### Acceptance Criteria

1. WHEN loading configuration, THE Server SHALL validate tick_interval_ms is between 100 and 10000
2. WHEN loading configuration, THE Server SHALL validate snapshot_interval_ticks is greater than 0
3. WHEN loading configuration, THE Server SHALL validate max_concurrent_players is between 1 and 1000
4. WHEN loading configuration, THE Server SHALL validate db_path is a non-empty string
5. WHEN loading configuration, THE Server SHALL validate world_config_path is a non-empty string
6. WHEN loading configuration, THE Server SHALL validate log_level is one of: "debug", "info", "warn", "error"
7. IF any validation fails, THEN THE Server SHALL log which field failed and exit with code 1
