package sshserver

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/BrianB-22/BlackSector/internal/registration"
	"github.com/rs/zerolog"
)

// RegistrationPrompter handles interactive registration prompts
type RegistrationPrompter struct {
	registrar *registration.Registrar
	logger    zerolog.Logger
}

// NewRegistrationPrompter creates a new registration prompter
func NewRegistrationPrompter(registrar *registration.Registrar, logger zerolog.Logger) *RegistrationPrompter {
	return &RegistrationPrompter{
		registrar: registrar,
		logger:    logger,
	}
}

// PromptForRegistration handles the interactive registration flow
// Returns the player token on success, or error if registration fails/is declined
func (rp *RegistrationPrompter) PromptForRegistration(conn io.ReadWriter, sshUsername string, remoteAddr string) (string, error) {
	reader := bufio.NewReader(conn)

	// Welcome message
	fmt.Fprintf(conn, "\nWelcome to Black Sector.\n")
	fmt.Fprintf(conn, "No account found for: %s\n\n", sshUsername)

	// Ask if they want to create an account
	fmt.Fprintf(conn, "Create a new account? (yes/no): ")
	response, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "yes" && response != "y" {
		fmt.Fprintf(conn, "\nRegistration cancelled. Goodbye.\n")
		return "", fmt.Errorf("registration declined by user")
	}

	// Prompt for display name
	fmt.Fprintf(conn, "\nChoose a display name (3-20 chars, letters/numbers/underscore): ")
	displayName, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read display name: %w", err)
	}
	displayName = strings.TrimSpace(displayName)

	// If empty, use SSH username as default
	if displayName == "" {
		displayName = sshUsername
		fmt.Fprintf(conn, "Using SSH username as display name: %s\n", displayName)
	}

	// Prompt for password
	fmt.Fprintf(conn, "\nCreate a password (min 8 chars): ")
	password, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	password = strings.TrimSpace(password)

	// Prompt for password confirmation
	fmt.Fprintf(conn, "Confirm password: ")
	passwordConfirm, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read password confirmation: %w", err)
	}
	passwordConfirm = strings.TrimSpace(passwordConfirm)

	// Verify passwords match
	if password != passwordConfirm {
		fmt.Fprintf(conn, "\nPasswords do not match. Registration failed.\n")
		return "", fmt.Errorf("password mismatch")
	}

	// Create registration request
	req := &registration.RegistrationRequest{
		SSHUsername: sshUsername,
		DisplayName: displayName,
		Password:    password,
		RemoteAddr:  remoteAddr,
	}

	// Register the player
	result, err := rp.registrar.RegisterNewPlayer(req)
	if err != nil {
		fmt.Fprintf(conn, "\nRegistration failed: %v\n", err)
		return "", fmt.Errorf("registration failed: %w", err)
	}

	// Display success message and token
	fmt.Fprintf(conn, "\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Fprintf(conn, "  Registration successful!\n\n")
	fmt.Fprintf(conn, "  Your player token (save this — shown only once):\n\n")
	fmt.Fprintf(conn, "  %s\n\n", result.PlayerToken)
	fmt.Fprintf(conn, "  Use this token with GUI clients or to recover\n")
	fmt.Fprintf(conn, "  your account if you change your SSH key.\n")
	fmt.Fprintf(conn, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Fprintf(conn, "  Press ENTER to continue...")

	// Wait for confirmation
	_, _ = reader.ReadString('\n')

	rp.logger.Info().
		Str("player_id", result.PlayerID).
		Str("player_name", displayName).
		Str("ssh_username", sshUsername).
		Str("remote_addr", remoteAddr).
		Msg("New player registered via interactive prompt")

	return result.PlayerToken, nil
}
