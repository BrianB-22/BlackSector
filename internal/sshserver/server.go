package sshserver

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gliderlabs/ssh"
	"github.com/rs/zerolog"
	"github.com/BrianB-22/BlackSector/internal/engine"
	"github.com/BrianB-22/BlackSector/internal/registration"
	"github.com/BrianB-22/BlackSector/internal/session"
	"github.com/BrianB-22/BlackSector/internal/tui"
)

// Server wraps the SSH server and manages connections
type Server struct {
	sshServer           *ssh.Server
	sessionManager      *session.SessionManager
	tickEngine          *engine.TickEngine
	registrar           *registration.Registrar
	registrationPrompter *RegistrationPrompter
	logger              zerolog.Logger
	maxPlayers          int
	activeConns         int
	mu                  sync.Mutex
	handshakeConfig     session.HandshakeConfig
}

// Config contains configuration for the SSH server
type Config struct {
	Port                 int
	MaxConcurrentPlayers int
	SessionManager       *session.SessionManager
	TickEngine           *engine.TickEngine
	Registrar            *registration.Registrar
	Logger               zerolog.Logger
	HandshakeConfig      session.HandshakeConfig
}

// NewServer creates a new SSH server
func NewServer(cfg Config) (*Server, error) {
	s := &Server{
		sessionManager:       cfg.SessionManager,
		tickEngine:           cfg.TickEngine,
		registrar:            cfg.Registrar,
		logger:               cfg.Logger,
		maxPlayers:           cfg.MaxConcurrentPlayers,
		handshakeConfig:      cfg.HandshakeConfig,
	}

	// Create registration prompter
	s.registrationPrompter = NewRegistrationPrompter(cfg.Registrar, cfg.Logger)

	// Create SSH server with gliderlabs/ssh
	sshServer := &ssh.Server{
		Addr: fmt.Sprintf(":%d", cfg.Port),
		Handler: s.handleSession,
		// Use wish middleware for better SSH handling
		// For now, we'll use password auth with a simple handler
		PasswordHandler: s.passwordHandler,
		// Public key auth can be added later
		PublicKeyHandler: nil,
	}

	s.sshServer = sshServer

	s.logger.Info().
		Int("port", cfg.Port).
		Int("max_players", cfg.MaxConcurrentPlayers).
		Msg("SSH server configured")

	return s, nil
}

// passwordHandler handles SSH password authentication
// For Milestone 1, we accept any password since authentication happens in the handshake
func (s *Server) passwordHandler(ctx ssh.Context, password string) bool {
	// We don't validate passwords here - authentication happens during the handshake protocol
	// This just allows the SSH connection to be established
	s.logger.Debug().
		Str("user", ctx.User()).
		Msg("SSH password auth accepted (handshake will validate)")
	return true
}

// handleSession is called for each SSH connection after authentication
// handleSession is called for each SSH connection after authentication
func (s *Server) handleSession(sess ssh.Session) {
	// Recover from panics in session handler to prevent affecting other sessions
	// Requirement 15.5: Session errors must not affect other sessions
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error().
				Interface("panic", r).
				Stack().
				Str("remote_addr", sess.RemoteAddr().String()).
				Msg("PANIC in session handler - closing session")

			// Attempt to close the session
			_ = sess.Close()
		}
	}()

	// Check if we've reached max concurrent players
	s.mu.Lock()
	if s.activeConns >= s.maxPlayers {
		s.mu.Unlock()
		s.logger.Warn().
			Int("active_connections", s.activeConns).
			Int("max_players", s.maxPlayers).
			Msg("Connection rejected: max concurrent players reached")

		// Send rejection message and close
		_, _ = io.WriteString(sess, "Server full. Maximum concurrent players reached.\n")
		_ = sess.Close()
		return
	}
	s.activeConns++
	connCount := s.activeConns
	s.mu.Unlock()

	// Decrement counter when session ends
	defer func() {
		s.mu.Lock()
		s.activeConns--
		s.mu.Unlock()
	}()

	remoteAddr := sess.RemoteAddr().String()
	sshUsername := sess.User()
	
	s.logger.Info().
		Str("remote_addr", remoteAddr).
		Str("user", sshUsername).
		Int("active_connections", connCount).
		Msg("SSH connection established")

	// Create a connection wrapper that implements io.ReadWriter
	conn := &sshConn{Session: sess}

	// Check if player exists by SSH username
	player, err := s.registrar.CheckPlayerExists(sshUsername)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("ssh_username", sshUsername).
			Str("remote_addr", remoteAddr).
			Msg("Failed to check player existence")
		_, _ = io.WriteString(sess, "Server error. Please try again later.\n")
		_ = sess.Close()
		return
	}

	// If player doesn't exist, run registration prompt
	if player == nil {
		s.logger.Info().
			Str("ssh_username", sshUsername).
			Str("remote_addr", remoteAddr).
			Msg("New player detected, starting registration flow")

		_, err := s.registrationPrompter.PromptForRegistration(conn, sshUsername, remoteAddr)
		if err != nil {
			s.logger.Info().
				Err(err).
				Str("ssh_username", sshUsername).
				Str("remote_addr", remoteAddr).
				Msg("Registration failed or declined")
			_ = sess.Close()
			return
		}

		// Registration successful - player now exists
		s.logger.Info().
			Str("ssh_username", sshUsername).
			Str("remote_addr", remoteAddr).
			Msg("Registration completed, proceeding to handshake")
	} else {
		// Returning player
		s.logger.Info().
			Str("player_id", player.PlayerID).
			Str("player_name", player.PlayerName).
			Str("ssh_username", sshUsername).
			Str("remote_addr", remoteAddr).
			Msg("Returning player detected")

		fmt.Fprintf(conn, "\nWelcome back, %s.\nAuthenticating...\n\n", player.PlayerName)
	}

	// Execute handshake protocol (SSH mode - uses SSH username for authentication)
	gameSession, err := s.sessionManager.HandleSSHHandshake(conn, sshUsername, s.handshakeConfig)
	if err != nil {
		s.logger.Info().
			Err(err).
			Str("remote_addr", remoteAddr).
			Msg("Handshake failed, closing connection")
		_ = sess.Close()
		return
	}

	// Handshake successful - keep connection open
	s.logger.Info().
		Str("session_id", gameSession.SessionID).
		Str("player_id", gameSession.PlayerID).
		Str("remote_addr", remoteAddr).
		Msg("Handshake successful, session active")

	// Register session for state updates from tick engine
	updateChan, err := s.sessionManager.RegisterSessionForUpdates(gameSession.SessionID)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("session_id", gameSession.SessionID).
			Msg("Failed to register session for updates")
		_ = sess.Close()
		return
	}

	// Unregister when session ends
	defer s.sessionManager.UnregisterSessionForUpdates(gameSession.SessionID)

	// Start bubbletea TUI program
	s.runTUISession(sess, gameSession, updateChan)

	// Wait for session context to be done (client disconnect or server shutdown)
	<-sess.Context().Done()

	// Update session state to disconnected
	if err := s.sessionManager.UpdateSessionState(gameSession.SessionID, "DISCONNECTED_LINGERING"); err != nil {
		s.logger.Error().
			Err(err).
			Str("session_id", gameSession.SessionID).
			Msg("Failed to update session state on disconnect")
	}

	s.logger.Info().
		Str("session_id", gameSession.SessionID).
		Str("player_id", gameSession.PlayerID).
		Str("remote_addr", remoteAddr).
		Msg("Player disconnected")
}

