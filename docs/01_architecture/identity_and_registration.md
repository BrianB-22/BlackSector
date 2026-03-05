# Identity & Registration Specification

## Version: 0.2

## Status: Draft

## Owner: Core Architecture

## Last Updated: 2026-03-05

---

# 1. Purpose

Defines the user identity, registration, authentication, and token model for BlackSector.

This specification governs:

* Account creation
* Credential storage
* Authentication flow
* Player token issuance
* Session binding
* Identity persistence
* Security controls

Identity is separate from simulation state but required before gameplay access.

---

# 2. Scope

IN SCOPE:

* New player registration flow (TEXT mode)
* Returning player login flow
* Player token model
* GUI client token usage
* Credential hashing
* Session binding
* Account persistence
* Registration rate limiting

OUT OF SCOPE:

* OAuth integration
* Third-party SSO
* Email verification (v1)
* Account recovery flows (future)

---

# 3. Design Principles

* Identity is server-controlled
* Passwords are never stored in plaintext
* Authentication must complete before command submission
* Player tokens are the credential used by all protocol handshakes
* Registration for new players occurs in-session via TEXT mode
* GUI clients require a token obtained from a prior TEXT session

---

# 4. Account Model

```go
type Account struct {
    PlayerID      string    // UUID — used as player_id throughout the system
    PlayerName    string    // unique, case-insensitive, 3–20 characters
    TokenHash     string    // bcrypt hash of the player_token
    PasswordHash  string    // bcrypt hash of the registration password
    Role          AccountRole
    Status        AccountStatus
    CreatedAt     int64     // Unix epoch
    LastLoginAt   int64
}

type AccountRole   string  // "player" | "admin"
type AccountStatus string  // "active" | "suspended" | "disabled"
```

Accounts are persisted independently of world/simulation state.

---

# 5. Player Token

The `player_token` is an opaque credential issued by the server at registration or login.

Properties:

* 32 bytes of cryptographically random data, base64url-encoded (43 characters)
* Issued once at registration; re-issued on explicit login with username + password
* Stored server-side as a bcrypt hash (`token_hash` in the `players` table)
* Used in every protocol handshake (`handshake_response.payload.player_token`)
* Functionally equivalent to a long-lived session key

The player must save their token after registration. It is displayed once in TEXT mode and must be recorded.

Token loss requires an admin to issue a replacement via `player token-reset <name>` in the admin CLI.

---

# 6. New Player Registration Flow (TEXT Mode)

New players always register through the SSH interface.

When a player SSHes to port 2222, the SSH transport authenticates the connection. After SSH auth succeeds, the server checks whether the SSH username matches an existing `player_name` in the database.

```
Player SSHes to port 2222
        ↓
SSH transport authentication completes
        ↓
Server looks up player_name = SSH username
        ↓
    [NOT FOUND]              [FOUND]
        ↓                       ↓
Registration flow           Login flow
(see 6.1)                   (see 7.1)
```

## 6.1 First-Time Registration

Before the JSON handshake begins, the server sends a TEXT-rendered registration prompt:

```
Welcome to Black Sector.
No account found for: nova

Create a new account? (yes/no): _
```

If the player confirms:

```
Choose a display name (3-20 chars, letters/numbers/underscore): _
```

The SSH username is used as the player name by default. The player may enter a different display name if they prefer (SSH username is transport identity only).

```
Create a password (min 8 chars): _
Confirm password: _
```

On success, the server:

1. Generates `player_id` (UUID)
2. Hashes password (bcrypt)
3. Generates `player_token` (32 random bytes, base64url)
4. Hashes token (bcrypt) and stores as `token_hash`
5. Creates player record in database
6. Creates starting ship at a default port
7. Logs `player_registered` event
8. Displays token to player:

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Your player token (save this — shown only once):

  bXlzZWNyZXR0b2tlbmhlcmVmb3JleGFtcGxl

  Use this token with GUI clients or to recover
  your account if you change your SSH key.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Press ENTER to continue...
```

After confirmation, the JSON handshake begins automatically:

```
Server → handshake_init  (interface_mode: TEXT)
Server → handshake_ack   (session_id, player_id, tick_interval_ms)
```

The `handshake_response` from the TEXT client is constructed internally by the server in this flow — the player does not manually type a token.

---

# 7. Returning Player Login Flow (TEXT Mode)

When a known SSH username reconnects:

```
Welcome back, nova.
Authenticating...
```

The server verifies the SSH credentials (handled by SSH transport layer — no password prompt needed if using SSH key auth).

If SSH password auth is in use, the SSH layer handles the prompt. The server does not issue a second password challenge.

After SSH auth succeeds on a known username, the server proceeds directly to the handshake:

```
Server → handshake_init  (interface_mode: TEXT)
Server → handshake_ack   (new session_id, player_id, tick_interval_ms)
```

Again, the `handshake_response` is constructed internally — the SSH username lookup is sufficient to identify the player.

---

# 8. Token-Based Login (GUI and Re-authentication)

GUI clients connect on TLS/TCP port 2223. There is no SSH transport layer — the player must supply their `player_token` directly in the handshake.

```
Client connects to TLS/TCP port 2223
        ↓
