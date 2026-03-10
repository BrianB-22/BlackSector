# Implementation Plan: Milestone 1 Foundation

## Overview

This plan implements the core server infrastructure for BlackSector Milestone 1. The implementation follows a bottom-up approach: starting with foundational components (config, database, logging), building up to core systems (session management, tick engine), and finally integrating everything into the complete server. Each task is designed to be testable independently and builds incrementally toward a working server that accepts SSH connections, manages sessions, runs the tick loop, and persists state.

## Tasks

- [x] 1. Set up project structure and core configuration
  - Create directory structure: cmd/server/, internal/config/, internal/db/, internal/engine/, internal/session/, config/, snapshots/, migrations/
  - Initialize Go module with required dependencies (modernc.org/sqlite, gliderlabs/ssh, charmbracelet/wish, rs/zerolog, google/uuid)
  - Create config/server.json with default configuration values
  - _Requirements: 1.1, 1.2_

- [x] 2. Implement configuration loading and validation
  - [x] 2.1 Create internal/config package with Config, ServerConfig, LoggingConfig, UniverseConfig structs
    - Implement LoadConfig() function to read and parse server.json
    - Add JSON struct tags for all configuration fields
    - _Requirements: 1.1, 1.3_
  
  - [x] 2.2 Implement configuration validation
    - Validate tick_interval_ms is between 100 and 10000
    - Validate ssh_port is between 1024 and 65535
    - Validate max_concurrent_players is between 1 and 1000
    - Validate snapshot_interval_ticks > 0
    - Validate required string fields are non-empty (db_path, world_config_path, log_level)
    - Return descriptive errors for each validation failure
    - _Requirements: 1.3, 1.4, 1.5, 1.6, 20.1, 20.2, 20.3, 20.4, 20.5, 20.6, 20.7_
  
  - [ ]* 2.3 Write property test for configuration validation
    - **Property 1: Configuration Validation Completeness**
    - **Validates: Requirements 1.3, 1.4, 1.5, 1.6, 20.1, 20.2, 20.3, 20.4, 20.5, 20.6**
    - Generate random Config structs with invalid values
    - Assert validation catches all constraint violations
  
  - [ ]* 2.4 Write unit tests for config loading
    - Test loading valid config file
    - Test missing config file returns error
    - Test invalid JSON returns error
    - Test out-of-range values are rejected
    - _Requirements: 1.2, 1.3, 1.4, 1.5, 1.6_

- [x] 3. Implement database layer with SQLite
  - [x] 3.1 Create database schema migration file
    - Create migrations/001_initial_schema.sql with tables: players, sessions, ships, sectors, ports
    - Define all columns, primary keys, foreign keys, and indexes per database_schema.md
    - _Requirements: 2.7_
  
  - [x] 3.2 Implement internal/db package with Database struct
    - Create InitDatabase() function to open SQLite connection
    - Apply WAL PRAGMAs: journal_mode=WAL, synchronous=NORMAL, foreign_keys=ON, busy_timeout=5000
    - Implement schema initialization from migration files
    - Add error handling for connection failures
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7_
  
  - [x] 3.3 Implement database query methods
    - GetPlayerByToken(tokenHash string) (*Player, error)
    - InsertSession(session *Session) error
    - UpdateSessionState(sessionID string, state SessionState) error
    - GetActiveSessionByPlayerID(playerID string) (*Session, error)
    - Use prepared statements for all queries
    - Wrap errors with context
    - _Requirements: 6.5, 7.5, 11.6, 17.4, 17.5, 15.3_
  
  - [ ]* 3.4 Write property test for database WAL mode
    - **Property 2: Database WAL Mode Invariant**
    - **Validates: Requirements 2.3, 2.4, 2.5, 2.6**
    - Open database connection and verify all PRAGMAs are set correctly
  
  - [ ]* 3.5 Write unit tests for database layer
    - Test database initialization with new file
    - Test schema creation
    - Test CRUD operations for players and sessions
    - Test prepared statement usage (no SQL injection)
    - Use go-sqlmock for mocking
    - _Requirements: 2.1, 2.2, 2.7, 17.4, 17.5_

- [x] 4. Implement logging infrastructure
  - [x] 4.1 Create logging initialization in main package
    - Initialize zerolog with configured log level
    - Set up file output to configured log_file
    - Configure structured logging format
    - Support debug_log_enabled flag
    - _Requirements: 14.1, 14.2, 14.3, 14.4_
  
  - [x] 4.2 Write unit tests for logging configuration
    - Test log level configuration
    - Test file output creation
    - Test structured field formatting
    - _Requirements: 14.1, 14.2, 14.5_