// Start starts the SSH server
func (s *Server) Start() error {
	s.logger.Info().
		Str("addr", s.sshServer.Addr).
		Msg("Starting SSH listener")

	// Try to bind to the port
	listener, err := net.Listen("tcp", s.sshServer.Addr)
	if err != nil {
		return fmt.Errorf("failed to bind SSH port: %w", err)
	}

	s.logger.Info().
		Str("addr", s.sshServer.Addr).
		Msg("SSH listener active")

	// Start serving in a goroutine
	go func() {
		if err := s.sshServer.Serve(listener); err != nil && err != ssh.ErrServerClosed {
			s.logger.Error().
				Err(err).
				Msg("SSH server error")
		}
	}()

	return nil
}

// Shutdown gracefully shuts down the SSH server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info().Msg("Shutting down SSH server")

	// Close the SSH server
	if err := s.sshServer.Close(); err != nil {
		return fmt.Errorf("failed to close SSH server: %w", err)
	}

	// Wait for all connections to close or timeout
	deadline, ok := ctx.Deadline()
	if ok {
		timeout := time.Until(deadline)
		s.logger.Info().
			Dur("timeout", timeout).
			Msg("Waiting for SSH connections to close")
	}

	// The SSH server will close all active sessions when Close() is called
	s.logger.Info().Msg("SSH server shutdown complete")
	return nil
}

// ActiveConnectionCount returns the current number of active SSH connections
func (s *Server) ActiveConnectionCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.activeConns
}

// runTUISession starts the bubbletea TUI program for a connected session
func (s *Server) runTUISession(sess ssh.Session, gameSession *session.GameSession, updateChan <-chan interface{}) {
	// Create command channel for sending commands to tick engine
	// This channel is buffered to avoid blocking the TUI
	commandChan := make(chan interface{}, 10)

	// Forward commands from TUI to tick engine
	go func() {
		for cmd := range commandChan {
			// Cast to engine.Command and enqueue
			if engineCmd, ok := cmd.(engine.Command); ok {
				s.tickEngine.EnqueueCommand(engineCmd)
				s.logger.Debug().
					Str("session_id", gameSession.SessionID).
					Str("command_type", engineCmd.CommandType).
					Msg("Command forwarded to tick engine")
			} else {
				s.logger.Warn().
					Str("session_id", gameSession.SessionID).
					Msg("Invalid command type from TUI")
			}
		}
	}()

	// Create GameView model
	gameView := tui.NewGameView(
		gameSession.SessionID,
		gameSession.PlayerID,
		updateChan,
		commandChan,
		s.logger,
	)

	// Start bubbletea program with SSH session as input/output
	program := tea.NewProgram(gameView, tea.WithInput(sess), tea.WithOutput(sess))
	
	// Run the program (blocks until quit)
	if _, err := program.Run(); err != nil {
		s.logger.Error().
			Err(err).
			Str("session_id", gameSession.SessionID).
			Msg("TUI program error")
	}

	s.logger.Info().
		Str("session_id", gameSession.SessionID).
		Msg("TUI session ended")
}

// sshConn wraps ssh.Session to implement io.ReadWriter for handshake protocol
type sshConn struct {
	ssh.Session
}

// Read implements io.Reader
func (c *sshConn) Read(p []byte) (n int, err error) {
	return c.Session.Read(p)
}

// Write implements io.Writer
func (c *sshConn) Write(p []byte) (n int, err error) {
	return c.Session.Write(p)
}