TLS handshake completes
        ↓
Server → handshake_init  (interface_mode: GUI)
        ↓
Client → handshake_response
         { "player_token": "bXlzZWNyZXR0..." }
        ↓
Server validates token (bcrypt compare)
        ↓
Server → handshake_ack  OR  handshake_reject
```

Token validation:

1. Server extracts `player_token` from `handshake_response`
2. Looks up player by brute-forcing token hashes (or via a lookup index — see Section 9)
3. Verifies token matches stored `token_hash`
4. Checks `account_status` is `active`
5. Issues `handshake_ack` with new `session_id`

If token is invalid: `handshake_reject` with reason `invalid_token`.

---

# 9. Token Lookup

Bcrypt comparison is expensive. To avoid scanning all player records on every GUI connect:

* A separate `token_lookup` index maps the first 8 characters of the raw token to `player_id`
* On login, the server uses the prefix to narrow candidates before bcrypt compare
* This reduces token validation to O(1) average case

---

# 10. Credential Storage

```
Passwords:   bcrypt, cost factor 12 minimum
Tokens:      bcrypt, cost factor 10 minimum (tokens are longer and random — lower cost acceptable)
```

Rules:

* Never logged
* Never returned to client after initial issuance
* Never transmitted over unencrypted connections (SSH and TLS enforce this)

---

# 11. Token Re-issuance

A player may re-issue their token (e.g., if they suspect it was compromised) by:

1. Connecting via SSH (TEXT mode)
2. Running the in-game command: `account token-refresh`
3. Authenticating with their password
4. Server generates and displays a new token; old token is invalidated immediately

Admin can force token reset via: `player token-reset <player_name>`

---

# 12. Session Binding

Upon successful authentication:

* `session_id` issued (new UUID each time)
* `player_id` bound to session
* Command submission enabled

Unauthenticated connections may only:

* Complete registration (TEXT mode)
* Complete the handshake (token validation)
* Not submit gameplay commands

See `session_management.md` for disconnect behavior, linger rules, and ship persistence.

---

# 13. Registration Rate Limiting

To prevent abuse:

* Max 3 new registrations per IP per hour
* Max 10 failed token validations per IP per minute — triggers IP cooldown
* Repeated failures logged at WARN level

---

# 14. Account Status

| Status    | Can Login | Notes                          |
| --------- | --------- | ------------------------------ |
| active    | Yes       | Normal gameplay                |
| suspended | No        | Temporary lock, admin reversible |
| disabled  | No        | Permanent block                |

All status changes are logged with admin_id, timestamp, and reason.

---

# 15. Admin Role

Admin accounts:

* Assigned by setting `role = "admin"` in the database directly (no in-game command)
* May execute admin-only CLI commands
* All admin actions logged at INFO level with player_id and parameters

---

# 16. Persistence

Account data is stored in the `players` table of the SQLite database (see `database_schema.md`).

Account writes are committed immediately — not batched with tick flushes.

Account reads (token validation) are synchronous but occur only during handshake, outside the tick loop.

---

# 17. Security Requirements

* No plaintext passwords or tokens stored
* No credentials logged
* No client-side hash acceptance
* All authentication server-side
* Tokens only transmitted over SSH or TLS

---

# 18. Logging

| Event                  | Level | Fields                              |
| ---------------------- | ----- | ----------------------------------- |
| player_registered      | INFO  | player_id, player_name, remote_addr |
| player_login_success   | INFO  | player_id, interface_mode           |
| player_login_failed    | WARN  | reason, remote_addr                 |
| player_token_reissued  | INFO  | player_id                           |
| account_suspended      | WARN  | player_id, admin_id, reason         |
| registration_rate_limit| WARN  | remote_addr                         |

---

# 19. Non-Goals (v1)

* Email verification
* Password reset via email
* Multi-factor authentication
* Social login
* Federated identity

---

# 20. Future Extensions

* MFA support
* Email-based account recovery
* API tokens for bots/tooling
* Fine-grained admin permissions

---

# End of Document