- [x] 5. Implement snapshot serialization and recovery
  - [x] 5.1 Create snapshot data structures
    - Define Snapshot, SnapshotState structs with JSON tags
    - Include snapshot_version, tick, timestamp, protocol_version fields
    - Include players and sessions arrays in state
    - _Requirements: 3.4, 3.5, 10.2, 10.3, 10.4, 10.5, 10.6_
  
  - [x] 5.2 Implement SaveSnapshot() function
    - Write snapshot to snapshots/snapshot_{tick}_{timestamp}.json
    - Use atomic write (temp file + rename)
    - Update snapshot_latest.json symlink
    - Delete old snapshots beyond retention count
    - Handle write errors gracefully (log but don't crash)
    - _Requirements: 10.7, 10.8, 10.9, 10.10, 10.11_
  
  - [x] 5.3 Implement LoadSnapshot() function
    - Read most recent snapshot from snapshots/ directory
    - Validate snapshot_version and protocol_version
    - Return nil if no snapshot exists (not an error)
    - Return error if snapshot is corrupted
    - _Requirements: 3.1, 3.2, 3.7, 3.8_
  
  - [ ]* 5.4 Write property test for snapshot round-trip
    - **Property 3: Snapshot Round-Trip Preservation**
    - **Validates: Requirements 3.2, 3.3, 3.4, 3.5, 10.2, 10.3, 10.4**
    - Generate random valid game state
    - Save snapshot and load it back
    - Assert original state equals loaded state (tick number, players, sessions)
  
  - [ ]* 5.5 Write unit tests for snapshot operations
    - Test saving snapshot creates correct filename
    - Test symlink is updated
    - Test old snapshots are deleted
    - Test loading non-existent snapshot returns nil
    - Test loading corrupted snapshot returns error
    - _Requirements: 3.1, 3.2, 3.7, 10.8, 10.9, 10.10_

- [x] 6. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 7. Implement protocol message types
  - [x] 7.1 Create protocol message structs
    - Define HandshakeInit, HandshakeResponse, HandshakeAck, HandshakeReject structs
    - Add JSON tags for all fields
    - Include type, timestamp, correlation_id, payload fields
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7, 19.1, 19.2, 19.3, 19.6, 19.7, 19.8_
  
  - [ ]* 7.2 Write property test for protocol message format
    - **Property 6: Handshake Protocol Message Format**
    - **Validates: Requirements 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7, 7.6, 7.7, 7.8, 7.9, 19.1, 19.2, 19.3, 19.6, 19.7**
    - Generate random handshake messages
    - Serialize to JSON and validate all required fields are present
    - Assert JSON is valid and parseable
  
  - [ ]* 7.3 Write unit tests for message serialization
    - Test HandshakeInit serialization includes all required fields
    - Test HandshakeAck serialization includes all required fields
    - Test HandshakeReject serialization includes reason
    - Test JSON validity
    - _Requirements: 5.7, 7.6, 7.7, 7.8, 7.9, 19.1, 19.7, 19.8_

- [x] 8. Implement session management
  - [x] 8.1 Create internal/session package with Session and SessionManager
    - Define Session struct with session_id, player_id, interface_mode, state, timestamps
    - Define SessionState enum: CONNECTED, DISCONNECTED_LINGERING, DOCKED_OFFLINE, TERMINATED
    - Create SessionManager with database reference and active sessions map
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 11.1_
  
  - [x] 8.2 Implement session lifecycle methods
    - CreateSession(playerID string) (*Session, error) - generates UUID, sets state to CONNECTED
    - GetActiveSession(playerID string) (*Session, error) - queries database for active sessions
    - UpdateSessionState(sessionID string, state SessionState) error - updates state and timestamps
    - TerminateSession(sessionID string) error - sets state to TERMINATED
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 11.2, 11.3, 11.4, 11.5, 11.6_
  
  - [ ]* 8.3 Write property test for session ID uniqueness
    - **Property 5: Session ID Uniqueness**
    - **Validates: Requirement 7.1**
    - Generate N concurrent session creation requests
    - Assert all session IDs are unique
  
  - [ ]* 8.4 Write property test for session state transitions
    - **Property 9: Session State Transitions**
    - **Validates: Requirements 7.2, 11.1, 11.2, 11.3, 11.4, 11.5**
    - Test valid state transitions: CONNECTED → DISCONNECTED_LINGERING → TERMINATED
    - Assert timestamps are set correctly for each transition
  
  - [ ]* 8.5 Write unit tests for session management
    - Test session creation generates unique UUID
    - Test session creation sets state to CONNECTED
    - Test GetActiveSession returns only CONNECTED sessions
    - Test session state updates are persisted
    - _Requirements: 7.1, 7.2, 11.1, 11.7, 16.5_

- [x] 9. Implement handshake protocol
  - [x] 9.1 Create HandleHandshake() function in session manager
    - Send HandshakeInit message with protocol_version "1.0", interface_mode "TEXT", server_name, motd
    - Wait for HandshakeResponse with 30-second timeout
    - Validate protocol_version matches "1.0"
    - Authenticate player_token against database
    - Check for existing active session
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 6.1, 6.2, 6.3, 6.5, 6.7_
  
  - [x] 9.2 Implement handshake response handling
    - On success: create session, send HandshakeAck with session_id, player_id, tick_interval_ms
    - On timeout: send HandshakeReject with reason "handshake_timeout", close connection
    - On version mismatch: send HandshakeReject with reason "version_mismatch", close connection
    - On invalid token: send HandshakeReject with reason "invalid_token", close connection
    - On session conflict: send HandshakeReject with reason "session_already_active", close connection
    - Log connection events at INFO level
    - _Requirements: 6.2, 6.4, 6.6, 6.8, 7.6, 7.7, 7.8, 7.9, 7.10, 14.6_
  
  - [ ]* 9.3 Write property test for protocol version validation
    - **Property 7: Protocol Version Validation**
    - **Validates: Requirements 6.3, 6.4**
    - Generate handshake responses with invalid protocol versions
    - Assert server sends handshake_reject with reason "version_mismatch"
  
  - [ ]* 9.4 Write property test for token authentication
    - **Property 8: Token Authentication**
    - **Validates: Requirements 6.5, 6.6**
    - Generate handshake responses with invalid tokens
    - Assert server sends handshake_reject with reason "invalid_token"
  
  - [x] 9.5 Write property test for session uniqueness enforcement
    - **Property 4: Session Creation Uniqueness**
    - **Validates: Requirements 6.7, 6.8, 16.1, 16.2, 16.3, 16.5**
    - Create active session for player A
    - Attempt second connection for player A
    - Assert handshake_reject with reason "session_already_active"
  
  - [ ]* 9.6 Write unit tests for handshake protocol
    - Test successful handshake flow
    - Test handshake timeout
    - Test version mismatch rejection
    - Test invalid token rejection
    - Test duplicate session rejection
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 6.7, 6.8_

- [x] 10. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 11. Implement tick engine core
  - [x] 11.1 Create internal/engine package with TickEngine struct
    - Define TickEngine with tickNumber, tickInterval, running flag, commandQueue channel
    - Add references to database, session manager, logger
    - Initialize with tick number from snapshot or 0
    - _Requirements: 8.1, 8.2, 8.3_
  
  - [x] 11.2 Implement RunTickLoop() method
    - Loop while running flag is true
    - Increment tickNumber each iteration
    - Record tick start time and calculate duration
    - Sleep for remaining interval time after tick completes
    - Log tick number at DEBUG level
    - Log warning if tick duration exceeds 500ms
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 8.6, 8.7, 8.8, 8.9, 14.8, 18.1, 18.2_
  
  - [x] 11.3 Implement command queue processing
    - Drain all commands from commandQueue at tick start
    - Process commands in FIFO order
    - Validate command structure
    - Emit command_reject event for invalid commands
    - Log valid commands at DEBUG level (no processing in M1)
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5, 9.6_
  
  - [ ]* 11.4 Write property test for tick monotonicity
    - **Property 11: Tick Monotonicity**
    - **Validates: Requirements 8.2, 8.9**
    - Simulate N ticks with random command inputs
    - Assert tick numbers form strictly increasing sequence with no gaps
  
  - [ ]* 11.5 Write property test for command queue FIFO ordering
    - **Property 13: Command Queue FIFO Ordering**
    - **Validates: Requirements 9.1, 9.2, 9.6**
    - Enqueue sequence of commands
    - Drain at tick start
    - Assert commands are processed in same order with no loss or duplication
  
  - [ ]* 11.6 Write unit tests for tick engine
    - Test tick loop increments tick number
    - Test tick duration is measured
    - Test slow tick warning is logged
    - Test command queue is drained each tick
    - Test invalid commands are rejected
    - _Requirements: 8.2, 8.4, 8.5, 8.6, 9.1, 9.4_

- [x] 12. Implement snapshot integration with tick engine
  - [x] 12.1 Add snapshot trigger logic to tick loop
    - Check if tickNumber % snapshotIntervalTicks == 0
    - Create snapshot with current tick, timestamp, players, sessions
    - Write snapshot asynchronously (goroutine) to avoid blocking tick
    - Log snapshot trigger at INFO level
    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5, 10.6, 10.7_
  
  - [x] 12.2 Implement snapshot recovery in server startup
    - Load most recent snapshot from snapshots/ directory
    - Restore tick number as snapshot.tick + 1
    - Restore player and session data
    - Initialize with empty state if no snapshot exists
    - Log recovery status at INFO level
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6_
  
  - [ ]* 12.3 Write property test for snapshot trigger timing
    - **Property 15: Snapshot Trigger Timing**
    - **Validates: Requirements 10.1, 10.2, 10.5, 10.6**
    - Run tick loop with configured snapshot interval
    - Assert snapshot is triggered at correct tick numbers (modulo interval)
  
  - [ ]* 12.4 Write integration test for snapshot recovery
    - Start server, run 100 ticks, trigger snapshot
    - Stop server
    - Restart server
    - Assert tick number resumes from snapshot.tick + 1
    - _Requirements: 3.1, 3.2, 3.3, 3.6_

- [x] 13. Implement SSH listener and connection handling
  - [x] 13.1 Create SSH server with gliderlabs/ssh and wish
    - Bind SSH listener to configured ssh_port
    - Set up SSH server with public key or password authentication
    - Create session handler for each incoming connection
    - Enforce max_concurrent_players limit
    - Log connection events at INFO level
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 14.6_
  
  - [x] 13.2 Integrate handshake protocol with SSH connections
    - On SSH connection established, call HandleHandshake()
    - Pass connection to session manager
    - Handle handshake success: keep connection open, add to active sessions
    - Handle handshake failure: close connection
    - _Requirements: 4.3, 4.4, 5.1, 6.1_
  
  - [ ]* 13.3 Write integration test for SSH connection flow
    - Start server with SSH listener
    - Open SSH connection
    - Complete handshake with valid token
    - Verify session created in database
    - Disconnect
    - Verify session state updated
    - _Requirements: 4.3, 4.4, 5.1, 6.1, 7.2, 11.2_

- [x] 14. Implement graceful shutdown
  - [x] 14.1 Add shutdown signal handling
    - Listen for SIGINT and SIGTERM signals
    - Set tick engine running flag to false
    - Wait for current tick to complete
    - _Requirements: 12.1, 12.2_
  
  - [x] 14.2 Implement shutdown sequence
    - Create final snapshot synchronously
    - Send server_shutdown messages to all active sessions
    - Close all SSH connections
    - Update all session states to TERMINATED
    - Close database connection
    - Log "Server stopped" at INFO level
    - Exit with code 0
    - _Requirements: 12.3, 12.4, 12.5, 12.6, 12.7, 12.8, 12.9, 12.10, 14.10_
  
  - [ ]* 14.3 Write property test for graceful shutdown completeness
    - **Property 19: Graceful Shutdown Completeness**
    - **Validates: Requirements 12.1, 12.2, 12.3, 12.4, 12.5, 12.6, 12.7, 12.8**
    - Initiate shutdown with active sessions
    - Assert current tick completes
    - Assert final snapshot is written
    - Assert all sessions receive shutdown message
    - Assert all connections are closed
    - Assert all session states are TERMINATED
  
  - [ ]* 14.4 Write integration test for server lifecycle
    - Start server
    - Create sessions
    - Send shutdown signal
    - Verify final snapshot written
    - Verify clean exit with code 0
    - _Requirements: 12.1, 12.2, 12.3, 12.9, 12.10_

- [x] 15. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 16. Implement admin CLI
  - [x] 16.1 Create admin command reader from stdin
    - Read commands from stdin in separate goroutine
    - Parse commands: "status", "sessions", "shutdown"
    - Send commands to tick engine via channel
    - Display "Unknown command" for unrecognized input
    - _Requirements: 13.1, 13.2, 13.3, 13.4, 13.5, 13.6_
  
  - [x] 16.2 Implement admin command handlers
    - "status": display current tick number and active session count
    - "sessions": display list of active sessions with player names
    - "shutdown": initiate graceful shutdown
    - _Requirements: 13.2, 13.3, 13.4_
  
  - [ ]* 16.3 Write unit tests for admin CLI
    - Test status command displays correct information
    - Test sessions command lists active sessions
    - Test shutdown command initiates graceful shutdown
    - Test unknown command displays error message
    - _Requirements: 13.2, 13.3, 13.4, 13.5_

- [x] 17. Implement error handling and recovery
  - [x] 17.1 Add fatal error handling for startup
    - Wrap all startup errors with context
    - Log fatal errors at ERROR level
    - Exit with code 1 on fatal errors (config missing, database failure, port bind failure)
    - _Requirements: 1.2, 2.2, 4.2, 15.1_
  
  - [x] 17.2 Add non-fatal error handling for runtime
    - Log non-fatal errors at ERROR level and continue
    - Handle snapshot write failures gracefully
    - Isolate session errors to prevent affecting other sessions
    - Reject invalid commands without crashing
    - _Requirements: 10.11, 15.2, 15.4, 15.5, 15.6_
  
  - [x] 17.3 Add panic recovery
    - Recover from panics in tick loop
    - Log panic with stack trace
    - Attempt graceful shutdown after panic
    - _Requirements: 15.7_
  
  - [ ]* 17.4 Write property test for session error isolation
    - **Property 20: Session Error Isolation**
    - **Validates: Requirement 15.5**
    - Simulate error in one session
    - Assert server closes only that session
    - Assert other sessions remain active
  
  - [x] 17.5 Write unit tests for error handling
    - Test fatal errors during startup exit with code 1
    - Test snapshot write failure logs error but continues
    - Test invalid command rejection
    - Test panic recovery
    - _Requirements: 15.1, 15.2, 15.4, 15.6, 15.7_

- [x] 18. Implement database transaction support
  - [x] 18.1 Add transaction wrapper methods to database layer
    - BeginTx() (*sql.Tx, error)
    - CommitTx(tx *sql.Tx) error
    - RollbackTx(tx *sql.Tx) error
    - Wrap related writes in single transaction
    - _Requirements: 17.1, 17.2, 17.3_
  
  - [ ]* 18.2 Write property test for database transaction atomicity
    - **Property 21: Database Transaction Atomicity**
    - **Validates: Requirements 17.1, 17.2, 17.3**
    - Execute transaction with multiple writes
    - Simulate failure mid-transaction
    - Assert all changes are rolled back
    - Execute successful transaction
    - Assert all changes are committed atomically
  
  - [ ]* 18.3 Write unit tests for transaction handling
    - Test successful transaction commits all changes
    - Test failed transaction rolls back all changes
    - Test nested transaction handling
    - _Requirements: 17.1, 17.2, 17.3_

- [x] 19. Wire all components together in main server binary
  - [x] 19.1 Create cmd/server/main.go
    - Parse command-line flags (--config path)
    - Load configuration
    - Initialize logger
    - Initialize database
    - Load or initialize state from snapshot
    - Create tick engine
    - Create session manager
    - Start admin CLI goroutine
    - Start SSH listener
    - Start tick loop
    - Handle shutdown signals
    - _Requirements: 1.1, 2.1, 3.1, 4.1, 8.1, 12.1, 13.1, 14.1_
  
  - [x] 19.2 Add startup and shutdown logging
    - Log "Server starting" at INFO level
    - Log configuration summary
    - Log "SSH listener on port X" at INFO level
    - Log "Tick loop starting" at INFO level
    - Log "Server stopped" at INFO level
    - _Requirements: 14.10_
  
  - [ ]* 19.3 Write end-to-end integration test
    - Start complete server with all components
    - Connect via SSH and complete handshake
    - Verify session created
    - Send shutdown command via admin CLI
    - Verify graceful shutdown
    - Restart server
    - Verify state restored from snapshot
    - _Requirements: All requirements integrated_

- [x] 20. Final checkpoint - Ensure all tests pass and server runs
  - Run all unit tests, property tests, and integration tests
  - Manually test SSH connection and handshake
  - Manually test admin CLI commands
  - Manually test graceful shutdown and recovery
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation at logical breakpoints
- Property tests validate universal correctness properties from the design document
- Unit tests validate specific examples and edge cases
- Integration tests validate end-to-end workflows
- The implementation follows Go conventions and uses the tech stack specified in tech_stack.md
- All database queries use prepared statements to prevent SQL injection
- All errors are wrapped with context for better debugging
- Structured logging with zerolog is used throughout
